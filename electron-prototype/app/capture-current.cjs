'use strict';

const NO_SNAPSHOT_WAIT_MS = 25;
const MAX_RETIREMENTS = 3;

function timeoutError(timeoutMs) {
  const error = new Error(`captureCurrent timed out after ${timeoutMs}ms waiting for a current painted overlay`);
  error.code = 'CAPTURE_TIMEOUT';
  return error;
}

function emptyImageError() {
  const error = new Error('captureCurrent received an empty image from the current overlay');
  error.code = 'CAPTURE_EMPTY';
  return error;
}

function retirementLimitError() {
  const error = new Error(`captureCurrent exceeded the ${MAX_RETIREMENTS}-retirement limit`);
  error.code = 'CAPTURE_RETIREMENT_LIMIT';
  return error;
}

function delay(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

function isThenable(value) {
  return value !== null && (typeof value === 'object' || typeof value === 'function') && typeof value.then === 'function';
}

/**
 * Capture one still image from the current decoded/painted overlay snapshot.
 * onRetired receives (snapshot, error); error is null when a successful image
 * becomes stale before the post-capture identity check.
 */
async function captureCurrent({ snapshot, isCurrent, timeoutMs = 10000, onRetired } = {}) {
  if (typeof snapshot !== 'function') throw new TypeError('captureCurrent requires snapshot()');
  if (typeof isCurrent !== 'function') throw new TypeError('captureCurrent requires isCurrent(snapshot)');
  if (!Number.isFinite(timeoutMs) || timeoutMs < 0) throw new RangeError('captureCurrent timeoutMs must be a non-negative finite number');

  const deadline = performance.now() + timeoutMs;
  let retirements = 0;

  const timedOut = () => performance.now() >= deadline;
  const remaining = () => Math.max(0, deadline - performance.now());
  const waitFor = async (promise, limitMs) => {
    if (limitMs <= 0) throw timeoutError(timeoutMs);
    let timer;
    try {
      return await Promise.race([
        Promise.resolve(promise),
        new Promise((_, reject) => { timer = setTimeout(() => reject(timeoutError(timeoutMs)), limitMs); })
      ]);
    } finally {
      if (timer !== undefined) clearTimeout(timer);
    }
  };
  const checkCurrent = currentSnapshot => {
    // Invoke the caller-owned guard synchronously. This keeps a synchronous
    // true result available to preserve an exact current-window capture error
    // even when the deadline has just been reached.
    const result = isCurrent(currentSnapshot);
    return isThenable(result) ? waitFor(result, remaining()) : result;
  };
  const retire = async (currentSnapshot, error) => {
    retirements++;
    if (retirements > MAX_RETIREMENTS) throw retirementLimitError();
    if (onRetired) {
      const result = onRetired(currentSnapshot, error ?? null);
      if (isThenable(result)) await waitFor(result, remaining());
    }
  };

  while (true) {
    if (timedOut()) throw timeoutError(timeoutMs);

    const currentSnapshot = await waitFor(Promise.resolve().then(() => snapshot()), remaining());
    if (!currentSnapshot) {
      await delay(Math.min(NO_SNAPSHOT_WAIT_MS, remaining()));
      continue;
    }

    // The caller owns identity and selection validity. Never invoke capturePage
    // for a snapshot that is already known to be retired.
    if (!await checkCurrent(currentSnapshot)) {
      await retire(currentSnapshot, null);
      continue;
    }

    let image;
    try {
      image = await waitFor(currentSnapshot.window.webContents.capturePage(currentSnapshot.rect), remaining());
    } catch (error) {
      // Preserve the capture error exactly when the caller still owns this
      // snapshot. The post-failure identity check must run even at the deadline.
      if (await checkCurrent(currentSnapshot)) throw error;
      await retire(currentSnapshot, error);
      continue;
    }

    // A capture can resolve after the overlay/window was replaced. Discard that
    // NativeImage and ask the caller for a fresh current snapshot.
    if (!await checkCurrent(currentSnapshot)) {
      await retire(currentSnapshot, null);
      continue;
    }
    if (timedOut()) throw timeoutError(timeoutMs);
    if (!image || (typeof image.isEmpty === 'function' && image.isEmpty())) throw emptyImageError();
    return { image, snapshot: currentSnapshot };
  }
}

module.exports = { captureCurrent };
