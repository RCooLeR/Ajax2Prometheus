const AJAX_EMBEDDED_DEVICE_CATALOG = __DEVICE_CATALOG__;

const ROOM_META = {
  "Central units": { icon: "mdi:router-wireless", accent: "#63d1ff", theme: "infrastructure" },
  "Fire and life safety": { icon: "mdi:fire-alert", accent: "#ff6b4a", theme: "fire" },
  "Intrusion protection": { icon: "mdi:shield-home", accent: "#ffb347", theme: "intrusion" },
  "Video surveillance": { icon: "mdi:cctv", accent: "#7dd3fc", theme: "video" },
  "Automation": { icon: "mdi:lightning-bolt", accent: "#9ae66e", theme: "automation" },
  "Controls and panic buttons": { icon: "mdi:gesture-tap-button", accent: "#fda4af", theme: "control" },
  "Integrations and modules": { icon: "mdi:puzzle", accent: "#c4b5fd", theme: "integration" },
  "Sirens": { icon: "mdi:bullhorn", accent: "#fb7185", theme: "siren" },
  "Accessories": { icon: "mdi:package-variant-closed", accent: "#cbd5e1", theme: "support" },
  "Flood protection": { icon: "mdi:water-alert", accent: "#5eead4", theme: "flood" }
};

const SIGNAL_META = {
  fire: { label: "Fire", icon: "mdi:fire-alert", tone: "critical", rank: 100 },
  smoke: { label: "Smoke", icon: "mdi:smoke-detector-alert", tone: "critical", rank: 96 },
  water_leak: { label: "Leak", icon: "mdi:water-alert", tone: "critical", rank: 95 },
  leak: { label: "Leak", icon: "mdi:water-alert", tone: "critical", rank: 95 },
  flood: { label: "Flood", icon: "mdi:water-alert", tone: "critical", rank: 95 },
  co: { label: "CO", icon: "mdi:molecule-co", tone: "critical", rank: 94 },
  gas: { label: "Gas", icon: "mdi:gas-cylinder", tone: "critical", rank: 94 },
  gas_or_co: { label: "Gas / CO", icon: "mdi:chemical-weapon", tone: "critical", rank: 93 },
  temperature: { label: "Heat", icon: "mdi:thermometer-alert", tone: "warning", rank: 78 },
  burglary: { label: "Burglary", icon: "mdi:shield-home-alert", tone: "critical", rank: 92 },
  panic: { label: "Panic", icon: "mdi:alert-octagon", tone: "critical", rank: 91 },
  duress: { label: "Duress", icon: "mdi:hand-back-right-off", tone: "critical", rank: 91 },
  emergency: { label: "Emergency", icon: "mdi:alarm-light", tone: "critical", rank: 91 },
  medical: { label: "Medical", icon: "mdi:medical-bag", tone: "critical", rank: 91 },
  alarm: { label: "Alarm", icon: "mdi:alarm-light", tone: "critical", rank: 90 },
  tamper: { label: "Tamper", icon: "mdi:shield-alert", tone: "warning", rank: 72 },
  tamper_bypass: { label: "Tamper bypass", icon: "mdi:shield-off", tone: "warning", rank: 50 },
  bypass: { label: "Bypass", icon: "mdi:shield-off-outline", tone: "warning", rank: 48 },
  connectivity: { label: "Connection lost", icon: "mdi:wifi-alert", tone: "warning", rank: 74 },
  battery: { label: "Battery low", icon: "mdi:battery-alert-variant-outline", tone: "warning", rank: 62 },
  power: { label: "Power issue", icon: "mdi:power-plug-off", tone: "warning", rank: 61 },
  interference: { label: "Interference", icon: "mdi:signal-distance-variant", tone: "warning", rank: 60 },
  hardware: { label: "Hardware", icon: "mdi:chip", tone: "warning", rank: 59 },
  configuration: { label: "Config", icon: "mdi:cog-alert", tone: "warning", rank: 45 },
  firmware: { label: "Firmware", icon: "mdi:update", tone: "warning", rank: 44 },
  supervision: { label: "Supervision", icon: "mdi:timeline-alert", tone: "warning", rank: 43 },
  accelerometer: { label: "Shock", icon: "mdi:vibrate", tone: "warning", rank: 73 },
  hold_up: { label: "Hold-up", icon: "mdi:alarm-panel", tone: "critical", rank: 91 },
  button: { label: "Button", icon: "mdi:radiobox-marked", tone: "warning", rank: 46 }
};

const KIND_ICON_HINTS = [
  ["hub", "mdi:router-wireless"],
  ["rex", "mdi:access-point-network"],
  ["fireprotect", "mdi:fire-alert"],
  ["manualcallpoint", "mdi:alarm-light"],
  ["motioncam", "mdi:cctv"],
  ["motionprotect", "mdi:walk"],
  ["doorprotect", "mdi:door-closed-lock"],
  ["glassprotect", "mdi:window-closed-variant"],
  ["curtain", "mdi:blinds"],
  ["cam", "mdi:cctv"],
  ["nvr", "mdi:video-input-component"],
  ["socket", "mdi:power-socket-eu"],
  ["outlet", "mdi:power-plug"],
  ["relay", "mdi:electric-switch"],
  ["wallswitch", "mdi:light-switch"],
  ["lightswitch", "mdi:light-switch"],
  ["button", "mdi:gesture-tap-button"],
  ["keypad", "mdi:dialpad"],
  ["siren", "mdi:bullhorn"],
  ["street", "mdi:bullhorn-outline"],
  ["waterstop", "mdi:water-pump"],
  ["leaksprotect", "mdi:water-alert"],
  ["lifequality", "mdi:home-thermometer"],
  ["transmitter", "mdi:radio-tower"],
  ["ocbridge", "mdi:lan-connect"],
  ["uartbridge", "mdi:serial-port"],
  ["multitransmitter", "mdi:expansion-card"],
  ["junctionbox", "mdi:package-variant"],
  ["speakerphone", "mdi:phone-in-talk"]
];

const DEFAULT_ROOM_ORDER = [];

const CYRILLIC_MAP = {
  а: "a",
  б: "b",
  в: "v",
  г: "g",
  ґ: "g",
  д: "d",
  е: "e",
  є: "ie",
  ж: "zh",
  з: "z",
  и: "i",
  і: "i",
  ї: "i",
  й: "i",
  к: "k",
  л: "l",
  м: "m",
  н: "n",
  о: "o",
  п: "p",
  р: "r",
  с: "s",
  т: "t",
  у: "u",
  ф: "f",
  х: "kh",
  ц: "ts",
  ч: "ch",
  ш: "sh",
  щ: "shch",
  ь: "",
  ю: "iu",
  я: "ia",
  ё: "io",
  ы: "i",
  э: "e",
  ъ: "",
  "'": "",
  "’": "",
  "`": ""
};

const NAME_STYLE_BASE_ENTITY_SUFFIXES = {
  alarm: { domain: "binary_sensor", suffixes: ["alarm_active"] },
  tamper: { domain: "binary_sensor", suffixes: ["tamper_active", "tamper"] },
  trouble: { domain: "binary_sensor", suffixes: ["trouble_active"] },
  lastEventName: { domain: "sensor", suffixes: ["last_event", "last_event_name"] },
  lastEventAt: { domain: "sensor", suffixes: ["last_event_time", "last_event_at"] },
  lastSignal: { domain: "sensor", suffixes: ["last_signal"] },
  alarmSignal: { domain: "sensor", suffixes: ["alarm_signal"] },
  alarmAction: { domain: "sensor", suffixes: ["alarm_action"] }
};

const ZONE_STYLE_BASE_ENTITY_SUFFIXES = {
  alarm: { domain: "binary_sensor", suffixes: ["alarm_active"] },
  tamper: { domain: "binary_sensor", suffixes: ["tamper_active"] },
  trouble: { domain: "binary_sensor", suffixes: ["trouble_active"] },
  lastEventName: { domain: "sensor", suffixes: ["last_event_name"] },
  lastEventAt: { domain: "sensor", suffixes: ["last_event_at"] },
  lastSignal: { domain: "sensor", suffixes: ["last_signal"] },
  alarmSignal: { domain: "sensor", suffixes: ["alarm_signal"] },
  alarmAction: { domain: "sensor", suffixes: ["alarm_action"] }
};

