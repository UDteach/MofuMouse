import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
const pkg=JSON.parse(fs.readFileSync('electron-prototype/package.json')),v=pkg.version;
const source=fs.readFileSync('electron-prototype/app/media/manifest.json'),catalog=JSON.parse(source);
const release=JSON.parse(fs.readFileSync('docs/release-status.json'));
assert.equal(release.version,v);
const suffixByProfile={win32:['win-x64.exe','win-x64.zip'],darwin:['mac-arm64.dmg','mac-arm64.zip','mac-x64.dmg','mac-x64.zip'],macos12:['macos12-arm64.dmg','macos12-arm64.zip','macos12-x64.dmg','macos12-x64.zip']};
assert.deepEqual([...release.availableProfiles,...release.pendingProfiles].sort(),Object.keys(suffixByProfile).sort());
assert(release.availableProfiles.includes('win32'));
const demo=JSON.parse(fs.readFileSync('docs/try/catalog.json'));
assert.equal(demo.version,v);assert.equal(demo.variants.length,catalog.variants.length);assert.equal(demo.variants.filter(v=>v.species==='degu').length,10);
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
assert.equal(hash(fs.readFileSync('docs/try/motion.mjs')),hash(fs.readFileSync('electron-prototype/app/motion.mjs')));
for(const f of demo.files)assert.equal(hash(fs.readFileSync(path.join('docs/try',f.path))),f.sha256,`Demo image changed: ${f.path}`);
const pages=['index.html','download.html','try/index.html'];
let downloads=new Set();
for(const page of pages){const html=fs.readFileSync(`docs/${page}`,'utf8'),dir=path.dirname(`docs/${page}`),ids=new Set([...html.matchAll(/\bid="([^"]+)"/g)].map(m=>m[1]));
 assert(html.includes(`v${v}`));assert(!/19種類|19 appearances|待機中は静止画|制作レーン|台帳|source of truth|(?<![a-z])[A-Z]:[\\/]/i.test(html));
 for(const m of html.matchAll(/(?:src|href)="([^"]+)"/g)){const url=m[1];if(/^https?:/.test(url)){if(url.includes('/releases/download/'))downloads.add(url);continue;}if(url.startsWith('#')){assert(ids.has(url.slice(1)),`Missing anchor: ${page}/${url}`);continue;}const [file,anchor]=url.split('#');const dest=path.resolve(dir,file||'.');assert(dest.startsWith(path.resolve('docs')+path.sep)||dest===path.resolve('docs'));let local=dest;if(fs.existsSync(local)&&fs.statSync(local).isDirectory())local=path.join(local,'index.html');assert(fs.existsSync(local),`Missing public file: ${page}/${url}`);if(anchor)assert(fs.readFileSync(local,'utf8').includes(`id="${anchor}"`));}
 for(const img of html.matchAll(/<img\b[^>]+>/g))assert(/\balt="[^"]*"/.test(img[0]));
}
const base=`https://github.com/UDteach/MofuMouse/releases/download/v${v}/`,suffixes=release.availableProfiles.flatMap(p=>suffixByProfile[p]);
assert.deepEqual([...downloads].sort(),[...suffixes.map(s=>base+`MofuMouse-${v}-${s}`),base+'SHA256SUMS.txt'].sort());
for(const dv of demo.variants){const original=catalog.variants.find(v=>v.id===dv.id);for(const action of ['walk','idle']){assert.deepEqual(dv.motions[action].durations,original.motions[action].durations);assert.deepEqual(dv.motions[action].frames,original.motions[action].tiers[96].map(p=>'media/'+p));assert.deepEqual(dv.motions[action].presentation,original.motions[action].presentation??{scale:1,x:0,y:0});}}
console.log(JSON.stringify({pass:true,version:v,availableProfiles:release.availableProfiles,downloadLinks:downloads.size,variants:demo.variants.length,hashedImages:demo.files.length,sharedMotion:true}));
