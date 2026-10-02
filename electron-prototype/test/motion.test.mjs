import test from 'node:test';
import assert from 'node:assert/strict';
import { Model, Sampler, Clock, Trail, RelaxedFollower, RELAXED, MAX_TRAIL_POINTS, validateSettings, walkRate, frameAt } from '../app/motion.mjs';
const timing = { walk: Array.from({ length: 30 }, (_, i) => i % 2 ? 18 : 19), idle: [2800, 120] };
const area = { x: 0, y: 0, width: 1280, height: 720 };
test('invalid controls cannot change the formation', () => {
  for (const count of [0, 11, 50, 1.5, NaN]) assert.throws(() => validateSettings({ count, size: 48, paused: false }));
  for (const size of [0, 40, 128, '48']) assert.throws(() => validateSettings({ count: 3, size, paused: false }));
  const model = new Model(timing); assert.throws(() => model.configure({ count: 11, size: 48, paused: false })); assert.equal(model.settings.count, 10);
});
test('cursor velocity uses elapsed time, has limits and resets after sleep', () => {
  const a = new Sampler(), b = new Sampler(); a.sample(0, { x: 0, y: 0 }); b.sample(0, { x: 0, y: 0 });
  let x, y;
  for (let t = 10; t <= 1000; t += 10) x = a.sample(t, { x: t * 0.6, y: 0 });
  for (let t = 20; t <= 1000; t += 20) y = b.sample(t, { x: t * 0.6, y: 0 });
  assert.ok(Math.abs(x.speed - y.speed) < 0.001); assert.ok(Math.abs(x.rate - 1) < 0.001);
  assert.equal(walkRate(0), 0.45); assert.equal(walkRate(99999), 1.5);
  const reset = a.sample(9000, { x: 8000, y: 0 }); assert.equal(reset.speed, 0); assert.equal(reset.moving, false);
});
test('speed changes keep phase continuous; idle begins after 220 ms', () => {
  const clock = new Clock(0, [100, 100, 100, 100], [2800, 120]);
  clock.frame(0, true, 1); clock.frame(100, true, 1); clock.frame(200, true, 0.5);
  assert.equal(clock.phase, 200); clock.frame(300, true, 0.5); assert.equal(clock.phase, 250);
  clock.frame(2500, true, 1); assert.equal(clock.frame(2700, false, 1).action, 'walk');
  assert.equal(clock.frame(2721, false, 1).action, 'idle'); assert.equal(clock.frame(5521, false, 1).frame, 1);
  assert.equal(frameAt(2920, [2800, 120]), 0);
});
test('all sizes seed ten separated animals inside positive and negative monitors', () => {
  for (const size of [32, 48, 64, 96]) for (const work of [area, { x: -1920, y: -100, width: 1920, height: 1080 }]) {
    const model = new Model(timing, { count: 10, size, paused: false });
    const s = model.step(0, { x: work.x + 70, y: work.y + 35 }, work);
    assert.equal(new Set(s.pets.map(p => `${p.x},${p.y}`)).size, 10);
    for (const p of s.pets) { assert.ok(p.x >= work.x && p.x + p.width <= work.x + work.width); assert.ok(p.y >= work.y && p.y + p.height <= work.y + work.height); }
  }
});
test('stopping and changing count retains the trail; size change resets spacing', () => {
  const model = new Model(timing); let state;
  for (let t = 0; t <= 4000; t += 16) state = model.step(t, { x: 400 + Math.min(t, 1000) / 3, y: 360 }, area);
  const before = model.trail.points.map(p => ({ ...p }));
  model.configure({ count: 1, size: 48, paused: false }); model.step(4016, { x: 733.333, y: 360 }, area);
  model.configure({ count: 10, size: 48, paused: false }); state = model.step(4032, { x: 733.333, y: 360 }, area);
  assert.deepEqual(model.trail.points, before); assert.equal(new Set(state.pets.map(p => `${p.x},${p.y}`)).size, 10);
  model.configure({ count: 10, size: 96, paused: false }); state = model.step(4048, { x: 733.333, y: 360 }, area);
  assert.ok(state.pets.every(p => p.height === 96)); assert.notDeepEqual(model.trail.points, before);
});
test('pause hides pets and monitor changes reseed safely', () => {
  const model = new Model(timing); model.step(0, { x: 400, y: 300 }, area);
  model.configure({ count: 10, size: 48, paused: true }); assert.equal(model.step(6000, { x: 500, y: 300 }, area).pets.length, 0);
  model.configure({ count: 10, size: 48, paused: false });
  const other = { x: -1280, y: 0, width: 1280, height: 720 }, s = model.step(7000, { x: -600, y: 300 }, other);
  assert.ok(s.pets.every(p => p.x < 0 && p.x >= -1280)); assert.equal(s.speed, 0);
});
test('trail follows corners and remains bounded during long motion', () => {
  const trail = new Trail(); trail.append({ x: 0, y: 0 }, 250); trail.append({ x: 100, y: 0 }, 250); trail.append({ x: 100, y: 100 }, 250);
  assert.deepEqual(trail.at(150), { x: 50, y: 0, left: false });
  for (let x = 101; x < 10000; x++) trail.append({ x, y: 100 }, 100);
  assert.ok(trail.points.length <= 102);
});

