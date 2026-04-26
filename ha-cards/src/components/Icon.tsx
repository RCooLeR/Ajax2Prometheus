import type { IconRef } from '../models/dashboard';
import { getIconAsset, getIconColor } from '../utils/assets';

interface IconProps {
  icon: IconRef;
  size?: number;
  className?: string;
}

export function Icon({ icon, size = 42, className }: IconProps) {
  const src = getIconAsset(icon);
  const tint = getIconColor(icon);

  return (
    <span
      className={`icon ${className ?? ''}`.trim()}
      style={{
        width: size,
        height: size,
        color: tint,
      }}
      aria-hidden="true"
    >
      {src ? <img src={src} alt="" width={size} height={size} loading="lazy" /> : null}
    </span>
  );
}
