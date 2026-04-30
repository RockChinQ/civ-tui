# civ-tui Web UI

> Status: local single-player Web UI, served by the Go binary.

The Web UI is an alternate front end for civ-tui. It does not fork the game rules: the browser client talks to a Go HTTP server, and the server uses the same `game/` package as the Bubble Tea TUI.

## Run

```bash
go run . -web
```

Then open:

```text
http://127.0.0.1:8080
```

To bind a different address:

```bash
go run . -web -addr 127.0.0.1:9090
```

The default bind address is loopback-only. Keep that default for normal local play.

## Runtime Model

- `main.go` selects Web mode with `-web`; without it, civ-tui starts the TUI.
- `web/server.go` serves the embedded files in `web/static/` and exposes JSON endpoints under `/api/*`.
- The server owns one active `game.Game` at a time.
- Access to the active game is serialized with a mutex, which is enough for local single-player use.
- Saves use the same JSON save system as the TUI: `~/.civ-tui/saves/save_1.json` through `save_10.json`.
- The old `save.json` path is migrated to slot 1 by the shared game startup path.

## Current Features

- New game setup with AI count, map size, and difficulty.
- Character-style browser map with terrain colors, fog of war, units, cities, cursor, selected unit, reachable tile hints, ranged targets, and destination markers.
- Side panels for tile details, selected unit/city details, civilization status, diplomacy relation summary, and recent log messages.
- Unit actions: move, attack by moving, found city, wait, ranged attack, set/clear destination, and build improvements.
- City actions: queue available units or buildings.
- Research menu for available technologies.
- Turn advancement, AI turns, victory/defeat/draw state display.
- Save/load through 10 save slots.

## Controls

| Input | Action |
|---|---|
| Click tile | Move cursor, select friendly unit, or move/attack adjacent tile |
| Arrow keys / `WASD` / `hjkl` | Move cursor |
| `Enter` | Move selected unit to cursor if adjacent |
| `F` | Found city with selected Settler |
| `B` | Build/train menu for current or first player city |
| `T` | Research menu |
| `R` | Enter ranged attack target mode |
| `G` | Enter destination target mode |
| `Z` | Wait selected unit |
| `I` | Open Worker improvement menu |
| `Space` | End turn |
| `Esc` | Cancel ranged/destination mode or deselect |

Top bar buttons also expose End Turn, Build, Research, Save, Load, and New Game.

## HTTP API

All action endpoints return the updated game-state JSON unless noted.

| Endpoint | Method | Body | Purpose |
|---|---|---|---|
| `/api/state` | `GET` | none | Return current state, or `{ "state": "none" }` before a game starts |
| `/api/new` | `POST` | `{ "numAI": 3, "mapSize": 1, "difficulty": 1 }` | Start a new game |
| `/api/saves` | `GET` | none | List all save slots |
| `/api/save` | `POST` | `{ "slot": 1 }` | Save active game to a slot |
| `/api/load` | `POST` | `{ "slot": 1 }` | Load a saved game |
| `/api/move` | `POST` | `{ "unitId": 12, "dx": 1, "dy": 0 }` | Move or melee-attack with a player unit |
| `/api/found-city` | `POST` | `{ "unitId": 12 }` | Found a city with a Settler |
| `/api/wait` | `POST` | `{ "unitId": 12 }` | Mark a unit as waiting |
| `/api/end-turn` | `POST` | `{}` | End the current turn |
| `/api/ranged` | `POST` | `{ "unitId": 12, "x": 20, "y": 8 }` | Fire at an enemy unit |
| `/api/improvement` | `POST` | `{ "unitId": 12, "type": "Farm" }` | Start Worker improvement construction |
| `/api/build` | `POST` | `{ "cityId": 5, "isUnit": true, "name": "Warrior" }` | Queue a unit or building |
| `/api/research` | `POST` | `{ "name": "Writing" }` | Set current research |
| `/api/set-dest` | `POST` | `{ "unitId": 12, "x": 25, "y": 11 }` | Set unit destination |
| `/api/clear-dest` | `POST` | `{ "unitId": 12 }` | Clear unit destination |

## Known Limitations

- Web mode is local single-player only. It has no authentication, multiplayer sessions, or per-user game state.
- One server process holds one active game. Starting a new game replaces the in-memory active game.
- The Web UI does not yet expose the full TUI surface: no dedicated diplomacy action screen, no in-browser help overlay, no language settings screen, and no advanced city production management.
- Reachable-tile hints in the browser are a lightweight client-side approximation; authoritative movement still happens in the Go game logic.
- The frontend is intentionally static HTML/CSS/JS and has no bundler, package step, or hot-reload workflow.

## Implementation Files

```text
main.go                 # -web / -addr flags and mode selection
web/server.go           # HTTP server, embedded static files, route registration
web/handlers.go         # JSON API handlers
web/state.go            # Game-to-JSON view model
web/static/index.html   # Browser shell
web/static/style.css    # Character-map UI styling
web/static/app.js       # Client rendering, input handling, API calls
```
