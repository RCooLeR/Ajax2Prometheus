import type { Device, GlowTone } from '../models/dashboard';
import { getToneClass } from '../utils/assets';
import { Icon } from '../components/Icon';
import { StatusBadge } from '../components/StatusBadge';

interface DeviceCardProps {
  device: Device;
  eventStatusLabel: string;
  eventStatusTone: GlowTone;
  selected: boolean;
  onSelect: (deviceId: string | null) => void;
}

function getEventStatusClass(tone: GlowTone): string {
  if (tone === 'green') {
    return 'device-card__status-value--good';
  }

  if (tone === 'red') {
    return 'device-card__status-value--critical';
  }

  return 'device-card__status-value--alert';
}

export function DeviceCard({ device, eventStatusLabel, eventStatusTone, selected, onSelect }: DeviceCardProps) {
  return (
    <button
      type="button"
      className={['device-card', getToneClass(device.tone), selected ? 'device-card--selected' : ''].join(' ')}
      onClick={() => onSelect(selected ? null : device.id)}
      aria-pressed={selected}
    >
      <div className="device-card__head">
        <div className="device-card__identity">
          <span className="device-card__icon-wrap">
            <Icon icon={device.icon} size={42} />
          </span>
          <div className="device-card__body">
            <div className="device-card__name">{device.name}</div>
            <div className="device-card__model">{device.model}</div>
          </div>
        </div>
        <StatusBadge label={device.isOnline ? 'Online' : 'Offline'} tone={device.isOnline ? 'green' : 'red'} />
      </div>
      <div className="device-card__status-list">
        <div className="device-card__status-row">
          <span className="device-card__status-label">Event status</span>
          <strong className={['device-card__status-value', getEventStatusClass(eventStatusTone)].join(' ')}>
            {eventStatusLabel}
          </strong>
        </div>
      </div>
    </button>
  );
}
