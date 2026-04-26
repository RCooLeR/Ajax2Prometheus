import type { Device, EventItem, EventType, GlowTone } from '../models/dashboard';
import { DeviceCard } from './DeviceCard';

const ISSUE_EVENT_TYPES = new Set<EventType>([
  'alarm',
  'device_offline',
  'fire_detected',
  'smoke_detected',
  'leak_detected',
  'gas_detected',
  'tamper_detected',
  'battery_low',
  'power_lost',
]);

interface DeviceGridProps {
  devices: Device[];
  events: EventItem[];
  selectedDeviceId: string | null;
  onSelectDevice: (deviceId: string | null) => void;
}

function getDeviceEventStatus(device: Device, events: EventItem[]): { label: string; tone: GlowTone } {
  const latestIssueEvent = events.find((event) => event.deviceId === device.id && ISSUE_EVENT_TYPES.has(event.type));

  if (latestIssueEvent) {
    return {
      label: latestIssueEvent.title,
      tone: latestIssueEvent.tone,
    };
  }

  if (!device.isOnline) {
    return {
      label: 'Device offline',
      tone: 'red',
    };
  }

  if (device.attention) {
    return {
      label: device.status,
      tone: device.tone === 'green' ? 'amber' : device.tone,
    };
  }

  return {
    label: 'Nominal',
    tone: 'green',
  };
}

export function DeviceGrid({ devices, events, selectedDeviceId, onSelectDevice }: DeviceGridProps) {
  return (
    <section className="device-grid glass-panel">
      <div className="section-heading">
        <span>Room devices</span>
        <strong>{devices.length}</strong>
      </div>
      <div className="device-grid__list">
        {devices.map((device) => {
          const eventStatus = getDeviceEventStatus(device, events);

          return (
            <DeviceCard
              key={device.id}
              device={device}
              eventStatusLabel={eventStatus.label}
              eventStatusTone={eventStatus.tone}
              selected={device.id === selectedDeviceId}
              onSelect={onSelectDevice}
            />
          );
        })}
      </div>
    </section>
  );
}
