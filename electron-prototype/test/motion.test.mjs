import test from 'node:test';
import assert from 'node:assert/strict';
import { Model, Sampler, Clock, Trail, validateSettings, walkRate, frameAt } from '../app/motion.mjs';
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
