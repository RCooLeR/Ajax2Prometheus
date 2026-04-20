# Ajax2Prometheus

Receives Ajax security hub SIA DC-09 events, keeps the current alarm state in memory, and exposes Prometheus metrics on `/metrics`.

<p style="text-align: center">
<img src="./ajax2prometheus.png" alt="Ajax 2 Prometheus" width="70%">
</p>

It is a small bridge made of:

- SIA DC-09 TCP receiver
- Ajax/SIA event normalizer
- in-memory state engine
- Prometheus exporter
- optional raw SIA forwarding to other receivers (Home Assistant, CMS, etc.)
- JSON device catalog for Grafana labels

No database is required. Event history and state are kept in memory and reset on restart. The device catalog is stored in `data/devices.json`.

## Docker Compose

```powershell
services:
  ajax2prometheus:
    image: rcooler/ajax2prometheus:latest
    container_name: ajax2prometheus
    restart: unless-stopped
    ports:
      - "8080:8080"
      - "8099:8099"
    environment:
      AJAX2PROM_HTTP_ADDR: ":8080"
      AJAX2PROM_SIA_ADDR: ":8099"
      AJAX2PROM_ACCOUNT: "0001"
      AJAX2PROM_LOG_LEVEL: "info"
      AJAX2PROM_LOG_PRETTY: "false"
      AJAX2PROM_STRICT_CRC: "true"
      AJAX2PROM_PING_INTERVAL: "60s"
      AJAX2PROM_OFFLINE_GRACE: "180s"
      AJAX2PROM_READ_TIMEOUT: "120s"
      AJAX2PROM_FORWARD_TIMEOUT: "5s"
      AJAX2PROM_FORWARD_REQUIRE_ACK: "false"
      AJAX2PROM_DEVICES_PATH: "/data/devices.json"

      # Optional: AES key configured in Ajax SIA monitoring settings.
      # AJAX2PROM_ENCRYPTION_KEY: "REPLACE_WITH_32_HEX_AES_KEY"

      # Optional: forward raw SIA frames to Home Assistant's SIA integration.
      # Use an IP/hostname reachable from inside this container.
      # AJAX2PROM_FORWARD_ADDR: "home-assistant.example:12345,cms.example:7700"
    volumes:
      # The app creates/updates /data/devices.json as new Ajax zones are seen.
      - ./data:/data

```

## Ajax Setup

Configure the Ajax hub monitoring station connection:

- Protocol: `SIA DC-09`
- Transport: `TCP`
- Receiver address: host running this service
- Receiver port: `8099` by default
- Account/object number: same value as `AJAX2PROM_ACCOUNT`
- Encryption: optional AES key matching `AJAX2PROM_ENCRYPTION_KEY`
- Connection mode: connect on demand or constant connection

## Configuration

| Variable | Default | Description |
| --- | --- | --- |
| `AJAX2PROM_SIA_ADDR` | `:8099` | SIA TCP listen address |
| `AJAX2PROM_HTTP_ADDR` | `:8080` | HTTP listen address |
| `AJAX2PROM_ACCOUNT` | empty | Expected Ajax account/object number |
| `AJAX2PROM_ENCRYPTION_KEY` | empty | Optional SIA AES key |
| `AJAX2PROM_FORWARD_ADDR` | empty | Comma-separated upstream SIA receivers |
| `AJAX2PROM_FORWARD_TIMEOUT` | `5s` | Upstream forwarding timeout |
| `AJAX2PROM_FORWARD_REQUIRE_ACK` | `false` | Return NAK to Ajax if an upstream receiver does not ACK |
| `AJAX2PROM_DEVICES_PATH` | `data/devices.json` | Device catalog path |
| `AJAX2PROM_PING_INTERVAL` | `60s` | Expected Ajax monitoring station ping interval |
| `AJAX2PROM_OFFLINE_GRACE` | `180s` | Grace period before marking account offline |
| `AJAX2PROM_READ_TIMEOUT` | `120s` | SIA connection read timeout |
| `AJAX2PROM_STRICT_CRC` | `true` | Validate SIA DC-09 CRC/length |
| `AJAX2PROM_LOG_LEVEL` | `info` | `debug`, `info`, `warn`, or `error` |
| `AJAX2PROM_LOG_PRETTY` | `false` | Human-readable console logs |

## HTTP Endpoints

- `GET /metrics` Prometheus metrics
- `GET /healthz` health check
- `GET /readyz` readiness check
- `GET /state` current account and zone state
- `GET /events?limit=100` latest in-memory events
- `GET /devices` device catalog

## Metrics

Core metrics:

- `ajax_sia_events_total`
- `ajax_sia_parse_errors_total`
- `ajax_sia_forward_total`
- `ajax_sia_forward_duration_seconds`
- `ajax_account_online`
- `ajax_account_armed`
- `ajax_account_night_mode`
- `ajax_account_partially_armed`
- `ajax_account_alarm_active`
- `ajax_account_tamper_active`
- `ajax_account_trouble_active`
- `ajax_zone_alarm_active`
- `ajax_zone_alarm_last_event_timestamp_seconds`
- `ajax_zone_tamper_active`
- `ajax_zone_tamper_last_event_timestamp_seconds`
- `ajax_zone_trouble_active`
- `ajax_zone_last_event_timestamp_seconds`

