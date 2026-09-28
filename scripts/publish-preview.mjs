import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import { execFileSync } from 'node:child_process';
const hash = b => crypto.createHash('sha256').update(b).digest('hex');
const pkg = JSON.parse(fs.readFileSync('electron-prototype/package.json', 'utf8'));
const tag = process.env.EXPECTED_TAG;
if (tag !== `v${pkg.version}` || !/^v\d+\.\d+\.\d+-preview\.\d+$/.test(tag)) throw new Error('Preview tag mismatch');
const expected = ['win-x64.exe', 'win-x64.zip', 'mac-arm64.dmg', 'mac-arm64.zip', 'mac-x64.dmg', 'mac-x64.zip'].map(x => `MofuMouse-${pkg.version}-${x}`);
const reports = ['win32', 'darwin'].map(platform => JSON.parse(fs.readFileSync(`release-assets/build-info-${platform}.json`, 'utf8')));
for (const report of reports) {
  if (report.version !== pkg.version || report.commit !== process.env.GITHUB_SHA || !report.packedSourceMatches || report.verifiedPngsPerArchitecture !== 2650) throw new Error('Unverified build report');
}
if (new Set(reports.map(r => r.sourceCatalogSha256)).size !== 1) throw new Error('Platforms contain different media');
const actual = reports.flatMap(r => r.files);
if (actual.length !== expected.length || expected.some(name => actual.filter(f => f.name === name).length !== 1)) throw new Error('Incomplete release');
for (const file of actual) {
  const bytes = fs.readFileSync(path.join('release-assets', file.name));
  if (bytes.length !== file.bytes || hash(bytes) !== file.sha256) throw new Error(`Release hash mismatch: ${file.name}`);
}
const attachments = [...expected, 'build-info-win32.json', 'build-info-darwin.json'];
fs.writeFileSync('release-assets/SHA256SUMS.txt', attachments.map(name => `${hash(fs.readFileSync(path.join('release-assets', name)))}  ${name}\n`).join(''), 'utf8');
attachments.push('SHA256SUMS.txt');
if (process.argv.includes('--check-only')) console.log(JSON.stringify({ tag, verifiedAssets: attachments.length, publish: false }));
else execFileSync('gh', ['release', 'create', tag, ...attachments.map(name => path.join('release-assets', name)), '--repo', process.env.GITHUB_REPOSITORY, '--target', process.env.GITHUB_SHA, '--prerelease', '--title', `MofuMouse ${tag} — 先行版`, '--notes-file', 'docs/publishing/preview-release-notes.md'], { stdio: 'inherit' });
