import type { CameraStreamProfile, Device } from '../models/dashboard';
import { Icon } from '../components/Icon';
import { StatusBadge } from '../components/StatusBadge';
import { getToneClass } from '../utils/assets';

interface CameraAudioState {
  muted: boolean;
  volume: number;
}

interface CameraStripProps {
  cameras: Device[];
  selectedCameraId: string | null;
  selectedProfile: CameraStreamProfile;
  audioByCameraId: Record<string, CameraAudioState>;
  onPlayCamera: (cameraId: string, profile: CameraStreamProfile) => void;
  onChangeAudio: (cameraId: string, audio: CameraAudioState) => void;
}

export function CameraStrip({
  cameras,
  selectedCameraId,
  selectedProfile,
  audioByCameraId,
  onPlayCamera,
  onChangeAudio,
}: CameraStripProps) {
  return (
    <section className="camera-strip glass-panel">
      <div className="section-heading">
        <span>Cameras</span>
        <strong>{cameras.length}</strong>
      </div>
      <div className="camera-strip__list">
        {cameras.length > 0 ? (
          cameras.map((camera) => {
            const selected = camera.id === selectedCameraId;
            const audio = audioByCameraId[camera.id] ?? { muted: true, volume: 1 };

            return (
              <article
                key={camera.id}
                className={[
                  'camera-card',
                  getToneClass(camera.tone),
                  selected ? 'camera-card--selected' : '',
                ]
                  .filter(Boolean)
                  .join(' ')}
              >
                <div className="camera-card__identity">
                  <span className="camera-card__icon-wrap">
                    <Icon icon={{ category: 'devices', key: 'camera' }} size={34} />
                  </span>
                  <div className="camera-card__copy">
                    <strong>{camera.name}</strong>
                    <span>{camera.model || camera.status}</span>
                  </div>
                </div>
                <StatusBadge label={camera.isOnline ? 'Online' : 'Offline'} tone={camera.isOnline ? 'green' : 'red'} />
                <div className="camera-card__controls">
                  <button type="button" onClick={() => onPlayCamera(camera.id, 'main')} disabled={!camera.isOnline}>
                    {selected && selectedProfile === 'main' ? 'Main live' : 'Main'}
                  </button>
                  <button type="button" onClick={() => onPlayCamera(camera.id, 'sub')} disabled={!camera.isOnline}>
                    {selected && selectedProfile === 'sub' ? 'Sub live' : 'Sub'}
                  </button>
                  <div className="camera-card__audio">
                    <button
                      type="button"
                      aria-label={audio.muted ? 'Unmute camera audio' : 'Mute camera audio'}
                      onClick={() => onChangeAudio(camera.id, { ...audio, muted: !audio.muted })}
                    >
                      {audio.muted ? 'Muted' : 'Audio'}
                    </button>
                    <input
                      type="range"
                      min="0"
                      max="1"
                      step="0.05"
                      value={audio.volume}
                      aria-label="Camera audio volume"
                      onChange={(event) => onChangeAudio(camera.id, { muted: false, volume: Number(event.currentTarget.value) })}
                    />
                  </div>
                </div>
              </article>
            );
          })
        ) : (
          <div className="camera-strip__empty">
            <strong>No cameras in this room</strong>
            <span>Assign Dahua or Home Assistant camera entities to this area.</span>
          </div>
        )}
      </div>
    </section>
  );
}
