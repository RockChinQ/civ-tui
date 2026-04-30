# civ-tui

**A full-featured Civilization-style 4X strategy game with both terminal and local Web UI modes.**

Build cities, research technologies, command armies, and negotiate diplomacy from a fast terminal UI. Prefer a browser? Run the built-in single-player Web UI locally and play the same core game through a character-map interface. No external game engine required.

![Game Screen](docs/gamescreen.png)

---

## Features

**Complete 4X Gameplay Loop** -- Explore, Expand, Exploit, Exterminate in ~5,900 lines of Go.

- **Procedural Maps** -- Fractal noise terrain generation with continent-style layouts. 9 terrain types, 3 map sizes, fog of war, coordinate grid labels
- **8 Unit Types** -- Settlers, Scouts, Warriors, Archers, Spearmen, Swordsmen, Horsemen, Workers -- each with unique stats
- **City Building** -- Found cities, manage population growth, construct buildings, queue production
- **11-Tech Tree** -- Research technologies to unlock advanced units, buildings, and improvements
- **Combat System** -- Melee & ranged combat, terrain defense bonuses, experience & leveling, combat logs with civ names and coordinates
- **Pathfinding & Goto** -- Set movement destinations with Dijkstra pathfinding; units auto-advance each turn. Reachable tiles highlighted via BFS flood-fill
- **Diplomacy** -- Declare war, negotiate peace. Multiple AI civilizations with strategic decision-making
- **5 Civilizations** -- Rome, Mongolia, Egypt, China, Greece -- each with historical city names
- **Workers & Improvements** -- Build farms, mines, roads, and lumber mills to boost your economy
- **Bilingual (EN/ZH)** -- Full English and Simplified Chinese support with automatic system language detection
- **TUI + Local Web UI** -- Play in the terminal by default, or start an embedded browser UI with `-web`
- **Save/Load** -- Full game state persistence via JSON with 10 save slots. Pick up where you left off
- **Multiple Victory Conditions** -- Domination, Science, or survive to turn 200

## Quick Start

### Install via Go

```bash
go install github.com/RockChinQ/civ-tui@latest
civ-tui
```

### Build from Source

```bash
git clone https://github.com/RockChinQ/civ-tui.git
cd civ-tui
make run
```

### Run the Web UI

The Web UI is a single-player local mode served by the same Go binary. It uses the shared `game/` package, keeps one active game in the server process, and stores saves in the same `~/.civ-tui/saves/` directory as the TUI.

```bash
go run . -web
# then open http://127.0.0.1:8080
```

Use a different bind address if needed:

```bash
go run . -web -addr 127.0.0.1:9090
```

### Download Prebuilt Binaries

Prebuilt binaries for Linux, macOS, and Windows are available on the [Releases](https://github.com/RockChinQ/civ-tui/releases) page.

**Requirements:** Go 1.24+ (for building from source), a terminal with 256-color support.

## TUI Controls

| Key | Action |
|-----|--------|
| `Arrow keys` / `hjkl` | Move cursor / selected unit |
| `Enter` | End turn |
| `F` | Found city (Settler) |
| `B` | Open build menu (on your city) |
| `T` | Tech research menu |
| `D` | Diplomacy menu |
| `R` | Ranged attack mode (Archer) |
| `I` | Build improvement (Worker) |
| `V` | View city details (own city) / inspect tile |
| `G` | Set movement destination (Goto mode) |
| `X` | Cancel unit's movement destination |
| `W` | Wait / skip unit turn |
| `N` | Cycle to next unit or city needing attention |
| `S` | Save game |
| `Esc` | Deselect / close menu / cancel mode |
| `?` | Help screen |
| `Q` | Quit |

Vim users rejoice -- `hjkl` navigation works everywhere.

## Web UI Controls

The Web UI supports mouse and keyboard input:

| Input | Action |
|-------|--------|
| `Click tile` | Move cursor / select friendly unit / move or attack adjacent tile |
| `Arrow keys` / `WASD` / `hjkl` | Move cursor |
| `Enter` | Move selected unit to cursor if adjacent |
| `F` | Found city (Settler) |
| `B` | Open build menu |
| `T` | Research menu |
| `R` | Ranged attack mode |
| `G` | Set movement destination |
| `Z` | Wait / skip unit turn |
| `I` | Build improvement (Worker) |
| `Space` | End turn |
| `Esc` | Cancel mode / deselect |

Top bar buttons provide New, Save, Load, Build, Research, and End Turn actions. See [Web UI docs](docs/WEB_UI.md) for API details and current limitations.

## Game Settings

Configure from the **New Game** screen before starting:

| Setting | Options |
|---------|---------|
| Map Size | Small (40x25), Medium (60x35), Large (80x48) |
| AI Opponents | 1 - 4 |
| Difficulty | 3 levels (TUI: Easy/Normal/Hard; Web UI: Normal/Hard/Brutal) |

Language can be changed in the **Settings** menu (English / 简体中文). The game auto-detects your system language on first launch and saves the preference to `~/.civ-tui/config.json`.

## Architecture

```
civ-tui/                 (~5,900 lines of Go)
├── main.go              # Entry point
├── i18n/                # Internationalization (EN / ZH)
├── game/                # Game logic (zero TUI dependencies)
│   ├── model/           #   Domain models (unit, city, civ, tech, tile)
│   └── worldmap/        #   Procedural map generation
├── tui/                 # Terminal UI (Bubble Tea)
└── web/                 # Local Web UI server, JSON API, embedded static assets
```

Game logic is fully decoupled from the UI layers -- the `game/` package has zero TUI or Web UI imports, making it independently testable and reusable by both front ends.

## Tech Stack

- **[Go](https://go.dev/)** -- Simple, fast, compiles to a single binary
- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)** -- Elm Architecture for terminal apps
- **[Lipgloss](https://github.com/charmbracelet/lipgloss)** -- Terminal styling and layout
- **Go `net/http` + embedded static assets** -- Single-binary local Web UI
- **Vanilla HTML/CSS/JS** -- Lightweight browser client without a frontend build step
- **Custom fractal noise** -- No external dependencies for map generation

## Development

```bash
make build    # Compile
make run      # Build and run
go run . -web # Run the local Web UI on 127.0.0.1:8080
make test     # Run tests
make lint     # Run go vet
make clean    # Clean build artifacts
```

## Roadmap

All core gameplay phases are complete, and the local Web UI mode is available. Future directions include:

- Strategic & luxury resources
- Naval units
- Rivers & advanced terrain
- Culture, religion, and policy systems
- World wonders
- Trade routes
- TUI mouse support
- Battle animations
- Web UI parity improvements (diplomacy screen, help overlay, richer city management)
- More unit types and civilizations

## License

[GPL-3.0](LICENSE)
