import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { spawnSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';
import store from '../app/settings-store.cjs';
test('ten individual choices and hidden slots survive file replacement and a fresh process', () => {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), 'mofumouse-settings-'));
  try {
    const file = path.join(root, 'profile', 'settings.json');
    const settings = { count: 10, size: 64, paused: false, followMode: 'relaxed', animalId: 'degu-agouti', animalIds: ['degu-agouti', ...Array(9).fill('hamster-dove')] };
    store.writeSettings(file, settings);
    store.writeSettings(file, { ...settings, count: 1, size: 32 });
    const module = fileURLToPath(new URL('../app/settings-store.cjs', import.meta.url));
    const result = spawnSync(process.execPath, ['-e', 'process.stdout.write(JSON.stringify(require(process.argv[1]).readSettings(process.argv[2])))', module, file], { encoding: 'utf8' });
    assert.equal(result.status, 0, result.stderr);
    assert.deepEqual(JSON.parse(result.stdout), { ...settings, count: 1, size: 32 });
    assert.equal(fs.existsSync(`${file}.tmp`), false);
    fs.writeFileSync(file, '{invalid', 'utf8');
    assert.throws(() => store.readSettings(file), SyntaxError);
  } finally {
    assert.equal(path.dirname(path.resolve(root)), path.resolve(os.tmpdir()));
    assert.ok(path.basename(root).startsWith('mofumouse-settings-'));
    fs.rmSync(root, { recursive: true, force: true });
  }
});
