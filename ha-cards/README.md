# Ajax Lovelace UI

`ha-cards` builds the Home Assistant Lovelace cards for AjaxBridge.

It ships two custom cards:

- `custom:ajax-lovelace-detailed-card`
- `custom:ajax-lovelace-chips-card`

When loaded inside Home Assistant, the cards use live Home Assistant data from the area, device, and entity registries plus current entity state. The standalone Vite preview still works for local UI development.

## Development

```bash
npm install
npm run dev
```

Useful commands:

- `npm run check`: TypeScript build check
- `npm run build`: production build into `dist/`
- `npm run preview`: serve the built `dist/` output locally

## Build output

`npm run build` writes:

- `dist/ajax-lovelace.js`: the Home Assistant module that registers both cards
- `dist/assets/*`: JS chunks, CSS, icons, and room assets used by the module
- `dist/index.html`: standalone browser preview

Copy the full `dist/` contents into Home Assistant, not only `ajax-lovelace.js`.

## Home Assistant install

1. Build the UI:

```bash
npm run build
```

2. Copy `dist/` into a folder under Home Assistant `www`, for example:

```text
<ha-config>/www/ajax-lovelace/
```

3. Register the resource:

```yaml
resources:
  - url: /local/ajax-lovelace/ajax-lovelace.js
    type: module
```

4. Add one of the cards.

Detailed card:

```yaml
title: Ajax
path: ajax
panel: true
cards:
  - type: custom:ajax-lovelace-detailed-card
```

Compact chips card:

```yaml
type: custom:ajax-lovelace-chips-card
max_chips: 7
```

Ready-to-paste examples live in [`examples/`](./examples/):

- [`ajax-lovelace-detailed-card.yaml`](./examples/ajax-lovelace-detailed-card.yaml)
- [`ajax-lovelace-chips-card.yaml`](./examples/ajax-lovelace-chips-card.yaml)

## Card behavior

- Rooms come from Home Assistant areas.
- Room hero backgrounds prefer area pictures and fall back to linked image or camera entities.
- Ajax devices come from the Home Assistant device/entity registries plus MQTT entities published by AjaxBridge.
- Dahua and Roller devices are grouped by Home Assistant device and rendered inside their assigned room.
- Event rows are synthesized from current or latest Home Assistant entity state. The UI does not query AjaxBridge `/events` history directly yet.

## Project structure

- `src/ha/`: Home Assistant card registration and types
- `src/data/liveDashboardData.ts`: live HA registry and state adapter
- `src/cards/` and `src/components/`: presentational UI building blocks
- `src/models/dashboard.ts`: shared typed dashboard model
- `src/styles/`: theme and layout CSS
- `public/assets/`: icons and room assets bundled by Vite

## Notes

- The module resolves icons and room assets relative to the module URL, so `/local/ajax-lovelace/` and similar install paths both work.
- `dist/index.html` is useful for local visual review, but Home Assistant custom-card usage is the main target.
