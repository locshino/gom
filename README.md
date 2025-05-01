# Gom - Go Package Alias Manager

Gom is a command-line interface (CLI) tool written in Go that simplifies managing Go dependencies. It allows you to use aliases for Go packages, stored both in a project-local `gom.json` file and an optional global cache for sharing aliases across projects. Gom acts as a convenient wrapper around the standard `go get` command, while still relying on the core Go Modules ecosystem.

**Current Version:** `0.1.0` (Update this manually or via build flags)

## Features

* **Project-Local Aliases (`gom.json`):** Define short aliases for dependencies specific to your current project in a `gom.json` file.
* **Global Alias Cache:** Share frequently used aliases across different projects by saving them to a global cache using the `-g` flag.
* **Initialize Project:** Quickly create a `gom.json` file using `gom init`.
* **Add & Install Aliases:** Use `gom get <link> as <alias>@<version> [-g]` to run `go get <link>` and save the alias mapping to either `gom.json` (default) or the global cache (`-g`).
* **Install by Alias (with Fallback):** Install a specific dependency using its alias (`gom install <alias>@<version>`). Gom first checks `gom.json`, then the global cache. If found in the global cache, the alias is automatically added to `gom.json` before installation.
* **Install All Project Dependencies:** Run `gom install` (or `gom i`) without arguments to install all dependencies listed *only* in the project's `gom.json`.
* **Uninstall Alias:** Remove an alias specifically from the project's `gom.json` using `gom uninstall <alias>@<version>`.
* **Show Aliases:** View alias mappings in `gom.json` (`gom show`) or the global cache (`gom cache list`).
* **Manage Global Cache:** Clear the global cache (`gom cache clear`).
* **Version Check:** Display the current version of the Gom tool (`gom version`).
* **Go Installation Check:** Verifies that Go is installed and accessible.
* **Standard Go Modules:** Relies entirely on Go modules (`go.mod`, `go.sum`) for dependency resolution. `gom.json` and the global cache are for alias lookup.

## Installation

