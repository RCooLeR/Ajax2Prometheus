DahuaBridge camera/channel playback instructions
================================================

Scope
-----

This document explains how a Home Assistant TypeScript/JavaScript card can play
a DahuaBridge camera/channel stream.

DahuaBridge creates Home Assistant camera entities for channels. It does not
create one Home Assistant media_player entity per camera/channel. A card should
use the camera entity and either:

  1. render the Home Assistant native camera stream player, or
  2. render its own browser media player from DahuaBridge media URLs.

An external Home Assistant media_player entity, such as a TV or speaker, is a
separate target configured by the HA user. DahuaBridge only supplies the camera
stream, snapshots, and archive playback URLs.


Discover the camera/channel
---------------------------

Start from DahuaBridge camera entities.

Default channel camera entity id:

  camera.<device_id>_camera

Example:

  camera.west20_nvr_channel_05_camera

Do not rely only on the default entity id. HA users can rename entities. Prefer
state attributes and the entity/device registry.

Useful camera attributes:

  bridge_device_kind
  bridge_device_id
  bridge_root_device_id
  bridge_channel
  bridge_base_url
  stream_source
  snapshot_url
  preview_url
  stream_available
  recommended_profile
  preferred_video_profile
  preferred_video_source
  video_fallbacks_enabled
  bridge_profiles
  bridge_playback_sessions_url
  bridge_archive_smd_ivs_url_template
  bridge_archive_recording_chunks_url_template

For an NVR channel, expect:

  bridge_device_kind = nvr_channel
  bridge_device_id = <nvr_device_id>_channel_XX
  bridge_root_device_id = <nvr_device_id>
  bridge_channel = positive channel number

The stream id used by bridge media endpoints is normally bridge_device_id for an
NVR channel.


Live playback choices
---------------------

A card can offer these live player modes:

  native
  hls
  dash
  mjpeg
  webrtc
  snapshot
  preview

Recommended order for a normal card:

  1. Use the user/configured preferred source when it is available.
  2. If the camera recommends native HA/ONVIF/RTSP, try native.
  3. Otherwise try HLS, then DASH, then MJPEG.
  4. Use snapshot as a non-streaming fallback.

For overview grids, prefer lower bandwidth:

  1. stable profile
  2. HLS or DASH
  3. snapshot fallback

For a single focused camera, prefer quality:

  1. quality profile
  2. native when recommended and available
  3. HLS, DASH, MJPEG fallback


Native Home Assistant player
----------------------------

Use Home Assistant's camera stream element when you want HA to own the stream
pipeline.

Lit example:

  html`
    <ha-camera-stream
      .hass=${hass}
      .stateObj=${cameraState}
      data-audio-muted="true"
      data-audio-volume="1"
    ></ha-camera-stream>
  `

Where:

  cameraState = hass.states["camera.<device_id>_camera"]

Native playback uses the camera entity's stream_source. The integration resolves
the source through bridge settings and HA's stream support.

Use native mode when:

  recommended_ha_integration from catalog/sensor data is onvif
  preferred_video_source is native, ha, homeassistant, onvif, rtsp, or direct_rtsp
  the card wants Home Assistant to transcode/proxy RTSP

Native mode is also the right path if the browser cannot play the bridge HLS/DASH
URLs directly but HA can stream the camera.


Bridge profile data
-------------------

The camera attribute bridge_profiles is a map keyed by profile name. The bridge
currently publishes these profile names:

  quality
  stable

Each profile can contain resolved URLs like:

  stream_url
  local_mjpeg_url
  local_hls_url
  local_dash_url
  local_webrtc_url

Use recommended_profile or preferred_video_profile when it is present. Fall back
to quality for focused playback and stable for overview playback.

Example profile lookup:

  const attrs = cameraState.attributes;
  const profiles = attrs.bridge_profiles ?? {};
  const profileKey =
    attrs.preferred_video_profile ||
    attrs.recommended_profile ||
    "stable";
  const profile = profiles[profileKey] ?? profiles.stable ?? profiles.quality;


