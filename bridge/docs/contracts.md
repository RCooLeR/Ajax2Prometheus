# Runtime Contracts

This project keeps SIA as the authoritative security source and uses Jeedom as an optional source for extra metrics and safe controls. The surfaces below are treated as compatibility contracts: change them only with tests and migration notes.

## HTTP JSON

- `GET /state` returns the current SIA-derived snapshot with `accounts` and `zones`.
- `GET /devices` returns the configured device catalog.
- `GET /jeedom/devices`, `GET /jeedom/commands`, and `GET /jeedom/actions` expose the Jeedom mirror when Jeedom is enabled.
- `GET /api/admin/bootstrap` combines current state, catalog, Jeedom mirror data, notifications, and file paths for the admin UI.

JSON field names are snake_case and should remain stable because the admin UI, Home Assistant cards, scripts, and troubleshooting workflows depend on them.

## MQTT State

Default state topics use the `ajaxbridge` prefix:

- `ajaxbridge/accounts/{account}/state`
- `ajaxbridge/accounts/{account}/zones/{zone}/state`
- `ajaxbridge/jeedom/devices/{device_slug}/state`
- `ajaxbridge/status`

State messages are retained when MQTT retain is enabled. On MQTT reconnect, AjaxBridge resets its publish caches and republishes retained SIA state plus cached Jeedom discovery/state so Home Assistant can recover without resubmitting Jeedom discovery.

## Home Assistant Discovery

Default discovery topics use:

- `homeassistant/{component}/ajaxbridge/{object_id}/config`
- `homeassistant/{component}/ajaxbridge/jeedom_cmd_{command_id}/config`
- `homeassistant/{component}/ajaxbridge/jeedom_control_{device_slug}/config`

SIA entity unique IDs are based on the discovery node plus the SIA object id. Jeedom measurement unique IDs are based on the Jeedom command id. These IDs must not be renamed casually because Home Assistant uses them for entity identity.

When a Jeedom device is linked to SIA, Jeedom metrics may attach to the SIA Home Assistant device. Jeedom security/status entities that duplicate SIA-owned state are published as empty retained discovery payloads to remove stale duplicates.

## Frontend Card Inputs

The Home Assistant cards consume Home Assistant device/entity registries and state objects. Ajax devices are recognized primarily by these stable identifiers:

- SIA account devices: `ajaxbridge_account_{account}`
- SIA zone devices: `ajaxbridge_{account}_zone_{zone}`
- Jeedom devices: `ajaxbridge_jeedom_{device_slug}` unless linked to a SIA identifier

Card-facing models live in `ha-cards/src/models/dashboard.ts`; keep those TypeScript shapes in sync with any bridge output or Home Assistant discovery changes.
