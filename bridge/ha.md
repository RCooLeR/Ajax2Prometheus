# Home Assistant MQTT Setup

AjaxBridge can publish Home Assistant MQTT discovery configs and retained device state. This makes every Ajax device from `bridge/data/devices.json` appear in Home Assistant as a separate device with multiple entities.

## Prerequisites

- Home Assistant has the MQTT integration enabled.
- Home Assistant and AjaxBridge can reach the same MQTT broker.
- `bridge/data/devices.json` contains your Ajax devices with real `zone`, `name`, `room`, `kind`, and `events` values.

## Enable MQTT

Add MQTT settings to the `ajaxbridge` service in `bridge/docker-compose.yml`.

```yaml
services:
  ajaxbridge:
    environment:
      AJAXBRIDGE_MQTT_BROKER: "tcp://homeassistant.local:1883"
      AJAXBRIDGE_MQTT_USERNAME: "your_mqtt_user"
      AJAXBRIDGE_MQTT_PASSWORD: "your_mqtt_password"
      AJAXBRIDGE_MQTT_CLIENT_ID: "ajaxbridge"
      AJAXBRIDGE_MQTT_TOPIC_PREFIX: "ajaxbridge"
      AJAXBRIDGE_MQTT_DISCOVERY: "true"
      AJAXBRIDGE_MQTT_DISCOVERY_PREFIX: "homeassistant"
      AJAXBRIDGE_MQTT_RETAIN: "true"
```

If the broker has no username or password, only the broker URL is required:

```yaml
AJAXBRIDGE_MQTT_BROKER: "tcp://HOME_ASSISTANT_IP:1883"
```

Common broker URLs:

```text
tcp://homeassistant.local:1883
tcp://HOME_ASSISTANT_IP:1883
tcp://mosquitto:1883
```

Use the hostname or IP address that is reachable from the AjaxBridge container.

## Restart

Recreate the container after changing environment variables.

```powershell
docker compose -f bridge/docker-compose.yml up -d --build
```

Check logs for MQTT messages:

```powershell
docker compose -f bridge/docker-compose.yml logs -f ajaxbridge
```

Successful MQTT connection logs include:

```text
MQTT connected
```

If MQTT cannot connect, SIA handling still continues. Ajax ACK responses are not blocked by MQTT.

## Home Assistant Discovery

In Home Assistant:

1. Open `Settings -> Devices & services`.
2. Open the MQTT integration.
3. Make sure discovery is enabled.
4. Search for your Ajax device names from `bridge/data/devices.json`.

Each Ajax account appears as one Home Assistant device. Each Ajax zone/device from `bridge/data/devices.json` also appears as one Home Assistant device.

Example Ajax device:

```json
{
  "account": "A0F80D",
  "zone": "3",
  "name": "Hall fire detector",
  "room": "Hall",
  "kind": "FireProtect",
  "events": ["fire", "smoke", "temperature", "tamper", "battery", "connectivity"]
}
```

Home Assistant device:

```text
Device: Hall fire detector
Manufacturer: Ajax Systems
Model: FireProtect
Area: Hall
```

Entities:

```text
Alarm active
Tamper active
Trouble active
Fire
Smoke
Temperature
Tamper
Battery low
Connection lost
Last event
Last event code
Last signal
Last event time
Alarm signal
Alarm action
```

## Published Topics

Availability:

```text
ajaxbridge/status
```

Account state:

```text
ajaxbridge/accounts/A0F80D/state
```

Zone/device state:

```text
ajaxbridge/accounts/A0F80D/zones/3/state
```

Home Assistant discovery:

```text
homeassistant/binary_sensor/ajaxbridge/zone_a0f80d_3_signal_fire/config
homeassistant/sensor/ajaxbridge/zone_a0f80d_3_last_event_name/config
```

## State Payload

Each zone publishes one retained JSON payload. Home Assistant entities read values from this payload.

```json
{
  "account": "A0F80D",
  "partition": "1",
  "group": "1",
  "zone": "3",
  "device": "3",
  "device_name": "Hall fire detector",
  "room": "Hall",
  "kind": "FireProtect",
  "device_events": ["fire", "smoke", "temperature", "tamper", "battery", "connectivity"],
  "alarm_active": false,
  "alarm_signal": "none",
  "alarm_action": "none",
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
  "last_event_at": "2026-04-20T12:00:00Z",
  "last_event_unix": 1776686400
}
```

## Entities Created

For every Ajax zone/device:

- `binary_sensor`: alarm active
- `binary_sensor`: tamper active
- `binary_sensor`: trouble active
- `binary_sensor`: one entity for every event in the device `events` list
- `sensor`: last event
- `sensor`: last event code
- `sensor`: last signal
- `sensor`: last event time
- `sensor`: alarm signal
- `sensor`: alarm action

For every Ajax account:

- `binary_sensor`: online
- `binary_sensor`: armed
- `binary_sensor`: night mode
- `binary_sensor`: partially armed
- `binary_sensor`: alarm active
- `binary_sensor`: tamper active
- `binary_sensor`: trouble active
- `sensor`: mode
- `sensor`: last event
- `sensor`: last event code
- `sensor`: last signal
- `sensor`: last event time
- `sensor`: last ping time

## Verify MQTT Manually

With `mosquitto_sub`:

```powershell
mosquitto_sub -h HOME_ASSISTANT_IP -u your_mqtt_user -P your_mqtt_password -t "ajaxbridge/#" -v
```

Discovery configs:

```powershell
mosquitto_sub -h HOME_ASSISTANT_IP -u your_mqtt_user -P your_mqtt_password -t "homeassistant/#" -v
```

With MQTT Explorer, connect to the same broker and inspect:

```text
ajaxbridge/
homeassistant/binary_sensor/ajaxbridge/
homeassistant/sensor/ajaxbridge/
```

## Troubleshooting

No devices in Home Assistant:

- Check `AJAXBRIDGE_MQTT_BROKER` is reachable from the AjaxBridge container.
- Check MQTT username and password.
- Check MQTT discovery is enabled in Home Assistant.
- Check `AJAXBRIDGE_MQTT_DISCOVERY_PREFIX` matches Home Assistant, usually `homeassistant`.
- Subscribe to `homeassistant/#` and confirm discovery configs are published.

Devices exist but entities are unavailable:

- Subscribe to `ajaxbridge/status`; it should be `online`.
- Subscribe to `ajaxbridge/#`; retained state JSON should be visible.
- Check Home Assistant and AjaxBridge use the same MQTT broker.

Wrong device names or rooms:

- Edit `bridge/data/devices.json`.
- Restart AjaxBridge so discovery configs are republished.

Old entities remain after changing zones or names:

- Remove stale MQTT entities from Home Assistant.
- Clear retained stale discovery topics from the broker if needed.
- Keep `account` and `zone` stable when editing `devices.json`.

## Dashboard Cards

See [../ha-cards/EXAMPLES.md](../ha-cards/EXAMPLES.md) for ready-to-paste Lovelace examples:

- a polished `button-card` + `stack-in-card` zone card
- a built-in Home Assistant fallback without custom cards
- a compact account overview grid
