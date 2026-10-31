"use strict";

const $ = (id) => document.getElementById(id);

const COLORS = { melody: "#1c1c1a", harmony: "#5a5a56", bass: "#8a8a84", drums: "#444440" };
const TAIL = 1.2;
const CROSSFADE = 1.0;
const OVERLAP = 0.9; // start the next piece this many seconds before the current one ends
const DRUM_LANE = 0.22; // fraction of the piano roll reserved for drums

const state = {
  rows: new Map(),
  order: [],
  loaded: false,
  current: null,
  loadToken: 0,
  playing: false,
  paused: false,
  autoplay: localStorage.getItem("dubdutduc-auto") !== "0",
  advancing: false,
  pieceStart: 0,
  endEvent: null,
  layer: null,
  retiring: [],
  prefetch: null,
  prefetchToken: 0,
  voices: [],
  parts: [],
  pendingShare: readShareParam(),
  shareFlash: 0,
};

/* ---------- catalog ---------- */

$("search-form").addEventListener("submit", async (e) => {
  e.preventDefault();
  const q = $("search").value.trim();
  if (!q) return;
  const msg = $("search-msg");
  msg.hidden = false;
  msg.classList.remove("error");
  msg.textContent = "searching...";
  try {
    const res = await fetch(`/api/search?q=${encodeURIComponent(q)}`, { cache: "no-store" });
    const text = await res.text();
    if (!res.ok) {
      msg.classList.add("error");
      msg.textContent = res.status === 404 ? "not found" : text.trim() || "search failed";
      return;
    }
    const data = JSON.parse(text);
    await refresh();
    if (data.block) {
      msg.hidden = true;
      Tone.start();
      load(data.block.height, true);
      scrollToRow(data.block.height);
      return;
    }
    if (data.blocks?.length) {
      msg.textContent = `${data.blocks.length} matches — pick one below`;
      scrollToRow(data.blocks[0].height);
      return;
    }
    msg.classList.add("error");
    msg.textContent = "not found";
  } catch {
    msg.classList.add("error");
    msg.textContent = "search failed";
  }
});

function scrollToRow(height) {
  const li = state.rows.get(height);
  if (li) li.scrollIntoView({ block: "nearest", behavior: "smooth" });
}

function readShareParam() {
  const b = new URLSearchParams(location.search).get("b");
  if (b != null && /^\d+$/.test(b)) return Number(b);
  return null;
}

function blockURL(height) {
  const u = new URL(location.href);
  u.searchParams.set("b", String(height));
  u.hash = "";
  return u.toString();
}

function syncShareURL(height) {
  const u = new URL(location.href);
  if (u.searchParams.get("b") === String(height)) return;
  u.searchParams.set("b", String(height));
  history.replaceState(null, "", `${u.pathname}${u.search}${u.hash}`);
}

async function openHeight(height, autoplay) {
  if (!state.order.includes(height)) {
    try {
      const res = await fetch(`/api/search?q=${encodeURIComponent(height)}`, { cache: "no-store" });
      if (!res.ok) {
        $("p-block").innerHTML = `<a class="block-link" href="${explorerUrl(height)}" target="_blank" rel="noopener noreferrer">Block #${fmtHeight(height)}</a>`;
        $("p-key").textContent = "Not found";
        $("p-meta").textContent = "Shared block could not be loaded.";
        if (state.order.length) load(state.order[0], false);
        return;
      }
      const cat = await fetch("/api/blocks", { cache: "no-store" });
      if (cat.ok) {
        const { blocks, status } = await cat.json();
        renderStatus(status);
        renderList(blocks, status);
      }
    } catch {
      $("p-meta").textContent = "Shared block could not be loaded.";
      if (state.order.length) load(state.order[0], false);
      return;
    }
  }
  await load(height, autoplay);
  scrollToRow(height);
}

async function refresh() {
  try {
    const res = await fetch("/api/blocks", { cache: "no-store" });
    if (!res.ok) throw new Error(res.statusText);
    const { blocks, status } = await res.json();
    renderStatus(status);
    renderList(blocks, status);
    if (!state.current && blocks.length && state.loadToken === 0) {
      if (state.pendingShare != null) {
        const height = state.pendingShare;
        state.pendingShare = null;
        openHeight(height, false);
      } else {
        load(blocks[0].height, false);
      }
    }
  } catch {
    renderStatus({ error: "Server unreachable" });
  }
}

function renderStatus(status) {
  const el = $("sync");
  el.className = "sync";
  if (status.error) {
    el.classList.add("error");
    $("sync-text").textContent = `offline // retrying`;
    el.title = status.error;
  } else if (status.tip) {
    el.classList.add(status.syncing ? "busy" : "live");
    $("sync-text").textContent = status.syncing
      ? `sync // tip #${fmtHeight(status.tip)}`
      : `live // tip #${fmtHeight(status.tip)}`;
    el.title = "";
  } else {
    el.classList.add("busy");
    $("sync-text").textContent = "sync // ...";
  }
}

