(() => {
  "use strict";
  // Study room: stage and elapsed time from the server's start time, a quiet
  // companion and synthesised ambient sound. Nothing here is sent or stored
  // except this viewer's sound preferences in localStorage.
  const store = {
    get(key, fallback) {
      try {
        const value = localStorage.getItem("workey.room." + key);
        return value === null ? fallback : value;
      } catch {
        return fallback;
      }
    },
    set(key, value) {
      try { localStorage.setItem("workey.room." + key, String(value)); } catch { /* private mode */ }
    },
  };
  const reducedMotion = matchMedia("(prefers-reduced-motion: reduce)");
  const baseTitle = document.title;
  const soundNames = { off: "Ses kapalı", library: "Kütüphane", cafe: "Kafe", rain: "Yağmur" };
  const rand = (min, max) => min + Math.random() * (max - min);
  const pick = (list) => list[Math.floor(Math.random() * list.length)];

  // ---- Sound engine --------------------------------------------------------
  let mode = store.get("sound", "off");
  if (!(mode in soundNames)) mode = "off";
  let volume = Math.min(100, Math.max(0, Number(store.get("volume", "45")) || 0));
  let chime = store.get("chime", "1") === "1";
  let ctx = null;
  let master = null;
  let scene = null;
  const buffers = {};

  function audio() {
    if (!ctx) {
      const Context = window.AudioContext || window.webkitAudioContext;
      if (!Context) return null;
      ctx = new Context();
      master = ctx.createGain();
      master.gain.value = level(volume);
      // Many layered events: keep peaks soft instead of clipping.
      const limiter = ctx.createDynamicsCompressor();
      limiter.threshold.value = -18;
      limiter.ratio.value = 6;
      master.connect(limiter);
      limiter.connect(ctx.destination);
    }
    if (ctx.state === "suspended") ctx.resume().catch(() => {});
    return ctx;
  }

  function level(value) { return Math.pow(value / 100, 2) * 1.4; }

  function noise(kind) {
    if (buffers[kind]) return buffers[kind];
    const length = ctx.sampleRate * 6;
    const buffer = ctx.createBuffer(1, length, ctx.sampleRate);
    const data = buffer.getChannelData(0);
    let last = 0, b0 = 0, b1 = 0, b2 = 0, b3 = 0, b4 = 0, b5 = 0, b6 = 0;
    for (let i = 0; i < length; i++) {
      const white = Math.random() * 2 - 1;
      if (kind === "brown") {
        last = (last + 0.02 * white) / 1.02;
        data[i] = last * 3.5;
      } else if (kind === "pink") {
        b0 = 0.99886 * b0 + white * 0.0555179;
        b1 = 0.99332 * b1 + white * 0.0750759;
        b2 = 0.969 * b2 + white * 0.153852;
        b3 = 0.8665 * b3 + white * 0.3104856;
        b4 = 0.55 * b4 + white * 0.5329522;
        b5 = -0.7616 * b5 - white * 0.016898;
        data[i] = (b0 + b1 + b2 + b3 + b4 + b5 + b6 + white * 0.5362) * 0.11;
        b6 = white * 0.115926;
      } else {
        data[i] = white;
      }
    }
    // Remove drift so the loop point does not click.
    const drift = data[length - 1] - data[0];
    for (let i = 0; i < length; i++) data[i] -= drift * (i / length);
    buffers[kind] = buffer;
    return buffer;
  }

  // A room impulse: decaying stereo noise gives each scene its space.
  function impulse(seconds, decay) {
    const length = Math.floor(ctx.sampleRate * seconds);
    const buffer = ctx.createBuffer(2, length, ctx.sampleRate);
    for (let channel = 0; channel < 2; channel++) {
      const data = buffer.getChannelData(channel);
      for (let i = 0; i < length; i++) data[i] = (Math.random() * 2 - 1) * Math.pow(1 - i / length, decay);
    }
    return buffer;
  }

  function chain(...nodes) {
    for (let i = 0; i < nodes.length - 1; i++) nodes[i].connect(nodes[i + 1]);
    return nodes[nodes.length - 1];
  }

  function filter(type, frequency, q) {
    const node = ctx.createBiquadFilter();
    node.type = type;
    node.frequency.value = frequency;
    if (q) node.Q.value = q;
    return node;
  }

  function gain(value) {
    const node = ctx.createGain();
    node.gain.value = value;
    return node;
  }

  function source(kind) {
    const node = ctx.createBufferSource();
    node.buffer = noise(kind);
    return node;
  }

  // An envelope that rises to peak and decays exponentially.
  function strike(env, start, peak, attack, release) {
    env.gain.setValueAtTime(0, start);
    env.gain.linearRampToValueAtTime(peak, start + attack);
    env.gain.exponentialRampToValueAtTime(0.0001, start + attack + release);
  }

  // A filtered noise gesture: paper, graphite, footfall, steam.
  function burst(dest, { kind = "white", type = "bandpass", frequency, q = 1, peak, attack = 0.005, release, at = 0 }) {
    const start = ctx.currentTime + at;
    const src = source(kind);
    const env = gain(0);
    chain(src, filter(type, frequency, q), env, dest);
    strike(env, start, peak, attack, release);
    src.start(start, Math.random() * 4);
    src.stop(start + attack + release + 0.05);
  }

  // Ceramic, glass and metal: a few inharmonic sine partials.
  function clink(dest, { base, peak, at = 0, decay = 0.35 }) {
    const start = ctx.currentTime + at;
    [[1, 1, 1], [1.58, 0.55, 0.7], [2.33, 0.35, 0.5], [3.11, 0.2, 0.35]].forEach(([ratio, amount, length]) => {
      const osc = ctx.createOscillator();
      const env = gain(0);
      osc.frequency.value = base * ratio * rand(0.995, 1.005);
      chain(osc, env, dest);
      strike(env, start, peak * amount, 0.002, decay * length);
      osc.start(start);
      osc.stop(start + decay + 0.05);
    });
  }

  // Speech-like murmur: a voiced (or breathy) source through two moving
  // vowel formants, syllable by syllable, with falling intonation. Heard at
  // a distance, through the room, it reads as people talking.
  const vowels = [[730, 1090], [530, 1840], [290, 2250], [570, 840], [320, 870], [660, 1700], [440, 1020], [400, 1900]];
  function speak(dest, { pitch, syllables, peak, whisper = false, at = 0, laugh = false }) {
    let t = ctx.currentTime + at;
    const start = t;
    let src;
    if (whisper) {
      src = source("pink");
      src.loop = true;
    } else {
      src = ctx.createOscillator();
      src.type = "sawtooth";
      src.frequency.setValueAtTime(pitch, t);
    }
    const env = gain(0);
    const mix = gain(1);
    const f1 = filter("bandpass", 600, whisper ? 5 : 9);
    const f2 = filter("bandpass", 1400, whisper ? 6 : 11);
    const air = filter("bandpass", 3200, 2.5);
    src.connect(f1);
    src.connect(f2);
    chain(f1, gain(1), mix);
    chain(f2, gain(0.6), mix);
    if (whisper) chain(src, air, gain(0.35), mix);
    chain(mix, env, dest);

    for (let i = 0; i < syllables; i++) {
      const length = laugh ? rand(0.1, 0.14) : rand(0.1, 0.24);
      const [a, b] = laugh ? [760, 1200] : pick(vowels);
      const shift = whisper ? 1.12 : 1;
      f1.frequency.setTargetAtTime(a * shift, t, 0.025);
      f2.frequency.setTargetAtTime(b * shift, t, 0.025);
      if (!whisper) {
        const contour = laugh ? 1.25 - (i / syllables) * 0.3 : 1.08 - (i / syllables) * 0.22;
        src.frequency.setTargetAtTime(pitch * contour * rand(0.92, 1.08), t, 0.04);
      }
      const loud = peak * rand(0.55, 1);
      env.gain.setTargetAtTime(loud, t, 0.018);
      env.gain.setTargetAtTime(loud * 0.18, t + length * 0.72, 0.03);
      if (whisper && Math.random() < 0.3) {
        burst(dest, { frequency: rand(4200, 6500), q: 1.5, peak: peak * 0.7, attack: 0.02, release: 0.09, at: t - ctx.currentTime });
      }
      t += length;
      if (laugh) {
        env.gain.setTargetAtTime(0, t, 0.015);
        t += rand(0.05, 0.09);
      } else if (Math.random() < 0.2) {
        env.gain.setTargetAtTime(0, t, 0.02);
        t += rand(0.06, 0.18);
      }
    }
    env.gain.setTargetAtTime(0, t, 0.05);
    src.start(start);
    src.stop(t + 0.5);
    return t - ctx.currentTime;
  }

  // ---- Library -----------------------------------------------------------
  function paperTurn(dest, amount = 1) {
    const start = ctx.currentTime;
    const src = source("white");
    const sweep = filter("bandpass", 900, 0.8);
    const env = gain(0);
    chain(src, sweep, env, dest);
    sweep.frequency.setValueAtTime(900, start);
    sweep.frequency.exponentialRampToValueAtTime(2600, start + 0.3);
    strike(env, start, 0.07 * amount, 0.07, 0.3);
    src.start(start, Math.random() * 4);
    src.stop(start + 0.45);
    const crackles = 8 + Math.floor(Math.random() * 9);
    for (let i = 0; i < crackles; i++) {
      burst(dest, { type: "highpass", frequency: rand(2200, 4200), peak: rand(0.015, 0.06) * amount, attack: 0.001, release: rand(0.004, 0.014), at: rand(0, 0.32) });
    }
    burst(dest, { type: "lowpass", frequency: 700, peak: 0.05 * amount, attack: 0.004, release: 0.07, at: 0.34 });
  }

  function pencil(dest) {
    let at = 0;
    const strokes = 3 + Math.floor(Math.random() * 8);
    for (let i = 0; i < strokes; i++) {
      const length = rand(0.06, 0.18);
      burst(dest, { frequency: rand(3200, 6000), q: 1.6, peak: rand(0.012, 0.024), attack: 0.015, release: length, at });
      at += length + rand(0.03, 0.2);
    }
  }

  function eraser(dest) {
    const start = ctx.currentTime;
    const length = rand(0.6, 1.4);
    const src = source("white");
    const rub = gain(0.5);
    const env = gain(0);
    const lfo = ctx.createOscillator();
    const depth = gain(0.5);
    lfo.frequency.value = rand(6.5, 10);
    chain(lfo, depth, rub.gain);
    chain(src, filter("bandpass", rand(1500, 2400), 1.3), rub, env, dest);
    env.gain.setValueAtTime(0, start);
    env.gain.linearRampToValueAtTime(0.05, start + 0.05);
    env.gain.setValueAtTime(0.05, start + length - 0.08);
    env.gain.linearRampToValueAtTime(0, start + length);
    src.start(start, Math.random() * 4);
    lfo.start(start);
    src.stop(start + length + 0.05);
    lfo.stop(start + length + 0.05);
    // Brushing the crumbs off the page.
    burst(dest, { frequency: 1100, q: 0.7, peak: 0.03, attack: 0.04, release: 0.18, at: length + rand(0.3, 0.7) });
  }

  // Someone walking past, trying not to be heard: soft heel-toe pairs that
  // approach, pass and fade from one side to the other.
  function footsteps(sceneNode) {
    const steps = 6 + Math.floor(Math.random() * 6);
    const pace = rand(0.55, 0.7);
    const from = Math.random() < 0.5 ? -0.9 : 0.9;
    const pan = panner(from);
    const near = filter("lowpass", 900);
    chain(pan, near, sceneNode);
    if (pan.pan) {
      pan.pan.setValueAtTime(from, ctx.currentTime);
      pan.pan.linearRampToValueAtTime(-from, ctx.currentTime + steps * pace);
    }
    for (let i = 0; i < steps; i++) {
      const at = i * pace + rand(-0.02, 0.02);
      const amount = Math.sin(Math.PI * (i + 0.5) / steps);
      burst(pan, { kind: "brown", type: "lowpass", frequency: rand(160, 240), peak: 0.22 * amount, attack: 0.006, release: 0.09, at });
      burst(pan, { kind: "brown", type: "lowpass", frequency: rand(220, 320), peak: 0.12 * amount, attack: 0.006, release: 0.06, at: at + 0.07 });
      burst(pan, { frequency: rand(1000, 1600), q: 1, peak: 0.008 * amount, attack: 0.01, release: 0.05, at: at + 0.05 });
    }
  }

  function bookDown(dest) {
    burst(dest, { kind: "brown", type: "lowpass", frequency: 180, peak: 0.25, attack: 0.003, release: 0.14 });
    burst(dest, { type: "bandpass", frequency: 1200, q: 0.8, peak: 0.03, attack: 0.002, release: 0.04 });
  }

  // ---- Café --------------------------------------------------------------
  function cup(dest) {
    const base = rand(2100, 3300);
    clink(dest, { base, peak: rand(0.02, 0.045), decay: rand(0.25, 0.45) });
    if (Math.random() < 0.55) clink(dest, { base: base * rand(0.7, 0.85), peak: rand(0.015, 0.03), decay: 0.3, at: rand(0.06, 0.14) });
  }

  function stir(dest) {
    const taps = 6 + Math.floor(Math.random() * 7);
    const base = rand(3600, 4600);
    for (let i = 0; i < taps; i++) clink(dest, { base: base * rand(0.97, 1.03), peak: rand(0.008, 0.016), decay: 0.09, at: i * rand(0.16, 0.22) });
  }

  // The bar: knock out the old puck, grind, pull, steam milk, tap the jug.
  function barista(dest) {
    let at = 0;
    for (let i = 0; i < 3; i++) {
      burst(dest, { kind: "brown", type: "lowpass", frequency: 140, peak: 0.3, attack: 0.003, release: 0.12, at });
      burst(dest, { frequency: 900, q: 1, peak: 0.02, attack: 0.002, release: 0.03, at });
      at += rand(0.22, 0.3);
    }
    at += rand(1, 2);
    grinder(dest, at, rand(2.6, 3.6));
    at += rand(9, 14);
    steam(dest, at, rand(4, 6.5));
    at += 8;
    clink(dest, { base: 1250, peak: 0.03, decay: 0.4, at });
    clink(dest, { base: 1250, peak: 0.025, decay: 0.4, at: at + 0.18 });
  }

  function grinder(dest, at, length) {
    const start = ctx.currentTime + at;
    const motor = ctx.createOscillator();
    motor.type = "sawtooth";
    motor.frequency.value = rand(90, 110);
    const grit = source("white");
    const body = filter("bandpass", 650, 1.1);
    const env = gain(0);
    motor.connect(body);
    chain(grit, gain(0.5), body);
    chain(body, env, dest);
    env.gain.setValueAtTime(0, start);
    env.gain.linearRampToValueAtTime(0.05, start + 0.15);
    env.gain.setValueAtTime(0.05, start + length - 0.4);
    env.gain.linearRampToValueAtTime(0, start + length);
    motor.frequency.setValueAtTime(motor.frequency.value, start + length - 0.4);
    motor.frequency.linearRampToValueAtTime(motor.frequency.value * 0.6, start + length);
    motor.start(start);
    grit.start(start, Math.random() * 4);
    motor.stop(start + length + 0.05);
    grit.stop(start + length + 0.05);
  }

  function steam(dest, at, length) {
    const start = ctx.currentTime + at;
    const hiss = source("white");
    const hissEnv = gain(0);
    chain(hiss, filter("highpass", 2200), filter("lowpass", 9000), hissEnv, dest);
    hissEnv.gain.setValueAtTime(0, start);
    hissEnv.gain.linearRampToValueAtTime(0.035, start + 0.2);
    hissEnv.gain.setValueAtTime(0.035, start + length - 0.3);
    hissEnv.gain.linearRampToValueAtTime(0, start + length);
    // Milk stretching: a gurgle that wanders, then rises as the jug fills.
    const milk = source("pink");
    const gurgle = filter("bandpass", 500, 3);
    const milkEnv = gain(0);
    chain(milk, gurgle, milkEnv, dest);
    for (let t = 0; t < length; t += 0.07) gurgle.frequency.setValueAtTime(rand(380, 900) + (t / length) * 700, start + t);
    milkEnv.gain.setValueAtTime(0, start);
    milkEnv.gain.linearRampToValueAtTime(0.06, start + 0.4);
    milkEnv.gain.setValueAtTime(0.06, start + length - 0.3);
    milkEnv.gain.linearRampToValueAtTime(0, start + length);
    hiss.start(start, Math.random() * 4);
    milk.start(start, Math.random() * 4);
    hiss.stop(start + length + 0.05);
    milk.stop(start + length + 0.05);
  }

  function chairScrape(dest) {
    const start = ctx.currentTime;
    const length = rand(0.25, 0.55);
    const src = source("white");
    const band = filter("bandpass", 450, 6);
    const env = gain(0);
    chain(src, band, env, dest);
    band.frequency.setValueAtTime(rand(350, 500), start);
    band.frequency.linearRampToValueAtTime(rand(550, 800), start + length);
    strike(env, start, 0.04, 0.05, length);
    src.start(start, Math.random() * 4);
    src.stop(start + length + 0.1);
  }

  // ---- Scenes ------------------------------------------------------------
  function panner(value) {
    if (ctx.createStereoPanner) {
      const node = ctx.createStereoPanner();
      node.pan.value = value;
      return node;
    }
    return gain(1);
  }

  function makeScene(name) {
    const out = gain(0);
    out.connect(master);
    out.gain.setValueAtTime(0, ctx.currentTime);
    out.gain.linearRampToValueAtTime(1, ctx.currentTime + 2);
    const input = gain(1);
    const verb = ctx.createConvolver();
    input.connect(out);
    chain(input, verb, gain(name === "library" ? 0.3 : 0.18), out);
    verb.buffer = impulse(name === "library" ? 1.8 : 0.9, name === "library" ? 3 : 3.5);

    const timers = new Set();
    let alive = true;
    const later = (seconds, fn) => {
      const id = setTimeout(() => { timers.delete(id); if (alive) fn(); }, seconds * 1000);
      timers.add(id);
    };
    const every = (min, max, fn) => {
      const loop = () => later(rand(min, max), () => { fn(); loop(); });
      loop();
    };
    // A place in the room: left/right and how far (a darker sound is farther).
    const at = (pan, cutoff) => {
      const node = panner(pan);
      chain(node, filter("lowpass", cutoff), input);
      return node;
    };
    const bed = (kind, filters, amount) => {
      const src = source(kind);
      src.loop = true;
      chain(src, ...filters, gain(amount), out);
      src.start(0, Math.random() * 4);
    };
    // Two people at one table taking turns.
    const table = (pan, cutoff, pitches, peak, pause) => {
      const place = at(pan, cutoff);
      let turn = 0;
      const talk = () => {
        const pitch = pitches[turn++ % pitches.length] * rand(0.97, 1.03);
        const laugh = Math.random() < 0.06;
        const length = speak(place, laugh ? { pitch: pitch * 1.3, syllables: 5 + Math.floor(Math.random() * 3), peak: peak * 1.2, laugh } : { pitch, syllables: 4 + Math.floor(Math.random() * 14), peak });
        later(length + rand(...pause), talk);
      };
      later(rand(0, 3), talk);
    };

    const sceneObject = { name, page: null, stop };
    if (name === "library") {
      // Background is a near-silent room; the events carry the scene.
      bed("brown", [filter("lowpass", 220)], 0.06);
      const desk = at(0.15, 9000);
      sceneObject.page = () => paperTurn(desk);
      every(10, 28, () => pencil(at(rand(-0.6, 0.6), rand(5000, 9000))));
      every(30, 70, () => eraser(at(rand(-0.6, 0.6), rand(4000, 8000))));
      every(15, 40, () => paperTurn(at(rand(-0.9, 0.9), rand(2200, 5000)), rand(0.4, 0.8)));
      every(50, 120, () => footsteps(input));
      every(90, 200, () => bookDown(at(rand(-0.9, 0.9), 2500)));
      // A whispered exchange somewhere between the shelves, then silence.
      every(45, 110, () => {
        const place = at(rand(-0.8, 0.8), rand(3500, 6000));
        const lines = 2 + Math.floor(Math.random() * 3);
        let delay = 0;
        for (let i = 0; i < lines; i++) {
          later(delay, () => speak(place, { pitch: 0, syllables: 3 + Math.floor(Math.random() * 7), peak: 0.05, whisper: true }));
          delay += rand(1.4, 2.6);
        }
      });
    } else if (name === "cafe") {
      // A light hubbub and two tables talking, with long pauses between turns.
      bed("pink", [filter("lowpass", 700)], 0.03);
      table(-0.7, 1200, [118, 205], 0.028, [1.2, 4.5]);
      table(0.6, 1100, [210, 120], 0.024, [1.8, 6]);
      const bar = at(0.55, 7000);
      every(6, 15, () => cup(at(rand(-0.8, 0.8), rand(5000, 9000))));
      every(40, 90, () => stir(at(rand(-0.6, 0.6), 8000)));
      every(70, 150, () => barista(bar));
      every(60, 140, () => chairScrape(at(rand(-0.9, 0.9), 2500)));
      later(rand(15, 30), () => barista(bar));
    } else if (name === "rain") {
      // Steady, soft rain heard through a window: dark wash, sparse drops.
      bed("pink", [filter("highpass", 300), filter("lowpass", 2400)], 0.2);
      every(0.25, 0.9, () => burst(out, { type: "bandpass", frequency: rand(1600, 3200), q: 0.8, peak: rand(0.004, 0.014), attack: 0.003, release: rand(0.04, 0.09) }));
    }

    function stop() {
      alive = false;
      timers.forEach(clearTimeout);
      timers.clear();
      const now = ctx.currentTime;
      out.gain.cancelScheduledValues(now);
      out.gain.setValueAtTime(out.gain.value, now);
      out.gain.linearRampToValueAtTime(0, now + 0.8);
      setTimeout(() => out.disconnect(), 1000);
    }
    return sceneObject;
  }

  function play(name) {
    if (scene && scene.name === name) return;
    if (scene) { scene.stop(); scene = null; }
    if (name === "off" || !audio()) return;
    scene = makeScene(name);
  }

  function ring() {
    if (!chime || !ctx || ctx.state !== "running") return;
    const now = ctx.currentTime;
    [[523.25, 0.16], [783.99, 0.08], [1046.5, 0.035]].forEach(([frequency, peak]) => {
      const osc = ctx.createOscillator();
      const env = gain(0);
      osc.frequency.value = frequency;
      chain(osc, env, master);
      strike(env, now, peak, 0.012, 2.6);
      osc.start(now);
      osc.stop(now + 2.7);
    });
  }

  // Browsers start audio only after a gesture; a saved choice waits for it.
  function unlockOnGesture() {
    const unlock = () => {
      document.removeEventListener("pointerdown", unlock);
      document.removeEventListener("keydown", unlock);
      if (!room) return;
      if (mode !== "off") play(mode);
      else if (chime) audio();
    };
    document.addEventListener("pointerdown", unlock);
    document.addEventListener("keydown", unlock);
  }

  // ---- Room ----------------------------------------------------------------
  let room = null;
  let segments = [];
  let started = 0;
  let offset = 0;
  let lastIndex = null;
  let tickTimer = 0;
  let turnTimer = 0;

  function stageAt(minute) {
    let start = 0;
    for (let i = 0; i < segments.length; i++) {
      if (minute < start + segments[i].min) return { index: i, left: start + segments[i].min - minute };
      start += segments[i].min;
    }
    return { index: segments.length, left: 0 };
  }

  function stateText(cell, parts) {
    cell.replaceChildren(...parts.map((part) => {
      if (typeof part === "string") return document.createTextNode(part);
      const el = part.check ? document.createElementNS("http://www.w3.org/2000/svg", "svg") : document.createElement(part.tag);
      if (part.className) el.className = part.className;
      if (part.text) el.textContent = part.text;
      if (part.check) {
        el.setAttribute("viewBox", "0 0 24 24");
        el.setAttribute("aria-hidden", "true");
        const path = document.createElementNS("http://www.w3.org/2000/svg", "path");
        path.setAttribute("d", "M20 6 9 17l-5-5");
        el.append(path);
      }
      return el;
    }));
  }

  function update() {
    if (!room) return;
    const total = segments.reduce((sum, seg) => sum + seg.min, 0);
    const minute = Math.max(0, Math.floor((Date.now() - started) / 60000)) + offset;
    const { index, left } = stageAt(minute);
    const over = index >= segments.length;
    const stale = minute > 24 * 60;
    const hint = room.querySelector("[data-room-hint]");

    let start = 0;
    room.querySelectorAll("[data-room-seg]").forEach((row, i) => {
      const state = over || i < index ? "past" : i === index ? "current" : "next";
      row.dataset.state = state;
      const now = row.querySelector("[data-room-now]");
      if (now) now.hidden = state !== "current";
      const cell = row.querySelector("[data-room-state]");
      if (cell) {
        if (state === "past") {
          stateText(cell, [{ tag: "svg", check: true }, "Bitti"]);
        } else if (state === "current") {
          stateText(cell, [{ tag: "span", className: "cn", text: String(minute) }, "dk / " + total]);
        } else {
          stateText(cell, [Math.max(0, start - minute) + " dk sonra"]);
        }
      }
      start += segments[i].min;
    });

    hint.hidden = !(over || stale);
    if (stale) {
      hint.textContent = "Bu çalışma bir günden uzun süredir açık. Bitirdiysen tamamla ya da kaldığın yeri işaretle.";
    } else if (over) {
      hint.textContent = "Planlı " + total + " dakika geçti (" + minute + " dk). Bitirdiysen seansı tamamla; sürüyorsa acele etme.";
    }
    document.title = (over || stale ? "Planlı süre doldu" : segments[index].label + " · " + left + " dk kaldı") + " · Workey";

    if (lastIndex !== null && index !== lastIndex && !stale) {
      ring();
      const announce = room.querySelector("[data-room-announce]");
      if (announce) announce.textContent = over ? "Planlı süre doldu." : segments[index].label + " aşamasına geçme zamanı.";
    }
    lastIndex = index;

    // Suggest the live stage in an untouched bookmark form.
    const form = room.querySelector("form[data-room-autostage]:not([data-touched])");
    const code = segments[Math.min(index, segments.length - 1)]?.stage;
    const radio = form && code && form.querySelector('input[name="stage"][value="' + code + '"]');
    if (radio) radio.checked = true;
  }

  function scheduleTurn() {
    clearTimeout(turnTimer);
    turnTimer = setTimeout(() => {
      const companion = room?.querySelector("[data-room-companion]");
      if (companion && !document.hidden) {
        if (!reducedMotion.matches) {
          companion.classList.remove("is-turning");
          void companion.getBoundingClientRect();
          companion.classList.add("is-turning");
          setTimeout(() => companion.classList.remove("is-turning"), 1400);
        }
        if (scene && scene.page) scene.page();
      }
      if (room) scheduleTurn();
    }, rand(16000, 42000));
  }

  function syncControls() {
    room.querySelectorAll("[data-sound]").forEach((button) => {
      button.setAttribute("aria-pressed", String(button.dataset.sound === mode));
    });
    const range = room.querySelector("[data-room-volume]");
    if (range) range.value = String(volume);
    const box = room.querySelector("[data-room-chime]");
    if (box) box.checked = chime;
  }

  function onClick(event) {
    const button = event.target.closest("[data-sound]");
    if (button && room.contains(button)) {
      mode = button.dataset.sound;
      store.set("sound", mode);
      syncControls();
      play(mode);
    }
  }

  function onInput(event) {
    if (event.target.matches("[data-room-volume]")) {
      volume = Number(event.target.value);
      store.set("volume", volume);
      if (master) master.gain.setTargetAtTime(level(volume), ctx.currentTime, 0.05);
    } else if (event.target.matches("[data-room-chime]")) {
      chime = event.target.checked;
      store.set("chime", chime ? "1" : "0");
      if (chime) audio();
    } else if (event.target.closest("form[data-room-autostage]")) {
      event.target.closest("form").dataset.touched = "true";
    }
  }

  function bind() {
    const next = document.querySelector("[data-room]");
    if (next === room) return;
    if (room) {
      room.removeEventListener("click", onClick);
      room.removeEventListener("input", onInput);
      room.removeEventListener("change", onInput);
    }
    clearInterval(tickTimer);
    clearTimeout(turnTimer);
    room = next;
    lastIndex = null;
    if (!room) {
      // Leaving the room (finished or closed) ends its ambience.
      play("off");
      document.title = baseTitle;
      return;
    }
    try { segments = JSON.parse(room.dataset.segments || "[]"); } catch { segments = []; }
    started = Date.parse(room.dataset.started);
    offset = Number(room.dataset.offset) || 0;
    if (!segments.length || Number.isNaN(started)) return;
    room.querySelectorAll("[data-room-sound]").forEach((el) => { el.hidden = false; });
    room.addEventListener("click", onClick);
    room.addEventListener("input", onInput);
    room.addEventListener("change", onInput);
    syncControls();
    update();
    tickTimer = setInterval(update, 10000);
    scheduleTurn();
    if (ctx && ctx.state === "running") play(mode);
  }

  document.addEventListener("visibilitychange", () => { if (!document.hidden) update(); });
  document.addEventListener("htmx:afterSettle", bind);
  bind();
  unlockOnGesture();
})();
