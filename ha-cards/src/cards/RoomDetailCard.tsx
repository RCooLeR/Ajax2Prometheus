import type { Device, EventItem, Room, RoomSummary } from '../models/dashboard';
import type { HomeAssistant } from '../ha/types';
import { DeviceGrid } from './DeviceGrid';
import { EventTimeline } from './EventTimeline';
import { RoomHero } from './RoomHero';
import { RoomSidebar } from './RoomSidebar';

interface RoomDetailCardProps {
  rooms: Room[];
  selectedRoom: Room;
  roomDevices: Device[];
  roomEvents: EventItem[];
  totalRoomEvents: number;
  roomSummaries: Record<string, RoomSummary>;
  selectedDeviceId: string | null;
  selectedDeviceName: string | null;
  selectedDevice: Device | null;
  onSelectDevice: (deviceId: string | null) => void;
  onClearDeviceFilter: () => void;
  onSelectRoom: (roomId: string) => void;
  embedded?: boolean;
  hass?: HomeAssistant;
}

export function RoomDetailCard({
  rooms,
  selectedRoom,
  roomDevices,
  roomEvents,
  totalRoomEvents,
  roomSummaries,
  selectedDeviceId,
  selectedDeviceName,
  selectedDevice,
  onSelectDevice,
  onClearDeviceFilter,
  onSelectRoom,
  embedded = false,
  hass,
}: RoomDetailCardProps) {
  return (
    <section className={`room-detail-layout ${embedded ? 'room-detail-layout--embedded' : ''}`}>
      <RoomSidebar
        rooms={rooms}
        roomSummaries={roomSummaries}
        selectedRoomId={selectedRoom.id}
        onSelectRoom={onSelectRoom}
      />
      <main className="room-detail-layout__main">
        <RoomHero room={selectedRoom} roomSummary={roomSummaries[selectedRoom.id]} selectedDevice={selectedDevice} hass={hass} />
        <DeviceGrid
          devices={roomDevices}
          events={roomEvents}
          selectedDeviceId={selectedDeviceId}
          onSelectDevice={onSelectDevice}
        />
      </main>
      <EventTimeline
        events={roomEvents}
        totalEvents={totalRoomEvents}
        activeDeviceName={selectedDeviceName}
        onClearFilter={onClearDeviceFilter}
      />
    </section>
  );
}