function renderList(blocks, status) {
  state.order = blocks.map((b) => b.height);
  const list = $("blocks");
  blocks.forEach((b, i) => {
    let li = state.rows.get(b.height);
    if (li && li.dataset.hash !== b.hash) {
      li.remove();
      li = null;
    }
    if (!li) {
      li = rowFor(b);
      if (state.loaded) li.classList.add("fresh");
      state.rows.set(b.height, li);
    }
    if (list.children[i] !== li) list.insertBefore(li, list.children[i] || null);
    li.querySelector(".ago").textContent = ago(b.time);
  });
  state.loaded = state.loaded || blocks.length > 0;

  $("count").textContent = blocks.length ? `${blocks.length} pieces` : "";
  const empty = $("empty");
  empty.hidden = blocks.length > 0;
  empty.classList.toggle("error", Boolean(status.error));
  empty.textContent = status.error
    ? "// can't reach mempool.space — retrying"
    : "// fetching blocks from mempool.space ...";
  markRows();
}

function explorerUrl(height) {
  return `https://www.blockchain.com/explorer/blocks/btc/${height}`;
}

function rowFor(b) {
  const li = document.createElement("li");
  li.dataset.height = b.height;
  li.dataset.hash = b.hash;

  const row = document.createElement("button");
  row.type = "button";
  row.className = "row";
  row.innerHTML = `
    <a class="height" target="_blank" rel="noopener noreferrer"></a>
    <span class="key"></span>
    <span class="feel">${b.beats}/4 // ${b.bpm} BPM</span>
    <span class="ago"></span>`;
  const heightLink = row.querySelector(".height");
  heightLink.href = explorerUrl(b.height);
  heightLink.textContent = `#${fmtHeight(b.height)}`;
  heightLink.setAttribute("aria-label", `Open block ${b.height} on blockchain.com`);
  heightLink.addEventListener("click", (e) => e.stopPropagation());
  row.querySelector(".key").textContent = b.key;
  row.setAttribute("aria-label", `Play block ${b.height}, ${b.key}`);
  row.addEventListener("click", () => choose(b.height));

  const dl = document.createElement("a");
  dl.className = "dl";
  dl.href = `/api/blocks/${b.height}/midi`;
  dl.download = `block-${b.height}.mid`;
  dl.textContent = "[MIDI]";
  dl.setAttribute("aria-label", `Download MIDI for block ${b.height}`);

  li.append(row, dl);
  return li;
}

function markRows() {
  const h = state.current?.height;
  for (const [height, li] of state.rows) {
    li.classList.toggle("active", height === h);
    li.classList.toggle("playing", height === h && state.playing);
  }
}

function choose(height) {
  Tone.start();
  if (state.current?.height === height) {
    toggle();
    return;
  }
  load(height, true);
}

/* ---------- player ---------- */

function nextHeight(height) {
  const order = state.order;
  if (!order.length) return null;
  const i = order.indexOf(height);
  if (i < 0) return order[0];
  return order[(i + 1) % order.length];
}

function prevHeight(height) {
  const order = state.order;
  if (!order.length) return null;
  const i = order.indexOf(height);
  if (i < 0) return order[0];
  return order[(i - 1 + order.length) % order.length];
}

function goAdjacent(delta) {
  if (!state.order.length || state.advancing) return;
  const from = state.current?.height ?? state.order[0];
  const height = delta > 0 ? nextHeight(from) : prevHeight(from);
  if (height == null || height === state.current?.height) return;
  load(height, true);
}

async function fetchScore(height) {
  const res = await fetch(`/api/blocks/${height}/score`, { cache: "no-store" });
  if (!res.ok) throw new Error(res.statusText);
  return res.json();
}

function prefetchNext(fromHeight) {
  if (!state.autoplay) return;
  const height = nextHeight(fromHeight);
  if (height == null) return;
  if (state.prefetch?.height === height) return;
  const token = ++state.prefetchToken;
  const ready = fetchScore(height)
    .then((score) => {
      if (state.prefetchToken === token) state.prefetch = { height, score, ready };
      return score;
    })
    .catch(() => {
      if (state.prefetchToken === token) state.prefetch = null;
      return null;
    });
  state.prefetch = { height, score: null, ready };
}

async function load(height, autoplay) {
  const token = ++state.loadToken;
  stop();
  $("p-block").innerHTML = `<a class="block-link" href="${explorerUrl(height)}" target="_blank" rel="noopener noreferrer">Block #${fmtHeight(height)}</a>`;
  $("p-key").textContent = "Loading...";
  $("p-meta").textContent = "";
  $("p-hash").textContent = "";
  $("chords").innerHTML = "";
  $("play").disabled = true;

  let score;
  try {
    score = await fetchScore(height);
  } catch {
    if (token === state.loadToken) $("p-meta").textContent = "Could not load this piece.";
    return;
  }
  if (token !== state.loadToken) return;

  state.current = score;
  renderPanel(score);
  drawRoll(score);
  updateClock(0);
  markRows();
  if (autoplay) play();
}

