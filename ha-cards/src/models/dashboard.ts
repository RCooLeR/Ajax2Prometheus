export type GlowTone = 'cyan' | 'green' | 'amber' | 'red' | 'violet' | 'slate';

export type RoomType = string;

export type DeviceType = string;

export type EventType = string;

export type IconCategory =
  | 'rooms'
  | 'devices'
  | 'system-states'
  | 'security-states'
  | 'sensors'
  | 'events'
  | 'connectivity'
  | 'misc';

export interface IconRef {
  category: IconCategory;
  key: string;
}

export interface IconRegistryEntry {
  svg: string;
  png?: string;
  color: string;
}

export type IconRegistry = Record<IconCategory, Record<string, IconRegistryEntry>>;

export interface DashboardChip {
  id: string;
  label: string;
  value: string;
  icon: IconRef;
  tone: GlowTone;
  active: boolean;
}

export interface Room {
  id: string;
  name: string;
  type: RoomType;
  summary: string;
  heroLabel: string;
  image: string;
  accent: string;
  icon: IconRef;
  statusTone: GlowTone;
  stateChips: DashboardChip[];
}

export interface Device {
  id: string;
  roomId: string;
  type: DeviceType;
  name: string;
  model: string;
  icon: IconRef;
  tone: GlowTone;
  status: string;
  connectivity: string;
  battery: string;
  signal: string;
  entityId: string;
  isOnline: boolean;
  attention: boolean;
}

export interface EventItem {
  id: string;
  roomId: string;
  deviceId?: string;
  type: EventType;
  title: string;
  description: string;
  occurredAt: string;
  source: string;
  icon: IconRef;
  tone: GlowTone;
}

export interface SystemState {
  chips: DashboardChip[];
}

export interface RoomSummary {
  roomId: string;
  deviceCount: number;
  onlineCount: number;
  attentionCount: number;
  latestEventLabel: string;
  tone: GlowTone;
}

export interface DashboardData {
  systemState: SystemState;
  rooms: Room[];
  devices: Device[];
  events: EventItem[];
}
