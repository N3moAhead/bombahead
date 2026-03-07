# Rust Bot Tutorial

This guide shows how to build a minimal BombAhead bot with Rust, test it locally, package it as a container image, and push it to `ghcr.io` so you can register it on the platform later.

<p><strong>Note:</strong> This page is a practical tutorial. For detailed API reference, consult the SDK's source or README.</p>

## 1. Prerequisites

- Rust 1.70+
- Docker
- A GitHub account (for `ghcr.io`)

<p>Podman works as well. You can replace <code>docker</code> with <code>podman</code> in the commands below.</p>

## 2. Create a new bot project

```bash
cargo new my-bombahead-rs-bot
cd my-bombahead-rs-bot
cargo add bombahead-rs
```

Recommended structure:

```txt
my-bombahead-rs-bot/
  Cargo.toml
  src/
    main.rs
```

## 3. Write a minimal bot

Replace the contents of `src/main.rs` with:

```rust
use bombahead_rs::{Action, Bot, GameState, GameHelpers};

struct SimpleBot;

impl Bot for SimpleBot {
    fn get_next_move(&mut self, state: &GameState, h: &GameHelpers) -> Action {
        let me = match &state.me {
            Some(player) => &player.pos,
            None => return Action::DoNothing,
        };

        // Run from explosions
        if !h.is_safe(me) {
            let safe = h.get_nearest_safe_position(me);
            return h.get_next_action_towards(me, &safe);
        }

        // Hunt boxes
        if let Some(box_pos) = h.find_nearest_box(me) {
            let dist = me.distance_to(&box_pos);
            if dist == 1 {
                return Action::PlaceBomb;
            } else if dist > 1 {
                return h.get_next_action_towards(me, &box_pos);
            }
        }

        Action::DoNothing
    }
}

fn main() {
    println!("Starting SimpleBot...");
    bombahead_rs::run(SimpleBot);
}
```

## 4. Learn the basic SDK concepts

The SDK handles connecting to the server over WebSocket, receiving game state messages, building helpers, and sending the returned action to the server. 

The only interface you must implement is the `Bot` trait:

```rust
pub trait Bot {
    fn get_next_move(&mut self, state: &GameState, helpers: &GameHelpers) -> Action;
}
```

Entry point:

```rust
bombahead_rs::run(user_bot)
```

Important actions (`Action` enum):
- `Action::MoveUp`
- `Action::MoveDown`
- `Action::MoveLeft`
- `Action::MoveRight`
- `Action::PlaceBomb`
- `Action::DoNothing`

Useful helpers to start with (`GameHelpers`):
- `is_safe(pos)`
- `get_nearest_safe_position(start)`
- `find_nearest_box(start)`
- `get_next_action_towards(start, target)`

## 5. Run and test locally

Run your bot:

```bash
cargo run
```

For a realistic local test, run server + a default opponent image, then your bot.

Default images used by BombAhead tooling:

- Server: `ghcr.io/n3moahead/bombahead/server:latest`
- Opponent (aggressive): `ghcr.io/n3moahead/bomber:self-destruct`
- Opponent (passive): `ghcr.io/n3moahead/bomber:idle`

```bash
# Terminal 1: game server
docker run --rm -p 8038:8038 ghcr.io/n3moahead/bombahead/server:latest

# Terminal 2: default opponent
docker run --rm --network host ghcr.io/n3moahead/bomber:idle

# Terminal 3: your bot from source
cargo run
```

Your bot should connect to `ws://localhost:8038/ws`.

## 6. Containerize your bot

Create a `Dockerfile`:

```dockerfile
FROM rust:1.75 AS builder
WORKDIR /app
COPY . .
RUN cargo build --release

FROM debian:bookworm-slim
WORKDIR /app
COPY --from=builder /app/target/release/my-bombahead-rs-bot ./bot
CMD ["./bot"]
```

Build locally:

```bash
docker build -t my-bombahead-rs-bot:local .
```

## 7. Push to GitHub Container Registry (`ghcr.io`)

Choose your final image name, for example:

`ghcr.io/<github-username>/bombahead-rs-bot:v1`

Login:

```bash
echo <YOUR_GITHUB_PAT> | docker login ghcr.io -u <github-username> --password-stdin
```

Tag and push:

```bash
docker tag my-bombahead-rs-bot:local ghcr.io/<github-username>/bombahead-rs-bot:v1
docker push ghcr.io/<github-username>/bombahead-rs-bot:v1
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
4. Paste your image URL, e.g. `ghcr.io/<github-username>/bombahead-rs-bot:v1`
5. Save

After that, BombAhead can schedule matches for your bot.

## 9. Recommended next improvements

- Add a bomb cooldown strategy (avoid placing bombs without an escape route).
- Use `find_nearest_box` plus safety checks before path commits.
- Version images (`v1`, `v2`, `v3`) instead of overwriting `latest`.
