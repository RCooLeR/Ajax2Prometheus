import { useMemo, useState } from 'react';
import type { Device, EventItem } from '../models/dashboard';
import type { HomeAssistant } from '../ha/types';
import { Icon } from '../components/Icon';
import { StatusBadge } from '../components/StatusBadge';
import { getToneClass } from '../utils/assets';
import { formatEventStamp } from '../utils/format';

interface RoomWorkspaceSidebarProps {
  devices: Device[];
  events: EventItem[];
  selectedDeviceId: string | null;
  onSelectDevice: (deviceId: string | null) => void;
  hass?: HomeAssistant;
}

type SidebarTab = 'devices' | 'events';

const DAHUA_EVENT_TYPES = new Set(['human_detected', 'vehicle_detected', 'tripwire_detected', 'intrusion_detected']);
const AJAX_EVENT_TYPES = new Set(['alarm', 'armed', 'disarmed', 'turn_on', 'turn_off', 'impulse']);
const USEFUL_METRIC_LABELS = new Set(['temperature', 'voltage', 'current', 'power']);

export function RoomWorkspaceSidebar({
  devices,
  events,
  selectedDeviceId,
  onSelectDevice,
  hass,
}: RoomWorkspaceSidebarProps) {
  const [activeTab, setActiveTab] = useState<SidebarTab>('devices');
  const deviceById = useMemo(() => new Map(devices.map((device) => [device.id, device])), [devices]);
  const ajaxDevices = devices.filter((device) => device.type !== 'camera');
  const meaningfulEvents = events.filter((event) => isMeaningfulEvent(event, deviceById));

  return (
    <aside className="room-workspace-sidebar glass-panel">
      <div className="room-workspace-sidebar__tabs" role="tablist" aria-label="Room sidebar">
        <button type="button" className={activeTab === 'devices' ? 'is-active' : ''} onClick={() => setActiveTab('devices')}>
          Devices
        </button>
        <button type="button" className={activeTab === 'events' ? 'is-active' : ''} onClick={() => setActiveTab('events')}>
          Events
        </button>
      </div>
      {activeTab === 'devices' ? (
        <div className="room-workspace-sidebar__list">
          {ajaxDevices.length > 0 ? (
            ajaxDevices.map((device) => (
              <AjaxDeviceListItem
                key={device.id}
                device={device}
                selected={device.id === selectedDeviceId}
                onSelect={onSelectDevice}
                hass={hass}
              />
            ))
          ) : (
            <EmptySidebarState title="No Ajax devices" copy="Assign Ajax entities to this area to control them here." />
          )}
        </div>
      ) : (
        <div className="room-workspace-sidebar__list">
          {meaningfulEvents.length > 0 ? (
            meaningfulEvents.map((event) => <SidebarEvent key={event.id} event={event} />)
          ) : (
            <EmptySidebarState title="No meaningful events" copy="Alarm, arm/disarm, relay, and Dahua SMD/IVS events will appear here." />
          )}
        </div>
      )}
    </aside>
  );
}

interface AjaxDeviceListItemProps {
  device: Device;
  selected: boolean;
  onSelect: (deviceId: string | null) => void;
  hass?: HomeAssistant;
}

function AjaxDeviceListItem({ device, selected, onSelect, hass }: AjaxDeviceListItemProps) {
  const [pending, setPending] = useState(false);
  const metrics = (device.metrics ?? []).filter((metric) => USEFUL_METRIC_LABELS.has(metric.label.toLowerCase()));
  const switchAction = getSwitchAction(device);
  const impulseAction = getImpulseAction(device);
  const isOn = deviceLooksOn(device);

  async function callAction(action: { domain: 'button' | 'switch' | 'lock'; service: string; entityId: string }) {
    if (!hass?.callService) {
      return;
    }
    setPending(true);
    try {
      await hass.callService(action.domain, action.service, { entity_id: action.entityId });
    } finally {
      setPending(false);
    }
  }

  return (
    <article className={['ajax-device-item', getToneClass(device.tone), selected ? 'ajax-device-item--selected' : ''].join(' ')}>
      <button type="button" className="ajax-device-item__main" onClick={() => onSelect(selected ? null : device.id)}>
        <span className="ajax-device-item__icon-wrap">
          <Icon icon={device.icon} size={34} />
        </span>
        <span className="ajax-device-item__copy">
          <strong>{device.name}</strong>
          <span>{device.model}</span>
        </span>
        <StatusBadge label={device.isOnline ? 'Online' : 'Offline'} tone={device.isOnline ? 'green' : 'red'} />
      </button>
      {metrics.length > 0 ? (
        <div className="ajax-device-item__metrics">
          {metrics.map((metric) => (
            <span key={metric.id} className={getToneClass(metric.tone)}>
              <Icon icon={metric.icon} size={16} />
              <small>{metric.label}</small>
              <strong>{metric.value}</strong>
            </span>
          ))}
        </div>
      ) : null}
      {switchAction || impulseAction ? (
        <div className="ajax-device-item__controls">
          {switchAction ? (
            <button
              type="button"
              className={['ajax-device-item__toggle', isOn ? 'ajax-device-item__toggle--on' : ''].join(' ')}
              role="switch"
              aria-checked={isOn}
              disabled={pending || !hass?.callService}
              onClick={() => callAction(switchAction)}
            >
              <span className="ajax-device-item__toggle-track">
                <span className="ajax-device-item__toggle-thumb" />
              </span>
              <span>{isOn ? 'On' : 'Off'}</span>
            </button>
          ) : null}
          {impulseAction ? (
            <button
              type="button"
              className="ajax-device-item__impulse"
              aria-label={`Send impulse to ${device.name}`}
              title="Impulse"
              disabled={pending || !hass?.callService}
              onClick={() => callAction(impulseAction)}
            >
              <Icon icon={{ category: 'devices', key: 'relay' }} size={20} />
            </button>
          ) : null}
        </div>
      ) : null}
    </article>
  );
}

