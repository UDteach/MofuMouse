import fs from 'node:fs';
import path from 'node:path';
import {fileURLToPath} from 'node:url';

// Preserve artifact directories and collect only distribution files for publishing.
export function normalizeReleaseAssets(root='release-assets') {
  const copied=[];
  for(const directory of ['release-build','release-build-macos12']) {
    const source=path.join(root,directory);
    if(!fs.existsSync(source))continue;
    for(const name of fs.readdirSync(source)) {
      if(!/^(MofuMouse-[\w.-]+\.(exe|zip|dmg)|build-info-(win32|darwin|macos12)\.json)$/.test(name))continue;
      const from=path.join(source,name),to=path.join(root,name);
      if(!fs.lstatSync(from).isFile())throw Error(`Invalid release file: ${name}`);
      if(fs.existsSync(to)) {
        if(!fs.readFileSync(from).equals(fs.readFileSync(to)))throw Error(`Conflicting release file: ${name}`);
      } else {fs.copyFileSync(from,to);copied.push(name);}
    }
  }
  return copied;
}
if(process.argv[1]&&path.resolve(process.argv[1])===fileURLToPath(import.meta.url))console.log(JSON.stringify({copied:normalizeReleaseAssets()}));
