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
func scan(root string, depth int) []RepoInfo {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil
	}

	type result struct {
		idx   int
		infos []RepoInfo
	}

	type job struct {
		idx  int
		name string
		path string
	}

	var jobs []job
	for i, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		jobs = append(jobs, job{i, e.Name(), filepath.Join(root, e.Name())})
	}

	resultsCh := make(chan result, len(jobs))
	var wg sync.WaitGroup

	for _, j := range jobs {
		wg.Add(1)
		go func(j job) {
			defer wg.Done()
			if isGitRepo(j.path) {
				if flagFetch {
					_ = gitFetch(j.path)
				}
				info := gitInfo(j.path)
				info.Name = j.name
				resultsCh <- result{j.idx, []RepoInfo{info}}
			} else if flagDepth == 0 || depth < flagDepth {
				sub := scan(j.path, depth+1)
				if len(sub) > 0 {
					resultsCh <- result{j.idx, sub}
				} else if flagAll {
					resultsCh <- result{j.idx, []RepoInfo{{Path: j.path, Name: j.name, IsRepo: false}}}
				}
			} else if flagAll {
				resultsCh <- result{j.idx, []RepoInfo{{Path: j.path, Name: j.name, IsRepo: false}}}
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
