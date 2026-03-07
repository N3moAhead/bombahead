# Python Bot Tutorial

This guide shows how to build a minimal BombAhead bot with Python, test it locally, package it as a container image, and push it to `ghcr.io` so you can register it on the platform later.

<p><strong>Note:</strong> This page is a practical tutorial. For detailed API reference, consult the SDK's source or README.</p>

## 1. Prerequisites

- Python 3.10+
- Docker
- A GitHub account (for `ghcr.io`)

<p>Podman works as well. You can replace <code>docker</code> with <code>podman</code> in the commands below.</p>

## 2. Create a new bot project

```bash
mkdir my-bombahead-py-bot
cd my-bombahead-py-bot
python3 -m venv venv
source venv/bin/activate
pip install bombahead-py
```

Recommended structure:

```txt
my-bombahead-py-bot/
  bot.py
  requirements.txt
```

Create a `requirements.txt`:

```txt
bombahead-py
```

## 3. Write a minimal bot

Create `bot.py`:

```python
from bombahead import Action, Bot, GameHelpers, GameState, run

class SimpleBot(Bot):
    def get_next_move(self, state: GameState, helpers: GameHelpers) -> Action:
        # If our bot is not initialized or dead, do nothing
        if not state.me:
            return Action.DO_NOTHING

        me_pos = state.me.pos

        # If our tile is unsafe, move to the nearest safe tile.
        if not helpers.is_safe(me_pos):
            safe_pos = helpers.get_nearest_safe_position(me_pos)
            return helpers.get_next_action_towards(me_pos, safe_pos)

        # Otherwise move toward the nearest box.
        box_pos, found = helpers.find_nearest_box(me_pos)
        if found:
            dist = me_pos.distance_to(box_pos)
            if dist == 1:
                return Action.PLACE_BOMB
            if dist > 1:
                return helpers.get_next_action_towards(me_pos, box_pos)

        return Action.DO_NOTHING

if __name__ == "__main__":
    print("Starting SimpleBot...")
    run(SimpleBot())
```

## 4. Learn the basic SDK concepts

The architecture of the SDK is built around a simple reactive event loop. Your bot operates as an asynchronous callback invoked upon receiving a state update from the server.

The only interface you must implement is the `Bot` protocol:

```python
class Bot(Protocol):
    def get_next_move(self, state: GameState, helpers: GameHelpers) -> Action: ...
```

Entry point:

```python
run(bot)
```

`run(...)` handles:
- WebSocket connection
- ready-state updates
- state parsing each tick
- helper construction
- sending your returned action

Important actions (`Action` enum):
- `Action.MOVE_UP`
- `Action.MOVE_DOWN`
- `Action.MOVE_LEFT`
- `Action.MOVE_RIGHT`
- `Action.PLACE_BOMB`
- `Action.DO_NOTHING`

Useful helpers to start with (`GameHelpers`):
- `is_safe(pos)`
- `get_nearest_safe_position(pos)`
- `find_nearest_box(pos)`
- `get_next_action_towards(start, target)`

## 5. Run and test locally

Run your bot:

```bash
python bot.py
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
python bot.py
```

Your bot should connect to `ws://localhost:8038/ws`.

## 6. Containerize your bot

Create a `Dockerfile`:

```dockerfile
FROM python:3.12-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY bot.py .

CMD ["python", "bot.py"]
```

Build locally:

```bash
docker build -t my-bombahead-py-bot:local .
```

## 7. Push to GitHub Container Registry (`ghcr.io`)

Choose your final image name, for example:

`ghcr.io/<github-username>/bombahead-py-bot:v1`

Login:

```bash
echo <YOUR_GITHUB_PAT> | docker login ghcr.io -u <github-username> --password-stdin
```

Tag and push:

```bash
docker tag my-bombahead-py-bot:local ghcr.io/<github-username>/bombahead-py-bot:v1
docker push ghcr.io/<github-username>/bombahead-py-bot:v1
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
4. Paste your image URL, e.g. `ghcr.io/<github-username>/bombahead-py-bot:v1`
5. Save

After that, BombAhead can schedule matches for your bot.

## 9. Recommended next improvements

- Add a bomb cooldown strategy (avoid placing bombs without an escape route).
- Use `find_nearest_box` plus safety checks before path commits.
- Version images (`v1`, `v2`, `v3`) instead of overwriting `latest`.
