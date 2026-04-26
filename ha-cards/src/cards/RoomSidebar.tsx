import type { Room, RoomSummary } from '../models/dashboard';
import { RoomListItem } from '../components/RoomListItem';

interface RoomSidebarProps {
  rooms: Room[];
  roomSummaries: Record<string, RoomSummary>;
  selectedRoomId: string;
  onSelectRoom: (roomId: string) => void;
}

export function RoomSidebar({ rooms, roomSummaries, selectedRoomId, onSelectRoom }: RoomSidebarProps) {
  return (
    <aside className="room-sidebar glass-panel">
      <div className="section-heading">
        <span>Protected rooms</span>
        <strong>{rooms.length}</strong>
      </div>
      <div className="room-sidebar__list">
        {rooms.map((room) => {
          const summary = roomSummaries[room.id];

          return (
            <RoomListItem
              key={room.id}
              room={room}
              summary={summary}
              selected={room.id === selectedRoomId}
              onSelect={onSelectRoom}
            />
          );
        })}
      </div>
    </aside>
  );
}
