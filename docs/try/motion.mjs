// Shared Windows/macOS motion. Coordinates and sizes are device-independent pixels.
export const SIZES = Object.freeze([32, 48, 64, 96]);
export const MAX_COUNT = 10;
export const FOLLOW_MODES = Object.freeze(['normal', 'relaxed']);
export const MAX_TRAIL_POINTS = 2048;
export const RELAXED = Object.freeze({ speed: 180, acceleration: 600, reactionMs: 150 });
const clamp = (v, min, max) => Math.max(min, Math.min(max, v));
const distance = (a, b) => Math.hypot(a.x - b.x, a.y - b.y);
const same = (a, b) => a && b && a.x === b.x && a.y === b.y;
export function validateSettings(value) {
  if (!value || !Number.isInteger(value.count) || value.count < 1 || value.count > MAX_COUNT ||
      !SIZES.includes(value.size) || typeof value.paused !== 'boolean') throw new Error('Invalid count, size or pause state');
  const followMode = value.followMode ?? 'normal';
  if (!FOLLOW_MODES.includes(followMode)) throw new Error('Invalid follow mode');
  return { count: value.count, size: value.size, paused: value.paused, followMode };
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
  constructor(start, walk, idle, warmupMs = 2400) { this.start = start; this.walk = walk; this.idle = idle; this.last = start; this.idleStart = start; this.warmupMs = warmupMs; }
  walking = false; phase = 0; rate = 1; lastMotion = -Infinity; idleStart = 0;
  frame(now, moving, rate) {
    if (moving) this.lastMotion = now;
    const walking = now - this.start < this.warmupMs || now - this.lastMotion < 220;
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
      if (this.length - d < keep && this.points.length <= MAX_TRAIL_POINTS) break;
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
// One current destination, never a queue of cursor events. Small integration steps
// keep acceleration and braking consistent across different update intervals.
export class RelaxedFollower {
  position = null; velocity = { x: 0, y: 0 }; last = null; wakeAt = null; resting = true;
  reset(now, p) {
    this.position = p ? { ...p } : null; this.velocity = { x: 0, y: 0 };
    this.last = now; this.wakeAt = null; this.resting = true;
  }
  step(now, target, size) {
    if (!this.position || this.last === null) { this.reset(now, this.position ?? target); return this.position; }
    const elapsed = now - this.last;
    if (elapsed <= 0) return this.position;
    this.last = now;
    if (elapsed > 250) { this.reset(now, this.position); return this.position; }
    const scale = size / 48, maxSpeed = RELAXED.speed * scale, acceleration = RELAXED.acceleration * scale;
    const radius = Math.max(2, size / 24);
    if (this.resting) {
      if (distance(this.position, target) <= radius * 2) { this.wakeAt = null; return this.position; }
      this.wakeAt ??= now + RELAXED.reactionMs;
      if (now < this.wakeAt) return this.position;
      this.wakeAt = null; this.resting = false;
    }
    const steps = Math.ceil(elapsed / (1000 / 120)), dt = elapsed / 1000 / steps;
    for (let i = 0; i < steps; i++) {
      const dx = target.x - this.position.x, dy = target.y - this.position.y, d = Math.hypot(dx, dy);
      const speed = Math.hypot(this.velocity.x, this.velocity.y);
      if (d <= radius && speed <= acceleration * dt) {
        this.velocity = { x: 0, y: 0 }; this.resting = true; break;
      }
      const wantedSpeed = Math.min(maxSpeed, d * 4);
      const wantedX = d ? dx / d * wantedSpeed : 0, wantedY = d ? dy / d * wantedSpeed : 0;
      const ax = (wantedX - this.velocity.x) / 0.06, ay = (wantedY - this.velocity.y) / 0.06;
      const gain = Math.min(1, acceleration / (Math.hypot(ax, ay) || 1));
      this.velocity.x += ax * gain * dt; this.velocity.y += ay * gain * dt;
      this.position.x += this.velocity.x * dt; this.position.y += this.velocity.y * dt;
    }
    return this.position;
  }
}
export class Model {
  constructor({ walk, idle }, settings = { count: 10, size: 48, paused: false }) {
    if (![walk, idle].every(a => Array.isArray(a) && a.length > 0 && a.length <= 4096) || [...walk, ...idle].some(t => !Number.isFinite(t) || t <= 0)) throw new Error('Invalid motion timings');
    this.timings = { walk, idle }; this.settings = validateSettings(settings);
  }
  sampler = new Sampler(); trail = new Trail(); members = []; head = null; left = false; areaKey = ''; resetLayout = true;
  follower = new RelaxedFollower(); targetLeft = false;
  profiles = [];
  setProfiles(profiles) {
    if (!Array.isArray(profiles) || profiles.length !== MAX_COUNT || profiles.some(p => !p?.id || ![p.walk, p.idle].every(a => Array.isArray(a) && a.length && a.every(t => Number.isFinite(t) && t > 0)))) throw new Error('Invalid animal profiles');
    this.profiles = profiles;
  }
  configure(next) {
    next = validateSettings(next);
    if (next.size !== this.settings.size) this.resetLayout = true;
    if (next.paused !== this.settings.paused || next.followMode !== this.settings.followMode) {
      this.resetMotion(); this.targetLeft = this.left;
    }
    this.settings = next;
  }
  resetMotion() {
    this.sampler.reset(); this.members.forEach(m => m.sampler.reset());
    this.follower.reset(null, this.head);
  }
  step(now, cursor, area) {
    const { count, size, paused, followMode } = this.settings, relaxed = followMode === 'relaxed';
    if (paused) return { pets: [], speed: 0, rate: 0.45, paused: true, settings: this.settings };
    const sample = this.sampler.sample(now, cursor), width = size * 1.5, spacing = width + 8, keep = spacing * (MAX_COUNT + 2);
    if (!relaxed && sample.direction) this.left = sample.direction < 0;
    const maxX = area.x + Math.max(0, area.width - width), maxY = area.y + Math.max(0, area.height - size);
    if (relaxed && this.head) {
      // Hold the approach side near arrival; changing it on every cursor reversal
      // would move the destination across the cursor and prevent settling.
      if (cursor.x < this.head.x - 24) this.targetLeft = true;
      else if (cursor.x > this.head.x + width + 24) this.targetLeft = false;
    }
    let side = relaxed ? this.targetLeft : this.left;
    let targetX = cursor.x + (side ? 10 : -width - 10);
    if (targetX < area.x) { side = true; targetX = cursor.x + 10; }
    if (targetX > maxX) { side = false; targetX = cursor.x - width - 10; }
    if (relaxed) this.targetLeft = side; else this.left = side;
    const target = { x: clamp(targetX, area.x, maxX), y: clamp(cursor.y - size / 2, area.y, maxY) };
    const key = JSON.stringify(area);
    const reset = this.resetLayout || this.areaKey !== key || !this.head;
    if (reset) {
      this.head = { ...target }; this.follower.reset(now, this.head);
      if (relaxed) this.left = side;
    } else if (relaxed) {
      this.head = this.follower.step(now, target, size);
      const x = clamp(this.head.x, area.x, maxX), y = clamp(this.head.y, area.y, maxY);
      if (x !== this.head.x) this.follower.velocity.x = 0;
      if (y !== this.head.y) this.follower.velocity.y = 0;
      this.head.x = x; this.head.y = y;
    } else {
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
      if (!this.members[i] || this.members[i].profileId !== profile.id) this.members[i] = { profileId: profile.id, sampler: new Sampler(), clock: new Clock(now, profile.walk, profile.idle, relaxed ? 0 : 2400) };
      const p = i === 0 ? { ...head, left: this.left } : this.trail.at(i * spacing);
      const m = this.members[i], local = m.sampler.sample(now, relaxed && i === 0 ? this.head : p);
      if (relaxed && i === 0 && local.direction) { this.left = local.direction < 0; p.left = this.left; }
      m.clock.warmupMs = relaxed ? 0 : 2400;
      const rate = relaxed ? walkRate(local.speed * 48 / size) : i === 0 ? sample.rate : local.rate;
      const moving = local.moving || (!relaxed && i === 0 && sample.moving);
      const frame = m.clock.frame(now, moving, now - m.clock.start < m.clock.warmupMs && !moving ? 1 : rate);
      pets.push({ id: i, animalId: profile.id, ...p, width, height: size, ...frame, rate });
    }
    // Hidden members keep their clocks, but resume without treating their time offscreen as movement.
    for (let i = count; i < this.members.length; i++) this.members[i].sampler.reset();
    return { pets, speed: sample.speed, movementSpeed: relaxed ? Math.hypot(this.follower.velocity.x, this.follower.velocity.y) : this.members[0].sampler.speed,
      rate: relaxed ? pets[0].rate : sample.rate, paused: false, settings: this.settings };
  }
}
