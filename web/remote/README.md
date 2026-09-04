# Remote web UI

React web UI embedded in the `remote.sh` binary and served by Remote. Source moved here from [`wizzomafizzo/mrext-client`](https://github.com/wizzomafizzo/mrext-client), with original Git history retained.

Remote communicates through its established REST and WebSocket APIs. See [`../../docs/remote-api.md`](../../docs/remote-api.md) for contracts.

## Development

Requires Node.js 24 or newer and npm 12.0.2.

```sh
npm ci
npm run dev
```

Development server can target another Remote instance through **Settings → Remote**.

## Checks

```sh
npm test
npm run build
npm run audit
```

Production build is written directly to `cmd/remote/_client/build` for Go embedding. Generated files and `node_modules` are not committed. Root Mage and release workflows build this client before compiling Remote.

Android and iOS wrappers were retired when source moved into this repository. Remote is distributed as embedded web UI only.
