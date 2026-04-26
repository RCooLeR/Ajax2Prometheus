import type { Room, RoomSummary } from '../models/dashboard';
import { getRoomImageAsset } from '../utils/assets';
import { StatusBadge } from '../components/StatusBadge';
import { RoomStateChips } from './RoomStateChips';

interface RoomHeroProps {
  room: Room;
  roomSummary: RoomSummary;
}

export function RoomHero({ room, roomSummary }: RoomHeroProps) {
  return (
    <section
      className="room-hero"
      style={{
        backgroundImage: `url(${getRoomImageAsset(room.image)})`,
      }}
    >
      <div className="room-hero__header">
        <div>
          <div className="room-hero__eyebrow">{room.heroLabel}</div>
          <h2>{room.name}</h2>
          <p>{room.summary}</p>
        </div>
        <div className="room-hero__stats">
          <StatusBadge label={`${roomSummary.deviceCount} devices`} tone="cyan" />
          <StatusBadge label={`${roomSummary.onlineCount} online`} tone="green" />
          <StatusBadge
            label={roomSummary.attentionCount > 0 ? `${roomSummary.attentionCount} warnings` : 'All clear'}
            tone={roomSummary.attentionCount > 0 ? 'amber' : 'green'}
          />
        </div>
      </div>
      <RoomStateChips chips={room.stateChips} />
    </section>
  );
}
