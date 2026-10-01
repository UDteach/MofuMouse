import { packager } from '@electron/packager';
import fs from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { prepareAssets } from './prepare-assets.mjs';
const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const [platform = process.platform, arch = process.arch, localLabel] = process.argv.slice(2);
if (!['win32', 'darwin'].includes(platform) || !['x64', 'arm64'].includes(arch)) throw new Error('Only Windows/macOS x64/arm64 targets are supported');
if (localLabel && !/^[a-z0-9]+(?:-[a-z0-9]+)*$/.test(localLabel)) throw new Error('Local build label must use lowercase words and hyphens');
const outputDirectory = path.join(root, 'release', ...(localLabel ? [localLabel] : []));
const catalog = await prepareAssets();
await fs.mkdir(path.join(root, 'stage'), { recursive: true });
const stage = await fs.mkdtemp(path.join(root, 'stage', `${platform}-${arch}-`));
await fs.cp(path.join(root, 'app'), path.join(stage, 'app'), { recursive: true, filter: source => source !== path.join(root, 'app/media') });
for (const file of ['manifest.json', ...catalog.files.map(f => f.path)]) {
  const dest = path.join(stage, 'app/media', file); await fs.mkdir(path.dirname(dest), { recursive: true });
  await fs.copyFile(path.join(root, 'app/media', file), dest);
}
await fs.writeFile(path.join(stage, 'package.json'), `${JSON.stringify({ name: 'mofumouse-electron', productName: 'MofuMouse Electron', version: '0.1.0', main: 'app/main.cjs', private: true }, null, 2)}\n`, 'utf8');
const result = await packager({ dir: stage, name: 'MofuMouseElectron', platform, arch, electronVersion: '44.4.5', out: outputDirectory, overwrite: false, asar: true, prune: false, appBundleId: 'com.kdevelopk.mofumouse.electron.prototype', appCategoryType: 'public.app-category.entertainment', darwinDarkModeSupport: true, executableName: 'MofuMouseElectron' });
if (!result.length) throw new Error('No application was produced. Check the platform warning above. Build macOS targets on a Mac when Windows cannot create symlinks.');
console.log(JSON.stringify({ platform, arch, output: result, signed: false, runtimeVerified: platform === process.platform ? 'pending smoke' : 'requires target OS' }));
