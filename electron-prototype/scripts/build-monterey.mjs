import {spawnSync} from 'node:child_process';
import {createRequire} from 'node:module';
const require=createRequire(import.meta.url);
if(process.platform!=='darwin')throw Error('Build the Monterey packages on macOS');
for(const arch of ['arm64','x64']){
 const result=spawnSync(process.execPath,[require.resolve('electron-builder/out/cli/cli.js'),'--mac','dmg','zip',`--${arch}`,'--publish','never',
  '--config.electronVersion=43.7.5','--config.mac.minimumSystemVersion=12.0','--config.directories.output=release-build-macos12',
  '--config.artifactName=MofuMouse-${version}-macos12-${arch}.${ext}'],{stdio:'inherit'});
 if(result.error)throw result.error;if(result.status!==0)process.exit(result.status??1);
}