const NAME_STYLE_SIGNAL_SUFFIXES = {
  fire: ["fire"],
  smoke: ["smoke"],
  temperature: ["temperature"],
  co: ["carbon_monoxide", "co"],
  gas: ["gas"],
  gas_or_co: ["carbon_monoxide", "gas_or_co"],
  fire_detector: ["fire_detector_fault", "fire_detector"],
  tamper: ["tamper"],
  battery: ["battery_low", "battery"],
  connectivity: ["connection_lost", "connectivity"],
  bypass: ["bypass"],
  tamper_bypass: ["tamper_bypass"],
  power: ["power_failure", "power"],
  hardware: ["hardware_fault", "hardware"],
  firmware: ["firmware"],
  interference: ["interference"],
  configuration: ["configuration"],
  burglary: ["burglary"],
  water_leak: ["water_leak"],
  leak: ["water_leak", "leak"],
  flood: ["water_leak", "flood"],
  supervision: ["supervision"],
  panic: ["panic"],
  duress: ["duress"],
  emergency: ["emergency"],
  medical: ["medical"],
  hold_up: ["hold_up", "panic"],
  night_mode: ["night_mode"],
  arming: ["arming"],
  accelerometer: ["accelerometer"],
  button: ["button"]
};

const ZONE_STYLE_SIGNAL_SUFFIXES = new Proxy({}, {
  get(_target, key) {
    return [`signal_${slugPart(key)}`];
  }
});

const DASHBOARD_BASE_WIDTH = 1920;
const DASHBOARD_BASE_HEIGHT = 1080;

function transliterate(value) {
  return String(value || "")
    .normalize("NFKD")
    .split("")
    .map((char) => {
      const lower = char.toLowerCase();
      if (CYRILLIC_MAP[lower] !== undefined) {
        return CYRILLIC_MAP[lower];
      }
      return lower;
    })
    .join("");
}

function slugPart(value) {
  return String(value || "")
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "_")
    .replace(/^_+|_+$/g, "");
}

function entitySlug(value) {
  return slugPart(transliterate(value));
}

function uniqueValues(list) {
  return Array.from(new Set(list.filter((value) => value !== null && value !== undefined && value !== "")));
}

function escapeHtml(value) {
  return String(value ?? "")
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#39;");
}

function humanizeKey(value) {
  return String(value || "")
    .replace(/_/g, " ")
    .replace(/\b\w/g, (match) => match.toUpperCase());
}

function formatShortTimestamp(value) {
  if (!value || value === "unknown" || value === "unavailable") {
    return "No timestamp";
  }

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return String(value);
  }

  return new Intl.DateTimeFormat(undefined, {
    month: "short",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit"
  }).format(date);
}

function formatRelative(value) {
  if (!value || value === "unknown" || value === "unavailable") {
    return "No recent event";
  }

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "No recent event";
  }

  const deltaSeconds = Math.round((date.getTime() - Date.now()) / 1000);
  const units = [
    ["day", 86400],
    ["hour", 3600],
    ["minute", 60],
    ["second", 1]
  ];

  for (const [unit, seconds] of units) {
    if (Math.abs(deltaSeconds) >= seconds || unit === "second") {
      const formatter = new Intl.RelativeTimeFormat(undefined, { numeric: "auto" });
      return formatter.format(Math.round(deltaSeconds / seconds), unit);
    }
  }

  return "Now";
}

function modeLabelFromAccount(accountState) {
  return accountState.armed
    ? "Armed"
    : accountState.partiallyArmed
      ? "Partially armed"
      : accountState.nightMode
        ? "Night mode"
        : "Disarmed";
}

function systemToneFromModel(model) {
  return model.totals.critical > 0 || model.accountState.alarmActive
    ? "critical"
    : model.totals.attention > 0 || model.accountState.troubleActive || model.accountState.tamperActive
      ? "warning"
      : "clear";
}

function commandStatusCopy(systemTone) {
  return systemTone === "critical"
    ? "Respond now"
    : systemTone === "warning"
      ? "Attention needed"
      : "Nominal";
}

function countBy(list, iteratee) {
  const map = new Map();
  for (const item of list) {
    const key = iteratee(item);
    map.set(key, (map.get(key) || 0) + 1);
  }
  return map;
}

function makeEntityId(domain, prefix, suffix) {
  return `${domain}.${prefix}_${suffix}`;
}

function firstAvailableState(hass, entityIds) {
  for (const entityId of entityIds) {
    const state = readState(hass, entityId);
    if (state) {
      return state;
    }
  }
  return null;
}

function baseEntityCandidates(source, key) {
  const mappings = source.style === "name" ? NAME_STYLE_BASE_ENTITY_SUFFIXES : ZONE_STYLE_BASE_ENTITY_SUFFIXES;
  const config = mappings[key];
  if (!config) {
    return [];
  }
  return config.suffixes.map((suffix) => makeEntityId(config.domain, source.prefix, suffix));
}

function signalEntityCandidates(source, signal) {
  const mappings = source.style === "name" ? NAME_STYLE_SIGNAL_SUFFIXES : ZONE_STYLE_SIGNAL_SUFFIXES;
  const suffixes = mappings[signal] || [slugPart(signal)];
  return suffixes.map((suffix) => makeEntityId("binary_sensor", source.prefix, suffix));
}

function scoreEntitySource(hass, source, device) {
  let score = 0;
  const anchors = [
    ...baseEntityCandidates(source, "lastSignal"),
    ...baseEntityCandidates(source, "lastEventName"),
    ...baseEntityCandidates(source, "lastEventAt"),
    ...baseEntityCandidates(source, "alarm"),
    ...baseEntityCandidates(source, "trouble")
  ];

  for (const entityId of anchors) {
    if (readState(hass, entityId)) {
      score += 4;
    }
  }

  for (const event of uniqueValues(device.events || [])) {
    if (firstAvailableState(hass, signalEntityCandidates(source, event))) {
      score += 1;
    }
  }

  return score;
}

function entitySourceCandidates(device, config) {
  const overridePrefix = config.entity_prefix_overrides?.[device.zone] || config.entity_prefix_overrides?.[device.name];
  const candidates = [];

  if (overridePrefix) {
    candidates.push({ prefix: slugPart(overridePrefix), style: "name" });
  }

  const nameSlug = entitySlug(device.name);
  if (nameSlug) {
    candidates.push({ prefix: nameSlug, style: "name" });
  }

  const rawNameSlug = slugPart(device.name);
  if (rawNameSlug && rawNameSlug !== nameSlug) {
    candidates.push({ prefix: rawNameSlug, style: "name" });
  }

  candidates.push({
    prefix: `zone_${slugPart(device.account)}_${slugPart(device.zone)}`,
    style: "zone"
  });

  return uniqueValues(candidates.map((candidate) => `${candidate.style}:${candidate.prefix}`)).map((key) => {
    const [style, prefix] = key.split(":");
    return { style, prefix };
  });
}

function resolveEntitySource(hass, device, config) {
  const candidates = entitySourceCandidates(device, config);
  let best = candidates[0];
  let bestScore = -1;

  for (const candidate of candidates) {
    const score = scoreEntitySource(hass, candidate, device);
    if (score > bestScore) {
      best = candidate;
      bestScore = score;
    }
  }

  return best;
}

function roomMeta(room) {
  if (ROOM_META[room]) {
    return ROOM_META[room];
  }

  const slug = entitySlug(room);
  if (slug.includes("budinok") || slug.includes("house")) {
    return { icon: "mdi:home-city", accent: "#7dd3fc", theme: "home" };
  }
  if (slug.includes("garazh") || slug.includes("garage")) {
    return { icon: "mdi:garage", accent: "#f59e0b", theme: "garage" };
  }
  if (slug.includes("kotel")) {
    return { icon: "mdi:boiler", accent: "#fb7185", theme: "boiler" };
  }
  if (slug.includes("gorish") || slug.includes("attic")) {
    return { icon: "mdi:home-roof", accent: "#c4b5fd", theme: "attic" };
  }
  if (slug.includes("khat")) {
    return { icon: "mdi:home-floor-1", accent: "#6ee7b7", theme: "annex" };
  }

  return { icon: "mdi:shield-home-outline", accent: "#7dd3fc", theme: "support" };
}

