import { createElement, useEffect, useRef, useState } from 'react';
import type { Device, GlowTone, IconRef, Room, RoomSummary } from '../models/dashboard';
import type { HomeAssistant, HomeAssistantState } from '../ha/types';
import { Icon } from '../components/Icon';
import { StatusBadge } from '../components/StatusBadge';
import { getRoomImageAsset, getToneClass } from '../utils/assets';

interface RoomHeroProps {
  room: Room;
  roomSummary: RoomSummary;
  selectedDevice: Device | null;
  hass?: HomeAssistant;
}

export function RoomHero({ room, roomSummary, selectedDevice, hass }: RoomHeroProps) {
  const [pendingActionId, setPendingActionId] = useState<string | null>(null);
  const [actionFeedback, setActionFeedback] = useState<string | null>(null);
  const [mediaSrc, setMediaSrc] = useState<string | null>(null);
  const heroMedia = selectedDevice?.heroMedia;
  const actions = selectedDevice?.actions ?? [];
  const videoMode = heroMedia?.kind === 'stream';
  const showDahuaStats = roomSummary.dahuaCameraCount > 0;

  useEffect(() => {
    setPendingActionId(null);
    setActionFeedback(null);
  }, [selectedDevice?.id]);

  useEffect(() => {
    setMediaSrc(heroMedia?.src ?? null);
  }, [heroMedia?.entityId, heroMedia?.src]);

  async function handleAction(actionId: string, domain: string, service: string, entityId: string) {
    if (!hass?.callService) {
      setActionFeedback('Home Assistant service API unavailable');
      return;
    }

    setPendingActionId(actionId);
    setActionFeedback(null);

    try {
      await hass.callService(domain, service, {}, { entity_id: entityId });
      setActionFeedback('Action sent');
    } catch {
      setActionFeedback('Action failed');
    } finally {
      setPendingActionId(null);
    }
  }

  return (
    <section
      className={[
        'room-hero',
        heroMedia ? 'room-hero--live' : '',
        videoMode ? 'room-hero--streaming' : '',
        actions.length > 0 ? 'room-hero--actionable' : '',
      ].join(' ')}
      style={!heroMedia ? { backgroundImage: `url(${getRoomImageAsset(room.image)})` } : undefined}
    >
      {heroMedia ? (
        <div className="room-hero__media-wrap">
          {videoMode && hass?.states[heroMedia.entityId] ? (
            <NativeCameraStream
              key={heroMedia.entityId}
              hass={hass}
              stateObj={hass.states[heroMedia.entityId]}
            />
          ) : (
            <img
              key={heroMedia.entityId}
              className="room-hero__media"
              src={mediaSrc ?? heroMedia.src}
              alt={heroMedia.title}
              onError={() => {
                if (heroMedia.posterSrc && mediaSrc !== heroMedia.posterSrc) {
                  setMediaSrc(heroMedia.posterSrc);
                }
              }}
            />
          )}
        </div>
      ) : null}
      <div className="room-hero__header">
        <div className="room-hero__copy">
          <div className="room-hero__eyebrow">{heroMedia ? `${room.heroLabel} - Live view` : room.heroLabel}</div>
          <h2>{room.name}</h2>
          <p>{room.summary}</p>
          {selectedDevice ? (
            <div className="room-hero__device-meta">
              <StatusBadge label={selectedDevice.name} tone={selectedDevice.isOnline ? 'cyan' : 'red'} />
              {heroMedia ? <StatusBadge label={heroMedia.kind === 'stream' ? 'Live camera' : 'Camera image'} tone="green" /> : null}
            </div>
          ) : null}
        </div>
        <div className="room-hero__stats">
          <RoomHeroStat
            label="Devices"
            value={roomSummary.deviceCount}
            icon={{ category: 'misc', key: 'info' }}
            tone="cyan"
          />
          <RoomHeroStat
            label="Online"
            value={`${roomSummary.onlineCount}/${roomSummary.deviceCount || 0}`}
            icon={{ category: 'system-states', key: roomSummary.onlineCount === roomSummary.deviceCount ? 'online' : 'offline' }}
            tone={roomSummary.onlineCount === roomSummary.deviceCount ? 'green' : 'amber'}
          />
          <RoomHeroStat
            label="Warnings"
            value={roomSummary.attentionCount}
            icon={{ category: 'misc', key: roomSummary.attentionCount > 0 ? 'alert' : 'check' }}
            tone={roomSummary.attentionCount > 0 ? 'amber' : 'green'}
          />
          {showDahuaStats ? (
            <>
              <RoomHeroStat
                label="Humans 24h"
                value={roomSummary.smdIvs.human}
                icon={{ category: 'events', key: 'human_detected' }}
                tone="violet"
              />
              <RoomHeroStat
                label="Vehicles 24h"
                value={roomSummary.smdIvs.vehicle}
                icon={{ category: 'events', key: 'vehicle_detected' }}
                tone="cyan"
              />
              <RoomHeroStat
                label="IVS 24h"
                value={roomSummary.smdIvs.ivs}
                icon={{ category: 'events', key: 'tripwire_detected' }}
                tone="amber"
              />
              {roomSummary.smdIvs.animal > 0 ? (
                <RoomHeroStat
                  label="Animals 24h"
                  value={roomSummary.smdIvs.animal}
                  icon={{ category: 'events', key: 'motion_detected' }}
                  tone="green"
                />
              ) : null}
            </>
          ) : null}
        </div>
      </div>
      {actions.length > 0 ? (
        <div className="room-hero__action-panel">
          <div className="room-hero__action-copy">
            <span className="room-hero__eyebrow">Device actions</span>
            <strong>{selectedDevice?.name}</strong>
            {actionFeedback ? <span className="room-hero__action-feedback">{actionFeedback}</span> : null}
          </div>
          <div className="room-hero__actions">
            {actions.map((action) => (
              <button
                key={action.id}
                type="button"
                className="room-hero__action-button"
                disabled={pendingActionId !== null}
                onClick={() => handleAction(action.id, action.domain, action.service, action.entityId)}
              >
                <span>{action.label}</span>
                {action.stateLabel ? <small>{action.stateLabel}</small> : null}
              </button>
            ))}
          </div>
        </div>
      ) : null}
    </section>
  );
}

