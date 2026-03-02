# General Tips

This page collects practical tips that help when building BombAhead bots.

## Coordinate system

When you receive a game state, the field starts at the **top-left** corner.

- Top-left is `(0,0)`
- `x` increases to the right
- `y` increases downward

## Start with survival first

A stable baseline strategy is:

1. If current position is unsafe, move to the nearest safe tile.
2. Otherwise, move toward a useful target (for example the nearest box or player).
3. If no safe/useful action exists, use `DO_NOTHING`.

## Avoid risky bomb placements

Do not place bombs unless you have a clear escape route. Many early bots lose because they trap themselves.

## Validate with local matches

Before pushing new versions, run local matches against baseline opponents. This catches regressions early.

Common baseline images:

- `ghcr.io/n3moahead/bomber:idle` Is just ideling
- `ghcr.io/n3moahead/bomber:self-destruct` Commits suicide
- `ghcr.io/n3moahead/bomber:simple-max-v1` A simple but worthy opponent

## Version your images

Use explicit tags like `:v1`, `:v2`, `:v3` instead of overwriting `:latest`. This makes rollbacks and result comparisons much easier.

## Keep images pullable

The match runner must be able to pull your image.

- Make sure repository and tag exist.
- For `ghcr.io`, set package visibility to public if required.
- Use fully qualified image references.
