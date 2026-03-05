# lsgit

`lsgit` is a command-line tool that works like `ls` — but for git repositories. Point it at any directory and it instantly shows you every git repo it finds, along with the current branch, working-tree status, and how far ahead or behind the remote tracking branch you are.

```
  /Users/you/workspace

  my-api              main                  ✓ clean
  frontend            feature/dark-mode     ● 3 changes   ↑2
  infra               main                  ● 1 change
  old-service         HEAD:3f9a12c          ● 7 changes   ↓14
  docs                main                  ✓ clean

  5 repos  2 clean  3 dirty
```

---

## Table of Contents

- [Features](#features)
- [Installation](#installation)
  - [macOS — Homebrew](#macos--homebrew)
  - [Linux — apt / deb](#linux--apt--deb)
  - [Build from Source](#build-from-source)
- [Usage](#usage)
- [Parameters & Flags](#parameters--flags)
- [Output Reference](#output-reference)
- [Contributing](#contributing)
- [Troubleshooting](#troubleshooting)
- [License](#license)

---

## Features

- Scans subdirectories and identifies git repositories
- Shows current branch name (or short SHA for detached HEAD)
- Shows working-tree status: clean or number of changed/staged/untracked files
- Shows ahead/behind count relative to the upstream tracking branch
- Parallel execution — all `git` calls run concurrently for speed
- Recursive scanning with configurable depth
- Colorized output (auto-disabled when piped)
- Optional `git fetch` before status check (`-f`)
- Zero runtime dependencies — single static binary

---

## Installation

### macOS — Homebrew

```bash
brew install itinance/tap/lsgit
```

To update:

```bash
brew upgrade lsgit
```

### Linux — apt / deb

Download the latest `.deb` package from the [Releases](https://github.com/itinance/lsgit/releases) page:

```bash
curl -sLO https://github.com/itinance/lsgit/releases/latest/download/lsgit_linux_amd64.deb
sudo dpkg -i lsgit_linux_amd64.deb
```

For ARM64 (e.g. Raspberry Pi, AWS Graviton):

```bash
curl -sLO https://github.com/itinance/lsgit/releases/latest/download/lsgit_linux_arm64.deb
sudo dpkg -i lsgit_linux_arm64.deb
```

RPM-based distros (Fedora, RHEL, CentOS):

```bash
sudo rpm -i https://github.com/itinance/lsgit/releases/latest/download/lsgit_linux_amd64.rpm
```

### Build from Source

Requires [Go](https://golang.org/dl/) 1.21 or later.

```bash
git clone https://github.com/itinance/lsgit.git
cd lsgit
make build
make install        # copies binary to /usr/local/bin
```

Or without `make`:

```bash
go build -o lsgit .
sudo mv lsgit /usr/local/bin/
```

---

## Usage

```bash
# Scan the current directory (one level deep)
lsgit

# Scan a specific directory
lsgit ~/workspace

# Also show non-git directories
lsgit -a ~/workspace

# Recurse up to 3 levels deep to find nested repos
lsgit -d 3 ~/workspace

# Recurse with no depth limit
lsgit -d 0 ~/workspace

# Fetch from remote before checking status
lsgit -f ~/workspace

# Combine flags
lsgit -a -d 2 -f ~/workspace

# Show remote origin URL for each repository
lsgit -u ~/workspace

# Disable color (useful for scripts and logging)
lsgit --no-color ~/workspace | tee repos.txt
```

---

## Parameters & Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--all` | `-a` | `false` | Show non-git directories too (displayed as `—`) |
| `--depth` | `-d` | `1` | Max depth for recursive scanning. `0` means unlimited. |
| `--fetch` | `-f` | `false` | Run `git fetch` before checking status. Makes results accurate against remote but is slower. |
| `--url` | `-u` | `false` | Show the remote origin URL below each repository name. |
| `--no-color` | | `false` | Disable ANSI color output. Automatically active when stdout is not a terminal. |
| `--version` | `-v` | | Print version and exit. |
| `--help` | `-h` | | Show help. |

### Depth explained

By default (`-d 1`) lsgit only looks at **direct subdirectories** of the target path. Increase depth to find repos nested inside non-repo directories:

```
workspace/               ← target
├── my-api/              ← found at depth 1 (repo)
├── clients/             ← not a repo, recurse if depth > 1
│   ├── client-a/        ← found at depth 2
│   └── client-b/        ← found at depth 2
└── tools/               ← not a repo, recurse if depth > 1
    └── linter/          ← found at depth 2
```

---

## Output Reference

```
  frontend    feature/dark-mode    ● 3 changes   ↑2 ↓1
  ^           ^                    ^             ^
  |           |                    |             |
  repo name   branch name          status        ahead/behind
```

### Status indicators

| Symbol | Meaning |
|--------|---------|
| `✓ clean` | Working tree is clean, nothing to commit |
| `● N changes` | N files are modified, staged, or untracked |
| `! error` | `git` returned an error (repo may be corrupt or inaccessible) |

### Branch indicators

| Value | Meaning |
|-------|---------|
| `main` / `feature/x` | Normal branch name |
| `HEAD:3f9a12c` | Detached HEAD state — shows short SHA |
| `(unknown)` | Could not determine branch or SHA |

### Ahead/behind indicators

| Symbol | Meaning |
|--------|---------|
| `↑N` | N commits ahead of the remote tracking branch (unpushed) |
| `↓N` | N commits behind the remote tracking branch (need to pull) |
| `↑N ↓M` | Diverged from remote |

> Ahead/behind is only shown when a remote tracking branch is configured. Repos with no remote or no tracking branch show nothing here.

---

## Contributing

Contributions are welcome. Please open an issue first for significant changes so we can discuss the approach.

### Setup

```bash
git clone https://github.com/itinance/lsgit.git
cd lsgit
go mod tidy
make build
```

### Running tests

```bash
make test
# or
go test ./...
```

### Local release simulation

To test the full GoReleaser pipeline without publishing:

```bash
# Install goreleaser if needed
brew install goreleaser

make snapshot
# Artifacts will appear in dist/
```

### Publishing a release

Maintainers with the required tokens:

```bash
git tag v1.2.3
make release
```

This requires two environment variables:
- `GITHUB_TOKEN` — a GitHub personal access token with `repo` scope
- `HOMEBREW_TAP_TOKEN` — a token with write access to the `itinance/homebrew-tap` repo

### Code structure

| File | Purpose |
|------|---------|
| `main.go` | CLI definition, flags, entry point |
| `scanner.go` | Directory traversal and git introspection |
| `display.go` | Colorized tabular output |
| `version.go` | Version string (overridden by ldflags at build time) |
| `.goreleaser.yml` | Cross-platform build and release pipeline |
| `Makefile` | Developer shortcuts |

### Guidelines

- Keep the binary dependency-free at runtime (`git` must be in `PATH` — that's it)
- New flags should have sensible defaults that preserve existing behavior
- Avoid adding dependencies; the current dep count (cobra + fatih/color) is intentional

---

## Troubleshooting

### `lsgit: command not found`

The binary is not on your `PATH`. If you installed manually:

```bash
# Check where it is
which lsgit || ls /usr/local/bin/lsgit

# Add /usr/local/bin to your PATH if missing
echo 'export PATH="$PATH:/usr/local/bin"' >> ~/.zshrc
source ~/.zshrc
```

### No repos found, but I know there are repos here

Check the depth. By default only direct subdirectories are scanned:

```bash
lsgit -d 3 .
```

If repos are inside hidden directories (`.dotfiles/myrepo`), they are skipped intentionally. Hidden directories (starting with `.`) are never scanned.

### `! error` shown for a repo

`git` returned a non-zero exit code for that repository. Common causes:

```bash
# Check manually
git -C /path/to/repo status

# Common issues:
# - Repo is corrupt:       git fsck
# - Permissions problem:   ls -la /path/to/repo/.git
# - Safe.directory issue:  git config --global --add safe.directory /path/to/repo
```

### Colors not showing in terminal

Ensure your terminal supports ANSI colors. If you are piping output, colors are disabled automatically. Force them off with `--no-color` or check:

```bash
echo $TERM          # should be xterm-256color or similar
echo $NO_COLOR      # if set, color libraries respect it
```

### Ahead/behind not showing for a repo

The repo has no remote tracking branch configured:

```bash
# Check
git -C /path/to/repo branch -vv

# Fix: set upstream for the current branch
git -C /path/to/repo branch --set-upstream-to=origin/main
```

### `git fetch` is slow with `-f`

`-f` runs `git fetch` on every repo sequentially before reporting. For large workspaces with many repos this can take a while. Consider running it only on a subdirectory:

```bash
lsgit -f ~/workspace/active-projects
```

### Build errors from source

Ensure your Go version is 1.21 or later:

```bash
go version
```

If you use `asdf`:

```bash
asdf install golang 1.23.5
asdf set golang 1.23.5
```

---

## License

MIT — see [LICENSE](LICENSE).
