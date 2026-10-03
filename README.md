# Projekt Redware

A prototype of a Discord-style chat client that runs peer to peer instead of through a central server. It is a Wails v2 desktop application: the backend is Go on top of go-libp2p, the UI is Svelte 3 with Tailwind.

**Status: in development.** The networking core, direct messages and local storage exist; channel chat and voice are not reachable from the current UI, and login is a placeholder. Read [What works today](#what-works-today) before trying it.

## What it is

- **Identity.** Each installation has an Ed25519 key pair generated on first run and stored in `identity.key`; the libp2p peer id derived from it doubles as the user id.
- **Discovery and transport.** A libp2p host listens on TCP (`-port`, default 3000), announces itself on the LAN with mDNS (service tag `redware-discovery`) and auto-connects to peers it finds. A Kademlia DHT is created but no bootstrap peers are configured, so discovery is LAN-only in practice.
- **Messaging.** Direct messages travel over a custom libp2p stream protocol `/redware/chat/1.0.0` as protobuf messages (schema in `backend/p2p/proto/redware.proto`); channel history is served over `/redware/history/1.0.0`; channel chat, profile exchange and WebRTC voice signalling are designed to run over GossipSub topics named after the channel.
- **Storage.** Messages, joined nodes, channels and peers are kept in a per-user SQLite database.
- **UI.** Server rail, channel list, chat pane with Markdown (marked + DOMPurify), member list, profile panel, settings with three themes, accent colour, font size and custom CSS, avatar cropping, and a layout editor.
- **Tracker (separate program).** `cmd/tracker` is a small HTTP + Postgres registry (register, login, heartbeat, peer list, channel membership) meant to help peers find each other across networks. The desktop app does not call it yet.

## What works today

Confirmed by reading the code at this commit (nothing here has been run as part of writing this README):

- libp2p host start-up with persistent identity, mDNS discovery, automatic connection to LAN peers, and a peer list refreshed every 5 seconds in the UI.
- Direct messages between two peers over `/redware/chat/1.0.0`, saved to SQLite on both ends.
- Channel history requests over `/redware/history/1.0.0`.
- Creating, editing and deleting channels locally; saving "my server" metadata.
- Themes, custom CSS, Markdown rendering.

Not wired yet or known to be broken:

- `Login` accepts any username and password and `Register` always succeeds; there is no account system.
- The mounted chat widget never calls `JoinChannel`, so no GossipSub topic is subscribed from the UI. Sending to a `#channel` therefore fails with "Not joined to topic", and voice signalling, which uses the same topics, cannot start.
- The Go side emits `CHAT_MESSAGE` events while the chat widget listens for `P2P_MESSAGE`; incoming direct messages only appear after the chat is reopened.
- Pubsub packets are JSON-encoded structs wrapping a protobuf `oneof`; decoding them on the receiving side is expected to fail with the current code.
- Direct-message history queries filter on an empty channel id while messages are saved with the peer id as channel id, so DM history lists stay empty.
- `UpdateAvatar`, `GetGuildMembers`, `GetActiveNow`, `ConnectPeer` and the heartbeat loop are stubs; the Steam "Game Tracker" widget shows hardcoded mock data.
- About a dozen Svelte components are not reachable from the mounted layout, and the generated Wails bindings are one method behind `app.go`.

## Requirements

- Go 1.24.6 or newer and a C toolchain (`go-sqlite3` uses cgo)
- Node.js and npm
- The Wails v2 CLI (`wails`)
- `protoc` with `protoc-gen-go` only if you change `redware.proto`
- For the tracker: Docker (Postgres) or a PostgreSQL instance

## Build and run

```
make dev        # wails dev: live-reloading desktop app
make build      # wails build: packaged binary ("Projekt Redware")
make proto      # regenerate backend/p2p/proto/redware.pb.go
```

Wails runs `npm install` and `npm run build` in `frontend/` for you (`wails.json`). The app accepts one flag:

```
./Projekt\ Redware -port 3000     # libp2p TCP listen port
```

Files created in the working directory at runtime: `identity.key` (private key, mode 0600) and the SQLite database for the logged-in user. **Never commit `identity.key`**; anyone holding it can impersonate that peer.

### Tracker

```
docker compose up -d                 # Postgres 15 on localhost:5433 (user/password/db: redware)
go run ./cmd/tracker                 # from the repository root; reads cmd/tracker/schema.sql
```

Environment: `PORT` (default `8080`), `DATABASE_URL` (default `postgres://redware:redware@localhost:5433/redware?sslmode=disable`). The schema is dropped and recreated on every start ("dev only"), passwords are stored in plain text and any bearer token is accepted as a user id, so the tracker is a development stub, not a service to expose.

## Repository layout

```
main.go, app.go          Wails bootstrap and the bound App (all UI-facing Go methods)
backend/p2p/             libp2p host, stream protocols, protobuf schema, unused tracker client
backend/store/           SQLite store
cmd/tracker/             HTTP + Postgres tracker and its schema
frontend/src/            App.svelte, stores, components/ (widgets, modals, views), utils/ (webrtc, markdown, canvas)
frontend/wailsjs/        generated bindings and Wails runtime (do not edit)
build/                   Wails packaging templates (macOS plist, Windows installer)
Makefile, wails.json, docker-compose.yml, go.mod
```

## Known issues

- Two libp2p hosts are created when logging in from the UI (one at start-up, one in `StartP2P`), on the same port.
- The SQLite file name gets a double prefix (`redware_redware_<user>.db.db`).
- Streams have no message framing (one read into a fixed buffer: 4 KiB for requests, 1 MiB for history responses).
- Tailwind is loaded from a CDN at runtime, so the UI is unstyled offline.
- Topic and subscription maps are updated from goroutines without a mutex.

## License

No license file is included; the code is all rights reserved by default until one is added. Third-party notices: Wails runtime (MIT), the Nunito font under `frontend/src/assets/fonts/` (SIL OFL, currently unreferenced by the CSS).
