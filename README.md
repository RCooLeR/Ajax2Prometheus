# AjaxBridge
# AjaxBridge

Receives Ajax security hub SIA DC-09 events, keeps the current alarm state in memory, and exposes Prometheus metrics on `/metrics`.

*If i ever get access to API will update to also have all sensors info, automation via Home Assistant etc.

<p style="text-align: center">
<img src="./ajax-bridge.png" alt="AjaxBridge" width="70%">
</p>

This repository is split into two main areas:

- [`bridge/`](./bridge/README.md): the Go service, Docker files, compose file, runtime data examples, and bridge-specific docs
- [`ha-cards/`](./ha-cards/README.md): Home Assistant card assets, build tooling, and Lovelace examples

The root stays minimal on purpose. It keeps only repo-level files such as `go.work`, CI/release config, ignores, and generated release output.
