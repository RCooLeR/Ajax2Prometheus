# Admin Panel

AjaxBridge includes a small Bootstrap admin panel for local operation.

Open:

```text
http://localhost:8080/admin
```

The panel uses the repository `logo.png` and Bootstrap from a CDN.

## Device Catalog Editing

The Devices tab edits `data/devices.json`.

Editable fields:

- SIA account
- SIA zone
- Ajax device id
- name
- room
- kind/model
- expected event signals
- Jeedom names
- Jeedom command ids

After saving, AjaxBridge reloads the catalog into the in-memory SIA state and republishes MQTT discovery/state for the current snapshot.

## SIA And Jeedom Matching

The SIA / Jeedom Matching tab shows:

- SIA catalog devices
- Jeedom mirror devices
- Jeedom command ids
- current linked SIA account/zone for Jeedom devices

Use this page to copy Jeedom command ids into the matching SIA catalog row. This prevents duplicate Home Assistant devices.

Matching fields in `data/devices.json`:

```json
{
  "account": "A0F80D",
  "zone": "8",
  "name": "Server power",
  "jeedom_names": ["Server power"],
  "jeedom_command_ids": ["52", "53", "54", "55", "56", "57", "58", "59"]
}
```

## Notification Editing

The Notifications tab edits `data/notifications.json`.

It can manage:

- channels
- threshold rules
- on/off state-change rules
- bridge-issued control rules
- arm mode filters
- cooldowns
- delivery history

See [Notifications](./notifications.md) for the full rule schema.

## API Endpoints

| Endpoint | Purpose |
| --- | --- |
| `GET /admin` | Admin UI. |
| `GET /admin/logo.png` | Logo asset used by the admin UI. |
| `GET /api/admin/bootstrap` | Full admin bootstrap payload. |
| `PUT /api/admin/devices` | Replace and persist the device catalog. |
| `PUT /api/admin/notifications` | Replace and persist notification config. |
| `GET /api/admin/notifications/history?limit=100` | Recent notification delivery attempts. |

## Security Note

The admin panel is intended for a trusted LAN or a reverse proxy that adds authentication. AjaxBridge does not currently implement built-in users, sessions, or CSRF protection.
