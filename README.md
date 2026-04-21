# Ajax2Prometheus

Receives Ajax security hub SIA DC-09 events, keeps the current alarm state in memory, and exposes Prometheus metrics on `/metrics`.

*If i ever get access to API will update to also have all sensors info, automation via Home Assistant etc.

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

      # Optional: publish Home Assistant MQTT discovery and retained state.
      # AJAX2PROM_MQTT_BROKER: "tcp://homeassistant.local:1883"
      # AJAX2PROM_MQTT_USERNAME: "mqtt-user"
      # AJAX2PROM_MQTT_PASSWORD: "mqtt-password"
      # AJAX2PROM_MQTT_TOPIC_PREFIX: "ajax2prometheus"
      # AJAX2PROM_MQTT_DISCOVERY: "true"

      # Optional: forward raw SIA frames to Home Assistant's SIA integration.
      # Use an IP/hostname reachable from inside this container.
      # AJAX2PROM_FORWARD_ADDR: "home-assistant.example:12345,cms.example:7700"
    volumes:
      # The app creates/updates /data/devices.json as new Ajax zones are seen.
      - ./data:/data

```

Published images:

- `rcooler/ajax2prometheus:latest`
- `rcooler/ajax2prometheus:vX.Y.Z`
- `rcooler/ajax2prometheus:vX.Y`
- `rcooler/ajax2prometheus:vX`

Git tags matching `v*` trigger `.github/workflows/release.yml`, which runs GoReleaser and pushes a multi-arch Docker image for `linux/amd64` and `linux/arm64` to Docker Hub.

Required GitHub Actions secrets:

- `DOCKERHUB_USERNAME`
- `DOCKERHUB_TOKEN`

Local dry-run:

```powershell
go run github.com/goreleaser/goreleaser/v2@latest release --snapshot --clean
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
| `AJAX2PROM_MQTT_BROKER` | empty | Optional MQTT broker URL, for example `tcp://homeassistant.local:1883` |
| `AJAX2PROM_MQTT_USERNAME` | empty | MQTT username |
| `AJAX2PROM_MQTT_PASSWORD` | empty | MQTT password |
| `AJAX2PROM_MQTT_CLIENT_ID` | `ajax2prometheus` | MQTT client ID |
| `AJAX2PROM_MQTT_TOPIC_PREFIX` | `ajax2prometheus` | MQTT state topic prefix |
| `AJAX2PROM_MQTT_DISCOVERY` | `true` | Publish Home Assistant MQTT discovery configs when MQTT is enabled |
| `AJAX2PROM_MQTT_DISCOVERY_PREFIX` | `homeassistant` | Home Assistant MQTT discovery prefix |
| `AJAX2PROM_MQTT_TIMEOUT` | `5s` | MQTT connect and publish timeout |
| `AJAX2PROM_MQTT_RETAIN` | `true` | Retain MQTT state messages |
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

## Home Assistant MQTT

Set `AJAX2PROM_MQTT_BROKER` to enable MQTT publishing. Home Assistant discovery is enabled by default, so each Ajax account and each device from `devices.json` appears as a separate Home Assistant device.

See [ha.md](./ha.md) for step-by-step Home Assistant setup and troubleshooting.

See [ha-cards.md](./ha-cards.md) for Lovelace card examples built around the discovered Ajax account and zone entities.

MQTT is intentionally non-blocking for SIA handling. If the MQTT broker is down or slow, Ajax ACK responses are still sent normally.

Published state topics:

```text
ajax2prometheus/status
ajax2prometheus/accounts/A0F80D/state
ajax2prometheus/accounts/A0F80D/zones/3/state
```

Home Assistant discovery topics use the configured discovery prefix:

```text
homeassistant/binary_sensor/ajax2prometheus/zone_a0f80d_3_signal_fire/config
homeassistant/sensor/ajax2prometheus/zone_a0f80d_3_last_event_name/config
```

Each zone/device publishes one retained JSON state payload with current status and signal states:

```json
{
  "account": "A0F80D",
  "zone": "3",
  "device_name": "Hall fire detector",
  "room": "Hall",
  "kind": "FireProtect",
  "device_events": ["fire", "smoke", "temperature", "tamper", "battery", "connectivity"],
  "alarm_active": false,
  "tamper_active": false,
  "trouble_active": false,
  "signal_active": {
    "fire": false,
    "smoke": false,
    "temperature": false,
    "tamper": false,
    "battery": false,
    "connectivity": false
  },
  "last_event_code": "BR",
  "last_event_name": "Burglary alarm restored",
  "last_signal": "burglary",
  "last_event_at": "2026-04-20T12:00:00Z"
}
```

For every zone, discovery creates:

- `binary_sensor` entities for `alarm_active`, `tamper_active`, and `trouble_active`
- one `binary_sensor` for every value in the device `events` list, such as `fire`, `smoke`, `battery`, `connectivity`, `power`, `bypass`, or `water_leak`
- `sensor` entities for last event name, code, signal, event time, alarm signal, and alarm action

For every account, discovery creates:

- `binary_sensor` entities for online, armed, night mode, partially armed, alarm, tamper, and trouble
- `sensor` entities for mode, last event, last signal, last event time, and last ping time

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
