# TypeScript Bot Tutorial

This guide shows how to build a minimal BombAhead bot with TypeScript, test it locally, package it as a container image, and push it to `ghcr.io` so you can register it on the platform later.

<p><strong>Note:</strong> The legacy <code>client_js/</code> example in this repository can be outdated. For new bots, use the SDK package <code>bombahead-js</code>.</p>

## 1. Prerequisites

- Node.js 20+
- npm
- Docker
- A GitHub account (for `ghcr.io`)

<p>Podman works as well. You can replace <code>docker</code> with <code>podman</code> in the commands below.</p>

## 2. Create a new bot project

```bash
mkdir my-bombahead-ts-bot
cd my-bombahead-ts-bot
npm init -y
```

Install dependencies:

```bash
npm install bombahead-js
npm install -D typescript tsx @types/node
```

Create a TypeScript config:

```bash
npx tsc --init --rootDir src --outDir dist --module nodenext --target es2022
```

Update `package.json` scripts:

```json
{
  "scripts": {
    "dev": "tsx src/bot.ts",
    "build": "tsc",
    "start": "node dist/bot.js"
  }
}
```

Recommended structure:

```txt
my-bombahead-ts-bot/
  src/
    bot.ts
  package.json
  tsconfig.json
```

## 3. Write a minimal bot

Create `src/bot.ts`:

```ts
import { Action, run, type IBot, type GameState, type GameHelpers } from "bombahead-js";

const bot: IBot = {
  getNextMove(state: GameState, helpers: GameHelpers): Action {
    // If our tile is unsafe, move to the nearest safe tile.
    if (!helpers.isSafe(state.me.pos)) {
      const safe = helpers.getNearestSafePosition(state.me.pos);
      if (safe) {
        return helpers.getNextActionTowards(state.me.pos, safe);
      }
    }

    // Otherwise move toward the nearest box.
    const box = helpers.findNearestBox(state.me.pos);
    if (box) {
      return helpers.getNextActionTowards(state.me.pos, box);
    }

    return Action.DO_NOTHING;
  },
};

await run(bot);
```

## 4. Learn the basic SDK concepts

Your bot gets called every tick via:

```ts
getNextMove(state, helpers)
```

Most important parts to know first:

- `Action`: `MOVE_UP`, `MOVE_DOWN`, `MOVE_LEFT`, `MOVE_RIGHT`, `PLACE_BOMB`, `DO_NOTHING`
- `GameState`: `me`, `opponents`, `players`, `field`, `bombs`, `explosions`
- `GameHelpers`:
  - `isSafe(pos)`
  - `getNearestSafePosition(pos)`
  - `findNearestBox(pos)`
  - `getNextActionTowards(start, target)`

The SDK is defensive. If your bot throws or returns an invalid action, it safely falls back to `DO_NOTHING`.

## 5. Run and test locally

Build and run locally:

```bash
npm run dev
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
npm run dev
```

Your bot should connect to `ws://localhost:8038/ws` (handled by the SDK runtime).

## 6. Containerize your bot

Create a `Dockerfile`:

```dockerfile
FROM docker.io/node:20-alpine

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

COPY tsconfig.json ./
COPY src ./src

RUN npm run build

CMD ["npm", "start"]
```

Build locally:

```bash
docker build -t my-bombahead-ts-bot:local .
```

## 7. Push to GitHub Container Registry (`ghcr.io`)

Choose your final image name, for example:

`ghcr.io/<github-username>/bombahead-ts-bot:v1`

Login:

```bash
echo <YOUR_GITHUB_PAT> | docker login ghcr.io -u <github-username> --password-stdin
```

Tag and push:

```bash
docker tag my-bombahead-ts-bot:local ghcr.io/<github-username>/bombahead-ts-bot:v1
docker push ghcr.io/<github-username>/bombahead-ts-bot:v1
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
4. Paste your image URL, e.g. `ghcr.io/<github-username>/bombahead-ts-bot:v1`
5. Save

After that, BombAhead can schedule matches for your bot.

## 9. Recommended next improvements

- Add bomb placement logic near boxes/opponents.
- Add danger prediction before committing to a path.
- Version your images (`v1`, `v2`, `v3`) instead of overwriting `latest`.
