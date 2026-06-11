# AjaxBridge

AjaxBridge receives Ajax security hub SIA DC-09 events, keeps the current alarm state in memory, and exposes the data through HTTP JSON, Prometheus, MQTT, and Home Assistant MQTT Discovery.

<p style="text-align: center">
<img src="./ajax-bridge.png" alt="AjaxBridge" width="70%">
</p>

The bridge is built around two inputs:

- SIA DC-09 from Ajax as the authoritative security source.
- Optional Jeedom MQTT for device metrics, allowlisted toggles, and relay impulse controls.

No database is required. Runtime SIA state is kept in memory, the optional device catalog lives in `data/devices.json`, and the optional Jeedom mirror cache lives in `data/jeedom.json`.

For your current production setup, use [docker-compose.production.yml](./docker-compose.production.yml). It uses account `A0F80D`, MQTT broker `tcp://192.168.100.100:1883`, the local `./data` catalog, Jeedom MQTT input, and allowlisted Jeedom controls.

CI runs on every pushed branch and pull request. It verifies the Go bridge with `go vet` and `go test`, and verifies the Home Assistant cards with TypeScript check plus production build.

## Disclaimer

AjaxBridge is an unofficial DIY open-source project for compatibility and integration. It is not affiliated with, endorsed by, or sponsored by Ajax Systems.

## Documentation

- [Documentation index](./docs/index.md)
- [Overview](./docs/overview.md)
- [SIA integration](./docs/sia.md)
- [Jeedom integration](./docs/jeedom.md)
- [Prometheus metrics](./docs/prometheus.md)
- [Notifications](./docs/notifications.md)
- [Admin panel](./docs/admin.md)
- [Home Assistant integration and technical guide](./docs/home-assistant.md)
- [Home Assistant card package](../ha-cards/README.md)
- [Disclaimer and trademark notice](../NOTICE)

## License

MIT License. See [../LICENSE](../LICENSE). See [../NOTICE](../NOTICE) for trademark and affiliation notice.