const relaxedSettings = { count: 10, size: 48, paused: false, followMode: 'relaxed' };
test('older settings default to normal and invalid follow modes cannot change the model', () => {
  assert.equal(validateSettings({ count: 3, size: 64, paused: false }).followMode, 'normal');
  const model = new Model(timing, relaxedSettings);
  for (const followMode of ['', 'slow', 1, {}]) assert.throws(() => model.configure({ ...relaxedSettings, followMode }));
  assert.equal(model.settings.followMode, 'relaxed');
});

test('relaxed speed and vector acceleration are bounded through sharp reversals at different update rates', () => {
  for (const size of [32, 48, 64, 96]) for (const interval of [4, 8, 16, 33, 50]) {
    const f = new RelaxedFollower(); f.reset(0, { x: 0, y: 0 });
    for (let now = interval; now < 6000; now += interval) {
      const previous = { ...f.velocity }, position = { ...f.position };
      const target = now < 2000 ? { x: 3800, y: 2100 } : now < 4000 ? { x: -3800, y: -2100 } : { x: 3800, y: -2100 };
      f.step(now, target, size);
      assert.ok(Math.hypot(f.velocity.x, f.velocity.y) <= RELAXED.speed * size / 48 + 1e-8);
      assert.ok(Math.hypot(f.velocity.x - previous.x, f.velocity.y - previous.y) <= RELAXED.acceleration * size / 48 * interval / 1000 + 1e-8);
      assert.ok(Math.hypot(f.position.x - position.x, f.position.y - position.y) <= RELAXED.speed * size / 48 * interval / 1000 + 1e-8);
    }
  }
});

test('relaxed reaction waits once while fresh cursor destinations keep arriving', () => {
  const model = new Model(timing, relaxedSettings);
  const initial = model.step(0, { x: 200, y: 300 }, area).pets[0];
  for (let now = 16; now <= 160; now += 16) {
    const state = model.step(now, { x: 900 + now / 4, y: 300 }, area);
    assert.equal(state.pets[0].x, initial.x); assert.equal(state.pets[0].action, 'idle');
  }
  const state = model.step(192, { x: 1100, y: 300 }, area);
  assert.ok(model.head.x > initial.x); assert.equal(state.pets[0].action, 'walk');
  assert.ok(state.pets[0].rate < 0.8, 'animation uses pet speed even when cursor speed is high');
});

