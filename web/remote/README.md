# mrext-client

This is the web UI component of [Remote](https://github.com/wizzomafizzo/mrext).
It is a standalone React app which is embedded into the Remote binary and
served statically from Remote's web server.

All communication between the
client and server is done through a combination of a REST API and WebSockets.
It can be hosted anywhere, as long as the API address is set correctly.

[Download Android APK](https://github.com/wizzomafizzo/mrext-client/releases/latest/download/mrext-client.apk)

## Setup

### Prerequisites

- [Node.js](https://nodejs.org/en/) (v16 LTS or higher)
- NPM (comes with Node.js)

These libraries are installed automatically but are notable:

- Vite is used to build the app
- All code is written in Typescript
- Prettier is used for code formatting
- React is used for basically everything
- Material UI is used for the UI components
- Axios for API requests
- React Router for routing
- React Query for data fetching
- Zustand for state management

### Dev Environment

1. Clone the repository
2. Run `npm install` to install dependencies
3. Run `npm run start` to start the development server
4. Browse to Settings > Remote and change the API settings to point to your
   MiSTer running Remote

### Production Build

1. Run `vite build` to build the app
2. Copy the entire `build` directory to `cmd/remote/_client` in the Remote
   repository

### Publishing a release

1. Create a tag `git tag vx.x` with your chosen version number
2. Push the tag: `git push origin --tags`
