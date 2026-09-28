import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { dirname, resolve, sep } from 'node:path';
import { fileURLToPath } from 'node:url';

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..');
const docs = resolve(root, 'docs');
const html = readFileSync(resolve(docs, 'index.html'), 'utf8');
const version = '0.1.0-preview.1';
const releaseBase = `https://github.com/UDteach/MofuMouse/releases/download/v${version}/`;
const hrefs = [...html.matchAll(/\bhref="([^"]+)"/g)].map(match => match[1]);
const expected = [
  `MofuMouse-${version}-win-x64.exe`,
  `MofuMouse-${version}-win-x64.zip`,
  `MofuMouse-${version}-mac-arm64.dmg`,
  `MofuMouse-${version}-mac-arm64.zip`,
  `MofuMouse-${version}-mac-x64.dmg`,
  `MofuMouse-${version}-mac-x64.zip`,
  'SHA256SUMS.txt',
].map(name => releaseBase + name);
const downloads = [...new Set(hrefs.filter(href => href.includes('/releases/download/')))];
assert.deepEqual(downloads.sort(), expected.sort(), 'Preview download contract changed');
assert(hrefs.includes(`https://github.com/UDteach/MofuMouse/releases/tag/v${version}`));
const publicHtml = html.replace(/<style>[\s\S]*?<\/style>/g, '');
assert(!/releases\/latest|download\/MofuMouse-windows|x86|全6毛色|1px|クイックキー|クイックスクロール|dev-local/.test(publicHtml), 'Stale Go release content');
assert(!/ImageGen|台帳|制作レーン|source of truth|\b[A-Z]:[\\/]|prompt/i.test(publicHtml), 'Internal production text exposed');

const ids = new Set([...html.matchAll(/\bid="([^"]+)"/g)].map(match => match[1]));
for (const href of hrefs.filter(href => href.startsWith('#'))) {
  assert(ids.has(href.slice(1)), `Missing anchor target: ${href}`);
}
const images = [...html.matchAll(/<img\b[^>]*>/g)].map(match => match[0]);
assert.equal(images.length, 2);
for (const image of images) {
  assert(/\balt="[^"]*"/.test(image), 'Image missing alternative text');
  const src = image.match(/\bsrc="([^"]+)"/)[1];
  const path = resolve(docs, src);
  assert(path.startsWith(docs + sep), 'Image outside public docs');
  const bytes = readFileSync(path);
  assert(bytes.subarray(0, 8).equals(Buffer.from([137, 80, 78, 71, 13, 10, 26, 10])), 'Invalid PNG');
  let offset = 8;
  while (offset + 12 <= bytes.length) {
    const size = bytes.readUInt32BE(offset);
    assert.notEqual(bytes.toString('ascii', offset + 4, offset + 8), 'acTL', 'Preview must remain static');
    offset += size + 12;
  }
}
for (const text of ['10種・19種類', '1〜10匹', '32・48・64・96px', 'macOS 13以降', 'ブルー／サンド', '待機中は静止画', 'コード署名なし', 'Mac実機の起動・表示は未確認']) {
  assert(html.includes(text), `Missing preview fact: ${text}`);
}
console.log(JSON.stringify({ pass: true, release: `v${version}`, downloadAssets: downloads.length, localImages: images.length, preview: 'static', externalAssetAvailability: 'not checked' }, null, 2));
