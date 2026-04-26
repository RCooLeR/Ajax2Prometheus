import type { DashboardChip } from '../models/dashboard';
import { SystemChip } from '../components/SystemChip';

interface RoomStateChipsProps {
  chips: DashboardChip[];
}

export function RoomStateChips({ chips }: RoomStateChipsProps) {
  return (
    <div className="room-state-chips">
      {chips.map((chip) => (
        <SystemChip key={chip.id} chip={chip} compact />
      ))}
    </div>
  );
}
