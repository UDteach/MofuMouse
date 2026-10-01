import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import crypto from 'node:crypto';
import {fileURLToPath} from 'node:url';
import {spawnSync} from 'node:child_process';
test('release refuses missing compatibility builds, wrong runtime/OS, cross-commit reports and corrupt assets',()=>{
 const root=fs.mkdtempSync(path.join(os.tmpdir(),'mofumouse-release-'));
 try{
  fs.mkdirSync(path.join(root,'electron-prototype/app/media'),{recursive:true});fs.mkdirSync(path.join(root,'release-assets'));
  const version='0.1.0-preview.2',commit='test-commit',hash=b=>crypto.createHash('sha256').update(b).digest('hex');
  const catalog=Buffer.from(JSON.stringify({files:[{path:'example'}]}));
  fs.writeFileSync(path.join(root,'electron-prototype/package.json'),JSON.stringify({version}));fs.writeFileSync(path.join(root,'electron-prototype/app/media/manifest.json'),catalog);
  const profiles=[['win32','44.4.5',null,['win-x64.exe','win-x64.zip']],['darwin','44.4.5','13.0',['mac-arm64.dmg','mac-arm64.zip','mac-x64.dmg','mac-x64.zip']],['macos12','43.7.5','12.0',['macos12-arm64.dmg','macos12-arm64.zip','macos12-x64.dmg','macos12-x64.zip']]];
  for(const [profile,electron,minimumMacOS,suffixes] of profiles){const files=suffixes.map(s=>{const name=`MofuMouse-${version}-${s}`,data=Buffer.from(name);fs.writeFileSync(path.join(root,'release-assets',name),data);return {name,bytes:data.length,sha256:hash(data)};});fs.writeFileSync(path.join(root,`release-assets/build-info-${profile}.json`),JSON.stringify({version,commit,profile,electron,minimumMacOS,packedSourceMatches:true,verifiedPngsPerArchitecture:1,sourceCatalogSha256:hash(catalog),files}));}
  const script=fileURLToPath(new URL('../../scripts/publish-release.mjs',import.meta.url)),run=()=>spawnSync(process.execPath,[script,'--check-only'],{cwd:root,env:{...process.env,EXPECTED_TAG:`v${version}`,GITHUB_SHA:commit},encoding:'utf8'});
  let result=run();assert.equal(result.status,0,result.stderr);assert.equal(JSON.parse(result.stdout).verifiedAssets,14);assert.equal(fs.readFileSync(path.join(root,'release-assets/SHA256SUMS.txt'),'utf8').trim().split('\n').length,13);
  const reportPath=path.join(root,'release-assets/build-info-macos12.json'),original=fs.readFileSync(reportPath,'utf8');
  for(const patch of [{electron:'44.4.5'},{minimumMacOS:'13.0'},{commit:'other'},{files:[]},{sourceCatalogSha256:'wrong'}]){fs.writeFileSync(reportPath,JSON.stringify({...JSON.parse(original),...patch}));result=run();assert.notEqual(result.status,0);}
  fs.writeFileSync(reportPath,original);fs.renameSync(reportPath,reportPath+'.hold');assert.notEqual(run().status,0);fs.renameSync(reportPath+'.hold',reportPath);
  fs.appendFileSync(path.join(root,`release-assets/MofuMouse-${version}-win-x64.zip`),'corrupt');result=run();assert.notEqual(result.status,0);assert.match(result.stderr,/hash mismatch/);
 }finally{assert.equal(path.dirname(path.resolve(root)),path.resolve(os.tmpdir()));assert.ok(path.basename(root).startsWith('mofumouse-release-'));fs.rmSync(root,{recursive:true,force:true});}
});
