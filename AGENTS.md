# AGENTS.md

mrext (MiSTer Extensions) is a single Go module that builds a set of standalone
apps for the MiSTer FPGA Linux userspace: Remote, BGM, Favorites, GamesMenu,
LastPlayed, LaunchSync, PlayLog, Random and Search. The
project is under active development again. Its job is stock-MiSTer workflows:
generating menu shortcuts, tracking the running core, controller-driven TUIs and
a network remote, all without requiring Zaparoo Core on the device.

Human docs live in [README.md](README.md), [docs/dev.md](docs/dev.md) and
`docs/<app>.md`. This file covers what an agent cannot infer from them.

## Scope and Zaparoo

- mrext is actively maintained and standalone. Preserve familiar workflows,
  configuration, files, CLI behavior, and existing Remote client contracts.
  Do not tell end users they must migrate or describe the tools as retired.
- For new media applications and integrations, including MiSTer-only projects,
  use the Zaparoo Core daemon's public API or CLI. Do not use mrext's internal
  indexing, tracking, launch packages, or Remote media API as the default
  foundation when Core provides the required interface. Read
  [MIGRATE.md](MIGRATE.md) for mappings and genuine workflow differences.
- Remote's API remains supported for existing clients and specialist MiSTer
  functions such as INI editing, wallpapers, and menu management. Use mrext
  where those functions are needed without claiming full Core parity.
- Keep maintenance of these tools separate from runtime integration changes.
  Do not add a mandatory Core dependency, silently switch backends, or claim
  proposed integration is shipped.
- Public app docs should explain usage and relevant Zaparoo features, not
  internal strategy, maintenance sequencing, or speculative architecture.
- `/tmp/ACTIVEGAME` is a MiSTer-specific compatibility file that both mrext and
  Zaparoo Core maintain. Keep existing readers working; new integrations should
  prefer Core's public state, notification or MQTT interfaces.
- Zaparoo Core's dependency-free `mister/catalog` and `mister/mgl` modules are
  the source of truth for system IDs, aliases, folders, extensions, RBF paths,
  set names, groups, MGL slots and MGL generation. mrext must not grow a
  parallel catalog. Fix catalog data upstream, then bump `coreRevision` in
  `internal/gensystemmetadata/main.go` and the module version in `go.mod`.

## Commands

Run from the repo root. Go 1.27.1 is pinned in `go.mod`, golangci-lint v2.13.2
in `.github/workflows/lint.yml`. `mage` is `go run github.com/magefile/mage`
if it is not installed.

```sh
mage test                 # go test ./... (generates system metadata first)
mage lint                 # golangci-lint run ./... (same generation step)
mage lintFix              # apply auto-fixes, then re-run mage lint
mage build <app>          # host binary into _bin/<os>_<arch>/
mage mister <app>         # static linux/arm GOARM=7 CGO_ENABLED=0 binary for MiSTer
mage mister all           # every app; also rebuilds the Remote web UI
mage deploy <ip>          # build everything and scp it to root@<ip>:/media/fat/Scripts
mage generateSystemMetadata   # refresh pkg/games/system_metadata.gen.json
mage genSystemsDoc        # regenerate docs/systems.md after a catalog bump
npm ci --prefix web/remote && npm test --prefix web/remote && npm run build --prefix web/remote
actionlint .github/workflows/*.yml   # after editing workflows
```

- Before finishing any Go change, `mage test` and `mage lint` must both pass,
  and `mage mister <app>` must build for each touched app.
- Build, test and lint first generate `pkg/games/system_metadata.gen.json`
  (gitignored) from a pinned Zaparoo Core revision. The first run needs Git and
  network access and caches a checkout under the user cache dir. Set
  `ZAPAROO_CORE_SOURCE` to a local Core checkout to generate from it instead.
- Cross-compiling needs no C toolchain, container or Docker. Everything is pure
  Go, including SQLite (`modernc.org/sqlite`).
- `mage release` and `mage prepRelease` need UPX (`UPX_BIN`) and are for
  maintainers. Do not run them, tag, or publish releases unless asked.

## Architecture

- `cmd/<app>/` is one binary per folder. Apps never import each other; shared
  code goes in `pkg/`. `cmd/remote` is the only app with subpackages, and it
  embeds the web UI built from `web/remote` into `cmd/remote/_client/build`.
- `internal/` holds build-time generators only (`gensystemmetadata`,
  `gensystemsdoc`).
- `pkg/config` owns every MiSTer path, `/tmp` state file and the per-app INI
  model (`UserConfig`, `LoadUserConfig`, `MREXT_CONFIG`, `MREXT_APP_PATH`).
  A hardcoded path or magic value belongs here, not in an app.
- `pkg/games` adapts the Zaparoo catalog plus generated display metadata into
  `games.System`, and does stock-filesystem scanning, folder-to-system matching
  and BIOS/setname hooks. `pkg/mister` wraps MiSTer itself: MGL generation and
  launching via `/dev/MiSTer_cmd`, `MiSTer.ini`, `user-startup.sh`, screen
  mode, shared memory and update helpers. `pkg/input` fakes keyboard and mouse
  with uinput.
