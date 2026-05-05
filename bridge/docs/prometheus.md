# Prometheus

AjaxBridge exposes Prometheus metrics on the HTTP server.

Default endpoint:

```text
http://localhost:8080/metrics
```

The endpoint is enabled whenever the HTTP server is enabled. Configure the HTTP listener with `AJAXBRIDGE_HTTP_ADDR`, default `:8080`.

## Scrape Config

Example Prometheus target:

```yaml
scrape_configs:
  - job_name: ajaxbridge
    static_configs:
      - targets:
          - ajaxbridge:8080
```

Use the hostname that Prometheus can reach. If Prometheus runs outside Docker, that may be `HOST_IP:8080`.

## SIA Ingestion Metrics

| Metric | Type | Labels | Meaning |
| --- | --- | --- | --- |
| `ajax_sia_events_total` | counter | `account`, `event_code`, `event_class`, `parse_status` | Total SIA events seen by the parser. |
| `ajax_sia_parse_errors_total` | counter | `reason` | Total rejected or degraded SIA frames by parse status. |

`parse_status` and `reason` values include:

- `ok`
- `crc_invalid`
- `length_invalid`
- `format_invalid`
- `account_invalid`
- `decrypt_invalid`

## SIA Forwarding Metrics

| Metric | Type | Labels | Meaning |
| --- | --- | --- | --- |
| `ajax_sia_forward_total` | counter | `target`, `status` | Forward attempts to an upstream SIA receiver. |
| `ajax_sia_forward_duration_seconds` | histogram | `target`, `status` | Duration of SIA forwarding attempts. |

Forwarding metrics are emitted only when `AJAXBRIDGE_FORWARD_ADDR` is set.

## Account Metrics

All account gauges have label `account`.

| Metric | Value |
| --- | --- |
| `ajax_account_online` | `1` when the account is online, otherwise `0`. |
| `ajax_account_armed` | `1` when armed, otherwise `0`. |
| `ajax_account_night_mode` | `1` when night mode is active. |
| `ajax_account_partially_armed` | `1` when partially armed. |
| `ajax_account_alarm_active` | `1` when any account alarm is active. |
| `ajax_account_tamper_active` | `1` when any account tamper is active. |
| `ajax_account_trouble_active` | `1` when any account trouble is active. |
| `ajax_account_last_event_timestamp_seconds` | Unix timestamp of the last event. |
| `ajax_account_last_ping_timestamp_seconds` | Unix timestamp of the last ping/test event. |

## Zone Metrics

Zone gauges use these labels:

| Label | Meaning |
| --- | --- |
| `account` | SIA account/object number. |
| `partition` | SIA partition/area or `unknown`. |
| `group` | SIA group or `unknown`. |
| `zone` | SIA zone. |
| `device` | Ajax device id, or zone when missing. |
| `device_name` | Catalog device name or fallback. |
| `room` | Catalog room or `unknown`. |
| `device_kind` | Catalog kind or inferred kind. |
| `device_events` | Comma-separated catalog signal list. |
| `alarm_signal` | Alarm signal label for alarm metrics. |
| `alarm_action` | Alarm action label for alarm metrics. |

| Metric | Extra labels | Value |
| --- | --- | --- |
| `ajax_zone_alarm_active` | `alarm_signal`, `alarm_action` | `1` when the zone alarm is active. |
| `ajax_zone_alarm_last_event_timestamp_seconds` | `alarm_signal`, `alarm_action` | Unix timestamp of the active/last zone alarm event. |
| `ajax_zone_tamper_active` | none | `1` when zone tamper is active. |
| `ajax_zone_tamper_last_event_timestamp_seconds` | none | Unix timestamp of the last tamper event. |
| `ajax_zone_trouble_active` | none | `1` when zone trouble is active. |
| `ajax_zone_last_event_timestamp_seconds` | none | Unix timestamp of the last event for the zone. |

## Jeedom Metrics

Jeedom metrics are emitted only when `AJAXBRIDGE_JEEDOM_ENABLED=true` and Jeedom MQTT messages are received.

| Metric | Type | Labels | Meaning |
| --- | --- | --- | --- |
| `ajax_jeedom_mqtt_messages_total` | counter | none | Total Jeedom MQTT messages received by the Jeedom service. |
| `ajax_jeedom_mqtt_parse_errors_total` | counter | none | Total Jeedom MQTT messages that could not be parsed. |
| `ajax_jeedom_empty_values_total` | counter | none | Jeedom messages with empty or null values. |
| `ajax_jeedom_last_update_timestamp_seconds` | gauge | `device`, `command`, `command_id`, `metric` | Last update timestamp for a Jeedom command. |
| `ajax_jeedom_command_value` | gauge | `device`, `command`, `command_id`, `metric` | Last numeric value for a Jeedom command. |
| `ajax_jeedom_device_power_watts` | gauge | `device` | Latest device power in watts. |
| `ajax_jeedom_device_current_amperes` | gauge | `device` | Latest device current in amperes. |
| `ajax_jeedom_device_voltage_volts` | gauge | `device` | Latest device voltage in volts. |
| `ajax_jeedom_device_temperature_celsius` | gauge | `device` | Latest device temperature in Celsius. |
| `ajax_jeedom_device_battery_percent` | gauge | `device` | Latest device battery percentage. |

Jeedom `command` labels are translated to English. The underlying debug JSON still keeps `raw_name`.

## Query Examples

Accounts offline:

```promql
ajax_account_online == 0
```

Any active alarm:

```promql
ajax_account_alarm_active == 1
```

Active zone alarms with device names:

```promql
ajax_zone_alarm_active == 1
```

SIA parse errors in the last 5 minutes:

```promql
increase(ajax_sia_parse_errors_total[5m])
```

Jeedom parse errors in the last 5 minutes:

```promql
increase(ajax_jeedom_mqtt_parse_errors_total[5m])
```

Power by Jeedom device:

```promql
ajax_jeedom_device_power_watts
```

Forward receiver latency:

```promql
histogram_quantile(
  0.95,
  sum(rate(ajax_sia_forward_duration_seconds_bucket[5m])) by (le, target)
)
```

## Dashboard Notes

- Use SIA metrics for alarm/security status.
- Use Jeedom metrics for power, voltage, current, temperature, battery, and signal diagnostics.
- Prefer `account`, `zone`, and `device_name` labels for panels and alerts.
- Timestamps are Unix seconds. A zero timestamp means the event has not been seen yet.
- Boolean gauges use `1` for true and `0` for false.
