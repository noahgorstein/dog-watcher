# Dog Watcher

A TUI to manage processes in Stardog.

https://user-images.githubusercontent.com/23270779/186049026-f15cc47d-cd26-4e24-9315-1bf74f099fa5.mov

`dog-watcher` currently supports:
- viewing processes in Stardog 
- killing processes in Stardog

## Installation

### homebrew

```bash
brew install noahgorstein/tap/dog-watcher
```

### Github releases

Download the relevant asset for your operating system from the latest Github release. Unpack it, then move the binary to somewhere accessible in your `PATH`, e.g. `mv ./dog-watcher /usr/local/bin`.

### Build from source

Clone this repo, build from source with `cd dog-watcher && go build`, then move the binary to somewhere accessible in your `PATH`, e.g. `mv ./dog-watcher /usr/local/bin`.

## Usage

Run the app by running `dog-watcher` in a terminal. See `dog-watcher --help` and [configuration](#configuration) section below for details.

## Controls

| Key | Description |
| ---- | ---------- |
| `up`/`down` | move table cursor |
| `left`/`right` | page table |
| `/` | filter table |
| `esc` | clear filter |
| `ctrl+x` | kill highlighted process |
| `d`/`i` | increase/decrease refresh rate |
| `ctrl+c` | exit |


## Configuration

`dog-watcher` can be configured in a yaml file at `$HOME/.dog-watcher.yaml`.

Example yaml file:

```yaml
# .dog-watcher.yaml
username: "admin"
password: "admin"
server: "http://localhost:5820"
```

Alternatively, `dog-watcher` can be configured via environment variables, or via command line args visible by running `dog-watcher --help`.

> Command line args take precedence over both the configuation file and environment variables. Environment variables take precedence over the configuration file.

`dog-watcher` will attempt to authenticate using the default superuser `admin` with password `admin` on `http://localhost:5820` if no credentials are provided.

### Environment Variables

| Environment Variable  |  Description |
|---|---|
| `DOG_WATCHER_USERNAME`  | username |
| `DOG_WATCHER_PASSWORD`  | password |
| `DOG_WATCHER_SERVER`  | Stardog server to connect to |


## Built With

- [bubbletea](https://github.com/charmbracelet/bubbletea)
- [bubbles](https://github.com/charmbracelet/bubbles)
- [bubble-table](https://github.com/Evertras/bubble-table)
- [lipgloss](https://github.com/charmbracelet/lipgloss)
- [go-stardog](https://github.com/noahgorstein/go-stardog)

## Notes

- 🚧 this project is under active development and is relatively unstable. Please file an issue if you see a bug or have suggestions on how to improve the app.

