import test from 'node:test';
import assert from 'node:assert/strict';
import crypto from 'node:crypto';
import api from '../app/catalog.cjs';
import { Model } from '../app/motion.mjs';
const bytes = Buffer.from('test image bytes');
const sha256 = crypto.createHash('sha256').update(bytes).digest('hex');
function fixture() {
  const variants = ['degu-blue', 'hamster-cream'].map(id => {
    const files = [64, 96].map(size => ({ path: `${id}/walk/${size}-000.png`, sha256 }));
    const motion = { durations: [100], tiers: { 64: [files[0].path], 96: [files[1].path] } };
    return { id, species: id.split('-')[0], speciesLabel: id.split('-')[0], coatLabel: id.split('-')[1], files, motions: { walk: motion, idle: structuredClone(motion) }, idleFallback: true };
  });
  return { schema: 2, defaultId: 'degu-blue', variants, files: variants.flatMap(v => v.files) };
}
test('old and unavailable saved animals use the default; an available choice persists', () => {
  const catalog = fixture();
  for (const old of [undefined, 'missing', null, '../bad']) assert.equal(api.resolveAnimalId(catalog, old), 'degu-blue');
  assert.equal(api.resolveAnimalId(catalog, 'hamster-cream'), 'hamster-cream');
});
test('per-animal settings migrate old files and retain hidden slots', () => {
  const catalog = fixture();
  assert.deepEqual(api.resolveAnimalIds(catalog, { animalId: 'hamster-cream' }), Array(10).fill('hamster-cream'));
  const saved = { animalIds: ['hamster-cream', 'degu-blue', 'missing', ...Array(7).fill('hamster-cream')] };
  const result = api.resolveAnimalIds(catalog, JSON.parse(JSON.stringify(saved)));
  assert.equal(result.length, 10); assert.equal(result[0], 'hamster-cream'); assert.equal(result[2], 'degu-blue'); assert.equal(result[9], 'hamster-cream');
  const one = api.animalMenu(catalog, result[0], () => {}, 'pet-1');
  const two = api.animalMenu(catalog, result[1], () => {}, 'pet-2');
  assert.notEqual(one[0].submenu[0].id, two[0].submenu[0].id);
});
test('mixed animals use their own timelines and changing one does not reset its neighbor', () => {
  const model = new Model({ walk: [100], idle: [1000] }, { count: 3, size: 64, paused: false });
  const a = { id: 'a', walk: [50,50,50,50], idle: [100,100] };
  const b = { id: 'b', walk: [250,250], idle: [1000] };
  const profiles = Array.from({ length: 10 }, (_, i) => i === 1 ? b : a);
  model.setProfiles(profiles);
  const area = { x: 0, y: 0, width: 1280, height: 720 };
  model.step(0, { x: 500, y: 300 }, area);
  let state = model.step(175, { x: 500, y: 300 }, area);
  assert.equal(state.pets[0].frame, 3); assert.equal(state.pets[1].frame, 0);
  assert.deepEqual(state.pets.map(p => p.animalId), ['a', 'b', 'a']);
  const clock = model.members[0].clock, trail = model.trail.points;
  model.setProfiles(profiles.map((p,i) => i === 1 ? a : p));
  state = model.step(190, { x: 500, y: 300 }, area);
  assert.equal(model.members[0].clock, clock); assert.equal(model.trail.points, trail); assert.equal(state.pets[1].animalId, 'a');
});
test('nested menus select exactly one animal and preserve static-idle disclosure', () => {
  let chosen;
  const menu = api.animalMenu(fixture(), 'hamster-cream', id => { chosen = id; });
  const entries = menu.flatMap(g => g.submenu);
  assert.equal(entries.filter(e => e.checked).length, 1);
  assert.ok(entries.every(e => e.label.includes('待機は静止画')));
  entries[0].click(); assert.equal(chosen, 'degu-blue');
});
test('asset validation rejects changed files, missing frames and cross-animal fallback', () => {
  const catalog = fixture(); api.verifyCatalog(catalog, () => bytes);
  assert.throws(() => api.verifyCatalog(catalog, () => Buffer.from('changed')), /changed/);
  const cross = structuredClone(catalog); cross.variants[0].motions.idle.tiers[64][0] = cross.variants[1].files[0].path;
  assert.throws(() => api.verifyCatalog(cross, () => bytes), /cross-animal/);
  const missing = structuredClone(catalog); missing.variants[0].motions.walk.durations.push(100);
  assert.throws(() => api.verifyCatalog(missing, () => bytes), /Missing/);
  const duplicate = structuredClone(catalog); duplicate.variants.push(duplicate.variants[0]);
  assert.throws(() => api.verifyCatalog(duplicate, () => bytes), /duplicate animal/);
});
test('different motion lengths and static idle preserve source timing', () => {
  for (const walk of [[74, 74, 74, 56, 74, 74, 74, 56], Array(22).fill(42)]) {
    const model = new Model({ walk, idle: [1000] });
    let state;
    for (let now = 0; now < 8000; now += 16) state = model.step(now, { x: 400, y: 300 }, { x: 0, y: 0, width: 1280, height: 720 });
    assert.ok(state.pets.every(p => p.action === 'idle' && p.frame === 0));
  }
  assert.throws(() => new Model({ walk: [], idle: [1] }), /Invalid/);
});

test('catalog binds tier and action, allowing only the declared static idle frame', () => {
  const wrongTier = fixture(); wrongTier.variants[0].motions.walk.tiers[64][0] = wrongTier.variants[0].motions.walk.tiers[96][0];
  assert.throws(() => api.verifyCatalog(wrongTier, () => bytes), /Wrong action or tier/);
  const wrongAction = fixture(); wrongAction.variants[0].idleFallback = false;
  assert.throws(() => api.verifyCatalog(wrongAction, () => bytes), /Wrong action or tier/);
  const swappedFallback = fixture(); swappedFallback.variants[0].motions.idle.tiers[64][0] = swappedFallback.variants[0].motions.walk.tiers[96][0];
  assert.throws(() => api.verifyCatalog(swappedFallback, () => bytes), /Wrong action or tier/);
  api.verifyCatalog(fixture(), () => bytes);
});
