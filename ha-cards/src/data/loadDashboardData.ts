import devicesData from './devices.json';
import eventsData from './events.json';
import roomsData from './rooms.json';
import systemStateData from './systemState.json';
import type { DashboardData, Device, EventItem, Room, RoomSmdIvsCounts, RoomSummary, SystemState } from '../models/dashboard';

export const dashboardData: DashboardData = {
  systemState: systemStateData as SystemState,
  rooms: roomsData as Room[],
  devices: devicesData as Device[],
  events: eventsData as EventItem[],
};

export function getDefaultRoomId(data: DashboardData): string {
  return data.rooms.find((room) => room.id === 'living-room')?.id ?? data.rooms[0]?.id ?? '';
}

export function getDevicesForRoom(data: DashboardData, roomId: string): Device[] {
  return data.devices.filter((device) => device.roomId === roomId);
}

export function getEventsForRoom(data: DashboardData, roomId: string): EventItem[] {
  return data.events
    .filter((event) => event.roomId === roomId)
    .sort((left, right) => Date.parse(right.occurredAt) - Date.parse(left.occurredAt));
}

export function getRoomSummaries(data: DashboardData): Record<string, RoomSummary> {
  return data.rooms.reduce<Record<string, RoomSummary>>((summaries, room) => {
    const roomDevices = getDevicesForRoom(data, room.id);
    const roomEvents = getEventsForRoom(data, room.id);
    const onlineCount = roomDevices.filter((device) => device.isOnline).length;
    const attentionCount = roomDevices.filter((device) => device.attention).length;
    const latestEvent = roomEvents[0];

    summaries[room.id] = {
      roomId: room.id,
      deviceCount: roomDevices.length,
      onlineCount,
      attentionCount,
      smdIvs: room.smdIvs ?? emptySmdIvsCounts(),
      dahuaCameraCount: room.dahuaCameraCount ?? 0,
      latestEventLabel: latestEvent?.title ?? 'No recent events',
      tone: attentionCount > 0 ? 'amber' : room.statusTone,
    };

    return summaries;
  }, {});
}

function emptySmdIvsCounts(): RoomSmdIvsCounts {
  return { total: 0, human: 0, vehicle: 0, animal: 0, ivs: 0 };
}
