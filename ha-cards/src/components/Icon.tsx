import { createElement } from 'react';
import type { IconRef } from '../models/dashboard';
import { getIconColor, getMaterialIconName } from '../utils/assets';

interface IconProps {
  icon: IconRef;
  size?: number;
  className?: string;
}

export function Icon({ icon, size = 42, className }: IconProps) {
  const tint = getIconColor(icon);
  const materialIcon = getMaterialIconName(icon);

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
      {createElement('ha-icon', { icon: materialIcon })}
      <span className="icon__fallback">{fallbackGlyph(materialIcon)}</span>
    </span>
  );
}

function fallbackGlyph(icon: string): string {
  if (/water|valve|leak/.test(icon)) {
    return 'W';
  }
  if (/fire|smoke|alarm|alert/.test(icon)) {
    return '!';
  }
  if (/battery|power|lightning|switch|socket/.test(icon)) {
    return 'P';
  }
  if (/door|window|lock|shield/.test(icon)) {
    return 'S';
  }
  if (/thermometer/.test(icon)) {
    return 'T';
  }
  return 'i';
}
