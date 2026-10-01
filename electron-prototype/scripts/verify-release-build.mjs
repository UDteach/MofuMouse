import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import { createRequire } from 'node:module';
const require = createRequire(import.meta.url), asar = require('@electron/asar');
const pkg = JSON.parse(fs.readFileSync('package.json', 'utf8'));
const catalog = JSON.parse(fs.readFileSync('app/media/manifest.json', 'utf8'));
const hash = data => crypto.createHash('sha256').update(data).digest('hex');
const platform = process.platform;
const monterey = process.argv.includes('--monterey');
if(monterey && platform!=='darwin')throw Error('Monterey verification requires macOS');
const directory = monterey ? 'release-build-macos12' : 'release-build';
const runtimeVersion = monterey ? '43.7.5' : pkg.devDependencies.electron;
if (!['win32', 'darwin'].includes(platform)) throw new Error('Verify on the build platform');
const architectures = platform === 'win32' ? ['x64'] : ['arm64', 'x64'];
const platformName = platform === 'win32' ? 'win' : monterey ? 'macos12' : 'mac';
for (const arch of architectures) {
  const archive = platform === 'win32' ? `${directory}/win-unpacked/resources/app.asar`
    : `${directory}/${arch === 'arm64' ? 'mac-arm64' : 'mac'}/MofuMouse.app/Contents/Resources/app.asar`;
  const packaged = JSON.parse(asar.extractFile(archive, 'package.json').toString('utf8'));
  if (packaged.version !== pkg.version) throw new Error('Packaged version mismatch');
  const sources = fs.readdirSync('app').filter(f => fs.statSync(path.join('app', f)).isFile());
  for (const f of [...sources, path.join('media', 'manifest.json')]) {
    if (hash(asar.extractFile(archive, path.join('app', f))) !== hash(fs.readFileSync(path.join('app', f)))) throw new Error(`Packaged source mismatch: ${f}`);
  }
  for (const file of catalog.files) if (hash(asar.extractFile(archive, path.join('app', 'media', file.path))) !== file.sha256) throw new Error(`Packaged image mismatch: ${file.path}`);
}
const files = architectures.flatMap(arch => (platform === 'win32' ? ['exe', 'zip'] : ['dmg', 'zip']).map(ext => `MofuMouse-${pkg.version}-${platformName}-${arch}.${ext}`));
const records = files.map(name => {
  const data = fs.readFileSync(path.join(directory, name));
  if (data.length < 1_000_000) throw new Error(`Unexpected small release file: ${name}`);
  return { name, bytes: data.length, sha256: hash(data) };
});
const report = { version: pkg.version, commit: process.env.GITHUB_SHA ?? null, platform, architectures,
  profile: monterey ? 'macos12' : platform, electron: runtimeVersion, minimumMacOS: platform==='darwin' ? (monterey ? '12.0' : '13.0') : null,
  sourceCatalogSha256: hash(fs.readFileSync('app/media/manifest.json')),
  verifiedPngsPerArchitecture: catalog.files.length, packedSourceMatches: true, files: records,
  physicalMacRuntimeVerified: false };
fs.writeFileSync(`${directory}/build-info-${monterey ? 'macos12' : platform}.json`, `${JSON.stringify(report, null, 2)}\n`, 'utf8');
console.log(JSON.stringify({ platform, architectures, images: catalog.files.length, files: records.map(f => f.name), pass: true }));
