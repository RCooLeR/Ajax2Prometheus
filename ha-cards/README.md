# AjaxBridge Lovelace UI

`ha-cards` builds the Home Assistant Lovelace cards for AjaxBridge.

It ships two custom cards:

- `custom:ajaxbridge-detailed-card`
- `custom:ajaxbridge-chips-card`

When loaded inside Home Assistant, the cards use live Home Assistant data from the area, device, and entity registries plus current entity state. The standalone Vite preview still works for local UI development.

## Disclaimer

AjaxBridge is an unofficial DIY open-source project for compatibility and integration. It is not affiliated with, endorsed by, or sponsored by Ajax Systems.

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

- `dist/ajaxbridge-lovelace.js`: the Home Assistant module that registers both cards
- `dist/assets/*`: JS chunks, CSS, icons, and room assets used by the module
- `dist/index.html`: standalone browser preview

Copy the full `dist/` contents into Home Assistant, not only `ajaxbridge-lovelace.js`.

## Home Assistant install

1. Build the UI:

```bash
npm run build
```

2. Copy `dist/` into a folder under Home Assistant `www`, for example:

```text
<ha-config>/www/ajaxbridge-lovelace/
```

3. Register the resource:

```yaml
resources:
  - url: /local/ajaxbridge-lovelace/ajaxbridge-lovelace.js
    type: module
```

4. Add one of the cards.

Detailed card:

```yaml
title: AjaxBridge
path: ajaxbridge
panel: true
cards:
  - type: custom:ajaxbridge-detailed-card
    dahua_base: https://ha.example.test/dahua-bridge
```

Use `dahua_base` when browser-side DahuaBridge calls need to go through a Home Assistant proxy path instead of the bridge URL published on camera attributes.

Compact chips card:

```yaml
type: custom:ajaxbridge-chips-card
max_chips: 7
```

Ready-to-paste examples live in [`examples/`](./examples/):

- [`ajaxbridge-detailed-card.yaml`](./examples/ajaxbridge-detailed-card.yaml)
- [`ajaxbridge-chips-card.yaml`](./examples/ajaxbridge-chips-card.yaml)

Reference notes for DahuaBridge camera discovery, live playback, and SMD/IVS room counters live in [`docs/`](./docs/).

## Card behavior

- Rooms come from Home Assistant areas.
- Room hero backgrounds prefer area pictures and fall back to linked image or camera entities.
- AjaxBridge devices come from the Home Assistant device/entity registries plus MQTT entities published by AjaxBridge.
- Dahua and Roller devices are grouped by Home Assistant device and rendered inside their assigned room.
- Dahua camera event rows are limited to SMD/IVS detection state and camera online/offline state. Unavailable SMD/IVS sensors, stream, codec, ONVIF/H.264, profile, capability, and other diagnostic entities are ignored for camera events and alerts.
- Dahua SMD/IVS 24 hour counters are shown only for rooms that contain a Dahua camera device. They are loaded from the DahuaBridge NVR `/events/summary` endpoint and can use `dahua_base` for browser-reachable proxy URLs.
- Event rows are synthesized from current or latest Home Assistant entity state. The UI does not query AjaxBridge `/events` history directly.

## Project structure

- `src/ha/`: Home Assistant card registration and types
- `src/data/liveDashboardData.ts`: live HA registry and state adapter
- `src/cards/` and `src/components/`: presentational UI building blocks
- `src/models/dashboard.ts`: shared typed dashboard model
- `src/styles/`: theme and layout CSS
- `public/assets/`: icons and room assets bundled by Vite

## Notes

- The module resolves icons and room assets relative to the module URL, so `/local/ajaxbridge-lovelace/` and similar install paths both work.
- `dist/index.html` is useful for local visual review, but Home Assistant custom-card usage is the main target.

## License

MIT License. See [../LICENSE](../LICENSE). See [../NOTICE](../NOTICE) for trademark and affiliation notice.