1. **Prerequisites:** [Go](https://go.dev) (version 1.11+) installed and in PATH ([https://go.dev/doc/install](https://go.dev/doc/install)).
2. **Get Source:** `git clone <your_repository_url>` & `cd <repository_directory_name>`
3. **Build:**

    ```bash
    # Simple build (version hardcoded in main.go)
    go build -o gom .

    # or more simply
    go build

    # Build with specific version (Recommended for releases)
    # Replace '0.2.0' with the desired version tag
    go build -ldflags="-X main.AppVersion=0.2.0" -o gom .
    ```

    This creates `gom` (Linux/macOS) or `gom.exe` (Windows).
4. **Add to PATH (Recommended):** To run `gom` from any directory, move the executable to a directory listed in your system's PATH environment variable.

    * **Linux/macOS:**
        * A common location is `/usr/local/bin`. Open your terminal in the directory where you built `gom` and run:

            ```bash
            sudo mv gom /usr/local/bin/gom
            ```

            (You might be prompted for your password).
        * Verify by opening a *new* terminal window and typing `gom version` or `which gom`.
        * Alternatively, create a directory like `~/bin`, move `gom` there (`mv gom ~/bin/`), and add `~/bin` to your PATH by editing your shell configuration file (`~/.bashrc`, `~/.zshrc`, etc.) and adding a line like `export PATH="$HOME/bin:$PATH"`. Remember to source the file (`source ~/.bashrc`) or open a new terminal.

    * **Windows:**
        1. **Create a Directory (Optional but Recommended):** Create a folder like `C:\bin` or `C:\Tools`.
        2. **Move Executable:** Move the `gom.exe` file into the directory you created (e.g., `C:\bin\gom.exe`).
        3. **Add to PATH (Command Line - requires Administrator):**
            * Open Command Prompt or PowerShell *as Administrator*.
            * Run the following command, replacing `C:\bin` with the actual path to your directory:

                ```cmd
                setx PATH "%PATH%;C:\bin" /M
                ```

                * `/M` makes the change system-wide. Omit `/M` to change it only for the current user.
                * **Important:** You need to **close and reopen** any terminal windows for the change to take effect.
        4. **Add to PATH (GUI):**
            * Search for "Environment Variables" in the Windows search bar and select "Edit the system environment variables".
            * Click the "Environment Variables..." button.
            * Under "System variables" (for all users) or "User variables" (for current user), find the `Path` variable, select it, and click "Edit...".
            * Click "New" and paste the full path to your directory (e.g., `C:\bin`).
            * Click "OK" on all open windows.
            * **Important:** Open a **new** Command Prompt or PowerShell window to test.
        5. **Verify:** In a *new* terminal window, type `gom version` or `where gom`.

## Usage

**Important:** Most `gom` commands should be run from within your Go project's root directory (containing `go.mod`). `cache` and `version` commands work globally.

### Commands

* **`gom init`**
  * Creates an empty `gom.json` in the current directory.

* **`gom get <link> as <pkg>@<ver> [-g]`**
  * Runs `go get <link>`.
  * If successful:
    * Without `-g`: Adds/updates alias `<pkg>@<ver>` -> `<link>` in `./gom.json`.
    * With `-g`: Adds/updates alias `<pkg>@<ver>` -> `<link>` in the global cache.
  * **Note:** `<ver>` must be a specific version (not `latest`).

* **`gom install [<pkg>@<ver>]`** (Alias: `i`)
  * **`gom install <pkg>@<ver>`**:
        1. Looks up alias in `./gom.json`.
        2. If not found, looks up alias in global cache.
        3. If found in global cache, adds the alias mapping to `./gom.json`.
        4. If found in either location, runs `go get <link>`.
        5. If not found anywhere, reports an error.
  * **`gom install`**: Runs `go get <link>` for all dependencies listed *only* in `./gom.json`.

* **`gom uninstall <pkg>@<ver>`**
  * Removes the alias `<pkg>@<ver>` from `./gom.json`.
  * Does *not* run `go mod tidy` or touch the global cache.

* **`gom show`**
  * Displays the contents of `./gom.json`.

* **`gom cache list`**
  * Displays the contents of the global alias cache.

* **`gom cache clear`**
  * Removes all entries from the global alias cache file.

* **`gom version`** (Aliases: `-v`, `--version`)
  * Displays the installed version of the Gom tool.

### File Locations

* **Project Aliases:** `gom.json` (In your project root, next to `go.mod`).
* **Global Cache:** Standard OS cache directory:
  * **Linux:** `~/.cache/gom/cache.json`
  * **macOS:** `~/Library/Caches/gom/cache.json`
  * **Windows:** `%LOCALAPPDATA%\gom\cache.json`

### Example Workflow

1. `cd /path/to/myproject`
2. `gom init`
3. `gom get github.com/gin-gonic/gin@v1.9.1 as gin@v1.9.1` (Adds to `./gom.json`)
4. `gom get gopkg.in/guregu/null.v3 as null@v3 -g` (Adds to global cache)
5. `cd /path/to/anotherproject`
6. `gom init`
7. `gom i null@v3` (Finds `null@v3` in global cache, adds it to `./gom.json`, runs `go get`)
8. `gom i gin@v1.9.1` (Fails - not in `./gom.json` or global cache for this project yet)
9. `gom get github.com/gin-gonic/gin@v1.9.1 as gin@v1.9.1` (Adds `gin` to this project's `./gom.json`)
10. `gom i` (Installs both `null@v3` and `gin@v1.9.1` based on the current project's `./gom.json`)
11. `gom uninstall null@v3` (Removes `null@v3` alias from `./gom.json`)
12. `gom cache list` (Shows `null@v3` is still in the global cache)
13. `gom version` (Shows the tool version)

## Contributing

Contributions welcome! Please open Issues or Pull Requests.

## License

MIT License - see `LICENSE` file.