function iconForKind(kind, room, events) {
  const lowered = String(kind || "").toLowerCase();
  for (const [hint, icon] of KIND_ICON_HINTS) {
    if (lowered.includes(hint)) {
      return icon;
    }
  }

  const eventSet = new Set(events || []);
  if (eventSet.has("fire") || eventSet.has("smoke")) {
    return "mdi:fire-alert";
  }
  if (eventSet.has("burglary")) {
    return "mdi:shield-home-alert";
  }
  if (eventSet.has("water_leak")) {
    return "mdi:water-alert";
  }

  return roomMeta(room).icon;
}

function readState(hass, entityId) {
  return hass?.states?.[entityId];
}

function isOn(hass, entityId) {
  return readState(hass, entityId)?.state === "on";
}

function sortDevices(a, b) {
  if (b.severity !== a.severity) {
    return b.severity - a.severity;
  }
  if (b.lastEventAtUnix !== a.lastEventAtUnix) {
    return b.lastEventAtUnix - a.lastEventAtUnix;
  }
  return a.name.localeCompare(b.name);
}

function sortRooms(a, b) {
  if (b.critical !== a.critical) {
    return b.critical - a.critical;
  }
  if (b.warning !== a.warning) {
    return b.warning - a.warning;
  }
  if (b.offline !== a.offline) {
    return b.offline - a.offline;
  }
  return a.position - b.position;
}

function summarizeHeadline(device) {
  if (device.fireLike) {
    return "Fire response";
  }
  if (device.intrusionLike) {
    return "Intrusion response";
  }
  if (device.waterLike) {
    return "Flood response";
  }
  if (device.alarmActive) {
    return "Alarm active";
  }
  if (device.tamperActive) {
    return "Tamper active";
  }
  if (device.troubleActive) {
    return "Trouble active";
  }
  if (device.offline) {
    return "Connectivity issue";
  }
  if (device.batteryIssue) {
    return "Battery attention";
  }
  if (device.warningSignals.length > 0) {
    return SIGNAL_META[device.warningSignals[0]]?.label || humanizeKey(device.warningSignals[0]);
  }
  return "Nominal";
}

function buildDeviceModel(hass, device, config) {
  const source = resolveEntitySource(hass, device, config || {});
  const alarmState = firstAvailableState(hass, baseEntityCandidates(source, "alarm"));
  const tamperState = firstAvailableState(hass, baseEntityCandidates(source, "tamper"));
  const troubleState = firstAvailableState(hass, baseEntityCandidates(source, "trouble"));
  const lastEventNameState = firstAvailableState(hass, baseEntityCandidates(source, "lastEventName"));
  const lastEventAtState = firstAvailableState(hass, baseEntityCandidates(source, "lastEventAt"));
  const lastSignalState = firstAvailableState(hass, baseEntityCandidates(source, "lastSignal"));
  const alarmSignalState = firstAvailableState(hass, baseEntityCandidates(source, "alarmSignal"));
  const alarmActionState = firstAvailableState(hass, baseEntityCandidates(source, "alarmAction"));

  const alarmActive = alarmState?.state === "on";
  const tamperActive = tamperState?.state === "on";
  const troubleActive = troubleState?.state === "on";
  const lastEventName = lastEventNameState?.state || "Awaiting event";
  const lastEventAt = lastEventAtState?.state || "";
  const lastSignal = lastSignalState?.state || "idle";
  const alarmSignal = alarmSignalState?.state || "none";
  const alarmAction = alarmActionState?.state || "none";

  const activeSignals = [];
  const warningSignals = [];
  const eventKeys = uniqueValues(device.events || []);

  for (const signal of eventKeys) {
    if (firstAvailableState(hass, signalEntityCandidates(source, signal))?.state === "on") {
      activeSignals.push(signal);
      warningSignals.push(signal);
    }
  }

  if (alarmActive && alarmSignal && alarmSignal !== "none" && !activeSignals.includes(alarmSignal)) {
    activeSignals.unshift(alarmSignal);
  }

  const fireLike = activeSignals.some((signal) => ["fire", "smoke", "co", "gas", "gas_or_co"].includes(signal));
  const waterLike = activeSignals.some((signal) => ["water_leak", "leak", "flood"].includes(signal));
  const intrusionLike = activeSignals.some((signal) => ["burglary", "panic", "duress", "emergency", "medical", "hold_up"].includes(signal));
  const offline = activeSignals.includes("connectivity");
  const batteryIssue = activeSignals.includes("battery");

  let severity = 0;
  if (fireLike || waterLike || alarmActive || intrusionLike) {
    severity = 4;
  } else if (tamperActive || troubleActive || offline) {
    severity = 3;
  } else if (batteryIssue || activeSignals.length > 0) {
    severity = 2;
  }

  const signalBadges = activeSignals
    .sort((left, right) => (SIGNAL_META[right]?.rank || 0) - (SIGNAL_META[left]?.rank || 0))
    .slice(0, 5)
    .map((signal) => ({
      key: signal,
      label: SIGNAL_META[signal]?.label || humanizeKey(signal),
      icon: SIGNAL_META[signal]?.icon || "mdi:alert-circle-outline",
      tone: SIGNAL_META[signal]?.tone || "warning"
    }));

  return {
    ...device,
    id: `${device.account}:${device.zone}`,
    entityPrefix: source.prefix,
    entityStyle: source.style,
    icon: iconForKind(device.kind, device.room, device.events),
    alarmActive,
    tamperActive,
    troubleActive,
    activeSignals,
    warningSignals,
    signalBadges,
    fireLike,
    waterLike,
    intrusionLike,
    offline,
    batteryIssue,
    severity,
    statusTone: severity >= 4 ? "critical" : severity >= 2 ? "warning" : "clear",
    headline: summarizeHeadline({
      alarmActive,
      tamperActive,
      troubleActive,
      offline,
      batteryIssue,
      warningSignals,
      fireLike,
      waterLike,
      intrusionLike
    }),
    kindLabel: humanizeKey(device.kind),
    lastEventName,
    lastEventAt,
    lastEventAtUnix: Date.parse(lastEventAt) || 0,
    lastSignal,
    alarmSignal,
    alarmAction,
    lastEventShort: formatShortTimestamp(lastEventAt),
    lastEventRelative: formatRelative(lastEventAt)
  };
}

