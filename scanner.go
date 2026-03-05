package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// RepoInfo holds the git metadata for one directory.
type RepoInfo struct {
	Path      string
	Name      string
	Branch    string
	Changes   int // number of changed/untracked/staged lines from git status --porcelain
	IsClean   bool
	IsRepo    bool
	Ahead     int // commits ahead of upstream
	Behind    int // commits behind upstream
	RemoteURL string
	Error     string
}

// scan returns RepoInfo entries for all git repos found under root.
// Depth is the current recursion depth; flagDepth is the cap (0 = unlimited).
// relBase is the path prefix accumulated so far relative to the original scan root.
func scan(root string, depth int) []RepoInfo {
	return scanRel(root, depth, "")
}

func scanRel(root string, depth int, relBase string) []RepoInfo {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil
	}

	type result struct {
		idx   int
		infos []RepoInfo
	}

	type job struct {
		idx     int
		name    string
		relName string // path relative to the original scan root
		path    string
	}

	var jobs []job
	for i, e := range entries {
		if !e.IsDir() {
			continue
		}
		if strings.HasPrefix(e.Name(), ".") && !flagHidden {
			continue
		}
		rel := e.Name()
		if relBase != "" {
			rel = relBase + "/" + e.Name()
		}
		jobs = append(jobs, job{i, e.Name(), rel, filepath.Join(root, e.Name())})
	}

	resultsCh := make(chan result, len(jobs))
	var wg sync.WaitGroup

	for _, j := range jobs {
		wg.Add(1)
		go func(j job) {
			defer wg.Done()
			var infos []RepoInfo
			if isGitRepo(j.path) {
				if flagFetch {
					_ = gitFetch(j.path)
				}
				info := gitInfo(j.path)
				info.Name = j.relName
				infos = append(infos, info)
				// When -H is active, also recurse into the repo to find nested
				// hidden dirs (e.g. .worktrees) that may contain more repos.
				if flagHidden && (flagDepth == 0 || depth < flagDepth) {
					infos = append(infos, scanRel(j.path, depth+1, j.relName)...)
				}
			} else if flagDepth == 0 || depth < flagDepth {
				sub := scanRel(j.path, depth+1, j.relName)
				if len(sub) > 0 {
					infos = sub
				} else if flagAll {
					infos = []RepoInfo{{Path: j.path, Name: j.relName, IsRepo: false}}
				}
			} else if flagAll {
				infos = []RepoInfo{{Path: j.path, Name: j.relName, IsRepo: false}}
			}
			if len(infos) > 0 {
				resultsCh <- result{j.idx, infos}
			}
		}(j)
	}

	wg.Wait()
	close(resultsCh)

	// Collect preserving order
	ordered := make(map[int][]RepoInfo)
	for r := range resultsCh {
		ordered[r.idx] = r.infos
	}

	var out []RepoInfo
	for i := range jobs {
		out = append(out, ordered[i]...)
	}
	return out
}

func isGitRepo(path string) bool {
	_, err := os.Stat(filepath.Join(path, ".git"))
	return err == nil
}

func gitFetch(path string) error {
	cmd := exec.Command("git", "-C", path, "fetch", "--quiet")
	return cmd.Run()
}

func gitInfo(path string) RepoInfo {
	info := RepoInfo{Path: path, IsRepo: true}

	// Branch
	out, err := exec.Command("git", "-C", path, "branch", "--show-current").Output()
	if err != nil || strings.TrimSpace(string(out)) == "" {
		// Detached HEAD
		out2, err2 := exec.Command("git", "-C", path, "rev-parse", "--short", "HEAD").Output()
		if err2 == nil {
			info.Branch = "HEAD:" + strings.TrimSpace(string(out2))
		} else {
			info.Branch = "(unknown)"
		}
	} else {
		info.Branch = strings.TrimSpace(string(out))
	}

	// Porcelain status
	out, err = exec.Command("git", "-C", path, "status", "--porcelain").Output()
	if err != nil {
		info.Error = err.Error()
		return info
	}
	lines := strings.Split(strings.TrimRight(string(out), "\n"), "\n")
	count := 0
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			count++
		}
	}
	info.Changes = count
	info.IsClean = count == 0

	// Ahead/behind tracking branch
	out, err = exec.Command("git", "-C", path, "rev-list", "--left-right", "--count", "@{u}...HEAD").Output()
	if err == nil {
		parts := strings.Fields(strings.TrimSpace(string(out)))
		if len(parts) == 2 {
			fmt.Sscanf(parts[0], "%d", &info.Behind)
			fmt.Sscanf(parts[1], "%d", &info.Ahead)
		}
	}

	// Remote URL (only when requested, to avoid the extra subprocess otherwise)
	if flagURL {
		out, err = exec.Command("git", "-C", path, "remote", "get-url", "origin").Output()
		if err == nil {
			info.RemoteURL = strings.TrimSpace(string(out))
		}
	}

	return info
}

func absolutePath(p string) (string, error) {
	return filepath.Abs(p)
}
