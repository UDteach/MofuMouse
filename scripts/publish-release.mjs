import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import {execFileSync} from 'node:child_process';
import {verifyWindowsInputs} from './windows-inputs.mjs';
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const pkg=JSON.parse(fs.readFileSync('electron-prototype/package.json','utf8'));
const catalogBytes=fs.readFileSync('electron-prototype/app/media/manifest.json');
const catalog=JSON.parse(catalogBytes),tag=process.env.EXPECTED_TAG;
if(tag!==`v${pkg.version}`||!/^v\d+\.\d+\.\d+-preview\.\d+$/.test(tag))throw Error('Preview tag mismatch');
const allProfiles=[{id:'win32',electron:'44.4.5',minimum:null,suffixes:['win-x64.exe','win-x64.zip']},
 {id:'darwin',electron:'44.4.5',minimum:'13.0',suffixes:['mac-arm64.dmg','mac-arm64.zip','mac-x64.dmg','mac-x64.zip']},
 {id:'macos12',electron:'43.7.5',minimum:'12.0',suffixes:['macos12-arm64.dmg','macos12-arm64.zip','macos12-x64.dmg','macos12-x64.zip']}];
const profiles=process.argv.includes('--windows-only')?allProfiles.filter(p=>p.id==='win32'):allProfiles;
const expected=profiles.flatMap(p=>p.suffixes.map(s=>`MofuMouse-${pkg.version}-${s}`));
let windowsReuse=false;
const reports=profiles.map(p=>{
 const report=JSON.parse(fs.readFileSync(`release-assets/build-info-${p.id}.json`));
 if(report.commit!==process.env.GITHUB_SHA){
  if(p.id!=='win32')throw Error(`Unverified build report: ${p.id}`);
  const reuse=JSON.parse(fs.readFileSync('release-assets/windows-reuse.json'));
  if(reuse.status!=='pass'||reuse.windowsCommit!==report.commit||reuse.releaseCommit!==process.env.GITHUB_SHA)throw Error('Windows reuse provenance differs');
  const inputs=verifyWindowsInputs(report.commit,process.env.GITHUB_SHA);
  if(JSON.stringify(inputs)!==JSON.stringify(reuse.inputs))throw Error('Windows reuse inputs differ');
  windowsReuse=true;
 }
 if(report.version!==pkg.version||report.profile!==p.id||report.electron!==p.electron||report.minimumMacOS!==p.minimum||!report.packedSourceMatches||report.verifiedPngsPerArchitecture!==catalog.files.length||report.sourceCatalogSha256!==hash(catalogBytes))throw Error(`Unverified build report: ${p.id}`);
 const names=report.files.map(f=>f.name).sort(),wanted=p.suffixes.map(s=>`MofuMouse-${pkg.version}-${s}`).sort();
 if(JSON.stringify(names)!==JSON.stringify(wanted))throw Error(`Incomplete release profile: ${p.id}`);
 return report;
});
const actual=reports.flatMap(r=>r.files);
if(actual.length!==expected.length||expected.some(n=>actual.filter(f=>f.name===n).length!==1))throw Error('Incomplete release');
for(const file of actual){const bytes=fs.readFileSync(path.join('release-assets',file.name));if(bytes.length!==file.bytes||hash(bytes)!==file.sha256)throw Error(`Release hash mismatch: ${file.name}`);}
const attachments=[...expected,...profiles.map(p=>`build-info-${p.id}.json`)];
if(windowsReuse)attachments.push('windows-reuse.json');
fs.writeFileSync('release-assets/SHA256SUMS.txt',attachments.map(name=>`${hash(fs.readFileSync(path.join('release-assets',name)))}  ${name}\n`).join(''));
attachments.push('SHA256SUMS.txt');
if(process.argv.includes('--check-only'))console.log(JSON.stringify({tag,verifiedAssets:attachments.length,publish:false}));
else execFileSync('gh',['release','create',tag,...attachments.map(n=>path.join('release-assets',n)),'--repo',process.env.GITHUB_REPOSITORY,'--target',process.env.GITHUB_SHA,'--prerelease','--title',`MofuMouse ${tag} — デグー全10色・Webデモ`,'--notes-file','docs/publishing/preview-release-notes.md'],{stdio:'inherit'});