function buildDashboardModel(hass, config, catalog) {
  const account = String(config.account || catalog[0]?.account || "");
  const accountSlug = slugPart(account);
  const filteredCatalog = catalog.filter((item) => String(item.account) === account);
  const devices = filteredCatalog.map((item) => buildDeviceModel(hass, item, config)).sort(sortDevices);
  const availableRooms = uniqueValues(filteredCatalog.map((item) => item.room));
  const roomOrder = uniqueValues((config.room_order || DEFAULT_ROOM_ORDER).concat(availableRooms));
  const orderedRooms = roomOrder.concat(availableRooms.filter((room) => !roomOrder.includes(room)));
  const roomPositions = new Map(orderedRooms.map((room, index) => [room, index]));

  const rooms = orderedRooms
    .map((room) => {
      const roomDevices = devices.filter((device) => device.room === room);
      if (roomDevices.length === 0) {
        return null;
      }

      const critical = roomDevices.filter((device) => device.severity >= 4).length;
      const warning = roomDevices.filter((device) => device.severity === 3 || device.severity === 2).length;
      const offline = roomDevices.filter((device) => device.offline).length;
      const alarmed = roomDevices.filter((device) => device.alarmActive).length;

      const meta = roomMeta(room);

      return {
        name: room,
        devices: roomDevices,
        total: roomDevices.length,
        critical,
        warning,
        offline,
        alarmed,
        accent: meta.accent,
        icon: meta.icon,
        theme: meta.theme,
        position: roomPositions.get(room) || 999
      };
    })
    .filter(Boolean)
    .sort(sortRooms);

  const activeDevices = devices.filter((device) => device.severity > 0).slice(0, 12);
  const recentDevices = devices
    .filter((device) => device.lastEventAtUnix > 0)
    .slice()
    .sort((left, right) => right.lastEventAtUnix - left.lastEventAtUnix)
    .slice(0, 12);

  const signalCount = countBy(
    devices.flatMap((device) => device.activeSignals),
    (signal) => signal
  );

  const accountOnlineState = readState(hass, `binary_sensor.account_${accountSlug}_online`);
  const accountArmedState = readState(hass, `binary_sensor.account_${accountSlug}_armed`);
  const accountPartialState = readState(hass, `binary_sensor.account_${accountSlug}_partially_armed`);
  const accountNightState = readState(hass, `binary_sensor.account_${accountSlug}_night_mode`);
  const accountAlarmState = readState(hass, `binary_sensor.account_${accountSlug}_alarm_active`);
  const accountTamperState = readState(hass, `binary_sensor.account_${accountSlug}_tamper_active`);
  const accountTroubleState = readState(hass, `binary_sensor.account_${accountSlug}_trouble_active`);

  const accountState = {
    online: accountOnlineState?.state === "on",
    armed: accountArmedState?.state === "on",
    partiallyArmed: accountPartialState?.state === "on",
    nightMode: accountNightState?.state === "on",
    alarmActive: accountAlarmState?.state === "on",
    tamperActive: accountTamperState?.state === "on",
    troubleActive: accountTroubleState?.state === "on",
    mode: readState(hass, `sensor.account_${accountSlug}_mode`)?.state || "unknown",
    lastEvent: readState(hass, `sensor.account_${accountSlug}_last_event_name`)?.state || "Awaiting event",
    lastEventAt: readState(hass, `sensor.account_${accountSlug}_last_event_at`)?.state || "",
    lastPingAt: readState(hass, `sensor.account_${accountSlug}_last_ping_at`)?.state || ""
  };

  if (!accountOnlineState) {
    accountState.online = devices.some((device) => device.lastEventAtUnix > 0 || device.lastEventName !== "Awaiting event");
  }
  if (accountState.mode === "unknown") {
    accountState.mode = accountState.armed
      ? "armed"
      : accountState.partiallyArmed
        ? "partially_armed"
        : accountState.nightMode
          ? "night_mode"
          : "disarmed";
  }
  if (accountState.lastEvent === "Awaiting event") {
    accountState.lastEvent = recentDevices[0]?.lastEventName || "Awaiting event";
    accountState.lastEventAt = recentDevices[0]?.lastEventAt || "";
  }

  return {
    account,
    devices,
    rooms,
    activeDevices,
    recentDevices,
    accountState,
    signalSummary: [
      "fire",
      "smoke",
      "burglary",
      "tamper",
      "connectivity",
      "battery"
    ].map((signal) => ({
      key: signal,
      label: SIGNAL_META[signal]?.label || humanizeKey(signal),
      icon: SIGNAL_META[signal]?.icon || "mdi:alert-circle-outline",
      count: signalCount.get(signal) || 0,
      tone: SIGNAL_META[signal]?.tone || "warning"
    })),
    totals: {
      devices: devices.length,
      rooms: rooms.length,
      critical: devices.filter((device) => device.severity >= 4).length,
      attention: devices.filter((device) => device.severity === 3 || device.severity === 2).length,
      offline: devices.filter((device) => device.offline).length,
      healthy: devices.filter((device) => device.severity === 0).length
    }
  };
}

class AjaxSecurityDashboard extends HTMLElement {
  constructor() {
    super();
    this.attachShadow({ mode: "open" });
    this._config = {};
    this._catalog = AJAX_EMBEDDED_DEVICE_CATALOG.slice();
    this._selectedRoom = null;
    this._model = null;
    this._resizeObserver = new ResizeObserver(() => this._updateScale());
  }

  connectedCallback() {
    this._resizeObserver.observe(this);
    this._updateScale();
  }

  disconnectedCallback() {
    this._resizeObserver.disconnect();
  }

  setConfig(config) {
    const merged = {
      title: "Ajax Security Command Deck",
      subtitle: "Purpose-built site overview",
      account: config.account || AJAX_EMBEDDED_DEVICE_CATALOG[0]?.account || "",
      default_room: config.default_room || "",
      room_order: config.room_order || DEFAULT_ROOM_ORDER,
      ...config
    };

    if (!merged.account) {
      throw new Error("Ajax Security Dashboard: account is required.");
    }

    this._config = merged;
    this._catalog = AJAX_EMBEDDED_DEVICE_CATALOG.filter((device) => String(device.account) === String(merged.account));
    if (this._catalog.length === 0) {
      throw new Error(`Ajax Security Dashboard: no catalog entries found for account ${merged.account}.`);
    }
    if (!this._selectedRoom) {
      this._selectedRoom = merged.default_room || null;
    }
    this._render();
  }

  set hass(hass) {
    this._hass = hass;
    if (!this._config.account) {
      return;
    }
    this._model = buildDashboardModel(hass, this._config, this._catalog);
    if (!this._model.rooms.some((room) => room.name === this._selectedRoom)) {
      const criticalRoom = this._model.rooms.find((room) => room.critical > 0 || room.warning > 0);
      this._selectedRoom = criticalRoom?.name || this._config.default_room || this._model.rooms[0]?.name || null;
    }
    this._render();
  }

  getCardSize() {
    return 20;
  }

  _updateScale() {
    const width = this.getBoundingClientRect().width || DASHBOARD_BASE_WIDTH;
    const scale = Math.max(0.42, width / DASHBOARD_BASE_WIDTH);
    this.style.setProperty("--ajax-dashboard-scale", scale.toFixed(4));
    this.style.setProperty("--ajax-dashboard-height", `${Math.round(DASHBOARD_BASE_HEIGHT * scale)}px`);
  }

  _bindEvents() {
    for (const button of this.shadowRoot.querySelectorAll("[data-room]")) {
      button.addEventListener("click", () => {
        this._selectedRoom = button.dataset.room;
        this._render();
      });
    }
  }

  _renderSummaryTile(label, value, detail, tone) {
    return `
      <section class="summary-tile tone-${tone}">
        <div class="summary-label">${escapeHtml(label)}</div>
        <div class="summary-value">${escapeHtml(value)}</div>
        <div class="summary-detail">${escapeHtml(detail)}</div>
      </section>
    `;
  }

  _renderRoomPill(room) {
    const selected = room.name === this._selectedRoom ? "selected" : "";
    const stateClass = room.critical > 0 ? "critical" : room.warning > 0 ? "warning" : "clear";
    return `
      <button class="room-pill ${selected} state-${stateClass}" data-room="${escapeHtml(room.name)}" type="button">
        <div class="room-pill-head">
          <ha-icon icon="${escapeHtml(room.icon)}"></ha-icon>
          <span>${escapeHtml(room.name)}</span>
        </div>
        <div class="room-pill-stats">
          <span>${room.total} devices</span>
          <span>${room.critical} critical</span>
          <span>${room.warning} warnings</span>
        </div>
      </button>
    `;
  }

  _renderDeviceCard(device) {
    const badges = device.signalBadges.length
      ? device.signalBadges
          .map(
            (badge) => `
              <span class="signal-badge tone-${badge.tone}">
                <ha-icon icon="${escapeHtml(badge.icon)}"></ha-icon>
                ${escapeHtml(badge.label)}
              </span>
            `
          )
          .join("")
      : `<span class="signal-badge tone-clear"><ha-icon icon="mdi:check-circle-outline"></ha-icon>Clear</span>`;

    return `
      <article class="device-card tone-${device.statusTone}">
        <div class="device-header">
          <div class="device-icon-wrap">
            <ha-icon icon="${escapeHtml(device.icon)}"></ha-icon>
          </div>
          <div class="device-titles">
            <div class="device-name">${escapeHtml(device.name)}</div>
            <div class="device-kind">${escapeHtml(device.kindLabel)} - ${escapeHtml(device.zone)}</div>
          </div>
          <div class="device-status">${escapeHtml(device.headline)}</div>
        </div>
        <div class="device-badges">${badges}</div>
        <div class="device-footer">
          <div>
            <span class="device-footer-label">Last event</span>
            <strong>${escapeHtml(device.lastEventName)}</strong>
          </div>
          <div>
            <span class="device-footer-label">Updated</span>
            <strong>${escapeHtml(device.lastEventShort)}</strong>
          </div>
        </div>
      </article>
    `;
  }

