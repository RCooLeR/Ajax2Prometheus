# Home Assistant Card Examples

AjaxBridge already publishes enough MQTT discovery entities to build a good-looking Lovelace dashboard without extra backend work.

## Entity ID Pattern

Discovery object IDs come from the Go code, so the usual entity IDs look like this:

```text
binary_sensor.account_a0f80d_online
binary_sensor.account_a0f80d_alarm_active
sensor.account_a0f80d_mode

binary_sensor.zone_a0f80d_3_alarm_active
binary_sensor.zone_a0f80d_3_signal_fire
binary_sensor.zone_a0f80d_3_signal_smoke
binary_sensor.zone_a0f80d_3_signal_tamper
binary_sensor.zone_a0f80d_3_signal_battery
sensor.zone_a0f80d_3_last_event_name
sensor.zone_a0f80d_3_last_event_at
```

Home Assistant may add suffixes if something already exists, so confirm the final entity IDs in `Settings -> Devices & services -> Entities`.

## Recommended Custom Cards

For the nicer card below, install:

- `button-card`
- `stack-in-card`

You can install both through HACS.

## Ajax Hero Card

This card gives one Ajax zone a strong status panel with alarm-aware colors and a compact signal row underneath.

Replace:

- `A0F80D` with your Ajax account ID
- `3` with the zone/device number from `devices.json`
- `Hall FireProtect` with the device name you want on the card

```yaml
type: custom:stack-in-card
mode: vertical
cards:
  - type: custom:button-card
    entity: binary_sensor.zone_a0f80d_3_alarm_active
    name: Hall FireProtect
    icon: mdi:shield-home
    show_state: false
    show_label: true
    tap_action:
      action: more-info
    hold_action:
      action: more-info
    label: |
      [[[
        const fire = states['binary_sensor.zone_a0f80d_3_signal_fire']?.state === 'on';
        const smoke = states['binary_sensor.zone_a0f80d_3_signal_smoke']?.state === 'on';
        const tamper = states['binary_sensor.zone_a0f80d_3_tamper_active']?.state === 'on';
        const trouble = states['binary_sensor.zone_a0f80d_3_trouble_active']?.state === 'on';
        const battery = states['binary_sensor.zone_a0f80d_3_signal_battery']?.state === 'on';
        const lastEvent = states['sensor.zone_a0f80d_3_last_event_name']?.state || 'No events yet';
        const lastAt = states['sensor.zone_a0f80d_3_last_event_at']?.state;

        if (fire || smoke) return 'Fire alarm';
        if (entity.state === 'on') return 'Alarm active';
        if (tamper) return 'Tamper active';
        if (battery) return 'Battery low';
        if (trouble) return 'Trouble active';
        if (!lastAt || lastAt === 'unknown' || lastAt === 'unavailable') return lastEvent;
        return `${lastEvent} - ${lastAt}`;
      ]]]
    custom_fields:
      mode: |
        [[[
          const signal = states['sensor.zone_a0f80d_3_last_signal']?.state || 'idle';
          const alarmSignal = states['sensor.zone_a0f80d_3_alarm_signal']?.state || 'none';
          return `
            <div style="font-size:12px; letter-spacing:0.12em; text-transform:uppercase; opacity:0.8;">Ajax zone</div>
            <div style="font-size:28px; font-weight:700; line-height:1.1;">${entity.state === 'on' ? 'Triggered' : 'Stable'}</div>
            <div style="font-size:13px; opacity:0.85;">signal: ${signal} | alarm: ${alarmSignal}</div>
          `;
        ]]]
    styles:
      grid:
        - grid-template-areas: '"i mode" "n n" "l l"'
        - grid-template-columns: 72px 1fr
        - grid-template-rows: min-content min-content min-content
        - column-gap: 16px
      card:
        - padding: 22px
        - border-radius: 24px
        - color: white
        - box-shadow: 0 18px 40px rgba(0, 0, 0, 0.22)
        - background: |
            [[[
              const fire = states['binary_sensor.zone_a0f80d_3_signal_fire']?.state === 'on';
              const smoke = states['binary_sensor.zone_a0f80d_3_signal_smoke']?.state === 'on';
              const tamper = states['binary_sensor.zone_a0f80d_3_tamper_active']?.state === 'on';
              const trouble = states['binary_sensor.zone_a0f80d_3_trouble_active']?.state === 'on';
              const battery = states['binary_sensor.zone_a0f80d_3_signal_battery']?.state === 'on';

              if (fire || smoke) return 'linear-gradient(160deg, #4f000b 0%, #9d0208 48%, #ffba08 100%)';
              if (entity.state === 'on') return 'linear-gradient(160deg, #370617 0%, #d00000 58%, #f77f00 100%)';
              if (tamper || trouble || battery) return 'linear-gradient(160deg, #3d2c00 0%, #8d6e00 55%, #e9c46a 100%)';
              return 'linear-gradient(160deg, #0b1f33 0%, #13505b 55%, #0fa3b1 100%)';
            ]]]
      icon:
        - width: 42px
        - color: white
      img_cell:
        - width: 72px
        - height: 72px
        - border-radius: 20px
        - background: rgba(255, 255, 255, 0.16)
      name:
        - justify-self: start
        - margin-top: 18px
        - font-size: 20px
        - font-weight: 700
      label:
        - justify-self: start
        - margin-top: 8px
        - font-size: 13px
        - line-height: 1.4
        - opacity: 0.9
      custom_fields:
        mode:
          - justify-self: start
          - align-self: center
  - type: grid
    columns: 4
    square: false
    cards:
      - type: custom:button-card
        entity: binary_sensor.zone_a0f80d_3_signal_fire
        name: Fire
        icon: mdi:fire
        show_state: false
        tap_action:
          action: more-info
        styles:
          card:
            - border-radius: 18px
            - padding: 14px 10px
            - background: |
                [[[
                  return entity.state === 'on'
                    ? 'linear-gradient(180deg, #9d0208 0%, #dc2f02 100%)'
                    : 'linear-gradient(180deg, #1f2937 0%, #334155 100%)';
                ]]]
            - color: white
          name:
            - font-size: 12px
      - type: custom:button-card
        entity: binary_sensor.zone_a0f80d_3_signal_smoke
        name: Smoke
        icon: mdi:smoke
        show_state: false
        tap_action:
          action: more-info
        styles:
          card:
            - border-radius: 18px
            - padding: 14px 10px
            - background: |
                [[[
                  return entity.state === 'on'
                    ? 'linear-gradient(180deg, #7f1d1d 0%, #b91c1c 100%)'
                    : 'linear-gradient(180deg, #1f2937 0%, #334155 100%)';
                ]]]
            - color: white
          name:
            - font-size: 12px
      - type: custom:button-card
        entity: binary_sensor.zone_a0f80d_3_signal_tamper
        name: Tamper
        icon: mdi:shield-alert
        show_state: false
        tap_action:
          action: more-info
        styles:
          card:
            - border-radius: 18px
            - padding: 14px 10px
            - background: |
                [[[
                  return entity.state === 'on'
                    ? 'linear-gradient(180deg, #7c2d12 0%, #ea580c 100%)'
                    : 'linear-gradient(180deg, #1f2937 0%, #334155 100%)';
                ]]]
            - color: white
          name:
            - font-size: 12px
      - type: custom:button-card
        entity: binary_sensor.zone_a0f80d_3_signal_battery
        name: Battery
        icon: mdi:battery-alert
        show_state: false
        tap_action:
          action: more-info
        styles:
          card:
            - border-radius: 18px
            - padding: 14px 10px
            - background: |
                [[[
                  return entity.state === 'on'
                    ? 'linear-gradient(180deg, #78350f 0%, #d97706 100%)'
                    : 'linear-gradient(180deg, #1f2937 0%, #334155 100%)';
                ]]]
            - color: white
          name:
            - font-size: 12px
```

