import {execFileSync} from 'node:child_process';
export const windowsInputs=['electron-prototype/app','electron-prototype/build','electron-prototype/package.json','electron-prototype/package-lock.json','electron-prototype/electron-builder.yml','electron-prototype/scripts/prepare-assets.mjs'];
export function verifyWindowsInputs(original,current,cwd=process.cwd()){
 if(![original,current].every(s=>/^[a-f0-9]{40}$/.test(s)))throw Error('Invalid source commit');
 return windowsInputs.map(input=>{
  const object=ref=>execFileSync('git',['rev-parse',`${ref}:${input}`],{cwd,encoding:'utf8'}).trim();
  const before=object(original),after=object(current);
  if(before!==after)throw Error(`Windows input differs: ${input}`);
  return {path:input,originalObject:before,currentObject:after};
 });
}