function renderPanel(s) {
  $("p-block").innerHTML =
    `<a class="block-link" href="${explorerUrl(s.height)}" target="_blank" rel="noopener noreferrer">Block #${fmtHeight(s.height)}</a>` +
    ` // mined ${ago(s.time)}`;
  $("p-key").textContent = s.key;
  const names = s.tracks.map((t) => t.name).join(" + ");
  $("p-meta").textContent = `${s.beats}/4 // ${s.bpm} BPM // ${names}`;

  const zeros = s.hash.match(/^0*/)[0];
  const hash = $("p-hash");
  hash.innerHTML = "";
  const z = document.createElement("span");
  z.className = "zeros";
  z.textContent = zeros;
  hash.append(z, s.hash.slice(zeros.length));

  const chords = $("chords");
  chords.innerHTML = "";
  for (const bar of s.bars) {
    const li = document.createElement("li");
    li.innerHTML = `<b></b><i></i>`;
    li.querySelector("b").textContent = bar.chord;
    li.querySelector("i").textContent = bar.roman;
    chords.append(li);
  }

  const dl = $("download");
  dl.href = `/api/blocks/${s.height}/midi`;
  dl.download = `block-${s.height}.mid`;
  dl.hidden = false;
  $("share").hidden = false;
  $("play").disabled = false;
  $("stop").disabled = false;
  $("prev").disabled = false;
  $("next").disabled = false;
  syncShareURL(s.height);
}

async function play() {
  const s = state.current;
  if (!s || state.playing) return;
  await Tone.start();
  const transport = Tone.getTransport();

  if (!state.paused) {
    schedule(s, { from: 0, continuous: false });
    transport.start("+0.05");
  } else {
    transport.start();
  }
  state.playing = true;
  state.paused = false;
  syncButtons();
  requestAnimationFrame(tick);
}

function pause() {
  if (!state.playing) return;
  Tone.getTransport().pause();
  state.voices.forEach((v) => v.release());
  state.playing = false;
  state.paused = true;
  syncButtons();
}

function toggle() {
  state.playing ? pause() : play();
}

function disposeLayer(layer) {
  if (!layer) return;
  layer.parts.forEach((p) => p.dispose());
  layer.voices.forEach((v) => {
    v.release();
    v.dispose();
  });
  try {
    layer.gain.dispose();
  } catch {
    /* already disposed */
  }
}

function retireLayer(layer, fade) {
  if (!layer) return;
  state.retiring.push(layer);
  const now = Tone.now();
  layer.gain.gain.cancelScheduledValues(now);
  layer.gain.gain.setValueAtTime(layer.gain.gain.value, now);
  layer.gain.gain.linearRampToValueAtTime(0, now + fade);
  setTimeout(() => {
    disposeLayer(layer);
    state.retiring = state.retiring.filter((l) => l !== layer);
  }, (fade + TAIL + 0.2) * 1000);
}

function stop() {
  state.advancing = false;
  state.prefetch = null;
  state.prefetchToken++;
  if (window.Tone) {
    const transport = Tone.getTransport();
    if (state.endEvent != null) {
      transport.clear(state.endEvent);
      state.endEvent = null;
    }
    transport.stop();
    transport.cancel();
  }
  disposeLayer(state.layer);
  state.retiring.forEach(disposeLayer);
  state.retiring = [];
  state.layer = null;
  state.parts = [];
  state.voices = [];
  state.pieceStart = 0;
  state.playing = false;
  state.paused = false;
  syncButtons();
  updateClock(0);
}

function schedule(s, { from = 0, continuous = false, fadeIn = 0 } = {}) {
  const transport = Tone.getTransport();
  if (!continuous) {
    transport.cancel();
    transport.position = 0;
    from = 0;
    state.endEvent = null;
    disposeLayer(state.layer);
    state.retiring.forEach(disposeLayer);
    state.retiring = [];
    state.layer = null;
    // Keep Transport BPM fixed after start — it is only a seconds clock.
    // Piece tempo is baked into note times via spb. Changing BPM mid-play
    // remaps ticks↔seconds and makes the next Part's notes land in the past.
    transport.bpm.value = 120;
  } else if (state.endEvent != null) {
    transport.clear(state.endEvent);
    state.endEvent = null;
  }

  const spb = 60 / s.bpm;
  const duration = s.length * spb;
  if (continuous) {
    // Start slightly ahead so note events are never already in the past.
    from = Math.max(from, transport.seconds + 0.05);
  }
  const now = Tone.now();
  const gain = new Tone.Gain(fadeIn > 0 ? 0 : 1).connect(master());
  if (fadeIn > 0) {
    gain.gain.setValueAtTime(0, now);
    gain.gain.linearRampToValueAtTime(1, now + fadeIn);
  }

  const voices = [];
  const parts = [];
  for (const track of s.tracks) {
    const voice = track.role === "drums" ? drumKit(gain) : instrument(track.instrument, gain, spb);
    voices.push(voice);
    const events = buildEvents(track, spb);
    const part = new Tone.Part((time, e) => voice.play(e, time), events);
    part.start(from);
    parts.push(part);
  }

  state.layer = { gain, voices, parts, score: s, from, duration };
  state.pieceStart = from;
  state.voices = voices;
  state.parts = parts;
  prefetchNext(s.height);

  const switchAt = state.autoplay ? Math.max(from, from + duration - OVERLAP) : from + duration + TAIL;
  state.endEvent = transport.scheduleOnce((time) => {
    Tone.getDraw().schedule(() => {
      if (state.autoplay) advanceContinuous();
      else stop();
    }, time);
  }, switchAt);
}