test('relaxed chase replaces old destinations and settles beside the stopped cursor', () => {
  const work = { x: 0, y: 0, width: 3840, height: 2160 };
  for (const size of [32, 48, 64, 96]) {
    const model = new Model(timing, { ...relaxedSettings, size });
    model.step(0, { x: 1800, y: 1080 }, work);
    for (let now = 16; now < 500; now += 16) model.step(now, { x: 3700, y: 1080 }, work);
    let state;
    for (let now = 512; now < 30000; now += 16) state = model.step(now, { x: 300, y: 700 }, work);
    assert.ok(Math.hypot(model.head.x - 310, model.head.y - (700 - size / 2)) <= Math.max(2, size / 24));
    assert.equal(state.movementSpeed, 0); assert.ok(state.pets.every(p => p.action === 'idle'));
    const stopped = structuredClone(state.pets.map(p => ({ x: p.x, y: p.y, left: p.left })));
    state = model.step(30000, { x: 300, y: 700 }, work);
    assert.deepEqual(state.pets.map(p => ({ x: p.x, y: p.y, left: p.left })), stopped);
  }
});

test('tiny cursor jitter leaves relaxed pets resting without facing flicker', () => {
  const model = new Model(timing, relaxedSettings);
  const initial = model.step(0, { x: 500, y: 300 }, area).pets;
  for (let now = 16; now < 5000; now += 16) {
    const state = model.step(now, { x: 500 + Math.sin(now) * 1.5, y: 300 + Math.cos(now) }, area);
    assert.equal(state.movementSpeed, 0);
    assert.ok(state.pets.every((p, i) => p.x === initial[i].x && p.y === initial[i].y && p.left === initial[i].left && p.action === 'idle'));
  }
});

test('relaxed mode switching, pause and sleep retain position and discard elapsed motion', () => {
  const model = new Model(timing, { ...relaxedSettings, followMode: 'normal' });
  model.step(0, { x: 400, y: 300 }, area); model.step(16, { x: 900, y: 300 }, area);
  const before = { ...model.head }, trail = structuredClone(model.trail.points);
  model.configure(relaxedSettings); model.step(32, { x: 1100, y: 300 }, area);
  assert.deepEqual(model.head, before); assert.deepEqual(model.trail.points, trail);
  for (let now = 48; now < 1000; now += 16) model.step(now, { x: 1100, y: 300 }, area);
  const sleeping = { ...model.head };
  const afterSleep = model.step(9000, { x: 200, y: 500 }, area);
  assert.deepEqual(model.head, sleeping); assert.equal(afterSleep.movementSpeed, 0);
  model.configure({ ...relaxedSettings, paused: true }); assert.equal(model.step(10000, { x: 200, y: 500 }, area).pets.length, 0);
  model.configure(relaxedSettings); model.step(20000, { x: 200, y: 500 }, area);
  assert.deepEqual(model.head, sleeping);
  const other = { x: -1920, y: -1080, width: 1920, height: 1080 };
  const state = model.step(20016, { x: -900, y: -300 }, other);
  assert.ok(state.pets.every(p => p.x >= other.x && p.x + p.width <= 0 && p.y >= other.y && p.y + p.height <= 0));
  assert.equal(state.movementSpeed, 0);
});

test('relaxed movement is consistent across update intervals', () => {
  const xs = [4, 8, 16, 33].map(interval => {
    const f = new RelaxedFollower(); f.reset(0, { x: 0, y: 0 });
    for (let now = interval; now < 2000; now += interval) f.step(now, { x: 3000, y: 0 }, 48);
    return f.step(2000, { x: 3000, y: 0 }, 48).x;
  });
  assert.ok(Math.max(...xs) - Math.min(...xs) < 12);
});

test('twenty minutes of fast input retains a bounded formation and constant member count', () => {
  const model = new Model(timing, { ...relaxedSettings, size: 96 });
  const work = { x: 0, y: 0, width: 3840, height: 2160 };
  for (let now = 0; now < 1200000; now += 16) {
    const state = model.step(now, { x: 1920 + 1700 * Math.sin(now / 100), y: 1080 + 900 * Math.cos(now / 80) }, work);
    assert.ok(model.trail.points.length <= MAX_TRAIL_POINTS);
    assert.equal(model.members.length, 10); assert.equal(state.pets.length, 10);
    assert.ok(Number.isFinite(model.head.x) && Number.isFinite(model.head.y));
  }
  const fractional = new Trail();
  for (let i = 0; i < 10000; i++) fractional.append({ x: i / 10000, y: 0 }, 100);
  assert.equal(fractional.points.length, MAX_TRAIL_POINTS);
});
