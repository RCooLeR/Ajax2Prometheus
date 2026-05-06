import { createElement, useEffect, useRef, useState } from 'react';
import type { CameraStreamProfile, Device, EventItem, GlowTone, IconRef, Room, RoomSummary } from '../models/dashboard';
import type { HomeAssistant, HomeAssistantState } from '../ha/types';
import { Icon } from '../components/Icon';
import { StatusBadge } from '../components/StatusBadge';
import { getRoomImageAsset, getToneClass } from '../utils/assets';

interface RoomHeroProps {
  room: Room;
  roomSummary: RoomSummary;
  roomEvents: EventItem[];
  selectedDevice: Device | null;
  streamProfile: CameraStreamProfile;
  audioMuted: boolean;
  audioVolume: number;
  hass?: HomeAssistant;
}

export function RoomHero({ room, roomSummary, roomEvents, selectedDevice, streamProfile, audioMuted, audioVolume, hass }: RoomHeroProps) {
  const [pendingActionId, setPendingActionId] = useState<string | null>(null);
  const [actionFeedback, setActionFeedback] = useState<string | null>(null);
  const [mediaSrc, setMediaSrc] = useState<string | null>(null);
  const heroMedia = selectedDevice?.heroMedia ?? buildFallbackCameraMedia(selectedDevice);
  const actions = selectedDevice?.actions ?? [];
  const videoMode = heroMedia?.kind === 'stream';
  const showDahuaStats = roomSummary.dahuaCameraCount > 0;
  const climate = roomSummary.climate ?? room.climate;
  const safety = roomSummary.safety ?? room.safety ?? { smokeHigh: 0, coHigh: 0 };
  const smokeHigh = safety.smokeHigh || countRoomEvents(roomEvents, ['smoke_detected', 'fire_detected']);
  const coHigh = safety.coHigh || countRoomEvents(roomEvents, ['gas_detected']);

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
      console.debug('[ajaxbridge] calling Home Assistant service', { domain, service, entityId });
      await hass.callService(domain, service, { entity_id: entityId });
      setActionFeedback(`Sent ${domain}.${service}`);
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
              profile={streamProfile}
              muted={audioMuted}
              volume={audioVolume}
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
            label="SMD 24h"
            value={roomSummary.smdIvs.human + roomSummary.smdIvs.vehicle + roomSummary.smdIvs.animal}
            icon={{ category: 'events', key: 'human_detected' }}
            tone="violet"
          />
          <RoomHeroStat
            label="IVS 24h"
            value={roomSummary.smdIvs.ivs}
            icon={{ category: 'events', key: 'tripwire_detected' }}
            tone="amber"
          />
          {climate?.temperature ? (
            <RoomHeroStat
              label="Temperature"
              value={climate.temperature}
              icon={{ category: 'sensors', key: 'temperature' }}
              tone="cyan"
            />
          ) : null}
          <RoomHeroStat
            label="Hi smoke"
            value={smokeHigh}
            icon={{ category: 'sensors', key: 'smoke' }}
            tone={smokeHigh > 0 ? 'amber' : 'green'}
          />
          <RoomHeroStat
            label="Hi CO"
            value={coHigh}
            icon={{ category: 'sensors', key: 'gas' }}
            tone={coHigh > 0 ? 'amber' : 'green'}
          />
          {showDahuaStats ? (
            <>
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
  profile: CameraStreamProfile;
  muted: boolean;
  volume: number;
}

type NativeCameraStreamElement = HTMLElement & {
  hass?: HomeAssistant;
  stateObj?: HomeAssistantState;
};

function NativeCameraStream({ hass, stateObj, profile, muted, volume }: NativeCameraStreamProps) {
  const streamRef = useRef<NativeCameraStreamElement | null>(null);

  useEffect(() => {
    const streamElement = streamRef.current;
    if (!streamElement) {
      return;
    }

    const assignStreamProps = () => {
      streamElement.hass = hass;
      streamElement.stateObj = preferFocusedCameraState(stateObj, profile);
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
  }, [hass, profile, stateObj]);

  useEffect(() => {
    const streamElement = streamRef.current;
    if (!streamElement) {
      return;
    }

    return forceNestedVideoObjectFit(streamElement, muted, volume);
  }, [muted, stateObj.entity_id, volume]);

  return createElement('ha-camera-stream', {
    ref: streamRef,
    className: 'room-hero__media room-hero__native-stream',
    'data-audio-muted': String(muted),
    'data-audio-volume': String(volume),
  });
}

function forceNestedVideoObjectFit(rootElement: HTMLElement, muted: boolean, volume: number): () => void {
  const observers: MutationObserver[] = [];
  const observedRoots = new WeakSet<Node>();
  let rafId = 0;
  let cleanupTimer = 0;
  let active = true;

  const scheduleApply = () => {
    window.cancelAnimationFrame(rafId);
    rafId = window.requestAnimationFrame(apply);
  };

  const observeRoot = (root: Node & ParentNode) => {
    if (!active || observedRoots.has(root)) {
      return;
    }

    observedRoots.add(root);
    const observer = new MutationObserver(scheduleApply);
    observer.observe(root, {
      childList: true,
      subtree: true,
    });
    observers.push(observer);
  };

  const visit = (root: Node & ParentNode) => {
    observeRoot(root);

    root.querySelectorAll('video').forEach((video) => {
      video.style.setProperty('object-fit', 'fill', 'important');
      video.style.setProperty('width', '100%', 'important');
      video.style.setProperty('height', '100%', 'important');
      video.muted = muted;
      video.volume = Math.min(1, Math.max(0, volume));
    });

    root.querySelectorAll('*').forEach((element) => {
      const shadowRoot = element.shadowRoot;
      if (shadowRoot) {
        visit(shadowRoot);
      }
    });
  };

  function apply() {
    if (!active) {
      return;
    }

    visit(rootElement);
  }

  apply();
  cleanupTimer = window.setInterval(scheduleApply, 500);

  return () => {
    active = false;
    window.cancelAnimationFrame(rafId);
    window.clearInterval(cleanupTimer);
    observers.forEach((observer) => observer.disconnect());
  };
}

function preferFocusedCameraState(stateObj: HomeAssistantState, profile: CameraStreamProfile): HomeAssistantState {
  const focusedProfile = resolveFocusedVideoProfile(stateObj.attributes, profile);
  const bridgeProfile = readBridgeProfile(stateObj.attributes, focusedProfile);
  const streamSource = readString(bridgeProfile?.stream_url) || readString(bridgeProfile?.local_stream_url);

  return {
    ...stateObj,
    attributes: {
      ...stateObj.attributes,
      preferred_video_profile: focusedProfile,
      recommended_profile: focusedProfile,
      ...(streamSource ? { stream_source: streamSource } : {}),
    },
  };
}

function resolveFocusedVideoProfile(attributes: Record<string, unknown>, profile: CameraStreamProfile): string {
  const profiles = readRecord(attributes.bridge_profiles);
  if (profile === 'sub') {
    if (profiles?.stable) {
      return 'stable';
    }
    if (profiles?.sub) {
      return 'sub';
    }
  }
  if (profiles?.quality) {
    return 'quality';
  }
  if (profiles?.main) {
    return 'main';
  }

  const preferred = readString(attributes.preferred_video_profile) || readString(attributes.recommended_profile);
  if (preferred && !/stable|sub|low|sd/i.test(preferred)) {
    return preferred;
  }

  return 'quality';
}

function countRoomEvents(events: EventItem[], types: string[]): number {
  const wanted = new Set(types);
  return events.filter((event) => wanted.has(event.type)).length;
}

function readBridgeProfile(attributes: Record<string, unknown>, profileKey: string): Record<string, unknown> | null {
  const profiles = readRecord(attributes.bridge_profiles);
  return readRecord(profiles?.[profileKey]);
}

function readRecord(value: unknown): Record<string, unknown> | null {
  return value && typeof value === 'object' && !Array.isArray(value) ? value as Record<string, unknown> : null;
}

function readString(value: unknown): string {
  return typeof value === 'string' ? value.trim() : '';
}

function buildFallbackCameraMedia(device: Device | null) {
  if (!device || device.type !== 'camera' || !device.entityId.startsWith('camera.')) {
    return undefined;
  }

  return {
    entityId: device.entityId,
    title: device.name,
    kind: 'stream' as const,
    src: `/api/camera_proxy_stream/${device.entityId}`,
    posterSrc: `/api/camera_proxy/${device.entityId}`,
  };
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