async function advanceContinuous() {
  if (!state.autoplay || !state.current || state.advancing || !state.playing) return;
  state.advancing = true;
  const fromHeight = state.current.height;
  const height = nextHeight(fromHeight);
  if (height == null) {
    state.advancing = false;
    stop();
    return;
  }

  let score = state.prefetch?.height === height ? state.prefetch.score : null;
  try {
    if (!score && state.prefetch?.height === height && state.prefetch.ready) {
      score = await state.prefetch.ready;
    }
    if (!score) score = await fetchScore(height);
  } catch {
    state.advancing = false;
    stop();
    return;
  }
  if (!state.playing || !state.autoplay) {
    state.advancing = false;
    return;
  }

  const transport = Tone.getTransport();
  retireLayer(state.layer, CROSSFADE);
  state.layer = null;

  state.current = score;
  state.prefetch = null;
  renderPanel(score);
  drawRoll(score);
  markRows();
  schedule(score, { from: transport.seconds, continuous: true, fadeIn: CROSSFADE * 0.75 });
  state.advancing = false;
  updateClock(transport.seconds);
}

function syncAutoButton() {
  const btn = $("auto");
  btn.setAttribute("aria-pressed", String(state.autoplay));
  btn.title = state.autoplay
    ? "Autoplay on — next block starts when this one ends"
    : "Autoplay off";
}

function syncButtons() {
  const play = $("play");
  play.textContent = state.playing ? "[ PAUSE ]" : state.paused ? "[ RESUME ]" : "[ PLAY ]";
  play.setAttribute("aria-pressed", String(state.playing));
  $("playhead").hidden = !(state.playing || state.paused);
  syncAutoButton();
  markRows();
}

function tick() {
  const s = state.current;
  if (!s || !(state.playing || state.paused)) return;
  updateClock(Tone.getTransport().seconds);
  if (state.playing) requestAnimationFrame(tick);
}

function updateClock(seconds) {
  const s = state.current;
  if (!s) return;
  const total = (s.length * 60) / s.bpm;
  const t = Math.min(Math.max(seconds - state.pieceStart, 0), total);
  const pct = total ? t / total : 0;
  $("progress").style.width = `${pct * 100}%`;
  $("clock").textContent = `${fmtTime(t)} / ${fmtTime(total)}`;
  $("playhead").style.transform = `translateX(${pct * $("roll-wrap").clientWidth}px)`;

  const bar = Math.min(Math.floor(((t * s.bpm) / 60) / s.beats), s.bars.length - 1);
  const active = state.playing || state.paused;
  [...$("chords").children].forEach((li, i) => li.classList.toggle("on", active && i === bar));
}

function seek(e) {
  const s = state.current;
  if (!s || !(state.playing || state.paused) || state.advancing) return;
  const rect = $("roll-wrap").getBoundingClientRect();
  const pct = (e.clientX - rect.left) / rect.width;
  const target = state.pieceStart + pct * ((s.length * 60) / s.bpm);
  state.voices.forEach((v) => v.release());
  Tone.getTransport().seconds = target;
  updateClock(Tone.getTransport().seconds);
}

/* ---------- piano roll ---------- */

function drawRoll(s) {
  const canvas = $("roll");
  const dpr = window.devicePixelRatio || 1;
  const w = canvas.clientWidth;
  const h = canvas.clientHeight;
  canvas.width = Math.round(w * dpr);
  canvas.height = Math.round(h * dpr);
  const ctx = canvas.getContext("2d");
  ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
  ctx.clearRect(0, 0, w, h);

  const pitched = s.tracks.filter((t) => t.role !== "drums").flatMap((t) => t.notes);
  const drumTrack = s.tracks.find((t) => t.role === "drums");
  const pitchH = drumTrack ? h * (1 - DRUM_LANE) : h;
  const lo = Math.min(...pitched.map((n) => n.midi)) - 2;
  const hi = Math.max(...pitched.map((n) => n.midi)) + 2;
  const rowH = pitchH / (hi - lo + 1);
  const x = (beat) => (beat / s.length) * w;

  for (let bar = 1; bar < s.bars.length; bar++) {
    ctx.fillStyle = bar % 4 === 0 ? "#8a8a84" : "#d0d0ca";
    ctx.fillRect(Math.round(x(bar * s.beats)), 0, 1, h);
  }

  if (drumTrack) {
    drawDrumLane(ctx, drumTrack, s, w, h, pitchH, x);
  }

  for (const role of ["bass", "harmony", "melody"]) {
    const track = s.tracks.find((t) => t.role === role);
    if (!track) continue;
    for (const n of track.notes) {
      ctx.globalAlpha = 0.55 + 0.45 * (n.velocity / 127);
      ctx.fillStyle = COLORS[role];
      const y = (hi - n.midi) * rowH;
      const noteW = Math.max(x(n.duration) - 1, 2);
      const noteH = Math.max(Math.floor(rowH) - 1, 3);
      ctx.fillRect(Math.floor(x(n.start)), Math.floor(y), noteW, noteH);
    }
  }
  ctx.globalAlpha = 1;
}

