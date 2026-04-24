# Ajax HA Cards

Purpose-built Home Assistant Lovelace cards for the Ajax2Prometheus device inventory in `data/devices.json`.

## What is included

- One full-tab custom card: `ajax-security-dashboard`
- One compact main-dashboard card: `ajax-security-overview`
- Automatic scaling from a 1920x1080 design stage
- Live state discovery from Ajax2Prometheus MQTT entities
- Layout tuned to the specific room and device set in this repository

## Build

```powershell
cd ajax-ha-cards
npm run build
npm run check
```

The build step embeds `../data/devices.json` when it exists, and falls back to `../devices.example.json` otherwise.

## Install in Home Assistant

1. Copy `ajax-ha-cards/dist/ajax-security-dashboard.js` into your Home Assistant `www/ajax/` folder.
2. Add it as a Lovelace resource:

```yaml
url: /local/ajax/ajax-security-dashboard.js
type: module
```

3. Use a dashboard view in panel mode for the intended full-screen layout.
4. Add the example card from `ajax-ha-cards/examples/ajax-security-dashboard.yaml`.
5. For a normal dashboard tile/section, use `ajax-ha-cards/examples/ajax-security-overview.yaml`.

## Recommended view config

```yaml
title: Ajax
path: ajax
panel: true
cards:
  - type: custom:ajax-security-dashboard
    title: Ajax Security Command Deck
    subtitle: Main site security overview
    account: "A0F80D"
```

## Compact overview example

```yaml
type: custom:ajax-security-overview
title: Ajax overview
account: "A0F80D"
navigation_path: /ajax
```

## Config

- `title`: Optional header title.
- `subtitle`: Optional header subtitle.
- `account`: Ajax account ID. Defaults to the first account in the embedded catalog.
- `navigation_path`: Optional Lovelace path to open when the overview card is clicked.
- `default_room`: Room shown in the main detail pane on first load.
- `room_order`: Optional explicit room order for the ribbon.
- `entity_prefix_overrides`: Optional mapping of device `zone` or exact `name` to a Home Assistant entity prefix if one device needs a manual override.

## Notes

- The card now prefers your actual Home Assistant name-based entities such as `binary_sensor.pozhezhnii_datchik_na_gorishchi_carbon_monoxide` and `sensor.detektor_elektrozhivlennia_merezhi_last_signal`.
- It still falls back to the repository MQTT discovery naming like `zone_<account>_<zone>_*` for any entities that use the newer discovery object IDs.
- The embedded catalog is intentionally specific to your current device set, not a generic public package.