All `ajax_zone_*` metrics include stable catalog labels where available: `account`, `partition`, `group`, `zone`, `device`, `device_name`, `room`, `device_kind`, and `device_events`. Alarm metrics also include `alarm_signal` and `alarm_action`.

## Device Catalog

The app creates `data/devices.json` automatically when it sees new zones. Edit this file to add human-readable names for Grafana.

Example:

```json
[
  {
    "account": "0001",
    "zone": "7",
    "name": "Hall smoke detector",
    "room": "Hall",
    "kind": "fireprotect",
    "events": ["smoke", "temperature", "tamper"]
  }
]
```

For Ajax SIA-DCS payloads like `Nri1/BA007`:

- `ri1` is the SIA area/partition
- `BA` is the event code
- `007` is the zone/device number to map in `devices.json`

In the Ajax app, the device number is shown in the device details at the very bottom of the screen.

Devices listed in `devices.json` are exported to `/state` and `/metrics` immediately after startup, even before the first event arrives for that device. Inactive device metrics start at `0` with a last-event timestamp of `0`.

The `events` field is a Grafana/Prometheus label, not an Ajax setting. The app also updates it automatically when real events arrive, but you can prefill it with the normalized signal labels below so dashboards do not need to wait for alarms or faults.

Common event labels supported by this app:

| Ajax device kind | Suggested `events` values |
| --- | --- |
| Hub | `supervision`, `connectivity`, `battery`, `power`, `tamper`, `interference`, `configuration`, `firmware` |
| User or mobile app | `arming`, `night_mode`, `duress`, `panic` |
| KeyPad, KeyPad Plus, KeyPad TouchScreen | `arming`, `night_mode`, `duress`, `access`, `tamper`, `battery`, `connectivity`, `bypass`, `tamper_bypass` |
| SpaceControl, control buttons, key fobs | `arming`, `night_mode`, `panic`, `duress`, `battery`, `connectivity`, `bypass` |
| Button, DoubleButton, panic/medical buttons | `panic`, `medical`, `emergency`, `tamper`, `battery`, `connectivity`, `bypass` |
| MotionProtect, MotionCam, DoorProtect, GlassProtect, CombiProtect, curtain/opening/motion detectors | `burglary`, `tamper`, `battery`, `connectivity`, `bypass`, `tamper_bypass`, `accelerometer` |
| MotionCam and detectors with photo verification | `burglary`, `tamper`, `battery`, `connectivity`, `bypass`, `tamper_bypass`, `accelerometer` |
| FireProtect, FireProtect Plus, FireProtect 2 | `fire`, `smoke`, `temperature`, `gas_or_co`, `co`, `fire_detector`, `tamper`, `battery`, `connectivity`, `bypass`, `tamper_bypass` |
| LeaksProtect | `water_leak`, `battery`, `connectivity`, `bypass`, `tamper_bypass` |
| Transmitter, MultiTransmitter, wired input modules | `burglary`, `fire`, `medical`, `panic`, `emergency`, `gas_or_co`, `water_leak`, `temperature`, `tamper`, `duress`, `accelerometer`, `hardware`, `arming`, `night_mode`, `battery`, `connectivity`, `power`, `bypass`, `tamper_bypass` |
| HomeSiren, StreetSiren, sirens | `tamper`, `battery`, `connectivity`, `power`, `bypass`, `tamper_bypass` |
| Relay, WallSwitch, Socket, automation modules | `power`, `connectivity`, `hardware`, `firmware` |
| ReX, ReX 2, range extenders | `connectivity`, `power`, `battery`, `tamper`, `firmware` |

Raw labels that may be auto-discovered from received SIA events are: `access`, `accelerometer`, `arming`, `battery`, `burglary`, `bypass`, `co`, `configuration`, `connectivity`, `duress`, `emergency`, `fire`, `fire_detector`, `firmware`, `gas`, `gas_or_co`, `hardware`, `interference`, `medical`, `night_mode`, `panic`, `power`, `smoke`, `supervision`, `tamper`, `tamper_bypass`, `temperature`, and `water_leak`.

Ajax documents direct SIA DC-09 event delivery and the SIA-DCS/ADM-CID payload shape in its support docs:

- <https://support.ajax.systems/en/how-to-use-sia-for-cms-connection/>
- <https://support.ajax.systems/en/manuals/cloud-signaling/>

## Logging

Default logs are JSON at `info` level. Accepted frames log a normalized event object without raw SIA frame data.

Use debug logs only while troubleshooting:

```powershell
--log-level debug --log-pretty
```

Debug logs include raw SIA frames, account numbers, internal addresses, timestamps, zones, and payloads. Do not publish debug logs.

## Forwarding

Set `AJAX2PROM_FORWARD_ADDR` to relay valid raw SIA frames to other receivers:

```text
AJAX2PROM_FORWARD_ADDR=home-assistant.example:12345,cms.example:7700
```

By default, forwarding errors are logged and exported as metrics but do not prevent ACK to Ajax. Set `AJAX2PROM_FORWARD_REQUIRE_ACK=true` if Ajax should receive NAK when any upstream receiver fails to ACK.
