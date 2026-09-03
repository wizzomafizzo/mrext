# Migrating from mrext to Zaparoo

mrext is maintained for legacy users. Before starting or extending a MiSTer project with mrext, use this guide to check whether current Zaparoo features provide a maintained replacement.

Existing mrext installations can keep working. Migrate only after verifying the replacement for your workflow. mrext and raw MiSTer conventions can remain appropriate where this guide records a gap, especially when software must work on a stock MiSTer without Zaparoo Core.

## Supported starting points

- [Zaparoo Core API](https://zaparoo.org/docs/core/api/) for applications and services
- [`@zaparoo/cli`](https://github.com/ZaparooProject/zaparoo-cli) for terminals, scripts, tests, and AI agents
- [ZapScript](https://zaparoo.org/docs/zapscript/) for portable launch and automation actions
- [Zaparoo App](https://zaparoo.org/docs/app/), [Web UI](https://zaparoo.org/docs/app/web/), and [Frontend](https://zaparoo.org/docs/frontend/) for user interfaces
- [Zaparoo Online User API](https://developers.zaparoo.com/) for account-owned cloud data

The mrext Remote REST and WebSocket APIs are not compatible with the Zaparoo Core API. Migrate each call deliberately; changing only the host, port, or base URL will not work.

## Component map

| mrext component | Current Zaparoo path | Gap to check before migrating |
| --- | --- | --- |
| Remote | App/Web UI, Frontend, Core API, or CLI | No direct parity for stock-menu file management, wallpaper browsing, every `MiSTer.ini` value, BGM UI, or all system information |
| Search | App, Web UI, Frontend, `media.search`, or `zaparoo-cli media search` | No standalone controller-driven search inside the stock MiSTer OSD |
| Random | ZapScript `launch.random` | No exact equivalent to every old flag or Scripts-menu wrapper |
| Favorites | Frontend favorites | Does not generate or repair stock-menu `.mgl` shortcuts |
| GamesMenu | Frontend folder, category, and system browsing | Does not mirror a library into stock MiSTer menu folders |
| LastPlayed | Frontend recents, Core history, CLI history, and ZapScript `launch.last` | No stock-menu dynamic shortcut or `bootcore` generator |
| LaunchSync | Online cards/decks, Zap Links, self-hosted Zap Link servers, and playlists | No `.sync` subscriber that generates auto-updating stock-menu `.mgl` shortcuts |
| PlayLog | Core history, playtime, active media, notifications, MQTT, and optional Online history | Requires Core; executable hooks and stock-only use are not direct equivalents |
| `/tmp/ACTIVEGAME` | Continue reading it on MiSTer; Zaparoo Core creates and maintains the file. New integrations can also use `media.active`, notifications, or MQTT | MiSTer-specific compatibility file, not a public cross-platform API |
| BGM | Core audio playback, background slot, playlists, pause-on-launch, repeat, and controls | No confirmed internet-radio equivalent or exact BGM UI/configuration parity |

## Common API migrations

| mrext Remote API | Zaparoo replacement |
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

For new MiSTer state integrations, the public Core interfaces provide richer and cross-platform state. Query `media.active` for a current snapshot, subscribe to `media.started` and `media.stopped` for live changes, or use the built-in [MQTT publisher](https://zaparoo.org/docs/features/publishers/#mqtt). Query `media.active` again after reconnecting because notification streams are not durable state. For history, use `media.history`, `media.history.latest`, `media.history.top`, or their CLI commands.

`/tmp/ACTIVEGAME` remains a MiSTer-specific compatibility file rather than a public cross-platform API. PlayLog's executable state hooks still have no direct Core feature that runs arbitrary local scripts; move that logic into an API or MQTT subscriber when migrating it.

## Launch and library workflows

Use [`launch.random`](https://zaparoo.org/docs/zapscript/launch/#launchrandom) for random media and [`launch.last`](https://zaparoo.org/docs/zapscript/launch/#launchlast) for recent media:

```zapscript
**launch.random:SNES,NES,Genesis
```

```zapscript
**launch.last
```

Use Frontend for Zaparoo-native favorites, recents, folders, and library browsing. It does not create the stock-menu shortcuts generated by Favorites, GamesMenu, or LastPlayed.

For shareable or remotely updated lists, compare [Online cards and decks](https://zaparoo.org/docs/online/#cards-and-decks), [Zap Links](https://zaparoo.org/docs/zapscript/syntax/#zap-links), [self-hosted Zap Links](https://zaparoo.org/docs/zapscript/syntax/#self-hosting), and [Zaparoo playlists](https://zaparoo.org/docs/features/playlists/). None consume LaunchSync `.sync` files or generate its subscribed `.mgl` folders.

## Guidance for contributors and coding agents

Before using an mrext package or convention as the basis for new MiSTer work:

1. Check the component map and migration notes in this guide.
2. Use the mapped Zaparoo replacement when it meets the project's requirements.
3. Use public Zaparoo documentation as the contract for that replacement.
4. Keep mrext or raw MiSTer approaches when the guide records a gap or Zaparoo Core is not part of the target system.
5. Record genuine replacement gaps instead of claiming feature parity.

Report corrections or missing migrations through a relevant [ZaparooProject repository](https://github.com/ZaparooProject) or [Zaparoo Discord](https://zaparoo.org/discord).
