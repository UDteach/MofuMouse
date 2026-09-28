import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import crypto from 'node:crypto';
import { fileURLToPath } from 'node:url';
import { spawnSync } from 'node:child_process';
test('publication dry check requires all six hash-matched downloads from one commit', () => {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), 'mofumouse-release-'));
  try {
    fs.mkdirSync(path.join(root, 'electron-prototype'));
    fs.mkdirSync(path.join(root, 'release-assets'));
    const version = '0.1.0-preview.1', commit = 'test-commit';
    fs.writeFileSync(path.join(root, 'electron-prototype/package.json'), JSON.stringify({ version }));
    const hash = b => crypto.createHash('sha256').update(b).digest('hex');
    for (const platform of ['win32', 'darwin']) {
      const suffixes = platform === 'win32' ? ['win-x64.exe', 'win-x64.zip'] : ['mac-arm64.dmg', 'mac-arm64.zip', 'mac-x64.dmg', 'mac-x64.zip'];
      const files = suffixes.map(suffix => {
        const name = `MofuMouse-${version}-${suffix}`, data = Buffer.from(`test fixture ${name}`);
        fs.writeFileSync(path.join(root, 'release-assets', name), data);
        return { name, bytes: data.length, sha256: hash(data) };
      });
      fs.writeFileSync(path.join(root, `release-assets/build-info-${platform}.json`), JSON.stringify({ version, commit, packedSourceMatches: true, verifiedPngsPerArchitecture: 2650, sourceCatalogSha256: 'same-catalog', files }));
    }
    const script = fileURLToPath(new URL('../../scripts/publish-preview.mjs', import.meta.url));
    const run = () => spawnSync(process.execPath, [script, '--check-only'], { cwd: root, env: { ...process.env, EXPECTED_TAG: `v${version}`, GITHUB_SHA: commit }, encoding: 'utf8' });
    let result = run(); assert.equal(result.status, 0, result.stderr); assert.equal(JSON.parse(result.stdout).publish, false);
    assert.equal(fs.readFileSync(path.join(root, 'release-assets/SHA256SUMS.txt'), 'utf8').trim().split('\n').length, 8);
    fs.appendFileSync(path.join(root, `release-assets/MofuMouse-${version}-win-x64.zip`), 'corrupted');
    result = run(); assert.notEqual(result.status, 0); assert.match(result.stderr, /hash mismatch/);
  } finally {
    assert.equal(path.dirname(path.resolve(root)), path.resolve(os.tmpdir()));
    assert.ok(path.basename(root).startsWith('mofumouse-release-'));
    fs.rmSync(root, { recursive: true, force: true });
  }
});