  _renderActivityItem(device) {
    return `
      <article class="activity-item tone-${device.statusTone}">
        <div class="activity-icon">
          <ha-icon icon="${escapeHtml(device.icon)}"></ha-icon>
        </div>
        <div class="activity-copy">
          <div class="activity-title">${escapeHtml(device.name)}</div>
          <div class="activity-subtitle">${escapeHtml(device.lastEventName)}</div>
        </div>
        <div class="activity-meta">
          <div>${escapeHtml(device.room)}</div>
          <div>${escapeHtml(device.lastEventRelative)}</div>
        </div>
      </article>
    `;
  }

  _renderSignalRow(signal) {
    return `
      <div class="signal-row tone-${signal.tone}">
        <div class="signal-label">
          <ha-icon icon="${escapeHtml(signal.icon)}"></ha-icon>
          <span>${escapeHtml(signal.label)}</span>
        </div>
        <strong>${signal.count}</strong>
      </div>
    `;
  }

  _renderMiniSummary(label, value, detail, tone) {
    return `
      <div class="mini-summary tone-${tone}">
        <div class="mini-summary-copy">
          <div class="mini-summary-label">${escapeHtml(label)}</div>
          <div class="mini-summary-detail">${escapeHtml(detail)}</div>
        </div>
        <strong>${escapeHtml(value)}</strong>
      </div>
    `;
  }