## Stock Home Assistant Version

If you do not want custom cards, this built-in version still works well:

```yaml
type: vertical-stack
cards:
  - type: tile
    entity: binary_sensor.zone_a0f80d_3_alarm_active
    name: Hall FireProtect
    color: red
    vertical: false
    features_position: bottom
  - type: glance
    columns: 4
    entities:
      - entity: binary_sensor.zone_a0f80d_3_signal_fire
        name: Fire
      - entity: binary_sensor.zone_a0f80d_3_signal_smoke
        name: Smoke
      - entity: binary_sensor.zone_a0f80d_3_signal_tamper
        name: Tamper
      - entity: binary_sensor.zone_a0f80d_3_signal_battery
        name: Battery
  - type: entities
    title: Ajax zone details
    entities:
      - entity: sensor.zone_a0f80d_3_last_event_name
        name: Last event
      - entity: sensor.zone_a0f80d_3_last_signal
        name: Last signal
      - entity: sensor.zone_a0f80d_3_last_event_at
        name: Last event time
      - entity: sensor.zone_a0f80d_3_alarm_signal
        name: Alarm signal
      - entity: sensor.zone_a0f80d_3_alarm_action
        name: Alarm action
```

## Account Overview Card

For the whole Ajax account, this compact grid works well:

```yaml
type: grid
columns: 4
square: false
cards:
  - type: tile
    entity: binary_sensor.account_a0f80d_online
    name: Online
    color: green
  - type: tile
    entity: binary_sensor.account_a0f80d_armed
    name: Armed
    color: blue
  - type: tile
    entity: binary_sensor.account_a0f80d_alarm_active
    name: Alarm
    color: red
  - type: tile
    entity: binary_sensor.account_a0f80d_trouble_active
    name: Trouble
    color: amber
```

## Practical Setup Flow

1. Let AjaxBridge publish MQTT discovery once.
2. Open the discovered device in Home Assistant and copy the final entity IDs.
3. Paste one of the YAML blocks above into a manual card.
4. Replace account and zone IDs with your real ones.
5. Duplicate the zone card for each important detector, siren, or keypad.

## Generated Detailed View

This repo also includes [ha-detailed-devices.yaml](./examples/ha-detailed-devices.yaml), a full detailed Lovelace view generated from the current `bridge/data/devices.json`.

It is grouped by room and includes one `entities` card per device using the discovered zone entities.
