import test from 'node:test';
import assert from 'node:assert/strict';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const { captureCurrent } = require('../app/capture-current.cjs');

function snapshot(id, capturePage) {
  return {
    window: { webContents: { capturePage } },
    rect: { x: 1, y: 2, width: 3, height: 4 },
    animalId: id,
    selection: id
  };
}

function image(label) {
  return { label, isEmpty: () => false };
}

test('a rejection from a retired overlay calls onRetired and recovers with the new owner', async () => {
  const retiredError = new Error('UnknownVizError: old HWND retired');
  const newImage = image('new');
  let current;
  const old = snapshot('old', async () => { current = next; throw retiredError; });
  const next = snapshot('new', async () => newImage);
  current = old;
  const retired = [];

  const result = await captureCurrent({
    snapshot: () => current,
    isCurrent: value => value === current,
    timeoutMs: 1000,
    onRetired: (value, error) => { retired.push({ value, error }); current = next; }
  });

  assert.equal(result.image, newImage);
  assert.equal(result.snapshot, next);
  assert.deepEqual(retired, [{ value: old, error: retiredError }]);
});

test('a successful capture from a retired overlay is discarded before retrying', async () => {
  const staleImage = image('stale');
  const freshImage = image('fresh');
  let current;
  const old = snapshot('old', async () => {
    current = next;
    return staleImage;
  });
  const next = snapshot('new', async () => freshImage);
  current = old;
  const retired = [];

  const result = await captureCurrent({
    snapshot: () => current,
    isCurrent: value => value === current,
    timeoutMs: 1000,
    onRetired: (value, error) => retired.push({ value, error })
  });

  assert.equal(result.image, freshImage);
  assert.equal(result.snapshot, next);
  assert.deepEqual(retired, [{ value: old, error: null }]);
});

test('a current capture failure is propagated exactly and is not retried', async () => {
  const failure = new Error('UnknownVizError: current capture failed');
  let calls = 0;
  const current = snapshot('current', () => { calls++; throw failure; });

  await assert.rejects(
    captureCurrent({ snapshot: () => current, isCurrent: value => value === current, timeoutMs: 1000 }),
    error => error === failure
  );
  assert.equal(calls, 1);
});

test('missing ready snapshot times out within the caller deadline', async () => {
  await assert.rejects(
    captureCurrent({ snapshot: () => null, isCurrent: () => false, timeoutMs: 10 }),
    error => error?.code === 'CAPTURE_TIMEOUT'
  );
});

test('an empty image from the current overlay is rejected', async () => {
  const current = snapshot('current', async () => ({ isEmpty: () => true }));

  await assert.rejects(
    captureCurrent({ snapshot: () => current, isCurrent: value => value === current, timeoutMs: 1000 }),
    error => error?.code === 'CAPTURE_EMPTY'
  );
});

test('repeated retired snapshots stop at the retirement bound', async () => {
  let captureCalls = 0;
  const retired = snapshot('retired', async () => { captureCalls++; return image('never-used'); });
  let callbacks = 0;

  await assert.rejects(
    captureCurrent({
      snapshot: () => retired,
      isCurrent: () => false,
      timeoutMs: 1000,
      onRetired: () => { callbacks++; }
    }),
    error => error?.code === 'CAPTURE_RETIREMENT_LIMIT'
  );
  assert.equal(callbacks, 3);
  assert.equal(captureCalls, 0);
});

test('a hanging onRetired callback is bounded by the remaining deadline', { timeout: 1000 }, async () => {
  const retiredError = new Error('UnknownVizError: retired HWND');
  let current;
  const old = snapshot('old', () => { current = next; throw retiredError; });
  const next = snapshot('new', () => image('new'));
  current = old;

  await assert.rejects(
    captureCurrent({
      snapshot: () => current,
      isCurrent: value => value === current,
      timeoutMs: 15,
      onRetired: () => new Promise(() => {})
    }),
    error => error?.code === 'CAPTURE_TIMEOUT'
  );
});

test('a hanging post-capture isCurrent thenable is bounded by the remaining deadline', { timeout: 1000 }, async () => {
  const current = snapshot('current', () => image('current'));
  let checks = 0;
  const isCurrent = () => {
    checks++;
    return checks === 1 ? true : new Promise(() => {});
  };

  await assert.rejects(
    captureCurrent({ snapshot: () => current, isCurrent, timeoutMs: 15 }),
    error => error?.code === 'CAPTURE_TIMEOUT'
  );
  assert.equal(checks, 2);
});

test('a capture that never settles is bounded by the remaining deadline', { timeout: 1000 }, async () => {
  let captureCalls = 0;
  const current = snapshot('current', () => {
    captureCalls++;
    return new Promise(() => {});
  });

  await assert.rejects(
    captureCurrent({ snapshot: () => current, isCurrent: () => true, timeoutMs: 15 }),
    error => error?.code === 'CAPTURE_TIMEOUT'
  );
  assert.equal(captureCalls, 1);
});