- `pkg/tracker` watches `/tmp/CORENAME` and friends, resolves the active core
  and game, writes `/tmp/ACTIVEGAME` and emits events. `pkg/service` is the
  daemon harness (pid file, `/tmp/<app>.log`, start/stop, logger). PlayLog,
  LastPlayed and Remote are built on both.
- `pkg/gamesdb` is the bbolt game-name index at `config.GamesDB`
  (`/media/fat/Scripts/.config/mrext/games.db`) used by Search, LaunchSync and
  Remote. PlayLog keeps its own SQLite database.
- `pkg/tui` is the shared controller-first terminal UI on tview/tcell: themes,
  responsive/CRT layout, page frame, button bar, menu list, dialogs, on-screen
  keyboard and the MiSTer TTY retry in `BuildAndRetry`. All interactive apps
  must be fully usable with only a controller.
- `pkg/bgm`, `pkg/favorites` and `pkg/gamesmenu` are domain packages for the Go ports;
  their `cmd/` folders contain only CLI parsing and TUI.
- `scripts/` is internal. `gamesmenu.sh`, `bgm.sh` and `favorites.sh` are the
  original Python implementations, kept as behavioural references for their Go
  ports; `generate_repo.py` builds Downloader databases in CI.
- `releases/*.json` and `releases/all.json` are Downloader databases written by
  `mage release` and the `repo.yml` workflow after a tagged release. Never edit
  them by hand. `docs/systems.md` is generated too.

## Conventions that are not enforced by tooling

- Release binaries are named `<app>.sh` although they are native ARM
  executables. The suffix is load-bearing: MiSTer's Scripts menu, Downloader
  entries and `user-startup.sh` hooks all depend on it. Keep it.
- Porting a Python script to Go: runtime behaviour must stay quirk-for-quirk
  compatible. Config file names, keys and locations, folder layouts, socket
  protocols, CLI arguments, startup-hook lines and generated file contents are
  all contracts with installed systems. Users must never need to change
  existing config or folders. Only the TUI may change, and new TUIs copy the
  Favorites shape: `cmd/favorites/ui.go` and `settings.go` (main page, staged
  Settings page with Change/Save/Cancel, footer help, themes). Every intentional
  deviation goes in the "Compatibility notes" section of `docs/<app>.md`.
- Existing user config files are never rewritten or reformatted wholesale.
  Create commented defaults only when missing, and save changes with atomic
  writes that preserve unknown keys and comments.
- Remote's REST and WebSocket API in `docs/remote-api.md` is a contract with
  existing clients. Extend it, do not change existing shapes.
- App-facing changes update `docs/<app>.md` in the same change. README.md holds
  only the per-app blurb and download links.
- Apps minimise SD-card writes, disk footprint and daemon CPU, and avoid
  depending on programs shipped with the MiSTer Linux image. BGM's audio
  players (`mpg123`, `ogg123`, `aplay`, `aplaymidi`, `vgmplay`) are the known
  exception; do not add new ones.
- Every new `.go` file needs the GPL header block; copy it from any existing
  file or `mage lint` fails on `goheader`. The `log` stdlib package is banned
  (use `pkg/service` logging or explicit command output), and so are `ioutil`
  and `http.DefaultClient`. Wrap returned errors with context.

## Testing

- Standard library `testing` only, no testify. Tests are table-light and
  fixture-heavy: build a fake MiSTer root under `t.TempDir()` and inject it
  (`bgm.RootedPaths`, `favorites.NewManagerWithPaths`,
  `gamesmenu.NewManagerWithPaths`, `config.LoadUserConfigAt`).
  Tests must never touch the real `/media/fat`, `/tmp/CORENAME` or
  `/dev/MiSTer_cmd`; if a package cannot be pointed at a temp root, add that
  seam first.
- TUI tests render through `tcell.NewSimulationScreen` and assert on the drawn
  cells or the focused widget; see `pkg/tui/foundation_test.go` and
  `cmd/favorites/ui_test.go`.
- Add or update tests for behaviour you change. Do not delete or skip a failing
  test to make lint or CI pass.
- Desktop runs of BGM, Favorites and GamesMenu take `--root PATH` to operate on a local
  fake MiSTer filesystem. Other apps mostly need a real device.
- Device testing: `mage mister <app>`, copy `_bin/linux_arm/<app>.sh` to
  `/media/fat/Scripts/` on a MiSTer, run it from the Scripts menu. Never run
  anything that mutates a live MiSTer (launching cores, editing `MiSTer.ini`,
  startup hooks, deleting menu entries) without the user's explicit go-ahead
  for that device.

## Git and CI

- Commits follow Conventional Commits with the app as scope, for example
  `feat(bgm): ...`, `fix(favorites): ...`, `docs(random): ...`, `ci: ...`.
  Branches are `<type>/<topic>`, such as `refactor/bgm-go`.
- CI on pull requests runs golangci-lint, the Remote web UI tests and build, and
  a dependency review that fails on any new advisory. Tags trigger
  `build-and-release.yml`, which runs `mage prepRelease`; `repo.yml` then
  regenerates the Downloader databases.
- Never commit `.env`, `_bin/`, `cmd/remote/_client/build`, generated metadata
  JSON, `.pi/` or `node_modules`. Never commit secrets or device addresses.
- GitHub issues are labelled per app (`bgm`, `favorites`, `remote`, `gamesmenu`
  and so on) and use the same `type(scope): summary` titles.
