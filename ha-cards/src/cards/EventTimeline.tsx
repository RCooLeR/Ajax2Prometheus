import type { EventItem } from '../models/dashboard';
import { Icon } from '../components/Icon';
import { getToneClass } from '../utils/assets';
import { formatEventStamp } from '../utils/format';

interface EventTimelineProps {
  events: EventItem[];
  totalEvents: number;
  activeDeviceName: string | null;
  onClearFilter: () => void;
}

export function EventTimeline({ events, totalEvents, activeDeviceName, onClearFilter }: EventTimelineProps) {
  return (
    <aside className="event-timeline glass-panel">
      <div className="section-heading event-timeline__heading">
        <div className="event-timeline__heading-copy">
          <span>Recent room events</span>
          {activeDeviceName ? <strong className="event-timeline__filter-label">{activeDeviceName}</strong> : null}
        </div>
        <div className="event-timeline__heading-actions">
          {activeDeviceName ? (
            <button type="button" className="event-timeline__clear" onClick={onClearFilter}>
              Show all
            </button>
          ) : null}
          <strong>{events.length === totalEvents ? events.length : `${events.length}/${totalEvents}`}</strong>
        </div>
      </div>
      <div className="event-timeline__list">
        {events.length > 0 ? (
          events.map((event) => (
            <article key={event.id} className={['event-item', getToneClass(event.tone)].join(' ')}>
              <div className="event-item__rail" />
              <span className="event-item__icon-wrap">
                <Icon icon={event.icon} size={42} />
              </span>
              <div className="event-item__copy">
                <div className="event-item__title">{event.title}</div>
                <div className="event-item__description">{event.description}</div>
                <div className="event-item__source">{event.source}</div>
              </div>
              <time className="event-item__time" dateTime={event.occurredAt}>
                {formatEventStamp(event.occurredAt)}
              </time>
            </article>
          ))
        ) : (
          <div className="event-timeline__empty">
            <strong>No matching events</strong>
            <span>Choose another device or show all room activity.</span>
          </div>
        )}
      </div>
    </aside>
  );
}
