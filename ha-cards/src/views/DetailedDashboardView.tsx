import { useEffect, useState } from 'react';
import { DashboardShell } from '../cards/DashboardShell';
import { RoomDetailCard } from '../cards/RoomDetailCard';
import { TopSystemBar } from '../cards/TopSystemBar';
import {
  dashboardData,
  getDefaultRoomId,
  getDevicesForRoom,
  getEventsForRoom,
  getRoomSummaries,
} from '../data/loadDashboardData';
import { useDashboardData } from '../data/liveDashboardData';
import type { HomeAssistant } from '../ha/types';

interface DetailedDashboardViewProps {
  mode?: 'standalone' | 'embedded';
  initialRoomId?: string;
  hass?: HomeAssistant;
  account?: string;
}

export function DetailedDashboardView({ mode = 'standalone', initialRoomId, hass, account }: DetailedDashboardViewProps) {
  const liveData = useDashboardData(hass, account);
  const data = hass ? liveData : dashboardData;
  const [selectedRoomId, setSelectedRoomId] = useState<string>('');
  const [selectedDeviceId, setSelectedDeviceId] = useState<string | null>(null);

  const resolvedInitialRoomId = resolveInitialRoomId(data, initialRoomId);
  const selectedRoom =
    data.rooms.find((room) => room.id === selectedRoomId) ??
    data.rooms.find((room) => room.id === resolvedInitialRoomId) ??
    data.rooms[0];

  useEffect(() => {
    if (!selectedRoom && data.rooms.length > 0) {
      setSelectedRoomId(resolveInitialRoomId(data, initialRoomId));
      return;
    }
    if (selectedRoom && selectedRoom.id !== selectedRoomId) {
      setSelectedRoomId(selectedRoom.id);
    }
  }, [data, initialRoomId, selectedRoom, selectedRoomId]);

  const roomDevices = selectedRoom ? getDevicesForRoom(data, selectedRoom.id) : [];
  const roomEvents = selectedRoom ? getEventsForRoom(data, selectedRoom.id) : [];
  const selectedDevice =
    roomDevices.find((device) => device.id === selectedDeviceId) ?? null;
  const visibleEvents = selectedDevice
    ? roomEvents.filter((event) => event.deviceId === selectedDevice.id)
    : roomEvents;
  const roomSummaries = getRoomSummaries(data);

  useEffect(() => {
    setSelectedDeviceId(null);
  }, [selectedRoomId]);

  if (!selectedRoom) {
    return (
      <DashboardShell
        mode={mode}
        topBar={<TopSystemBar systemState={data.systemState} />}
        detail={
          <section className="room-detail-layout">
            <main className="device-grid glass-panel">
              <div className="section-heading">
                <span>Protected rooms</span>
                <strong>0</strong>
              </div>
              <div className="event-timeline__empty">
                <strong>No mapped areas</strong>
                <span>Assign Ajax or Dahua devices to Home Assistant areas to populate the dashboard.</span>
              </div>
            </main>
          </section>
        }
      />
    );
  }

  return (
    <DashboardShell
      mode={mode}
      topBar={<TopSystemBar systemState={data.systemState} />}
      detail={
        <RoomDetailCard
          rooms={data.rooms}
          selectedRoom={selectedRoom}
          roomDevices={roomDevices}
          roomEvents={visibleEvents}
          totalRoomEvents={roomEvents.length}
          roomSummaries={roomSummaries}
          selectedDeviceId={selectedDeviceId}
          selectedDeviceName={selectedDevice?.name ?? null}
          onSelectDevice={setSelectedDeviceId}
          onClearDeviceFilter={() => setSelectedDeviceId(null)}
          onSelectRoom={setSelectedRoomId}
          embedded={mode === 'embedded'}
        />
      }
    />
  );
}

function resolveInitialRoomId(data: typeof dashboardData, initialRoomId?: string): string {
  if (initialRoomId) {
    const normalized = initialRoomId.trim().toLowerCase();
    const directMatch = data.rooms.find((room) => room.id === initialRoomId || room.name.toLowerCase() === normalized);
    if (directMatch) {
      return directMatch.id;
    }
    const slugMatch = data.rooms.find((room) => slugPart(room.name) === slugPart(initialRoomId));
    if (slugMatch) {
      return slugMatch.id;
    }
  }
  return getDefaultRoomId(data);
}

function slugPart(value: string): string {
  return value
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '_')
    .replace(/^_+|_+$/g, '');
}
