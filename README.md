# Gom - Go Package Alias Manager (Local Project + Global Cache)

Gom is a command-line interface (CLI) tool written in Go that simplifies managing Go dependencies and documenting project environment variables. It allows you to use aliases for Go packages, stored both in a project-local `gom.json` file and an optional global cache. It also helps document required environment variables within `gom.json`. Gom acts as a convenient wrapper around the standard `go get` command, while still relying on the core Go Modules ecosystem.

**Current Version:** `0.1.0` (Update this manually or via build flags)

## Features

* **Project-Local Aliases & Env Docs (`gom.json`):** Define short aliases for dependencies and document required environment variables (with descriptions and optional defaults) specific to your project.
* **Global Alias Cache:** Share frequently used aliases across different projects using the `-g` flag with `gom get`.
* **Initialize Project:** Quickly create a `gom.json` file using `gom init`.
* **Add & Install Aliases:** Use `gom get <link> as <alias>@<version> [-g]` to run `go get <link>` and save the alias mapping.
* **Install by Alias (with Fallback):** Install a specific dependency using its alias (`gom install <alias>@<version>`). Checks `gom.json` first, then the global cache. Adds to `gom.json` if found globally before installing.
* **Install All Project Dependencies:** Run `gom install` (or `gom i`) to install all dependencies listed *only* in the project's `gom.json`.
* **Uninstall Alias:** Remove an alias from the project's `gom.json` using `gom uninstall <alias>@<version>`.
* **Manage Env Vars:** Add, remove, and list environment variable documentation within `gom.json` using `gom env ...` commands.
* **Show Project Config:** View the entire contents of `gom.json` (aliases and environments) using `gom show`.
* **Manage Global Cache:** List (`gom cache list`) or clear (`gom cache clear`) the global alias cache.
* **Version Check:** Display the current version of the Gom tool (`gom version`).
* **Go Installation Check:** Verifies Go installation on startup (using cache) and allows explicit check (`gom check go`).
* **Standard Go Modules:** Relies entirely on Go modules (`go.mod`, `go.sum`). `gom.json` and the global cache are for alias lookup and environment documentation.

## Installation

1. **Prerequisites:** Go (version 1.11+) installed and in PATH ([https://go.dev/doc/install](https://go.dev/doc/install)).
2. **Get Source:** `git clone <your_repository_url>` & `cd <repository_directory_name>`
3. **Build:**

    ```bash
    # Simple build (version hardcoded in config.go)
    go build -o gom .

    # or more simply
    go build

    # Build with specific version (Recommended for releases)
    # Replace '0.2.0' with the desired version tag
    go build -ldflags="-X main.AppVersion=0.2.0" -o gom .
    ```

4. **Add to PATH (Recommended):** Move the `gom` (or `gom.exe`) executable to a directory in your system's PATH (see previous README examples for detailed steps per OS).

## Usage

**Important:** Most `gom` commands require running inside a Go project root (containing `go.mod`). `cache` and `version` commands work globally. `check go` also works globally.

### Commands

* **`gom init`**
  * Creates an empty `gom.json` in the current directory.

* **`gom get <link> as <pkg>@<ver> [-g]`**
  * Runs `go get <link>`.
  * On success, adds/updates alias `<pkg>@<ver>` -> `<link>`:
    * To `./gom.json` (default).
    * To global cache (with `-g`).
  * **Note:** `<ver>` must be specific (not `latest`).

* **`gom install [<pkg>@<ver>]`** (Alias: `i`)
  * **`gom install <pkg>@<ver>`**: Installs specific alias (checks `./gom.json`, then global cache, adds to `./gom.json` if found globally).
  * **`gom install`**: Installs all dependencies listed in `./gom.json`.

* **`gom uninstall <pkg>@<ver>`**
  * Removes the alias `<pkg>@<ver>` from `./gom.json`. Does not run `go mod tidy`.

* **`gom show`**
  * Displays the full contents of `./gom.json` (dependencies and environments).

* **`gom env list`** (Alias: `ls`)
  * Lists environment variables documented in `./gom.json`.
* **`gom env add <VAR_NAME> --desc "Description" [--default "Default Value"]`**
  * Adds or updates documentation for `<VAR_NAME>` in `./gom.json`. `--desc` is required.
* **`gom env remove <VAR_NAME>`** (Alias: `rm`)
  * Removes documentation for `<VAR_NAME>` from `./gom.json`.

* **`gom cache list`**
  * Displays the contents of the global alias cache.
* **`gom cache clear`**
  * Removes all entries from the global alias cache file.

* **`gom check go`** (Alias: `doctor go`)
  * Verifies Go installation by running `go version` and updates the cached result.

* **`gom version`** (Aliases: `-v`, `--version`)
  * Displays the installed version of the Gom tool.

### File Locations

* **Project Config:** `gom.json` (In your project root, next to `go.mod`). Contains `dependencies` and `environments`.
* **Global Cache:** Standard OS cache directory:
  * **Linux:** `~/.cache/gom/cache.json`
  * **macOS:** `~/Library/Caches/gom/cache.json`
  * **Windows:** `%LOCALAPPDATA%\gom\cache.json`
    (Contains `aliases` and `go_version`).

### Example Workflow

1. `cd /path/to/myproject`
2. `gom init`
3. `gom get github.com/gin-gonic/gin@v1.9.1 as gin@v1.9.1`
4. `gom env add PORT --desc "Port the server listens on" --default "8080"`
5. `gom show` (Shows both gin dependency and PORT env)
6. `gom i gin@v1.9.1`
7. `gom check go` (Verifies Go and updates cache)
8. `gom version`

## Contributing

Contributions welcome! Please open Issues or Pull Requests.

## License

MIT License - see `LICENSE` file.