Bridge live media endpoints
---------------------------

If bridge_profiles is missing or a card needs to build URLs manually, use:

  GET /api/v1/media/snapshot/{stream_id}?profile={profile}
  GET /api/v1/media/preview/{stream_id}?profile={profile}
  GET /api/v1/media/mjpeg/{stream_id}?profile={profile}
  GET /api/v1/media/hls/{stream_id}/{profile}/index.m3u8
  GET /api/v1/media/dash/{stream_id}/{profile}/manifest.mpd
  GET /api/v1/media/webrtc/{stream_id}/{profile}
  POST /api/v1/media/webrtc/{stream_id}/{profile}/offer

Build absolute URLs from bridge_base_url:

  const url = `${bridgeBaseUrl}/api/v1/media/hls/${encodeURIComponent(streamId)}/${profile}/index.m3u8`;

Prefer URLs from bridge_profiles when available because they are already resolved
for the bridge/HA public base URL configuration.


HLS player
----------

HLS is the best browser-first live player for most dashboards.

Use:

  profile.local_hls_url

or:

  /api/v1/media/hls/{stream_id}/{profile}/index.m3u8

Browser behavior:

  Safari can usually play HLS directly with video.src.
  Chromium/Firefox usually need hls.js.

Minimal logic:

  const video = document.createElement("video");
  video.controls = true;
  video.autoplay = true;
  video.muted = true;
  video.playsInline = true;

  if (video.canPlayType("application/vnd.apple.mpegurl")) {
    video.src = hlsUrl;
  } else {
    const hls = new Hls();
    hls.loadSource(hlsUrl);
    hls.attachMedia(video);
  }

  await video.play().catch(() => undefined);


DASH player
-----------

DASH is the second browser-first choice.

Use:

  profile.local_dash_url

or:

  /api/v1/media/dash/{stream_id}/{profile}/manifest.mpd

Use dash.js:

  const video = document.createElement("video");
  video.controls = true;
  video.autoplay = true;
  video.muted = true;
  video.playsInline = true;

  const player = dashjs.MediaPlayer().create();
  player.initialize(video, dashUrl, true);


MJPEG player
------------

MJPEG is simple and useful as a fallback, but it is bandwidth-heavy and has no
normal audio.

Use:

  profile.local_mjpeg_url

or:

  /api/v1/media/mjpeg/{stream_id}?profile={profile}

Render with an image element:

  const img = document.createElement("img");
  img.src = mjpegUrl;
  img.alt = cameraLabel;

For snapshots, use:

  snapshot_url

or:

  /api/v1/media/snapshot/{stream_id}?profile={profile}&width={pixels}

Add a cache buster when refreshing snapshots:

  img.src = `${snapshotUrl}${snapshotUrl.includes("?") ? "&" : "?"}_=${Date.now()}`;


WebRTC player
-------------

The bridge exposes a helper page and an offer endpoint.

Simple helper-page mode:

  const iframe = document.createElement("iframe");
  iframe.src = profile.local_webrtc_url;

Manual WebRTC mode:

  1. Create RTCPeerConnection.
  2. Create an offer.
  3. POST the offer to /api/v1/media/webrtc/{stream_id}/{profile}/offer.
  4. Set the returned SDP answer as the remote description.
  5. Attach remote tracks to a video element.

Use WebRTC only when the card is ready to manage WebRTC lifecycle, ICE failures,
and browser autoplay rules. HLS/DASH are simpler for a Lovelace card.


Preview page
------------

The bridge preview page is useful for diagnostics or a quick embedded view.

Use:

  preview_url

or:

  /api/v1/media/preview/{stream_id}?profile={profile}

Render it in an iframe:

  const iframe = document.createElement("iframe");
  iframe.src = previewUrl;

