SMD/IVS counts by Home Assistant room
=====================================

Goal
----

A separate TypeScript/JavaScript Lovelace card can show SMD/IVS counters per
Home Assistant room by combining two data sources:

1. Home Assistant entity/device/area registries: tells which HA room (area) a
   DahuaBridge camera/channel belongs to.
2. DahuaBridge event summary endpoint: returns indexed SMD/IVS counts per NVR
   channel for a time window.

Use "room" as Home Assistant "area". The DahuaBridge card code calls this
value roomLabel, but it is resolved from HA areas.


Required DahuaBridge camera attributes
--------------------------------------

The Home Assistant integration exposes these attributes on camera entities:

- bridge_base_url: base URL for bridge HTTP API.
- bridge_device_id: DahuaBridge device id for this camera entity.
- bridge_root_device_id: root NVR id for an NVR channel.
- bridge_device_kind: "nvr_channel", "ipc", or "vto".
- bridge_device_name: display name from the bridge catalog.
- bridge_channel: positive NVR channel number.
- bridge_archive_smd_ivs_url_template: per-channel SMD/IVS event-list URL.

For counters, prefer bridge_root_device_id + bridge_channel and call the event
summary endpoint. Do not count rows from bridge_archive_smd_ivs_url_template
unless you specifically need event rows; that is heavier and paginated.

Only NVR channels have archive SMD/IVS summary counts. Standalone IPC cameras
can still be assigned to a room, but they should be skipped for this counter
unless a future bridge API exposes IPC archive summaries.


Bridge counter endpoint
-----------------------

Call once per NVR:

GET {bridge_base_url}/api/v1/nvr/{bridge_root_device_id}/events/summary?start={iso}&end={iso}&event=all

Optional channel-filtered call:

GET {bridge_base_url}/api/v1/nvr/{bridge_root_device_id}/events/summary?channel={channel}&start={iso}&end={iso}&event=all

The endpoint reads the local SQLite archive index. It does not query the NVR
live. Timestamps can be ISO/RFC3339 strings. The existing cards use a rolling
24 hour window:

- end = new Date()
- start = new Date(end.getTime() - 24 * 60 * 60 * 1000)

Response shape:

{
  "device_id": "west20_nvr",
  "start_time": "2026-05-04T12:00:00Z",
  "end_time": "2026-05-05T12:00:00Z",
  "total_count": 12,
  "items": [
    {"code": "human", "label": "Human", "count": 5},
    {"code": "vehicle", "label": "Vehicle", "count": 3},
    {"code": "tripwire", "label": "Cross Line", "count": 4}
  ],
  "channels": [
    {
      "channel": 1,
      "total_count": 7,
      "items": [
        {"code": "human", "label": "Human", "count": 5},
        {"code": "tripwire", "label": "Cross Line", "count": 2}
      ]
    }
  ]
}

Normalized codes currently emitted by the summary are:

- human
- vehicle
- animal
- tripwire
- intrusion

Recommended display grouping:

- SMD humans: human
- SMD vehicles: vehicle
- SMD animals: animal
- IVS: tripwire + intrusion
- total: total_count


Resolving HA room/area
----------------------

Inside a Lovelace custom card, the card receives hass. Use:

- hass.states for entity states and attributes.
- hass.callWS(...) or hass.connection.sendMessagePromise(...) for registries.

Load registries:

const [entities, devices, areas] = await Promise.all([
  hass.callWS({type: "config/entity_registry/list"}),
  hass.callWS({type: "config/device_registry/list"}),
  hass.callWS({type: "config/area_registry/list"}),
]);

Build maps:

const entityById = new Map(entities.map((entry) => [entry.entity_id, entry]));
const deviceById = new Map(devices.map((entry) => [entry.id, entry]));
const areaById = new Map(areas.map((entry) => [entry.area_id, entry]));

Resolve an entity's room:

function areaNameForEntity(entityId) {
  const entityEntry = entityById.get(entityId);
  if (!entityEntry) return "Unassigned";

  let areaId = entityEntry.area_id || null;
  let deviceId = entityEntry.device_id || null;

  while (!areaId && deviceId) {
    const device = deviceById.get(deviceId);
    if (!device) break;
    areaId = device.area_id || null;
    deviceId = device.via_device_id || null;
  }

  return areaId ? (areaById.get(areaId)?.name || "Unassigned") : "Unassigned";
}

This is the same precedence used by the existing cards:

1. Entity area_id.
2. Device area_id.
3. Parent device area_id through via_device_id.
4. "Unassigned".


Discover DahuaBridge NVR channels
---------------------------------

Scan camera entities and keep only DahuaBridge NVR channel cameras:

