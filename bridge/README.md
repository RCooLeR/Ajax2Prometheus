# AjaxBridge

AjaxBridge receives Ajax security hub SIA DC-09 events, keeps the current alarm state in memory, and exposes the data through HTTP JSON, Prometheus, MQTT, and Home Assistant MQTT Discovery.

<p style="text-align: center">
<img src="./ajax-bridge.png" alt="AjaxBridge" width="70%">
</p>

The bridge is built around two inputs:

- SIA DC-09 from Ajax as the authoritative security source.
- Optional Jeedom MQTT for device metrics and allowlisted on/off controls.

No database is required. Runtime state is kept in memory, and the optional device catalog lives in `data/devices.json`.

## Documentation

- [Documentation index](./docs/index.md)
- [Overview](./docs/overview.md)
- [SIA integration](./docs/sia.md)
- [Jeedom integration](./docs/jeedom.md)
- [Prometheus metrics](./docs/prometheus.md)
- [Home Assistant integration and technical guide](./docs/home-assistant.md)
- [Home Assistant card package](../ha-cards/README.md)
