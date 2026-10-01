import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import {execFileSync} from 'node:child_process';
const repo='UDteach/MofuMouse';
const pkg=JSON.parse(fs.readFileSync('electron-prototype/package.json'));
const status=JSON.parse(fs.readFileSync('docs/release-status.json'));
const tag=`v${pkg.version}`;
if(status.version!==pkg.version||(process.env.EXPECTED_TAG&&process.env.EXPECTED_TAG!==tag))throw Error('Published version mismatch');
const release=JSON.parse(execFileSync('gh',['api',`repos/${repo}/releases/tags/${tag}`],{encoding:'utf8'}));
if(release.draft||!release.prerelease||release.tag_name!==tag)throw Error('Preview release is not public');
const assets=new Map(release.assets.map(a=>[a.name,a]));
const suffixes={win32:['win-x64.exe','win-x64.zip'],darwin:['mac-arm64.dmg','mac-arm64.zip','mac-x64.dmg','mac-x64.zip'],macos12:['macos12-arm64.dmg','macos12-arm64.zip','macos12-x64.dmg','macos12-x64.zip']};
const wanted=status.availableProfiles.flatMap(p=>suffixes[p].map(s=>`MofuMouse-${pkg.version}-${s}`));
for(const name of wanted)if(!assets.has(name)||assets.get(name).size<1_000_000)throw Error(`Missing published download: ${name}`);
const temporary=fs.mkdtempSync(path.join(os.tmpdir(),'mofumouse-published-'));
try {
 execFileSync('gh',['release','download',tag,'--repo',repo,'--pattern','SHA256SUMS.txt','--dir',temporary],{stdio:'pipe'});
 const lines=fs.readFileSync(path.join(temporary,'SHA256SUMS.txt'),'utf8').trim().split('\n');
 const checksums=new Map();
 for(const line of lines){const match=/^([a-f0-9]{64})  ([\w.-]+)$/.exec(line);if(!match||checksums.has(match[2]))throw Error('Invalid published checksums');checksums.set(match[2],match[1]);}
 for(const [name,digest] of checksums){const asset=assets.get(name);if(!asset||asset.digest!==`sha256:${digest}`)throw Error(`Published digest mismatch: ${name}`);}
 for(const name of wanted)if(!checksums.has(name))throw Error(`Download not checksummed: ${name}`);
 console.log(JSON.stringify({pass:true,tag,url:release.html_url,availableProfiles:status.availableProfiles,verifiedDownloads:wanted.length,verifiedDigests:checksums.size}));
} finally {
 if(path.dirname(temporary)!==os.tmpdir()||!path.basename(temporary).startsWith('mofumouse-published-'))throw Error('Unexpected temporary path');
 fs.rmSync(temporary,{recursive:true,force:true});
}