function drawDrumLane(ctx, drumTrack, s, w, h, pitchH, x) {
  const laneTop = Math.floor(pitchH) + 1;
  const laneH = h - laneTop;
  const rowH = laneH / 3;
  const labels = ["K", "S", "H"];

  ctx.fillStyle = "#f3f3f1";
  ctx.fillRect(0, laneTop - 1, w, laneH + 1);

  // ASCII-style separator between pitched notes and drums
  ctx.fillStyle = "#8a8a84";
  for (let i = 0; i < w; i += 6) {
    ctx.fillRect(i, laneTop - 1, 4, 1);
  }

  ctx.font = "10px \"IBM Plex Mono\", monospace";
  ctx.textBaseline = "middle";
  for (let i = 0; i < 3; i++) {
    const y = laneTop + i * rowH;
    if (i > 0) {
      ctx.fillStyle = "#d8d8d2";
      ctx.fillRect(0, Math.floor(y), w, 1);
    }
    ctx.fillStyle = "#9a9a94";
    ctx.fillText(labels[i], 6, y + rowH / 2);
  }

  const lanes = { 36: 0, 38: 1, 42: 2, 46: 2 };
  for (const n of drumTrack.notes) {
    const lane = lanes[n.midi] ?? 2;
    const open = n.midi === 46;
    const cx = Math.floor(x(n.start));
    const cy = laneTop + lane * rowH + rowH / 2;
    const alpha = 0.45 + 0.55 * (n.velocity / 127);
    ctx.globalAlpha = alpha;
    ctx.fillStyle = COLORS.drums;
    ctx.strokeStyle = COLORS.drums;
    ctx.lineWidth = 1.25;

    if (lane === 0) {
      // kick — filled block
      const bw = Math.max(5, Math.floor(x(0.2)));
      const bh = Math.max(5, Math.floor(rowH * 0.45));
      ctx.fillRect(cx, Math.floor(cy - bh / 2), bw, bh);
    } else if (lane === 1) {
      // snare — outlined square
      const bw = Math.max(4, Math.floor(x(0.16)));
      const bh = Math.max(4, Math.floor(rowH * 0.4));
      ctx.strokeRect(cx + 0.5, Math.floor(cy - bh / 2) + 0.5, bw, bh);
    } else if (open) {
      // open hat — wider dash
      const bw = Math.max(6, Math.floor(x(0.28)));
      ctx.fillRect(cx, Math.floor(cy - 1), bw, 2);
    } else {
      // closed hat — small tick, not a filled slab
      ctx.fillRect(cx, Math.floor(cy - 1), 2, 2);
    }
  }
  ctx.globalAlpha = 1;

  // keep bar lines readable through the drum strip
  for (let bar = 1; bar < s.bars.length; bar++) {
    ctx.fillStyle = bar % 4 === 0 ? "#8a8a84" : "#d0d0ca";
    ctx.fillRect(Math.round(x(bar * s.beats)), laneTop, 1, laneH);
  }
}

/* ---------- instruments ---------- */

const NTSC_FRAME = 1 / 60;
const SEMI_DOWN = Math.pow(2, -1 / 12);

let bus;

function master() {
  if (bus) return bus;
  const limiter = new Tone.Limiter(-1).toDestination();
  const comp = new Tone.Compressor({ threshold: -12, ratio: 3.5, attack: 0.003, release: 0.08 }).connect(limiter);
  bus = new Tone.Filter(11000, "lowpass").connect(comp);
  return bus;
}

function dacShaper(steps = 16) {
  const n = 512;
  const curve = new Float32Array(n);
  const half = steps / 2;
  for (let i = 0; i < n; i++) {
    const x = (i * 2) / (n - 1) - 1;
    curve[i] = Math.round(x * half) / half;
  }
  return () => new Tone.WaveShaper(curve, n);
}

function trackerEcho(delaySec = 0.16, wet = 0.2, feedback = 0.12) {
  const delayTime = Math.min(0.6, Math.max(0.05, delaySec));
  return () =>
    new Tone.FeedbackDelay({
      delayTime,
      feedback,
      wet,
      maxDelay: 1.0,
    });
}

