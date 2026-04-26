import type { GlowTone } from '../models/dashboard';
import { getToneClass } from '../utils/assets';

interface StatusBadgeProps {
  label: string;
  tone: GlowTone;
}

export function StatusBadge({ label, tone }: StatusBadgeProps) {
  return (
    <span className={`status-badge ${getToneClass(tone)}`}>
      <span className="status-badge__dot" />
      {label}
    </span>
  );
}
