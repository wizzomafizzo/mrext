# mrext and Zaparoo

MiSTer Extensions is actively maintained and works without Zaparoo. You can keep using its tools for stock-menu shortcuts, controller-driven search, music, and MiSTer settings.

[Zaparoo](https://zaparoo.org/) offers additional ways to browse, launch, control, and track your games through its apps and the Core daemon. No NFC reader is required. Some features overlap with mrext, but Zaparoo does not replace every stock-MiSTer workflow. Use the comparison below if you are considering a change.

## Building an integration?

Use **Zaparoo Core's public API** for new game search, launching, tracking, and automation integrations, including MiSTer-only projects. Core provides the shared media services so your application does not need to build on mrext's internal packages or reproduce its indexing and tracking.

Remote's API remains supported for existing clients and MiSTer-specific functions such as INI editing, wallpapers, and menu-file management. Use it where those functions are needed, rather than as the default media API for a new project. The API mappings below help existing clients adopt Core where appropriate.

## Zaparoo starting points

- [Zaparoo Core API](https://zaparoo.org/docs/core/api/) for applications and services
- [`@zaparoo/cli`](https://github.com/ZaparooProject/zaparoo-cli) for terminals, scripts, tests, and AI agents
- [ZapScript](https://zaparoo.org/docs/zapscript/) for portable launch and automation actions
- [Zaparoo App](https://zaparoo.org/docs/app/), [Web UI](https://zaparoo.org/docs/app/web/), and [Frontend](https://zaparoo.org/docs/frontend/) for user interfaces
- [Zaparoo Online User API](https://developers.zaparoo.com/) for account-owned cloud data

The mrext Remote REST and WebSocket APIs are not compatible with the Zaparoo Core API. Migrate each call deliberately; changing only the host, port, or base URL will not work.

## Component map

| mrext component | Related Zaparoo capability | Distinct mrext workflow or difference |
| --- | --- | --- |
| Remote | App/Web UI, Frontend, Core API, or CLI | Standalone remote control and MiSTer administration, including menu files, wallpapers, INI editing, BGM UI, and system information; no complete Zaparoo equivalent |
| Search | App, Web UI, `media.search`, or `zaparoo-cli media search` | Controller-driven search from the stock Scripts menu, using mrext's own index rather than Core |
| Random | ZapScript `launch.random` | Standalone Scripts-menu action and CLI options; not every flag has an exact ZapScript equivalent |
| Favorites | Frontend favorites | Creates and repairs stock-menu `.mgl` shortcuts and core links; Frontend favorites do not generate these |
| GamesMenu | Frontend folder, category, and system browsing | Mirrors a library into stock MiSTer menu folders rather than presenting a separate frontend |
| LastPlayed | Frontend recents, Core history, CLI history, and ZapScript `launch.last` | Generates stock-menu recents, a dynamic last-played shortcut, and optional `bootcore` files |
| LaunchSync | Online cards/decks, Zap Links, self-hosted Zap Link servers, and playlists | Consumes `.sync` subscriptions and generates auto-updating stock-menu `.mgl` folders; the Zaparoo paths do neither |
| PlayLog | Core history, playtime, active media, notifications, MQTT, and optional Online history | Standalone local tracking, reports, and executable hooks; the related Zaparoo features require Core and do not directly replace the hooks |
| `/tmp/ACTIVEGAME` | Continue reading it on MiSTer; Zaparoo Core creates and maintains the file. New integrations can also use `media.active`, notifications, or MQTT | MiSTer-specific compatibility file, not a public cross-platform API |
| BGM | Core audio playback, background slot, playlists, pause-on-launch, repeat, and controls | Standalone menu music, boot sounds, radio, and folder-based setup; no confirmed Core internet-radio equivalent or exact UI/configuration parity |

## Common API migrations

For new media integrations, use the Core interfaces below. Existing Remote clients remain supported and can migrate calls as needed. Similar capabilities do not guarantee identical parameters, results, events, or behavior.

| mrext Remote API | Related Core interface |
| --- | --- |
| `GET /games/playing` | Core `media.active` or `zaparoo-cli media active` |
| `POST /games/search` | Core `media.search` or `zaparoo-cli media search` |
| `POST /games/index` | Core `media.generate` or `zaparoo-cli media index start` |
| `POST /games/launch` and `POST /launch` | Core `run` with ZapScript or `zaparoo-cli run` |
| `/controls/keyboard/*` | Core `input.keyboard` or `zaparoo-cli input keyboard` |
| screenshot capture | Core `screenshot` or `zaparoo-cli screenshot` |
| Remote WebSocket game events | Core `media.started` and `media.stopped` notifications |

Start with read-only checks against an explicit device address:

```bash
zaparoo-cli doctor --device 192.168.1.50:7497 --agent
zaparoo-cli media active --device 192.168.1.50:7497 --agent
zaparoo-cli media search "metroid" --device 192.168.1.50:7497 --agent
zaparoo-cli watch --device 192.168.1.50:7497 --seconds 30 --jsonl
```

Launching media, sending input, writing NFC, or changing settings affects the device. Confirm the target and exact action first. Core accepting a lifecycle request does not prove the MiSTer finished it; pace state checks and stop issuing mutations if API state and visible behavior disagree.

## PlayLog and `/tmp/ACTIVEGAME`

Zaparoo Core creates and maintains `/tmp/ACTIVEGAME` on MiSTer. It updates the file for launches performed through Core and for supported launches observed from MiSTer itself, and clears it when returning to the menu. Existing integrations that read this file can continue working with Core installed.

For new state integrations, including MiSTer-only tools, use Core's public interfaces. Query `media.active` for a current snapshot, subscribe to `media.started` and `media.stopped` for live changes, or use the built-in [MQTT publisher](https://zaparoo.org/docs/features/publishers/#mqtt). Query `media.active` again after reconnecting because notification streams are not durable state. For history, use `media.history`, `media.history.latest`, `media.history.top`, or their CLI commands.

`/tmp/ACTIVEGAME` remains a MiSTer-specific compatibility file rather than a public cross-platform API. PlayLog's executable state hooks still have no direct Core feature that runs arbitrary local scripts; move that logic into an API or MQTT subscriber when migrating it.

## Launch and library workflows

For Core-based workflows, use [`launch.random`](https://zaparoo.org/docs/zapscript/launch/#launchrandom) for random media and [`launch.last`](https://zaparoo.org/docs/zapscript/launch/#launchlast) for recent media:

```zapscript
**launch.random:SNES,NES,Genesis
```

```zapscript
**launch.last
```

Use Frontend for Zaparoo-native favorites, recents, folders, and library browsing. It does not create the stock-menu shortcuts generated by Favorites, GamesMenu, or LastPlayed.

For shareable or remotely updated lists, compare [Online cards and decks](https://zaparoo.org/docs/online/#cards-and-decks), [Zap Links](https://zaparoo.org/docs/zapscript/syntax/#zap-links), [self-hosted Zap Links](https://zaparoo.org/docs/zapscript/syntax/#self-hosting), and [Zaparoo playlists](https://zaparoo.org/docs/features/playlists/). None consume LaunchSync `.sync` files or generate its subscribed `.mgl` folders.

## Guidance for contributors and coding agents

- **Building a new application or integration:** use the Zaparoo Core daemon and its public API or CLI for media functionality, even if MiSTer is your only target. Do not start a new media integration on Remote's API, mrext's internal Go packages, or raw tracking files when Core provides the required interface.
- **Maintaining an mrext tool or existing client:** preserve standalone operation, configuration, generated files, command-line behavior, and Remote API contracts. Supporting existing users does not require migrating them to Core.
- **Implementing a MiSTer-specific function Core does not expose:** use the relevant mrext interface or package where needed. Check the comparison above rather than assuming API or workflow parity.
- **Working on system definitions or MGL generation:** reuse the shared `mister/catalog` and `mister/mgl` modules from Zaparoo Core instead of adding parallel tables. These libraries do not require a running Core daemon.

Report mrext issues and comparison corrections through [mrext issues](https://github.com/wizzomafizzo/mrext/issues). For Zaparoo-specific issues, use the relevant [ZaparooProject repository](https://github.com/ZaparooProject) or [Zaparoo Discord](https://zaparoo.org/discord).