  _render() {
    if (!this._config.account) {
      return;
    }

    const model = this._model;
    if (!model) {
      this.shadowRoot.innerHTML = `
        <ha-card>
          <div style="padding:24px;">Waiting for Home Assistant state...</div>
        </ha-card>
      `;
      return;
    }

    const selectedRoom =
      model.rooms.find((room) => room.name === this._selectedRoom) ||
      model.rooms[0] || {
        name: "No rooms",
        devices: [],
        total: 0,
        critical: 0,
        warning: 0,
        offline: 0,
        alarmed: 0,
        icon: "mdi:shield-home-outline"
      };

    const modeLabel = model.accountState.armed
      ? "Armed"
      : model.accountState.partiallyArmed
        ? "Partially armed"
        : model.accountState.nightMode
          ? "Night mode"
          : "Disarmed";

    const lastPing = model.accountState.lastPingAt ? formatShortTimestamp(model.accountState.lastPingAt) : "No ping";
    const systemTone = model.totals.critical > 0 || model.accountState.alarmActive
      ? "critical"
      : model.totals.attention > 0 || model.accountState.troubleActive || model.accountState.tamperActive
        ? "warning"
        : "clear";

    this.shadowRoot.innerHTML = `
      <style>
        :host {
          display: block;
          --ajax-dashboard-scale: 1;
          --ajax-dashboard-height: 1080px;
          --bg-0: #07111a;
          --bg-1: #0b1723;
          --bg-2: #102434;
          --bg-3: #153349;
          --line: rgba(255, 255, 255, 0.08);
          --text-0: #f5f7fb;
          --text-1: rgba(245, 247, 251, 0.84);
          --text-2: rgba(245, 247, 251, 0.58);
          --critical: #ff6b4a;
          --warning: #ffb347;
          --clear: #6fe3a2;
          --accent: #7dd3fc;
          --shadow: 0 28px 80px rgba(0, 0, 0, 0.42);
          color: var(--text-0);
        }

        * {
          box-sizing: border-box;
        }

        ha-card {
          border: 0;
          box-shadow: none;
          background: transparent;
          overflow: visible;
        }

        .viewport {
          position: relative;
          width: 100%;
          height: var(--ajax-dashboard-height);
          overflow: hidden;
          border-radius: 34px;
        }

        .stage {
          position: relative;
          width: ${DASHBOARD_BASE_WIDTH}px;
          height: ${DASHBOARD_BASE_HEIGHT}px;
          transform: scale(var(--ajax-dashboard-scale));
          transform-origin: top left;
          overflow: hidden;
          border-radius: 34px;
          background:
            radial-gradient(circle at top right, rgba(255, 107, 74, 0.18), transparent 28%),
            radial-gradient(circle at bottom left, rgba(125, 211, 252, 0.14), transparent 26%),
            linear-gradient(145deg, #061018 0%, #0b1723 42%, #07111a 100%);
          box-shadow: var(--shadow);
        }

        .stage::before {
          content: "";
          position: absolute;
          inset: 0;
          background-image:
            linear-gradient(rgba(255, 255, 255, 0.02) 1px, transparent 1px),
            linear-gradient(90deg, rgba(255, 255, 255, 0.02) 1px, transparent 1px);
          background-size: 48px 48px;
          pointer-events: none;
        }

        .dashboard {
          position: relative;
          z-index: 1;
          display: grid;
          grid-template-rows: 272px 150px 1fr;
          gap: 20px;
          padding: 26px;
          height: 100%;
        }

        .glass {
          background: linear-gradient(180deg, rgba(16, 36, 52, 0.92) 0%, rgba(7, 17, 26, 0.92) 100%);
          border: 1px solid var(--line);
          border-radius: 28px;
          backdrop-filter: blur(14px);
        }

        .header {
          display: grid;
          grid-template-columns: 340px minmax(0, 1fr);
          gap: 20px;
          padding: 15px;
        }

        .brand {
          display: flex;
          flex-direction: column;
          justify-content: flex-start;
          gap: 10px;
          min-width: 0;
        }

        .brand-lockup {
          display: flex;
          align-items: flex-start;
        }

        .ajax-logo {
          display: flex;
          align-items: center;
          justify-content: center;
          width: 100%;
          min-height: 128px;
          padding: 14px 18px;
          border-radius: 24px;
          background: linear-gradient(145deg, rgba(255, 107, 74, 0.18), rgba(255, 107, 74, 0.03) 58%, rgba(255, 255, 255, 0.02));
          border: 1px solid rgba(255, 107, 74, 0.22);
          box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.06);
        }

        .ajax-logo svg {
          display: block;
          width: min(100%, 184px);
          height: auto;
          color: #fff8f5;
          filter: drop-shadow(0 10px 24px rgba(255, 107, 74, 0.22));
        }

        .header-content {
          display: grid;
          gap: 12px;
          min-width: 0;
        }

        .top-strip {
          display: flex;
          flex-wrap: wrap;
          gap: 8px;
        }

        .account-chip {
          display: inline-flex;
          align-items: center;
          gap: 8px;
          padding: 8px 10px;
          border-radius: 999px;
          font-size: 13px;
          color: rgba(255, 250, 247, 0.96);
          background: linear-gradient(180deg, rgba(8, 19, 29, 0.94), rgba(13, 28, 41, 0.94));
          border: 1px solid rgba(255, 255, 255, 0.12);
          box-shadow: 0 10px 28px rgba(0, 0, 0, 0.24);
        }

        .account-chip.tone-critical {
          border-color: rgba(255, 107, 74, 0.3);
          background: linear-gradient(180deg, rgba(79, 0, 11, 0.92), rgba(126, 15, 16, 0.9));
        }

        .account-chip.tone-warning {
          border-color: rgba(255, 179, 71, 0.28);
          background: linear-gradient(180deg, rgba(61, 44, 0, 0.92), rgba(96, 72, 8, 0.9));
        }

        .account-chip.tone-clear {
          border-color: rgba(111, 227, 162, 0.2);
          background: linear-gradient(180deg, rgba(7, 58, 43, 0.9), rgba(9, 83, 61, 0.9));
        }

        .account-chip .dot {
          width: 10px;
          height: 10px;
          border-radius: 50%;
          background: var(--clear);
          box-shadow: 0 0 14px rgba(111, 227, 162, 0.45);
        }

        .account-chip.warning .dot {
          background: var(--warning);
          box-shadow: 0 0 14px rgba(255, 179, 71, 0.45);
        }

        .account-chip.critical .dot {
          background: var(--critical);
          box-shadow: 0 0 14px rgba(255, 107, 74, 0.45);
        }

        .account-chip.tone-warning .dot {
          background: var(--warning);
          box-shadow: 0 0 14px rgba(255, 179, 71, 0.45);
        }

        .account-chip.tone-critical .dot {
          background: var(--critical);
          box-shadow: 0 0 14px rgba(255, 107, 74, 0.45);
        }

        .header-mini-grid {
          display: grid;
          grid-template-columns: repeat(4, minmax(0, 1fr));
          gap: 10px;
        }

        .mini-summary {
          display: grid;
          grid-template-columns: 1fr auto;
          gap: 10px;
          align-items: center;
          padding: 12px 14px;
          border-radius: 18px;
          border: 1px solid rgba(255, 255, 255, 0.06);
          background: rgba(255, 255, 255, 0.04);
        }

        .mini-summary-copy {
          min-width: 0;
        }

        .mini-summary-label {
          font-size: 12px;
          text-transform: uppercase;
          letter-spacing: 0.14em;
          color: var(--text-2);
        }

        .mini-summary-detail {
          margin-top: 5px;
          font-size: 12px;
          color: var(--text-1);
        }

        .mini-summary strong {
          font-size: 24px;
          font-weight: 800;
          letter-spacing: -0.04em;
        }

        .mini-summary.tone-critical {
          background: linear-gradient(180deg, rgba(255, 107, 74, 0.16), rgba(255, 107, 74, 0.06));
          border-color: rgba(255, 107, 74, 0.28);
        }

        .mini-summary.tone-warning {
          background: linear-gradient(180deg, rgba(255, 179, 71, 0.16), rgba(255, 179, 71, 0.06));
          border-color: rgba(255, 179, 71, 0.28);
        }

        .mini-summary.tone-clear {
          background: linear-gradient(180deg, rgba(111, 227, 162, 0.14), rgba(111, 227, 162, 0.05));
          border-color: rgba(111, 227, 162, 0.22);
        }

        .summary-grid {
          display: grid;
          grid-template-columns: repeat(3, minmax(0, 1fr));
          gap: 10px;
        }

        .ribbon {
          display: grid;
          gap: 12px;
          padding: 18px;
        }

        .ribbon-head {
          display: flex;
          align-items: baseline;
          justify-content: space-between;
        }

        .ribbon-title {
          font-size: 17px;
          font-weight: 700;
        }

        .ribbon-copy {
          color: var(--text-2);
          font-size: 13px;
        }

        .room-pill-row {
          display: grid;
          grid-template-columns: repeat(5, minmax(0, 1fr));
          gap: 12px;
        }

        .room-pill {
          display: grid;
          gap: 12px;
          padding: 16px;
          border-radius: 22px;
          color: inherit;
          cursor: pointer;
          text-align: left;
          border: 1px solid rgba(255, 255, 255, 0.08);
          background: rgba(255, 255, 255, 0.035);
          transition: transform 160ms ease, border-color 160ms ease, background 160ms ease;
        }

        .room-pill:hover {
          transform: translateY(-2px);
          border-color: rgba(255, 255, 255, 0.18);
        }

        .room-pill.selected {
          border-color: rgba(125, 211, 252, 0.52);
          background: linear-gradient(180deg, rgba(125, 211, 252, 0.13), rgba(125, 211, 252, 0.05));
        }

        .room-pill.state-critical {
          border-color: rgba(255, 107, 74, 0.3);
        }

        .room-pill.state-warning {
          border-color: rgba(255, 179, 71, 0.22);
        }

        .room-pill-head {
          display: flex;
          align-items: center;
          gap: 10px;
          font-size: 15px;
          font-weight: 600;
        }

        .room-pill-stats {
          display: flex;
          flex-wrap: wrap;
          gap: 8px 14px;
          color: var(--text-2);
          font-size: 12px;
        }

        .main {
          display: grid;
          grid-template-columns: 360px 1fr 360px;
          gap: 20px;
          min-height: 0;
        }

        .panel {
          display: grid;
          grid-template-rows: auto 1fr;
          min-height: 0;
          overflow: hidden;
        }

        .panel-head {
          display: flex;
          align-items: baseline;
          justify-content: space-between;
          padding: 20px 22px 12px;
        }

        .panel-title {
          font-size: 18px;
          font-weight: 700;
        }

        .panel-copy {
          color: var(--text-2);
          font-size: 13px;
        }

        .panel-body {
          min-height: 0;
          padding: 0 18px 18px;
        }

        .list-stack {
          display: grid;
          gap: 10px;
          max-height: 100%;
          overflow: auto;
          padding-right: 4px;
        }

        .activity-item,
        .subsystem-row,
        .signal-row {
          display: grid;
          gap: 10px;
          align-items: center;
          padding: 14px 16px;
          border-radius: 18px;
          border: 1px solid rgba(255, 255, 255, 0.06);
          background: rgba(255, 255, 255, 0.04);
        }

        .activity-item {
          grid-template-columns: 40px 1fr auto;
        }

        .activity-icon {
          width: 40px;
          height: 40px;
          display: grid;
          place-items: center;
          border-radius: 14px;
          background: rgba(255, 255, 255, 0.06);
        }

        .activity-title {
          font-size: 14px;
          font-weight: 600;
        }

        .activity-subtitle,
        .activity-meta {
          color: var(--text-2);
          font-size: 12px;
        }

        .activity-meta {
          text-align: right;
        }

        .signal-row {
          grid-template-columns: 1fr auto;
        }

        .signal-label,
        .subsystem-name {
          display: flex;
          align-items: center;
          gap: 10px;
        }

        .subsystem-row {
          grid-template-columns: 1fr auto;
        }

        .subsystem-metrics {
          display: flex;
          gap: 14px;
          color: var(--text-2);
          font-size: 12px;
        }

        .panel-center {
          display: grid;
          grid-template-rows: auto auto 1fr;
          gap: 14px;
          min-height: 0;
          padding: 18px;
        }

        .room-hero {
          display: grid;
          grid-template-columns: 1fr auto;
          gap: 12px;
          align-items: center;
          padding: 20px 22px;
          border-radius: 24px;
          background: linear-gradient(135deg, rgba(255, 255, 255, 0.06), rgba(255, 255, 255, 0.03));
          border: 1px solid rgba(255, 255, 255, 0.08);
        }

        .room-hero-title {
          display: flex;
          align-items: center;
          gap: 12px;
          font-size: 24px;
          font-weight: 800;
          letter-spacing: -0.04em;
        }

        .room-hero-copy {
          margin-top: 8px;
          color: var(--text-2);
          font-size: 14px;
        }

        .room-stat-row {
          display: flex;
          gap: 10px;
          flex-wrap: wrap;
          justify-content: flex-end;
        }

        .mini-stat {
          min-width: 108px;
          padding: 12px 14px;
          border-radius: 18px;
          background: rgba(255, 255, 255, 0.04);
          border: 1px solid rgba(255, 255, 255, 0.06);
        }

        .mini-stat strong {
          display: block;
          margin-top: 4px;
          font-size: 24px;
          font-weight: 800;
        }

        .mini-stat span {
          color: var(--text-2);
          font-size: 12px;
          text-transform: uppercase;
          letter-spacing: 0.12em;
        }

        .device-grid {
          display: grid;
          grid-template-columns: repeat(2, minmax(0, 1fr));
          gap: 12px;
          min-height: 0;
          overflow: auto;
          padding-right: 4px;
        }

        .device-card {
          display: grid;
          gap: 14px;
          padding: 16px;
          border-radius: 22px;
          border: 1px solid rgba(255, 255, 255, 0.06);
          background: rgba(255, 255, 255, 0.04);
        }

        .device-card.tone-critical {
          border-color: rgba(255, 107, 74, 0.28);
          background: linear-gradient(180deg, rgba(255, 107, 74, 0.13), rgba(255, 107, 74, 0.03));
        }

        .device-card.tone-warning {
          border-color: rgba(255, 179, 71, 0.24);
          background: linear-gradient(180deg, rgba(255, 179, 71, 0.1), rgba(255, 179, 71, 0.03));
        }

        .device-card.tone-clear {
          background: rgba(255, 255, 255, 0.03);
        }

        .device-header {
          display: grid;
          grid-template-columns: 56px 1fr auto;
          gap: 14px;
          align-items: start;
        }

        .device-icon-wrap {
          width: 56px;
          height: 56px;
          display: grid;
          place-items: center;
          border-radius: 18px;
          background: rgba(255, 255, 255, 0.08);
        }

        .device-icon-wrap ha-icon,
        .activity-icon ha-icon,
        .room-pill ha-icon,
        .room-hero ha-icon,
        .signal-row ha-icon,
        .subsystem-row ha-icon {
          --mdc-icon-size: 22px;
        }

        .device-name {
          font-size: 15px;
          font-weight: 700;
          line-height: 1.35;
        }

        .device-kind {
          margin-top: 4px;
          color: var(--text-2);
          font-size: 12px;
        }

        .device-status {
          padding: 8px 10px;
          border-radius: 14px;
          background: rgba(255, 255, 255, 0.08);
          font-size: 12px;
          white-space: nowrap;
        }

        .device-badges {
          display: flex;
          flex-wrap: wrap;
          gap: 8px;
        }

        .signal-badge {
          display: inline-flex;
          align-items: center;
          gap: 6px;
          padding: 7px 10px;
          border-radius: 999px;
          font-size: 12px;
          background: rgba(255, 255, 255, 0.06);
          border: 1px solid rgba(255, 255, 255, 0.07);
        }

        .signal-badge ha-icon {
          --mdc-icon-size: 15px;
        }

        .signal-badge.tone-critical {
          color: #ffd6cf;
          border-color: rgba(255, 107, 74, 0.28);
          background: rgba(255, 107, 74, 0.14);
        }

        .signal-badge.tone-warning {
          color: #ffe5b8;
          border-color: rgba(255, 179, 71, 0.22);
          background: rgba(255, 179, 71, 0.12);
        }

        .signal-badge.tone-clear {
          color: #d7fbe4;
          border-color: rgba(111, 227, 162, 0.2);
          background: rgba(111, 227, 162, 0.1);
        }

        .device-footer {
          display: grid;
          grid-template-columns: 1fr 148px;
          gap: 12px;
          color: var(--text-1);
          font-size: 12px;
        }

        .device-footer strong {
          display: block;
          margin-top: 4px;
          color: var(--text-0);
          font-size: 13px;
          line-height: 1.4;
        }

        .device-footer-label {
          color: var(--text-2);
          text-transform: uppercase;
          letter-spacing: 0.12em;
        }

        .empty-note {
          display: grid;
          place-items: center;
          min-height: 220px;
          color: var(--text-2);
          border: 1px dashed rgba(255, 255, 255, 0.12);
          border-radius: 24px;
        }

        ::-webkit-scrollbar {
          width: 8px;
          height: 8px;
        }

        ::-webkit-scrollbar-thumb {
          background: rgba(255, 255, 255, 0.12);
          border-radius: 999px;
        }
      </style>
      <ha-card>
        <div class="viewport">
          <div class="stage">
            <div class="dashboard">
              <section class="header glass">
                <div class="brand">
                  <div class="brand-lockup">
                    <div class="ajax-logo" aria-label="Ajax Security">
                      <svg xmlns="http://www.w3.org/2000/svg" aria-label="Ajax logo" viewBox="0 0 96 20" fill="currentColor" role="img">
                        <path d="M74.69.88h-5.3l21.25 19.11h5.3L74.69.88m6.14 13.6-6.13 5.51h-5.3l6.12-5.51h5.31m8.99-8.09L95.95.88h-5.3l-6.13 5.51h5.3M13.28.88l-2.17 3.11 11.07 16h4.38L13.28.88M7.66 8.97H12L4.38 19.99H0L7.66 8.97M53.24.88l-2.17 3.11 11.07 16h4.38L53.24.88m-5.62 8.09h4.34l-7.62 11.02h-4.38l7.66-11.02M35.23.88l.01 11.7c-.01 1.9-.9 4.27-3.89 4.47h-1.09v2.94l1.57.01a7.01 7.01 0 0 0 4.81-2.06 8.11 8.11 0 0 0 2.03-5.55V.88h-3.44"></path>
                      </svg>
                    </div>
                  </div>
                  <div class="top-strip">
                    <span class="account-chip ${model.accountState.online ? "" : "warning"}">
                      <span class="dot"></span>
                      ${model.accountState.online ? "Online" : "Offline"}
                    </span>
                    <span class="account-chip ${model.accountState.alarmActive ? "critical" : systemTone === "warning" ? "warning" : ""}">
                      <span class="dot"></span>
                      ${escapeHtml(modeLabel)}
                    </span>
                    <span class="account-chip tone-${systemTone}">
                      <span class="dot"></span>
                      ${systemTone === "critical" ? "Respond now" : systemTone === "warning" ? "Attention needed" : "Nominal"}
                    </span>
                  </div>
                </div>
                <div class="header-content">
                  <div class="header-mini-grid">
                    ${this._renderMiniSummary("Devices", String(model.totals.devices), `${model.totals.healthy} nominal`, "clear")}
                    ${this._renderMiniSummary("Critical", String(model.totals.critical), "Immediate response", model.totals.critical > 0 ? "critical" : "clear")}
                    ${this._renderMiniSummary("Attention", String(model.totals.attention), `${model.totals.offline} connectivity`, model.totals.attention > 0 ? "warning" : "clear")}
                    ${this._renderMiniSummary("Rooms", String(model.totals.rooms), `${selectedRoom.name} selected`, "clear")}
                  </div>
                  <div class="summary-grid">
                    ${model.signalSummary.map((signal) => this._renderSignalRow(signal)).join("")}
                  </div>
                </div>
              </section>

              <section class="ribbon glass">
                <div class="ribbon-head">
                  <div class="ribbon-title">Room command ribbon</div>
                  <div class="ribbon-copy">Tap a room to pivot the main device grid.</div>
                </div>
                <div class="room-pill-row">
                  ${model.rooms.map((room) => this._renderRoomPill(room)).join("")}
                </div>
              </section>

              <section class="main">
                <section class="panel glass">
                  <div class="panel-head">
                    <div class="panel-title">Active queue</div>
                    <div class="panel-copy">${model.activeDevices.length} surfaced devices</div>
                  </div>
                  <div class="panel-body">
                    <div class="list-stack">
                      ${model.activeDevices.length
                        ? model.activeDevices.map((device) => this._renderActivityItem(device)).join("")
                        : `<div class="empty-note">No active alarms or warnings.</div>`}
                    </div>
                  </div>
                </section>
 
                <section class="panel glass">
                  <div class="panel-center">
                    <section class="room-hero">
                      <div>
                        <div class="room-hero-title">
                          <ha-icon icon="${escapeHtml(selectedRoom.icon)}"></ha-icon>
                          ${escapeHtml(selectedRoom.name)}
                        </div>
                        <div class="room-hero-copy">
                          ${selectedRoom.total} devices in this zone cluster - ${selectedRoom.critical} critical - ${selectedRoom.warning} warnings - ${selectedRoom.offline} connectivity issues
                        </div>
                      </div>
                      <div class="room-stat-row">
                        <div class="mini-stat"><span>Devices</span><strong>${selectedRoom.total}</strong></div>
                        <div class="mini-stat"><span>Critical</span><strong>${selectedRoom.critical}</strong></div>
                        <div class="mini-stat"><span>Warnings</span><strong>${selectedRoom.warning}</strong></div>
                        <div class="mini-stat"><span>Offline</span><strong>${selectedRoom.offline}</strong></div>
                      </div>
                    </section>
                    <div class="panel-head" style="padding: 0 4px;">
                      <div class="panel-title">Device detail grid</div>
                      <div class="panel-copy">Sorted by severity, then latest event.</div>
                    </div>
                    <div class="device-grid">
                      ${selectedRoom.devices.length
                        ? selectedRoom.devices.map((device) => this._renderDeviceCard(device)).join("")
                        : `<div class="empty-note">No devices found for this room.</div>`}
                    </div>
                  </div>
                </section>

                <section class="panel glass">
                  <div class="panel-head">
                    <div class="panel-title">Recent activity</div>
                    <div class="panel-copy">${model.recentDevices.length} latest device events</div>
                  </div>
                  <div class="panel-body">
                    <div class="list-stack">
                      ${model.recentDevices.length
                        ? model.recentDevices.map((device) => this._renderActivityItem(device)).join("")
                        : `<div class="empty-note">No recent device events yet.</div>`}
                    </div>
                  </div>
                </section>
              </section>
            </div>
          </div>
        </div>
      </ha-card>
    `;

    this._bindEvents();
  }
}