interface NativeCameraStreamProps {
  hass: HomeAssistant;
  stateObj: HomeAssistantState;
}

type NativeCameraStreamElement = HTMLElement & {
  hass?: HomeAssistant;
  stateObj?: HomeAssistantState;
};

function NativeCameraStream({ hass, stateObj }: NativeCameraStreamProps) {
  const streamRef = useRef<NativeCameraStreamElement | null>(null);

  useEffect(() => {
    const streamElement = streamRef.current;
    if (!streamElement) {
      return;
    }

    const assignStreamProps = () => {
      streamElement.hass = hass;
      streamElement.stateObj = stateObj;
    };

    if (customElements.get('ha-camera-stream')) {
      assignStreamProps();
      return;
    }

    let active = true;
    void customElements.whenDefined('ha-camera-stream').then(() => {
      if (active) {
        assignStreamProps();
      }
    });

    return () => {
      active = false;
    };
  }, [hass, stateObj]);

  return createElement('ha-camera-stream', {
    ref: streamRef,
    className: 'room-hero__media room-hero__native-stream',
    'data-audio-muted': 'true',
    'data-audio-volume': '1',
  });
}

interface RoomHeroStatProps {
  label: string;
  value: number | string;
  icon: IconRef;
  tone: GlowTone;
}

function RoomHeroStat({ label, value, icon, tone }: RoomHeroStatProps) {
  return (
    <span className={`room-hero__stat ${getToneClass(tone)}`}>
      <span className="room-hero__stat-icon">
        <Icon icon={icon} size={22} />
      </span>
      <span className="room-hero__stat-copy">
        <strong>{typeof value === 'number' ? formatCount(value) : value}</strong>
        <span>{label}</span>
      </span>
    </span>
  );
}

function formatCount(value: number): string {
  return new Intl.NumberFormat('en-GB', { maximumFractionDigits: 0 }).format(value);
}