function buildEvents(track, spb) {
  if (track.instrument === "chip-chord") {
    const byStart = new Map();
    for (const n of track.notes) {
      let g = byStart.get(n.start);
      if (!g) {
        g = {
          time: n.start * spb,
          dur: n.duration * spb,
          vel: n.velocity / 127,
          midis: [],
        };
        byStart.set(n.start, g);
      }
      g.midis.push(n.midi);
    }
    return [...byStart.values()].map((g) => {
      g.midis.sort((a, b) => a - b);
      const freqs = g.midis.map((m) => Tone.Frequency(m, "midi").toFrequency());
      return {
        time: g.time,
        midi: g.midis[0],
        freq: freqs[0],
        freqs,
        dur: g.dur,
        vel: g.vel,
      };
    });
  }
  return track.notes.map((n) => {
    const freq = Tone.Frequency(n.midi, "midi").toFrequency();
    return {
      time: n.start * spb,
      midi: n.midi,
      freq,
      freqs: [freq],
      dur: n.duration * spb,
      vel: n.velocity / 127,
    };
  });
}

const PRESETS = {
  square: (spb) => ({
    volume: -8,
    vibrato: { freq: 6.5, depth: 0.04, delay: 0.14 },
    duty: [
      [0, 0.125],
      [NTSC_FRAME * 1.5, 0.5],
    ],
    scoop: SEMI_DOWN,
    fx: [
      dacShaper(32),
      () => new Tone.Filter(8500, "lowpass"),
      trackerEcho(spb * 0.75, 0.2, 0.12),
    ],
    options: {
      oscillator: { type: "pulse", width: 0.5 },
      envelope: { attack: 0.001, decay: 0.09, sustain: 0.65, release: 0.025 },
    },
  }),
  pulse: (spb) => ({
    volume: -8,
    vibrato: { freq: 6.8, depth: 0.045, delay: 0.13 },
    duty: [
      [0, 0.125],
      [NTSC_FRAME, 0.25],
      [NTSC_FRAME * 3, 0.375],
    ],
    scoop: SEMI_DOWN,
    fx: [
      dacShaper(32),
      () => new Tone.Filter(8200, "lowpass"),
      trackerEcho(spb * 0.5, 0.22, 0.14),
    ],
    options: {
      oscillator: { type: "pulse", width: 0.25 },
      envelope: { attack: 0.001, decay: 0.1, sustain: 0.6, release: 0.025 },
    },
  }),
  saw: (spb) => ({
    volume: -10,
    vibrato: { freq: 6.2, depth: 0.038, delay: 0.15 },
    scoop: SEMI_DOWN,
    fx: [
      dacShaper(14),
      () => new Tone.Filter(5400, "lowpass"),
      trackerEcho(spb * 0.75, 0.18, 0.1),
    ],
    options: {
      oscillator: { type: "sawtooth" },
      envelope: { attack: 0.001, decay: 0.11, sustain: 0.58, release: 0.03 },
    },
  }),
  "chip-arp": (spb) => ({
    volume: -12,
    duty: [
      [0, 0.125],
      [NTSC_FRAME, 0.25],
    ],
    fx: [
      dacShaper(24),
      () => new Tone.Filter(7200, "lowpass"),
      trackerEcho(spb * 0.25, 0.18, 0.1),
    ],
    options: {
      oscillator: { type: "pulse", width: 0.25 },
      envelope: { attack: 0.001, decay: 0.07, sustain: 0.35, release: 0.02 },
    },
  }),
  "tri-bass": () => ({
    volume: -5,
    pitchKick: { mult: 2.0, dur: 0.012 },
    fx: [
      dacShaper(16),
      () => new Tone.Filter(3800, "lowpass"),
    ],
    options: {
      oscillator: { type: "triangle" },
      envelope: { attack: 0.001, decay: 0.04, sustain: 0.92, release: 0.008 },
    },
  }),
  "pulse-bass": () => ({
    volume: -8,
    duty: [
      [0, 0.5],
      [NTSC_FRAME * 1.5, 0.18],
    ],
    pitchKick: { mult: 1.5, dur: 0.012 },
    fx: [
      dacShaper(16),
      () => new Tone.Filter(3000, "lowpass"),
    ],
    options: {
      oscillator: { type: "pulse", width: 0.18 },
      envelope: { attack: 0.001, decay: 0.12, sustain: 0.55, release: 0.015 },
    },
  }),
};

