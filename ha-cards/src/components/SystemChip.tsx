import type { DashboardChip } from '../models/dashboard';
import { getToneClass } from '../utils/assets';
import { Icon } from './Icon';

interface SystemChipProps {
  chip: DashboardChip;
  compact?: boolean;
}

export function SystemChip({ chip, compact = false }: SystemChipProps) {
  return (
    <article
      className={[
        'system-chip',
        compact ? 'system-chip--compact' : '',
        chip.active ? 'system-chip--active' : 'system-chip--inactive',
        getToneClass(chip.tone),
      ]
        .filter(Boolean)
        .join(' ')}
    >
      <div className="system-chip__icon-wrap">
      <Icon icon={chip.icon} size={42} className="system-chip__icon" />
      </div>
      <div className="system-chip__copy">
        <span className="system-chip__label">{chip.label}</span>
        <strong className="system-chip__value">{chip.value}</strong>
      </div>
    </article>
  );
}
