# Changelog

## 2.0.1 - 2026-08-05

- Split Home Assistant JSON attributes from dynamic Ajax/Jeedom state topics. Entities now keep compact stable device metadata while full MQTT state, HTTP APIs, entity identities, controls, and dedicated timestamp sensors remain compatible.
- Clear a zone's latched SIA alarm when the device is bypassed/deactivated or turned off, while preserving last-event and measurement telemetry.

## 2.0.0 - 2026-05-06

### Added

- Home Assistant Lovelace dashboard card for AjaxBridge rooms, devices, cameras, safety, grid power, and controls.
- Home Assistant MQTT discovery for Ajax devices, sensors, actions, valves, switches, hub modes, and Jeedom-derived entities.
- Jeedom integration support, including device catalog resolution, value seeding, event parsing, and retained MQTT topic cleanup.
- Dahua camera support in the card, including room camera selection, direct IPC model detection, native HA camera stream handling, and SMD/IVS summaries.
- Device image catalog based on Jeedom and Dahua device assets.
- Push notification support and admin UI improvements.

### Changed

- Renamed and documented the project as AjaxBridge while keeping compatibility-oriented release images.
- Improved Ajax event/state detection for wall switches, WaterStop valves, transmitters, grid power, hub arming actions, fire mute actions, and duplicate sensor cleanup.
- Reworked the card device layout to use real device images, Material Design icons, meaningful device controls, and room-level summaries.
- Wall switch power shown in the card is now calculated from voltage and current instead of using total power.

### Fixed

- Cleaned legacy Jeedom and Ajax2Prometheus retained MQTT discovery topics for unlinked or renamed devices.
- Hid Ajax app devices from the card device list.
- Hid room smoke/CO status when the room has no FireProtect, LifeQuality, or equivalent measuring device.
- Improved nested Home Assistant camera player mute and sizing behavior after the native video element is created.

## 1.0.1 - 2026-04-25

- Migrated Ajax2Prometheus release naming and metadata to AjaxBridge.

## 1.0.0 - 2026-04-25

- Initial Docker release.