function SidebarEvent({ event }: { event: EventItem }) {
  return (
    <article className={['sidebar-event', getToneClass(event.tone)].join(' ')}>
      <span className="sidebar-event__icon-wrap">
        <Icon icon={normalizeEventIcon(event)} size={30} />
      </span>
      <div className="sidebar-event__copy">
        <strong>{event.title}</strong>
        <span>{event.description}</span>
        <small>{event.source}</small>
      </div>
      <time dateTime={event.occurredAt}>{formatEventStamp(event.occurredAt)}</time>
    </article>
  );
}

function EmptySidebarState({ title, copy }: { title: string; copy: string }) {
  return (
    <div className="event-timeline__empty">
      <strong>{title}</strong>
      <span>{copy}</span>
    </div>
  );
}

function isMeaningfulEvent(event: EventItem, deviceById: Map<string, Device>): boolean {
  if (DAHUA_EVENT_TYPES.has(event.type)) {
    return true;
  }
  if (event.type === 'motion_detected' && event.deviceId && deviceById.get(event.deviceId)?.type === 'camera') {
    return true;
  }
  return AJAX_EVENT_TYPES.has(event.type);
}

function normalizeEventIcon(event: EventItem) {
  if (event.type === 'armed') {
    return { category: 'events' as const, key: 'armed_event' };
  }
  if (event.type === 'disarmed') {
    return { category: 'events' as const, key: 'disarmed_event' };
  }
  if (event.type === 'turn_on') {
    return { category: 'misc' as const, key: 'energy' };
  }
  if (event.type === 'turn_off') {
    return { category: 'system-states' as const, key: 'power_loss' };
  }
  if (event.type === 'impulse') {
    return { category: 'devices' as const, key: 'relay' };
  }
  return event.icon;
}

function getSwitchAction(device: Device): { domain: 'switch'; service: string; entityId: string } | null {
  if (!isSwitchControlledDevice(device)) {
    return null;
  }
  const action = device.actions?.find((candidate) => candidate.domain === 'switch');
  if (action) {
    return { domain: 'switch', service: action.service, entityId: action.entityId };
  }
  if (device.entityId.startsWith('switch.')) {
    return { domain: 'switch', service: deviceLooksOn(device) ? 'turn_off' : 'turn_on', entityId: device.entityId };
  }
  return null;
}

function getImpulseAction(device: Device): { domain: 'button' | 'switch' | 'lock'; service: string; entityId: string } | null {
  if (!isImpulseRelay(device)) {
    return null;
  }
  const action = device.actions?.find((candidate) => candidate.domain === 'button') ?? device.actions?.[0];
  if (action) {
    return { domain: action.domain, service: action.service, entityId: action.entityId };
  }
  if (device.entityId.startsWith('switch.')) {
    return { domain: 'switch', service: 'turn_on', entityId: device.entityId };
  }
  return null;
}

function isSwitchControlledDevice(device: Device): boolean {
  const text = deviceText(device);
  return !isImpulseRelay(device)
    && (device.type === 'smart_plug'
      || device.type === 'wall_switch'
      || device.type === 'light_switch'
      || device.type === 'waterstop'
      || /wallswitch|wall switch|lightswitch|light switch|waterstop|water stop|socket|plug|outlet/.test(text));
}

function isImpulseRelay(device: Device): boolean {
  const text = deviceText(device);
  return device.type === 'relay' && !/wallswitch|wall switch|lightswitch|light switch|waterstop|water stop/.test(text);
}

function deviceLooksOn(device: Device): boolean {
  if (device.actions?.some((action) => action.service === 'turn_off')) {
    return true;
  }
  if (device.actions?.some((action) => action.service === 'turn_on')) {
    return false;
  }
  return /\bon\b|active|enabled|load|w\b/.test(`${device.status} ${device.signal}`.toLowerCase());
}

function deviceText(device: Device): string {
  return `${device.type} ${device.name} ${device.model} ${device.entityId}`.toLowerCase();
}
