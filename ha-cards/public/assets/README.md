# Ajax Neon Icon System

Material/Phosphor-inspired semantic icons restyled for the dark neon Ajax Lovelace dashboard.

## Folders

- `assets/icons/<category>/*.svg` — individual editable SVGs
- `assets/icons-png/<category>/*.png` — 512px PNG exports
- `assets/sprites/ajax-icons.svg` — SVG symbol sprite
- `icon-manifest.json` — category/name/path mapping

## Categories

rooms, devices, system-states, security-states, sensors, events, automation, navigation, connectivity, misc.

## React usage

```tsx
<img src="/assets/icons/devices/motion_sensor.svg" />
```

## Sprite usage

```html
<svg class="icon"><use href="/assets/sprites/ajax-icons.svg#devices-motion-sensor" /></svg>
```
