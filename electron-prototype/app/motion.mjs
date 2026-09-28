// Shared Windows/macOS motion. Coordinates and sizes are device-independent pixels.
export const SIZES = Object.freeze([32, 48, 64, 96]);
export const MAX_COUNT = 10;
const clamp = (v, min, max) => Math.max(min, Math.min(max, v));
const distance = (a, b) => Math.hypot(a.x - b.x, a.y - b.y);
const same = (a, b) => a && b && a.x === b.x && a.y === b.y;
export function validateSettings(value) {
  if (!value || !Number.isInteger(value.count) || value.count < 1 || value.count > MAX_COUNT ||
      !SIZES.includes(value.size) || typeof value.paused !== 'boolean') throw new Error('Invalid count, size or pause state');
  return { count: value.count, size: value.size, paused: value.paused };
}
export const walkRate = speed => clamp(0.45 + 0.55 * speed / 600, 0.45, 1.5);
export function frameAt(phase, durations) {
  const total = durations.reduce((a, b) => a + b, 0);
  let t = ((phase % total) + total) % total;
  for (let i = 0; i < durations.length; i++) { if (t < durations[i]) return i; t -= durations[i]; }
  return 0;
}
export class Sampler {
  previous = null; speed = 0; directionX = 0;
  reset() { this.previous = null; this.speed = 0; }
  sample(now, p) {
    const result = { elapsed: 0, moving: false, speed: this.speed, rate: walkRate(this.speed), direction: 0 };
    if (!this.previous) { this.previous = { ...p, now }; this.directionX = p.x; return result; }
    const dt = now - this.previous.now;
    if (dt <= 0) return result;
    const moved = distance(this.previous, p);
    this.previous = { ...p, now };
    result.elapsed = Math.min(dt, 250);
    if (dt > 250) { this.speed = 0; this.directionX = p.x; return { ...result, speed: 0, rate: 0.45 }; }
    const instant = Math.min(moved / (dt / 1000), 2400);
    this.speed += (instant - this.speed) * (1 - Math.exp(-dt / 100));
    if (Math.abs(p.x - this.directionX) >= 3) {
      result.direction = p.x < this.directionX ? -1 : 1; this.directionX = p.x;
    }
    return { ...result, moving: moved > 0.1, speed: this.speed, rate: walkRate(this.speed) };
  }
}
export class Clock {
  constructor(start, walk, idle) { this.start = start; this.walk = walk; this.idle = idle; this.last = start; }
  walking = false; phase = 0; rate = 1; lastMotion = -Infinity; idleStart = 0;
  frame(now, moving, rate) {
    if (moving) this.lastMotion = now;
    const walking = now - this.start < 2400 || now - this.lastMotion < 220;
    if (walking !== this.walking) {
      if (walking) this.phase = 0; else this.idleStart = now;
    } else if (walking) this.phase += Math.max(0, Math.min(250, now - this.last)) * this.rate;
    this.walking = walking; this.last = now; this.rate = clamp(rate, 0.45, 1.5);
    this.phase %= this.walk.reduce((a, b) => a + b, 0);
    return { action: walking ? 'walk' : 'idle', frame: frameAt(walking ? this.phase : now - this.idleStart, walking ? this.walk : this.idle) };
  }
}
export class Trail {
  points = []; length = 0;
  seed(head, area, width, height, left, keep) {
    const back = [{ ...head }]; let p = { ...head }, dir = left ? 1 : -1;
    let vertical = head.y > area.y + (area.height - height) / 2 ? -1 : 1, length = 0;
    const maxX = area.x + Math.max(0, area.width - width), maxY = area.y + Math.max(0, area.height - height);
    const add = q => { if (!same(p, q)) { length += distance(p, q); back.push(q); p = q; } };
    for (let n = 0; length < keep && n < 128; n++) {
      add({ x: dir > 0 ? maxX : area.x, y: p.y });
      let y = p.y + vertical * (height + 12);
      if (y < area.y || y > maxY) { vertical *= -1; y = p.y + vertical * (height + 12); }
      add({ x: p.x, y: clamp(y, area.y, maxY) }); dir *= -1;
    }
    this.points = back.reverse(); this.length = length; this.trim(keep);
  }
  append(p, keep) {
    if (same(this.points.at(-1), p)) return;
    if (this.points.length) this.length += distance(this.points.at(-1), p);
    this.points.push({ ...p }); this.trim(keep);
  }
  trim(keep) {
    while (this.points.length > 2) {
      const d = distance(this.points[0], this.points[1]);
      if (this.length - d < keep) break;
      this.length -= d; this.points.shift();
    }
  }
  at(gap) {
    for (let i = this.points.length - 1; i > 0; i--) {
      const a = this.points[i], b = this.points[i - 1], len = distance(a, b);
      if (gap <= len) return { x: Math.round(a.x + (b.x - a.x) * gap / len), y: Math.round(a.y + (b.y - a.y) * gap / len), left: a.x < b.x };
      gap -= len;
    }
    return { ...(this.points[0] ?? { x: 0, y: 0 }), left: false };
  }
}
export class Model {
  constructor({ walk, idle }, settings = { count: 10, size: 48, paused: false }) {
    if (![walk, idle].every(a => Array.isArray(a) && a.length > 0 && a.length <= 4096) || [...walk, ...idle].some(t => !Number.isFinite(t) || t <= 0)) throw new Error('Invalid motion timings');
    this.timings = { walk, idle }; this.settings = validateSettings(settings);
  }
  sampler = new Sampler(); trail = new Trail(); members = []; head = null; left = false; areaKey = ''; resetLayout = true;
  profiles = [];
  setProfiles(profiles) {
    if (!Array.isArray(profiles) || profiles.length !== MAX_COUNT || profiles.some(p => !p?.id || ![p.walk, p.idle].every(a => Array.isArray(a) && a.length && a.every(t => Number.isFinite(t) && t > 0)))) throw new Error('Invalid animal profiles');
    this.profiles = profiles;
  }
  configure(next) {
    next = validateSettings(next);
    if (next.size !== this.settings.size) this.resetLayout = true;
    if (next.paused !== this.settings.paused) { this.sampler.reset(); this.members.forEach(m => m.sampler.reset()); }
    this.settings = next;
  }
  step(now, cursor, area) {
    const { count, size, paused } = this.settings;
    if (paused) return { pets: [], speed: 0, rate: 0.45, paused: true, settings: this.settings };
    const sample = this.sampler.sample(now, cursor), width = size * 1.5, spacing = width + 8, keep = spacing * (MAX_COUNT + 2);
    if (sample.direction) this.left = sample.direction < 0;
    const maxX = area.x + Math.max(0, area.width - width), maxY = area.y + Math.max(0, area.height - size);
    let targetX = cursor.x + (this.left ? 10 : -width - 10);
    if (targetX < area.x) { this.left = true; targetX = cursor.x + 10; }
    if (targetX > maxX) { this.left = false; targetX = cursor.x - width - 10; }
    const target = { x: clamp(targetX, area.x, maxX), y: clamp(cursor.y - size / 2, area.y, maxY) };
    const key = JSON.stringify(area);
    const reset = this.resetLayout || this.areaKey !== key || !this.head;
    if (reset) this.head = { ...target };
    else {
      const response = 140 - 90 * Math.min(1, sample.speed / 1200), gain = 1 - Math.exp(-sample.elapsed / response);
      this.head.x = clamp(this.head.x + (target.x - this.head.x) * gain, area.x, maxX);
      this.head.y = clamp(this.head.y + (target.y - this.head.y) * gain, area.y, maxY);
    }
    const head = { x: Math.round(this.head.x), y: Math.round(this.head.y) };
    if (reset) { this.trail.seed(head, area, width, size, this.left, keep); this.members = []; }
    this.areaKey = key; this.resetLayout = false; this.trail.append(head, keep);
    const pets = [];
    for (let i = 0; i < count; i++) {
      const profile = this.profiles[i] ?? { id: 'default', ...this.timings };
      if (!this.members[i] || this.members[i].profileId !== profile.id) this.members[i] = { profileId: profile.id, sampler: new Sampler(), clock: new Clock(now, profile.walk, profile.idle) };
      const p = i === 0 ? { ...head, left: this.left } : this.trail.at(i * spacing);
      const m = this.members[i], local = m.sampler.sample(now, p);
      const rate = i === 0 ? sample.rate : local.rate;
      const moving = local.moving || (i === 0 && sample.moving);
      const frame = m.clock.frame(now, moving, now - m.clock.start < 2400 && !moving ? 1 : rate);
      pets.push({ id: i, animalId: profile.id, ...p, width, height: size, ...frame, rate });
    }
    // Hidden members keep their clocks, but resume without treating their time offscreen as movement.
    for (let i = count; i < this.members.length; i++) this.members[i].sampler.reset();
    return { pets, speed: sample.speed, rate: sample.rate, paused: false, settings: this.settings };
  }
}
