# Go Bot Tutorial

This guide shows how to build a minimal BombAhead bot with Go, test it locally, package it as a container image, and push it to `ghcr.io` so you can register it on the platform later.

<p><strong>Note:</strong> This page is a practical tutorial. The old content was only added for testing and has been replaced.</p>

## 1. Prerequisites

- Go 1.23+
- Docker
- A GitHub account (for `ghcr.io`)

<p>Podman works as well. You can replace <code>docker</code> with <code>podman</code> in the commands below.</p>

## 2. Create a new bot project

```bash
mkdir my-bombahead-go-bot
cd my-bombahead-go-bot
go mod init github.com/<your-user>/my-bombahead-go-bot
go get github.com/N3moAhead/bombahead-go
```

Recommended structure:

```txt
my-bombahead-go-bot/
  go.mod
  go.sum
  cmd/
    bot/
      main.go
```

## 3. Write a minimal bot

Create `cmd/bot/main.go`:

```go
package main

import (
	"log"

	bombahead "github.com/N3moAhead/bombahead-go"
)

type SimpleBot struct{}

func (b *SimpleBot) GetNextMove(state *bombahead.GameState, h *bombahead.GameHelpers) bombahead.Action {
	if state == nil || state.Me == nil {
		return bombahead.DoNothing
	}

	me := state.Me.Pos

	if !h.IsSafe(me) {
		safe := h.GetNearestSafePosition(me)
		return h.GetNextActionTowards(me, safe)
	}

	boxPos, found := h.FindNearestBox(me)
	if found {
		dist := me.DistanceTo(boxPos)
		if dist == 1 {
			return bombahead.PlaceBomb
		}
		if dist > 1 {
			return h.GetNextActionTowards(me, boxPos)
		}
	}

	return bombahead.DoNothing
}

func main() {
	log.Println("Starting SimpleBot...")
	bombahead.Run(&SimpleBot{})
}
```

## 4. Learn the basic SDK concepts

The only interface you must implement:

```go
type Bot interface {
    GetNextMove(state *GameState, helpers *GameHelpers) Action
}
```

Entry point:

```go
func Run(userBot Bot)
```

`Run(...)` handles:

- WebSocket connection
- ready-state updates
- state parsing each tick
- helper construction
- sending your returned action

Important actions:

- `MoveUp`
- `MoveDown`
- `MoveLeft`
- `MoveRight`
- `PlaceBomb`
- `DoNothing`

Useful helpers to start with:

- `IsSafe(pos)`
- `GetNearestSafePosition(pos)`
- `FindNearestBox(pos)`
- `GetNextActionTowards(start, target)`

## 5. Run and test locally

Run your bot:

```bash
go run ./cmd/bot
```

For a realistic local test, run server + a default opponent image, then your bot.

Default images used by BombAhead tooling:

- Server: `ghcr.io/n3moahead/bombahead/os-server:latest`
- Opponent (aggressive): `ghcr.io/n3moahead/bomber:self-destruct`
- Opponent (passive): `ghcr.io/n3moahead/bomber:idle`

```bash
# Terminal 1: game server
docker run --rm -p 8038:8038 ghcr.io/n3moahead/bombahead/os-server:latest

# Terminal 2: default opponent
docker run --rm --network host ghcr.io/n3moahead/bomber:idle

# Terminal 3: your bot from source
go run ./cmd/bot
```

Your bot should connect to `ws://localhost:8038/ws`.

## 6. Containerize your bot

Create a `Dockerfile`:

```dockerfile
FROM docker.io/golang:1.25-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /app/bot ./cmd/bot/main.go

FROM docker.io/alpine:latest
WORKDIR /root/
COPY --from=builder /app/bot .

CMD ["./bot"]
```

Build locally:

```bash
docker build -t my-bombahead-go-bot:local .
```

## 7. Push to GitHub Container Registry (`ghcr.io`)

Choose your final image name, for example:

`ghcr.io/<github-username>/bombahead-go-bot:v1`

Login:

```bash
echo <YOUR_GITHUB_PAT> | docker login ghcr.io -u <github-username> --password-stdin
```

Tag and push:

```bash
docker tag my-bombahead-go-bot:local ghcr.io/<github-username>/bombahead-go-bot:v1
docker push ghcr.io/<github-username>/bombahead-go-bot:v1
```

<p><strong>Important:</strong> <code>ghcr.io</code> images are usually private by default. Make your package/image public, otherwise the BombAhead match runner cannot pull it and your matches will fail before game start.</p>

GitHub visibility path (typical):

- GitHub profile/org
- Packages
- Your container package
- Package settings
- Change visibility to <strong>Public</strong>

## 8. Register bot on BombAhead

After your image is public and pullable:

1. Log in to BombAhead
2. Open `/bots/new`
3. Enter bot name + description
4. Paste your image URL, e.g. `ghcr.io/<github-username>/bombahead-go-bot:v1`
5. Save

After that, BombAhead can schedule matches for your bot.

## 9. Recommended next improvements

- Add a bomb cooldown strategy (avoid placing bombs without an escape route).
- Use `FindNearestBox` plus safety checks before path commits.
- Version images (`v1`, `v2`, `v3`) instead of overwriting `latest`.