Do not use preview iframe as the main polished player if the card can render HLS
or native HA playback directly.


Archive playback with bridge playback sessions
----------------------------------------------

For archive rows or selected time ranges, a card can ask the bridge to create a
finite playback session.

Endpoint:

  POST /api/v1/nvr/{root_device_id}/playback/sessions

Request body:

  {
    "channel": 1,
    "start_time": "2026-05-03T14:11:22Z",
    "end_time": "2026-05-03T14:11:48Z",
    "seek_time": "2026-05-03T14:11:22Z",
    "file_path": "optional original DAV path",
    "source": "optional source from archive row",
    "type": "optional type from archive row",
    "video_stream": "optional video stream from archive row"
  }

The response includes:

  id
  stream_id
  channel
  start_time
  end_time
  seek_time
  recommended_profile
  snapshot_url
  profiles

Playback session profile fields:

  hls_url
  dash_url
  mjpeg_url
  webrtc_offer_url

Recommended archive player order:

  1. HLS
  2. DASH
  3. MJPEG

Example:

  const response = await fetch(playbackSessionsUrl, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Accept: "application/json",
    },
    body: JSON.stringify({
      channel,
      start_time: startIso,
      end_time: endIso,
      seek_time: seekIso,
    }),
  });

  const session = await response.json();
  const profileKey = session.recommended_profile || "quality";
  const profile = session.profiles[profileKey] ?? Object.values(session.profiles)[0];
  const playUrl = profile.hls_url ?? profile.dash_url ?? profile.mjpeg_url;

Use the selected session URL in the same HLS/DASH/MJPEG player described above.


Seeking an archive playback session
-----------------------------------

Existing playback sessions can be recreated/seeked.

Endpoint:

  POST /api/v1/nvr/playback/sessions/{session_id}/seek

Request body:

  {
    "seek_time": "2026-05-03T14:12:00Z"
  }

The response has the same shape as the create-session response. Replace the
current player source with the new profile URL.


Native archive playback through HA camera stream
------------------------------------------------

The integration also supports native archive playback by temporarily replacing
the camera entity stream source with a Dahua RTSP playback URL.

Service:

  dahuabridge.set_native_playback_source

Target:

  entity_id: camera.<device_id>_camera

Data:

  stream_source: rtsp://user:pass@host:554/cam/playback?channel=1&subtype=0&starttime=2026_05_03_14_11_22&endtime=2026_05_03_14_11_48

Then render the same camera entity with <ha-camera-stream>.

Clear native archive playback when the user exits archive mode:

  dahuabridge.clear_native_playback_source

Target:

  entity_id: camera.<device_id>_camera

The service only accepts rtsp:// URLs. It resets the cached HA stream so HA picks
up the archive source. Live playback returns after clear_native_playback_source.

Use native archive playback only when the integration was configured to include
RTSP credentials or the card has another valid RTSP playback URL. Normal cards
should prefer bridge playback sessions because the session response provides
browser-playable HLS/DASH/MJPEG URLs.


Home Assistant timeframe proxy
------------------------------

The integration exposes a Home Assistant camera proxy timeframe route:

  /api/camera_proxy/{entity_id}/timeframe?starttime={time}&endtime={time}&seektime={time}&profile={profile}
  /api/camera_proxy/{entity_id}/timeframe/snapshot?starttime={time}&endtime={time}&seektime={time}&profile={profile}

This route creates a bridge playback session internally and proxies MJPEG or a
snapshot through Home Assistant.

Time query values can use Dahua-style:

  2026_05_03_14_11_22

Use this when the card wants to stay inside the HA camera proxy path. For a rich
custom player, direct bridge playback sessions provide more control.


Bridge MP4 clips
----------------

For short saved clips, the bridge media layer supports MP4 recording.

Start a live clip:

  POST /api/v1/media/streams/{stream_id}/recordings