function discoverBridgeChannels(hass) {
  const channels = [];

  for (const entity of Object.values(hass.states)) {
    if (!entity.entity_id.startsWith("camera.")) continue;

    const attrs = entity.attributes || {};
    const kind = String(attrs.bridge_device_kind || "").trim();
    const rootDeviceId = String(attrs.bridge_root_device_id || "").trim();
    const deviceId = String(attrs.bridge_device_id || "").trim();
    const bridgeBaseUrl = String(attrs.bridge_base_url || "").trim();
    const channel = Number(attrs.bridge_channel);

    if (kind !== "nvr_channel") continue;
    if (!rootDeviceId || !deviceId || !bridgeBaseUrl) continue;
    if (!Number.isFinite(channel) || channel <= 0) continue;

    channels.push({
      entityId: entity.entity_id,
      deviceId,
      rootDeviceId,
      channel,
      bridgeBaseUrl,
      room: areaNameForEntity(entity.entity_id),
      label:
        String(attrs.bridge_device_name || "").trim() ||
        String(attrs.friendly_name || "").trim() ||
        entity.entity_id,
    });
  }

  return channels;
}


Fetch summaries efficiently
---------------------------

Group discovered channels by rootDeviceId + bridgeBaseUrl and make one summary
request per NVR. Then merge the returned channel counts back into the HA room
groups.

function summaryUrl(baseUrl, rootDeviceId) {
  const base = String(baseUrl || "").replace(/\/+$/, "");
  return `${base}/api/v1/nvr/${encodeURIComponent(rootDeviceId)}/events/summary`;
}

async function fetchSummary(baseUrl, rootDeviceId, start, end, signal) {
  const url = new URL(summaryUrl(baseUrl, rootDeviceId));
  url.searchParams.set("start", start.toISOString());
  url.searchParams.set("end", end.toISOString());
  url.searchParams.set("event", "all");

  const response = await fetch(url, {
    method: "GET",
    headers: {Accept: "application/json"},
    signal,
  });
  if (!response.ok) {
    throw new Error(`DahuaBridge summary failed: HTTP ${response.status}`);
  }
  return response.json();
}

function emptyCounts() {
  return {total: 0, human: 0, vehicle: 0, animal: 0, ivs: 0};
}

function addSummaryItems(target, items) {
  for (const item of items || []) {
    const code = String(item.code || "").trim().toLowerCase();
    const count = Number(item.count) || 0;

    if (code === "human") target.human += count;
    else if (code === "vehicle") target.vehicle += count;
    else if (code === "animal") target.animal += count;
    else if (code === "tripwire" || code === "intrusion") target.ivs += count;
  }
}

async function loadSmdIvsCountsByRoom(hass, signal) {
  const channels = discoverBridgeChannels(hass);
  const end = new Date();
  const start = new Date(end.getTime() - 24 * 60 * 60 * 1000);

  const nvrGroups = new Map();
  for (const camera of channels) {
    const key = `${camera.bridgeBaseUrl}|${camera.rootDeviceId}`;
    const group = nvrGroups.get(key) || {
      bridgeBaseUrl: camera.bridgeBaseUrl,
      rootDeviceId: camera.rootDeviceId,
      cameras: [],
    };
    group.cameras.push(camera);
    nvrGroups.set(key, group);
  }

  const roomCounts = new Map();

  await Promise.all([...nvrGroups.values()].map(async (group) => {
    const summary = await fetchSummary(
      group.bridgeBaseUrl,
      group.rootDeviceId,
      start,
      end,
      signal,
    );

    const summaryByChannel = new Map(
      (summary.channels || []).map((entry) => [Number(entry.channel), entry]),
    );

    for (const camera of group.cameras) {
      const channelSummary = summaryByChannel.get(camera.channel);
      if (!channelSummary) continue;

      const counts = roomCounts.get(camera.room) || emptyCounts();
      counts.total += Number(channelSummary.total_count) || 0;
      addSummaryItems(counts, channelSummary.items);
      roomCounts.set(camera.room, counts);
    }
  }));

  return [...roomCounts.entries()].map(([room, counts]) => ({
    room,
    ...counts,
  })).sort((left, right) => left.room.localeCompare(right.room));
}


Card lifecycle recommendations
------------------------------

- Fetch the registry snapshot once and cache it for about 5 minutes.
- Refresh SMD/IVS summaries every 60 seconds for live dashboard counters.
- Use AbortController and abort the previous request when hass/config changes or
  the card disconnects.
- If browser_bridge_url is supported by the new card, rewrite bridge_base_url
  before building summary URLs, same as the existing cards do.
- Prefer one summary request per NVR, not one request per camera, when rendering
  room totals.
- If a camera has no HA area, put it in "Unassigned" instead of dropping it.
- If a summary request fails for one NVR, keep the other NVR room counters and
  mark that NVR as stale/error in card state.


Minimal rendering data model
----------------------------

The card can keep state like this:

type RoomSmdIvsCounts = {
  room: string;
  total: number;
  human: number;
  vehicle: number;
  animal: number;
  ivs: number;
};

Render one row/card per room:

- room name
- total SMD/IVS events in the current window
- human count
- vehicle count
- animal count, if nonzero or if you want a fixed layout
- IVS count

For drill-down, keep the discovered camera list and the per-channel summary so a
room row can expand into its assigned cameras/channels.