function chipChordVoice(dest = master(), spb = 0.4) {
  // Soft pulse pad: frame-rate chord cycle without a second oscillator into the DAC shaper.
  const synth = new Tone.Synth({
    oscillator: { type: "pulse", width: 0.5 },
    envelope: { attack: 0.002, decay: 0.1, sustain: 0.28, release: 0.04 },
  });
  synth.volume.value = -18;

  const shaper = dacShaper(48)();
  const filter = new Tone.Filter(4200, "lowpass");
  const echo = trackerEcho(spb * 0.5, 0.1, 0.06)();

  synth.chain(shaper, filter, echo, dest);

  return {
    play: (e, time) => {
      const vel = Math.min(0.85, Math.max(0, e.vel * 0.75));
      const freqs = e.freqs?.length ? e.freqs : [e.freq];
      const widthParam = synth.oscillator.width;
      if (widthParam) {
        widthParam.setValueAtTime(0.5, time);
        if (e.dur > NTSC_FRAME * 3) {
          widthParam.setValueAtTime(0.25, time + NTSC_FRAME * 3);
        }
      }
      synth.triggerAttackRelease(freqs[0], e.dur, time, vel);

      if (freqs.length > 1) {
        const steps = Math.max(1, Math.floor(e.dur / NTSC_FRAME));
        for (let i = 1; i < steps; i++) {
          synth.frequency.setValueAtTime(freqs[i % freqs.length], time + i * NTSC_FRAME);
        }
      }
    },
    release: () => {
      const now = Tone.now();
      synth.frequency.cancelScheduledValues(now);
      synth.triggerRelease();
    },
    dispose: () => {
      synth.dispose();
      shaper.dispose();
      filter.dispose();
      echo.dispose();
    },
  };
}

function instrument(id, dest = master(), spb = 0.4) {
  if (id === "chip-chord") return chipChordVoice(dest, spb);
  const preset = (PRESETS[id] || PRESETS.square)(spb);
  const synth = new Tone.Synth(preset.options);
  synth.volume.value = preset.volume;

  const nodes = [];
  let vib = null;
  if (preset.vibrato) {
    vib = new Tone.Vibrato({ frequency: preset.vibrato.freq, depth: 0, wet: 1 });
    nodes.push(vib);
  }
  for (const make of preset.fx || []) {
    nodes.push(make());
  }
  synth.chain(...nodes, dest);

  return {
    play: (e, time) => {
      const vel = Math.min(1, Math.max(0, e.vel));
      const widthParam = synth.oscillator.width;
      if (preset.duty && widthParam) {
        for (const [offset, w] of preset.duty) {
          if (offset === 0 || offset < e.dur * 0.75) {
            widthParam.setValueAtTime(w, time + offset);
          }
        }
      }

      if (preset.pitchKick && e.dur > preset.pitchKick.dur * 1.5) {
        synth.triggerAttackRelease(e.freq * preset.pitchKick.mult, e.dur, time, vel);
        synth.frequency.setValueAtTime(e.freq, time + preset.pitchKick.dur);
      } else if (preset.scoop && e.dur >= 0.11) {
        synth.triggerAttackRelease(e.freq * preset.scoop, e.dur, time, vel);
        synth.frequency.setValueAtTime(e.freq, time + NTSC_FRAME);
      } else {
        synth.triggerAttackRelease(e.freq, e.dur, time, vel);
      }

      if (vib && preset.vibrato) {
        const { depth, delay } = preset.vibrato;
        vib.depth.cancelScheduledValues(time);
        vib.depth.setValueAtTime(0, time);
        if (e.dur > delay + 0.05) {
          vib.depth.setValueAtTime(0, time + delay);
          vib.depth.linearRampToValueAtTime(depth, time + delay + 0.07);
          vib.depth.setValueAtTime(0, time + e.dur);
        }
      }
    },
    release: () => {
      const now = Tone.now();
      synth.frequency.cancelScheduledValues(now);
      if (vib) vib.depth.cancelScheduledValues(now);
      synth.triggerRelease();
    },
    dispose: () => {
      synth.dispose();
      nodes.forEach((n) => n.dispose());
    },
  };
}