Request body:

  {
    "profile": "quality",
    "duration_seconds": 30
  }

Response fields include playback/download URLs.

List clips:

  GET /api/v1/media/recordings

Play completed clip:

  GET /api/v1/media/recordings/{clip_id}/play

Download completed clip:

  GET /api/v1/media/recordings/{clip_id}/download

Stop active clip:

  POST /api/v1/media/recordings/{clip_id}/stop

Use a normal HTML video element for MP4 playback:

  const video = document.createElement("video");
  video.controls = true;
  video.src = clip.playback_url;


Home Assistant camera recording services
----------------------------------------

The camera entity exposes DahuaBridge services for bridge-owned MP4 clips:

  dahuabridge.start_recording
  dahuabridge.stop_recording

Start recording:

  await hass.callService(
    "dahuabridge",
    "start_recording",
    {
      profile: "quality",
      duration_seconds: 30,
    },
    {
      entity_id: cameraEntityId,
    },
  );

Stop recording:

  await hass.callService(
    "dahuabridge",
    "stop_recording",
    {},
    {
      entity_id: cameraEntityId,
    },
  );

Camera attributes can expose:

  bridge_recording_active
  bridge_start_recording_url
  bridge_stop_recording_url
  bridge_recordings_url


Playing on an external HA media_player
--------------------------------------

DahuaBridge does not create the external media_player. If a user has a target
such as media_player.living_room_tv, the card should treat it as a separate HA
entity selected by the user.

For live camera casting, prefer Home Assistant's camera streaming/casting path
using the DahuaBridge camera entity. The card should pass the camera entity id,
not a raw bridge URL, when it wants HA to manage the stream.

For MP4 clips, use the bridge clip playback URL as a media URL only if the target
player can reach bridge_base_url from its network. Otherwise, use HA-native
camera streaming or an HA-accessible proxy URL.


Authentication and reachability
-------------------------------

Cards run in the user's browser. Any direct bridge media URL must be reachable
from that browser.

Use camera attributes first because the integration resolves URLs with the
configured bridge and HA public base URL settings.

Avoid include_credentials=true in normal browser calls. It can expose RTSP URLs
with embedded credentials and is intended for diagnostics or native RTSP mode.

For direct bridge fetch/video/img requests, expect normal browser networking
rules: CORS, mixed-content blocking, HTTPS certificates, and LAN reachability all
matter.


Minimal card algorithm
----------------------

1. Discover a DahuaBridge camera entity.

     const cameraState = hass.states[entityId];
     const attrs = cameraState.attributes;
     if (attrs.bridge_device_kind !== "nvr_channel") return;

2. Resolve channel identity.

     const deviceId = attrs.bridge_device_id;
     const rootDeviceId = attrs.bridge_root_device_id;
     const channel = Number(attrs.bridge_channel);
     const bridgeBaseUrl = attrs.bridge_base_url;

3. Resolve profile.

     const profiles = attrs.bridge_profiles ?? {};
     const profileKey =
       attrs.preferred_video_profile ||
       attrs.recommended_profile ||
       "stable";
     const profile =
       profiles[profileKey] ||
       profiles.stable ||
       profiles.quality ||
       null;

4. Pick source.

     native if HA camera stream is preferred and cameraState exists
     else profile.local_hls_url
     else profile.local_dash_url
     else profile.local_mjpeg_url
     else attrs.snapshot_url

5. Render.

     native -> <ha-camera-stream>
     hls    -> <video> with hls.js fallback
     dash   -> <video> with dash.js
     mjpeg  -> <img>
     image  -> <img> snapshot

6. On archive click, create a playback session from
   bridge_playback_sessions_url and play the returned HLS/DASH/MJPEG URL.

7. Stop media cleanly when the card changes camera, changes mode, or disconnects.

     destroy hls.js/dash.js players
     clear video src
     stop WebRTC peer connection
     call clear_native_playback_source after native archive playback
