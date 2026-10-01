import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import crypto from 'node:crypto';
import {execFileSync,spawnSync} from 'node:child_process';
import {fileURLToPath} from 'node:url';
import {windowsInputs,verifyWindowsInputs} from '../../scripts/windows-inputs.mjs';
import {normalizeReleaseAssets} from '../../scripts/normalize-release-assets.mjs';

function temporary(prefix,fn) {
 const root=fs.mkdtempSync(path.join(os.tmpdir(),prefix));
 try{return fn(root);}finally{
  assert.equal(path.dirname(path.resolve(root)),path.resolve(os.tmpdir()));
  assert.ok(path.basename(root).startsWith(prefix));
  fs.rmSync(root,{recursive:true,force:true});
 }
}
test('artifact collection preserves nested originals and rejects conflicting downloads',()=>temporary('mofumouse-artifacts-',root=>{
 for(const dir of ['release-build','release-build-macos12'])fs.mkdirSync(path.join(root,dir));
 for(const [dir,name] of [['release-build','build-info-win32.json'],['release-build-macos12','MofuMouse-0.1.0-preview.2-macos12-arm64.zip']])fs.writeFileSync(path.join(root,dir,name),'fixture');
 assert.equal(normalizeReleaseAssets(root).length,2);assert.equal(normalizeReleaseAssets(root).length,0);
 assert.equal(fs.readFileSync(path.join(root,'release-build/build-info-win32.json'),'utf8'),'fixture');
 fs.writeFileSync(path.join(root,'build-info-win32.json'),'different');
 assert.throws(()=>normalizeReleaseAssets(root),/Conflicting/);
}));

test('Windows reuse requires unchanged Git objects and retains the original build commit',()=>temporary('mofumouse-reuse-',root=>{
 const git=args=>execFileSync('git',['-c','core.autocrlf=false',...args],{cwd:root,encoding:'utf8',stdio:['ignore','pipe','pipe']}).trim();
 git(['init','--quiet']);git(['config','user.name','Release test']);git(['config','user.email','release-test@example.invalid']);
 fs.mkdirSync(path.join(root,'electron-prototype/app/media'),{recursive:true});fs.mkdirSync(path.join(root,'electron-prototype/build'));
 const hash=b=>crypto.createHash('sha256').update(b).digest('hex'),version='0.1.0-preview.2';
 const catalog=Buffer.from(JSON.stringify({files:[{path:'fixture.png'}]}));
 fs.writeFileSync(path.join(root,'electron-prototype/app/media/manifest.json'),catalog);
 fs.writeFileSync(path.join(root,'electron-prototype/build/icon.txt'),'icon');
 for(const file of windowsInputs.filter(p=>!['electron-prototype/app','electron-prototype/build'].includes(p))){fs.mkdirSync(path.dirname(path.join(root,file)),{recursive:true});fs.writeFileSync(path.join(root,file),file.endsWith('package.json')?JSON.stringify({version}):'build input');}
 git(['add','.']);git(['commit','--quiet','-m','Original Windows build']);const original=git(['rev-parse','HEAD']);
 fs.writeFileSync(path.join(root,'mac-fix.txt'),'typed config');git(['add','.']);git(['commit','--quiet','-m','Mac-only repair']);const current=git(['rev-parse','HEAD']);
 const inputs=verifyWindowsInputs(original,current,root);assert.equal(inputs.length,windowsInputs.length);
 fs.mkdirSync(path.join(root,'release-assets'));
 for(const [profile,electron,minimumMacOS,suffixes] of [['win32','44.4.5',null,['win-x64.exe','win-x64.zip']],['darwin','44.4.5','13.0',['mac-arm64.dmg','mac-arm64.zip','mac-x64.dmg','mac-x64.zip']],['macos12','43.7.5','12.0',['macos12-arm64.dmg','macos12-arm64.zip','macos12-x64.dmg','macos12-x64.zip']]]){
  const files=suffixes.map(s=>{const name=`MofuMouse-${version}-${s}`,data=Buffer.from(name);fs.writeFileSync(path.join(root,'release-assets',name),data);return {name,bytes:data.length,sha256:hash(data)};});
  fs.writeFileSync(path.join(root,`release-assets/build-info-${profile}.json`),JSON.stringify({version,commit:profile==='win32'?original:current,profile,electron,minimumMacOS,packedSourceMatches:true,verifiedPngsPerArchitecture:1,sourceCatalogSha256:hash(catalog),files}));
 }
 const script=fileURLToPath(new URL('../../scripts/publish-release.mjs',import.meta.url));
 const run=(sha,windowsOnly=false)=>spawnSync(process.execPath,[script,'--check-only',...(windowsOnly?['--windows-only']:[])],{cwd:root,env:{...process.env,EXPECTED_TAG:`v${version}`,GITHUB_SHA:sha},encoding:'utf8'});
 assert.notEqual(run(current).status,0,'cross-commit Windows needs attestation');
 const reusePath=path.join(root,'release-assets/windows-reuse.json'),reuse={status:'pass',originalRun:'123',windowsCommit:original,releaseCommit:current,inputs};
 fs.writeFileSync(reusePath,JSON.stringify(reuse));let result=run(current);assert.equal(result.status,0,result.stderr);assert.equal(JSON.parse(result.stdout).verifiedAssets,15);
 assert.equal(fs.readFileSync(path.join(root,'release-assets/SHA256SUMS.txt'),'utf8').trim().split('\n').length,14);
 assert.equal(JSON.parse(fs.readFileSync(path.join(root,'release-assets/build-info-win32.json'))).commit,original);
 result=run(current,true);assert.equal(result.status,0,result.stderr);assert.equal(JSON.parse(result.stdout).verifiedAssets,5);
 assert.equal(fs.readFileSync(path.join(root,'release-assets/SHA256SUMS.txt'),'utf8').trim().split('\n').length,4);
 const macReport=path.join(root,'release-assets/build-info-darwin.json');fs.renameSync(macReport,macReport+'.hold');
 assert.equal(run(current,true).status,0,'explicit Windows-only release does not require pending Mac binaries');
 assert.notEqual(run(current).status,0,'complete release still requires Mac binaries');fs.renameSync(macReport+'.hold',macReport);
 fs.writeFileSync(reusePath,JSON.stringify({...reuse,inputs:[]}));assert.notEqual(run(current).status,0);fs.writeFileSync(reusePath,JSON.stringify(reuse));
 fs.appendFileSync(path.join(root,'electron-prototype/app/media/manifest.json'),'changed');git(['add','.']);git(['commit','--quiet','-m','Changed runtime input']);const changed=git(['rev-parse','HEAD']);
 assert.throws(()=>verifyWindowsInputs(original,changed,root),/Windows input differs/);
 fs.writeFileSync(reusePath,JSON.stringify({...reuse,releaseCommit:changed}));assert.notEqual(run(changed).status,0);
}));