function drumKit(dest = master()) {
  const kickShaper = dacShaper(16)();
  const kick = new Tone.MembraneSynth({
    pitchDecay: 0.022,
    octaves: 5.5,
    oscillator: { type: "triangle" },
    envelope: { attack: 0.001, decay: 0.14, sustain: 0, release: 0.03 },
  });
  kick.volume.value = -3;

  const kickClick = new Tone.NoiseSynth({
    noise: { type: "white" },
    envelope: { attack: 0.001, decay: 0.01, sustain: 0, release: 0.005 },
  });
  const kickClickFilter = new Tone.Filter(2200, "bandpass");
  kickClick.volume.value = -12;

  const snareShaper = dacShaper(8)();
  const snare = new Tone.NoiseSynth({
    noise: { type: "white" },
    envelope: { attack: 0.001, decay: 0.13, sustain: 0, release: 0.03 },
  });
  const snareHP = new Tone.Filter(1200, "highpass");
  const snareLP = new Tone.Filter(7200, "lowpass");
  const snareTone = new Tone.MembraneSynth({
    pitchDecay: 0.016,
    octaves: 2.4,
    oscillator: { type: "triangle" },
    envelope: { attack: 0.001, decay: 0.06, sustain: 0, release: 0.02 },
  });
  snare.volume.value = -8;
  snareTone.volume.value = -9;

  const hatShaper = dacShaper(8)();
  const hat = new Tone.NoiseSynth({
    noise: { type: "white" },
    envelope: { attack: 0.001, decay: 0.032, sustain: 0, release: 0.01 },
  });
  const hatFilter = new Tone.Filter(8200, "highpass");
  hat.volume.value = -16;

  const openHat = new Tone.NoiseSynth({
    noise: { type: "white" },
    envelope: { attack: 0.001, decay: 0.19, sustain: 0, release: 0.05 },
  });
  const openFilter = new Tone.Filter(6800, "highpass");
  openHat.volume.value = -14;

  kick.chain(kickShaper, dest);
  kickClick.chain(kickClickFilter, dest);
  snare.chain(snareShaper, snareHP, snareLP, dest);
  snareTone.chain(kickShaper, dest);
  hat.chain(hatShaper, hatFilter, dest);
  openHat.chain(hatShaper, openFilter, dest);

  return {
    play: (e, time) => {
      switch (e.midi) {
        case 36:
          kick.triggerAttackRelease(52, Math.min(e.dur, 0.15), time, e.vel);
          kickClick.triggerAttackRelease(0.01, time, e.vel * 0.8);
          break;
        case 38:
          snare.triggerAttackRelease(Math.min(e.dur, 0.13), time, e.vel);
          snareTone.triggerAttackRelease(195, 0.055, time, e.vel * 0.85);
          break;
        case 46:
          openHat.triggerAttackRelease(Math.min(e.dur, 0.22), time, e.vel);
          break;
        default:
          hat.triggerAttackRelease(Math.min(e.dur, 0.035), time, e.vel);
      }
    },
    release: () => {
      kick.triggerRelease();
      kickClick.triggerRelease();
      snare.triggerRelease();
      snareTone.triggerRelease();
      hat.triggerRelease();
      openHat.triggerRelease();
    },
    dispose: () => {
      kick.dispose();
      kickShaper.dispose();
      kickClick.dispose();
      kickClickFilter.dispose();
      snare.dispose();
      snareShaper.dispose();
      snareHP.dispose();
      snareLP.dispose();
      snareTone.dispose();
      hat.dispose();
      hatShaper.dispose();
      hatFilter.dispose();
      openHat.dispose();
      openFilter.dispose();
    },
  };
}

/* ---------- formatting ---------- */

function fmtHeight(h) {
  return Number(h).toLocaleString("en-US");
}

function fmtTime(sec) {
  const s = Math.round(sec);
  return `${Math.floor(s / 60)}:${String(s % 60).padStart(2, "0")}`;
}

function ago(unix) {
  const d = Date.now() / 1000 - unix;
  if (d < 60) return "just now";
  if (d < 3600) return `${Math.floor(d / 60)} min ago`;
  if (d < 86400) return `${Math.floor(d / 3600)} h ago`;
  return new Date(unix * 1000).toLocaleDateString(undefined, { month: "short", day: "numeric" });
}

/* ---------- wiring ---------- */

$("play").addEventListener("click", toggle);
$("stop").addEventListener("click", () => stop());
$("prev").addEventListener("click", () => goAdjacent(-1));
$("next").addEventListener("click", () => goAdjacent(1));
$("auto").addEventListener("click", () => {
  state.autoplay = !state.autoplay;
  localStorage.setItem("dubdutduc-auto", state.autoplay ? "1" : "0");
  syncAutoButton();
  if (state.playing && state.layer) {
    const transport = Tone.getTransport();
    if (state.endEvent != null) {
      transport.clear(state.endEvent);
      state.endEvent = null;
    }
    const { from, duration, score } = state.layer;
    const switchAt = state.autoplay ? Math.max(from, from + duration - OVERLAP) : from + duration + TAIL;
    state.endEvent = transport.scheduleOnce((time) => {
      Tone.getDraw().schedule(() => {
        if (state.autoplay) advanceContinuous();
        else stop();
      }, time);
    }, switchAt);
    if (state.autoplay) prefetchNext(score.height);
    else {
      state.prefetch = null;
      state.prefetchToken++;
    }
  }
});
$("roll-wrap").addEventListener("click", seek);
$("share").addEventListener("click", async () => {
  if (!state.current) return;
  const url = blockURL(state.current.height);
  const btn = $("share");
  const flash = ++state.shareFlash;
  try {
    await navigator.clipboard.writeText(url);
    btn.textContent = "[ COPIED ]";
  } catch {
    window.prompt("Copy link to this block", url);
    btn.textContent = "[ SHARE ]";
    return;
  }
  setTimeout(() => {
    if (state.shareFlash === flash) btn.textContent = "[ SHARE ]";
  }, 1400);
});
syncAutoButton();
document.addEventListener("keydown", (e) => {
  if (e.target.closest("button, a, input, textarea")) return;
  if (e.code === "Space") {
    e.preventDefault();
    toggle();
    return;
  }
  if (e.key === "n" || e.key === "N") {
    e.preventDefault();
    goAdjacent(1);
    return;
  }
  if (e.key === "p" || e.key === "P") {
    e.preventDefault();
    goAdjacent(-1);
  }
});
window.addEventListener("resize", () => {
  if (state.current) drawRoll(state.current);
  updateClock(window.Tone && (state.playing || state.paused) ? Tone.getTransport().seconds : 0);
});

refresh();
setInterval(refresh, 5000);
