import type { SystemState } from '../models/dashboard';
import { SystemChip } from '../components/SystemChip';
import { getStaticAsset } from '../utils/assets';

interface TopSystemBarProps {
  systemState: SystemState;
  maxChips?: number;
  compact?: boolean;
}

export function TopSystemBar({ systemState, maxChips, compact = false }: TopSystemBarProps) {
  const chips = typeof maxChips === 'number' ? systemState.chips.slice(0, maxChips) : systemState.chips;
  const logoSrc = getStaticAsset('ajax-systems-logo.svg');

  return (
    <header
      className={[
        'top-system-bar',
        'glass-panel',
        compact ? 'top-system-bar--compact' : '',
      ]
        .filter(Boolean)
        .join(' ')}
    >
      <div className="brand-logo" aria-label="Ajax Systems">
        <img className="brand-logo__image" src={logoSrc} alt="Ajax Systems" />
      </div>

      <div className="top-system-bar__chips">
        {chips.map((chip) => (
          <SystemChip key={chip.id} chip={chip} />
        ))}
      </div>
    </header>
  );
}
