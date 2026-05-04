import iconRegistryData from '../data/iconRegistry.json';
import type { GlowTone, IconRef, IconRegistry } from '../models/dashboard';

const iconRegistry = iconRegistryData as IconRegistry;
let assetBaseUrl = '/assets/';

export function setAssetBaseUrl(baseUrl: string): void {
  assetBaseUrl = ensureTrailingSlash(baseUrl);
}

export function getIconAsset(icon: IconRef): string {
  const entry = iconRegistry[icon.category]?.[icon.key];
  return entry?.svg ? resolveAssetPath(entry.svg) : '';
}

export function getIconColor(icon: IconRef): string | undefined {
  return iconRegistry[icon.category]?.[icon.key]?.color;
}

export function getRoomImageAsset(fileName: string): string {
  if (!fileName) {
    return '';
  }
  if (/^(https?:)?\/\//.test(fileName) || fileName.startsWith('/')) {
    return normalizeHomeAssistantImageUrl(fileName);
  }
  return `${assetBaseUrl}rooms/${fileName}`;
}

export function getStaticAsset(path: string): string {
  return resolveAssetPath(path);
}

export function getToneClass(tone: GlowTone): string {
  return `tone-${tone}`;
}

export function withOpacity(hexColor: string, alpha: number): string {
  const normalized = hexColor.replace('#', '');

  if (normalized.length !== 6) {
    return `rgba(45, 226, 230, ${alpha})`;
  }

  const red = Number.parseInt(normalized.slice(0, 2), 16);
  const green = Number.parseInt(normalized.slice(2, 4), 16);
  const blue = Number.parseInt(normalized.slice(4, 6), 16);

  return `rgba(${red}, ${green}, ${blue}, ${alpha})`;
}

function resolveAssetPath(path: string): string {
  return `${assetBaseUrl}${stripAssetPrefix(path)}`;
}

function stripAssetPrefix(path: string): string {
  return path.replace(/^\/?assets\//, '');
}

function ensureTrailingSlash(path: string): string {
  return path.endsWith('/') ? path : `${path}/`;
}

function normalizeHomeAssistantImageUrl(url: string): string {
  return url.replace(/(\/api\/image\/serve\/[^/?]+)\/\d+x\d+(\?.*)?$/i, '$1/original$2');
}
