import test from 'node:test';
import assert from 'node:assert/strict';
import {createRequire} from 'node:module';
import {fileURLToPath} from 'node:url';
const require=createRequire(import.meta.url);
test('Monterey configuration preserves the typed minimum OS string and exact runtime',async()=>{
 const {getConfig}=require('app-builder-lib/out/util/config/config');
 const root=fileURLToPath(new URL('../',import.meta.url));
 const c=await getConfig(root,'electron-builder.monterey.yml',undefined);
 assert.equal(c.mac.minimumSystemVersion,'12.0');assert.equal(typeof c.mac.minimumSystemVersion,'string');
 assert.equal(c.electronVersion,'43.7.5');assert.equal(c.mac.identity,'-');assert.equal(c.directories.output,'release-build-macos12');
 assert.equal(c.artifactName,'MofuMouse-${version}-macos12-${arch}.${ext}');
});
