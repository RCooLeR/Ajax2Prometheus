import { useEffect, useState } from 'react';
import type { CameraStreamProfile, Device, EventItem, Room, RoomSummary } from '../models/dashboard';
import type { HomeAssistant } from '../ha/types';
import { CameraStrip } from './CameraStrip';
import { RoomHero } from './RoomHero';
import { RoomSidebar } from './RoomSidebar';
import { RoomWorkspaceSidebar } from './RoomWorkspaceSidebar';

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
  const [selectedCameraId, setSelectedCameraId] = useState<string | null>(null);
  const [streamProfile, setStreamProfile] = useState<CameraStreamProfile>('main');
  const [audioByCameraId, setAudioByCameraId] = useState<Record<string, { muted: boolean; volume: number }>>({});
  const cameras = roomDevices.filter((device) => device.type === 'camera');
  const selectedCamera = cameras.find((device) => device.id === selectedCameraId) ?? null;
  const selectedCameraAudio = selectedCamera ? audioByCameraId[selectedCamera.id] ?? { muted: true, volume: 1 } : { muted: true, volume: 1 };

  useEffect(() => {
    setSelectedCameraId(null);
    setStreamProfile('main');
  }, [selectedRoom.id]);

  function handlePlayCamera(cameraId: string, profile: CameraStreamProfile) {
    setSelectedCameraId(cameraId);
    setStreamProfile(profile);
  }

  return (
    <section className={`room-detail-layout ${embedded ? 'room-detail-layout--embedded' : ''}`}>
      <RoomSidebar
        rooms={rooms}
        roomSummaries={roomSummaries}
        selectedRoomId={selectedRoom.id}
        onSelectRoom={onSelectRoom}
      />
      <main className="room-detail-layout__main">
        <RoomHero
          room={selectedRoom}
          roomSummary={roomSummaries[selectedRoom.id]}
          roomEvents={roomEvents}
          selectedDevice={selectedCamera}
          streamProfile={streamProfile}
          audioMuted={selectedCameraAudio.muted}
          audioVolume={selectedCameraAudio.volume}
          hass={hass}
        />
        <CameraStrip
          cameras={cameras}
          selectedCameraId={selectedCameraId}
          selectedProfile={streamProfile}
          audioByCameraId={audioByCameraId}
          onPlayCamera={handlePlayCamera}
          onChangeAudio={(cameraId, audio) => setAudioByCameraId((current) => ({ ...current, [cameraId]: audio }))}
        />
      </main>
      <RoomWorkspaceSidebar
        devices={roomDevices}
        events={roomEvents}
        selectedDeviceId={selectedDeviceId}
        onSelectDevice={onSelectDevice}
        hass={hass}
      />
    </section>
  );
}