class AjaxSecurityOverview extends HTMLElement {
  constructor() {
    super();
    this.attachShadow({ mode: "open" });
    this._config = {};
    this._catalog = AJAX_EMBEDDED_DEVICE_CATALOG.slice();
    this._model = null;
  }

  setConfig(config) {
    const merged = {
      account: config.account || AJAX_EMBEDDED_DEVICE_CATALOG[0]?.account || "",
      ...config
    };

    if (!merged.account) {
      throw new Error("Ajax Security Overview: account is required.");
    }

    this._config = merged;
    this._catalog = AJAX_EMBEDDED_DEVICE_CATALOG.filter((device) => String(device.account) === String(merged.account));
    if (this._catalog.length === 0) {
      throw new Error(`Ajax Security Overview: no catalog entries found for account ${merged.account}.`);
    }
    this._render();
  }

  set hass(hass) {
    this._hass = hass;
    if (!this._config.account) {
      return;
    }
    this._model = buildDashboardModel(hass, this._config, this._catalog);
    this._render();
  }

  getCardSize() {
    return 6;
  }

  _render() {
    if (!this._config.account) {
      return;
    }

    const model = this._model;
    if (!model) {
      this.shadowRoot.innerHTML = `
        <ha-card>
          <div style="padding:20px;">Waiting for Home Assistant state...</div>
        </ha-card>
      `;
      return;
    }

    const systemTone = systemToneFromModel(model);
    const modeLabel = modeLabelFromAccount(model.accountState);
    const lastPing = model.accountState.lastPingAt ? formatShortTimestamp(model.accountState.lastPingAt) : "No ping";
    const commandStatus = commandStatusCopy(systemTone);
    const chipClass = systemTone === "clear" ? "tone-clear" : systemTone === "warning" ? "tone-warning" : "tone-critical";

    this.shadowRoot.innerHTML = `
      <style>
        :host {
          display: block;
          color: #f5f7fb;
          --line: rgba(255, 255, 255, 0.08);
          --text-1: rgba(245, 247, 251, 0.84);
          --text-2: rgba(245, 247, 251, 0.58);
          --critical: #ff6b4a;
          --warning: #ffb347;
          --clear: #6fe3a2;
        }

        * {
          box-sizing: border-box;
        }

        ha-card {
          border: 0;
          border-radius: 28px;
          overflow: hidden;
          background:
            radial-gradient(circle at top right, rgba(255, 107, 74, 0.14), transparent 30%),
            linear-gradient(180deg, rgba(16, 36, 52, 0.96) 0%, rgba(7, 17, 26, 0.96) 100%);
          box-shadow: 0 22px 52px rgba(0, 0, 0, 0.24);
        }

        .wrap {
          padding: 18px;
        }

        .overview-row {
          display: flex;
          align-items: stretch;
          gap: 8px;
          overflow-x: auto;
          overflow-y: hidden;
          padding-bottom: 2px;
          scrollbar-width: thin;
          scrollbar-color: rgba(255, 255, 255, 0.18) transparent;
        }

        .chip {
          flex: 0 0 auto;
          display: inline-flex;
          align-items: center;
          gap: 8px;
          padding: 10px 13px;
          border-radius: 999px;
          font-size: 12px;
          color: #fffaf7;
          background: linear-gradient(180deg, rgba(8, 19, 29, 0.94), rgba(13, 28, 41, 0.94));
          border: 1px solid rgba(255, 255, 255, 0.12);
          box-shadow: 0 10px 24px rgba(0, 0, 0, 0.18);
        }

        .chip.tone-critical {
          border-color: rgba(255, 107, 74, 0.3);
          background: linear-gradient(180deg, rgba(79, 0, 11, 0.92), rgba(126, 15, 16, 0.9));
        }

        .chip.tone-warning {
          border-color: rgba(255, 179, 71, 0.28);
          background: linear-gradient(180deg, rgba(61, 44, 0, 0.92), rgba(96, 72, 8, 0.9));
        }

        .chip.tone-clear {
          border-color: rgba(111, 227, 162, 0.22);
          background: linear-gradient(180deg, rgba(7, 58, 43, 0.9), rgba(9, 83, 61, 0.9));
        }

        .chip-dot {
          width: 10px;
          height: 10px;
          border-radius: 50%;
          background: var(--clear);
          box-shadow: 0 0 14px rgba(111, 227, 162, 0.45);
        }

        .chip.tone-critical .chip-dot {
          background: var(--critical);
          box-shadow: 0 0 14px rgba(255, 107, 74, 0.45);
        }

        .chip.tone-warning .chip-dot {
          background: var(--warning);
          box-shadow: 0 0 14px rgba(255, 179, 71, 0.45);
        }

        .signal-row {
          flex: 0 0 auto;
          display: inline-flex;
          align-items: center;
          gap: 10px;
          min-width: 150px;
          padding: 10px 14px;
          border-radius: 18px;
          border: 1px solid rgba(255, 255, 255, 0.06);
          background: rgba(255, 255, 255, 0.04);
        }

        .signal-row strong {
          font-size: 20px;
          letter-spacing: -0.04em;
        }

        .signal-label {
          display: flex;
          align-items: center;
          gap: 10px;
          white-space: nowrap;
        }

        .signal-label span {
          font-size: 13px;
          font-weight: 600;
        }

        .signal-label ha-icon {
          --mdc-icon-size: 18px;
        }

        .overview-row::-webkit-scrollbar {
          height: 8px;
        }

        .overview-row::-webkit-scrollbar-thumb {
          background: rgba(255, 255, 255, 0.14);
          border-radius: 999px;
        }

        .tone-critical {
          border-color: rgba(255, 107, 74, 0.22);
          background: linear-gradient(180deg, rgba(255, 107, 74, 0.12), rgba(255, 107, 74, 0.03));
        }

        .tone-warning {
          border-color: rgba(255, 179, 71, 0.2);
          background: linear-gradient(180deg, rgba(255, 179, 71, 0.1), rgba(255, 179, 71, 0.03));
        }

        .tone-clear {
          border-color: rgba(111, 227, 162, 0.18);
          background: linear-gradient(180deg, rgba(111, 227, 162, 0.08), rgba(111, 227, 162, 0.03));
        }
      </style>
      <ha-card>
        <div class="wrap">
          <div class="overview-row">
            <span class="chip ${model.accountState.online ? "tone-clear" : "tone-warning"}">
              <span class="chip-dot"></span>
              Ajax ${model.accountState.online ? "online" : "offline"}
            </span>
            <span class="chip ${chipClass}">
              <span class="chip-dot"></span>
              ${escapeHtml(modeLabel)}
            </span>
            <span class="chip ${chipClass}">
              <span class="chip-dot"></span>
              Command status: ${escapeHtml(commandStatus)}
            </span>
            ${model.signalSummary.map((signal) => this._renderSignalSummary(signal)).join("")}
          </div>
        </div>
      </ha-card>
    `;
  }

  _renderSignalSummary(signal) {
    return `
      <div class="signal-row tone-${signal.tone}">
        <div class="signal-label">
          <ha-icon icon="${escapeHtml(signal.icon)}"></ha-icon>
          <span>${escapeHtml(signal.label)}</span>
        </div>
        <strong>${signal.count}</strong>
      </div>
    `;
  }
}

if (!customElements.get("ajax-security-dashboard")) {
  customElements.define("ajax-security-dashboard", AjaxSecurityDashboard);
}

if (!customElements.get("ajax-security-overview")) {
  customElements.define("ajax-security-overview", AjaxSecurityOverview);
}

window.customCards = window.customCards || [];
window.customCards.push({
  type: "ajax-security-dashboard",
  name: "Ajax Security Dashboard",
  description: "Full-tab purpose-built Ajax Security command dashboard."
});
window.customCards.push({
  type: "ajax-security-overview",
  name: "Ajax Security Overview",
  description: "Compact Ajax summary card for a main dashboard."
});
