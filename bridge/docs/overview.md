# Overview

AjaxBridge is a small Go service for exposing Ajax security hub data to monitoring and home automation tools. It receives Ajax SIA DC-09 TCP events, maintains current account and zone state in memory, and publishes that state through HTTP JSON, Prometheus, MQTT, and Home Assistant MQTT Discovery.

The bridge has two input sources:

| Source | Required | Role |
| --- | --- | --- |
| SIA DC-09 | Yes | Authoritative security state from the Ajax hub. |
| Jeedom MQTT | No | Additional metrics and allowlisted on/off controls from Jeedom Ajax equipment. |

SIA always has priority. If SIA and Jeedom report different security state for the same physical device, dashboards and integrations must use SIA. Jeedom is used for values SIA does not provide well, such as power, current, voltage, energy, battery, signal diagnostics, and safe outlet/relay/switch controls.

## Data Flow

1. Ajax sends SIA DC-09 frames to the bridge on TCP port `8099`.
2. AjaxBridge validates, decrypts if configured, parses, normalizes, and ACKs or NAKs the frame.
3. The in-memory state engine updates account and zone state.
4. HTTP JSON endpoints, Prometheus metrics, MQTT state, and Home Assistant discovery are updated.
5. If Jeedom input is enabled, Jeedom MQTT Manager events are parsed into a separate Jeedom mirror store.
6. When Jeedom devices are linked to SIA catalog entries, their Home Assistant entities are attached to the same HA device as the SIA zone.

## Runtime Ports

| Port | Protocol | Purpose |
| --- | --- | --- |
| `8099` | TCP | SIA DC-09 receiver for Ajax. |
| `8080` | HTTP | Health, JSON state, debug endpoints, and Prometheus metrics. |

## Install And Run

From `bridge/`:

```bash
docker compose up -d
```

For the current production setup with the collected `data/devices.json`, Jeedom enabled, and MQTT broker `tcp://192.168.100.100:1883`:

```bash
docker compose -f docker-compose.production.yml up -d
```

Check the service:

```bash
curl http://localhost:8080/healthz
curl http://localhost:8080/state
curl http://localhost:8080/metrics
```

Build locally:

```bash
go test ./...
go build ./cmd/ajaxbridge
./ajaxbridge --sia-addr :8099 --http-addr :8080 --account 0001
```

## Core Configuration

| Variable | Default | Description |
| --- | --- | --- |
| `AJAXBRIDGE_SIA_ADDR` | `:8099` | SIA TCP listen address. |
| `AJAXBRIDGE_HTTP_ADDR` | `:8080` | HTTP listen address. |
| `AJAXBRIDGE_ACCOUNT` | empty | Expected Ajax account/object number. Empty accepts all accounts. |
| `AJAXBRIDGE_ENCRYPTION_KEY` | empty | Optional SIA AES key. |
| `AJAXBRIDGE_STRICT_CRC` | `true` | Reject frames with invalid CRC. |
| `AJAXBRIDGE_DEVICES_PATH` | `data/devices.json` | Device catalog path. |
| `AJAXBRIDGE_MQTT_BROKER` | empty | MQTT broker URL. Enables MQTT when set. |
| `AJAXBRIDGE_MQTT_TOPIC_PREFIX` | `ajaxbridge` | MQTT state prefix. |
| `AJAXBRIDGE_MQTT_DISCOVERY` | `true` | Publish Home Assistant discovery when MQTT is enabled. |
| `AJAXBRIDGE_JEEDOM_ENABLED` | `false` | Enable Jeedom MQTT input. |
| `AJAXBRIDGE_FORWARD_ADDR` | empty | Optional comma-separated raw SIA forward targets. |
| `AJAXBRIDGE_NOTIFICATIONS_PATH` | `data/notifications.json` | Notification channel/rule config path. |

Legacy `AJAX2PROM_*` variables are still accepted as compatibility aliases.

## HTTP Endpoints

| Endpoint | Purpose |
| --- | --- |
| `GET /healthz` | Process health. |
| `GET /readyz` | Store readiness. |
| `GET /state` | Current SIA-derived account and zone state. |
| `GET /events?limit=100` | Latest in-memory SIA events. |
| `GET /devices` | Device catalog. |
| `GET /metrics` | Prometheus metrics. |
| `GET /admin` | Bootstrap admin panel. |
| `GET /api/admin/bootstrap` | Admin data for catalog, Jeedom, notifications, and current state. |
| `PUT /api/admin/devices` | Replace and persist the device catalog. |
| `PUT /api/admin/notifications` | Replace and persist notification rules/channels. |
| `GET /jeedom/devices` | Jeedom mirror devices, when Jeedom is enabled. |
| `GET /jeedom/devices/{slug}` | One Jeedom mirror device. |
| `GET /jeedom/commands` | Jeedom command metadata. |
| `GET /jeedom/actions` | Jeedom action metadata. |
| `GET /jeedom/control-audit?limit=100` | Recent Jeedom control attempts. |
| `POST /jeedom/devices/{slug}/control` | Jeedom on/off control, when enabled. |

## Data Files

| Path | Purpose |
| --- | --- |
| `data/devices.json` | Optional catalog with stable names, rooms, kinds, SIA zones, Jeedom aliases, and Jeedom command ids. |
| `data/notifications.json` | Notification channels and rules. |
| `tmp-jeedom/*.json` | Optional raw Jeedom MQTT sample envelopes when `AJAXBRIDGE_JEEDOM_SAMPLE_DIR` is set. Disabled by default for production. |

## Recommended Setup Order

1. Run AjaxBridge with SIA only and confirm `/state` receives Ajax events.
2. Edit `data/devices.json` with stable device names, rooms, kinds, and expected signals.
3. Enable MQTT and Home Assistant discovery.
4. Enable Jeedom only after SIA devices are stable.
5. Add `jeedom_names`, `jeedom_command_ids`, and `AJAXBRIDGE_JEEDOM_ACCOUNT_NAMES` to prevent duplicate HA devices.
6. Enable Jeedom controls only after `/jeedom/actions` shows the expected allowlisted actions.
7. Open `/admin` to edit catalog links and configure notifications.

## External References

- Ajax direct SIA DC-09 setup: https://support.ajax.systems/en/how-to-use-sia-for-cms-connection/
- Jeedom installation: https://doc.jeedom.com/en_US/installation/index.html
- Jeedom MQTT Manager: https://doc.jeedom.com/en_US/plugins/programming/mqtt2/
- Jeedom Ajax System plugin: https://doc.jeedom.com/en_US/plugins/security/ajaxSystem/
- Home Assistant MQTT integration: https://www.home-assistant.io/integrations/mqtt
