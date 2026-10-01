import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import catalogApi from '../electron-prototype/app/catalog.cjs';
const source = 'electron-prototype/app/media', out = 'docs/try';
const catalog = catalogApi.loadCatalog(source);
const hash = b => crypto.createHash('sha256').update(b).digest('hex');
fs.mkdirSync(`${out}/media`, { recursive: true });
const variants = catalog.variants.map(v => ({ id:v.id, species:v.species, speciesLabel:v.speciesLabel, coatLabel:v.coatLabel,
  motions:Object.fromEntries(Object.entries(v.motions).map(([action,m]) => [action, {
    durations:m.durations, presentation:m.presentation ?? {scale:1,x:0,y:0}, frames:m.tiers[96].map(p=>`media/${p}`)
  }])) }));
const files = catalog.files.filter(f => /\/(?:walk|idle)\/96-/.test(f.path));
for (const file of files) {
  const bytes=fs.readFileSync(path.join(source,file.path));
  if(hash(bytes)!==file.sha256) throw Error(`Demo source changed: ${file.path}`);
  const dest=path.join(out,'media',file.path); fs.mkdirSync(path.dirname(dest),{recursive:true}); fs.writeFileSync(dest,bytes);
}
fs.copyFileSync('electron-prototype/app/motion.mjs',`${out}/motion.mjs`);
fs.writeFileSync(`${out}/catalog.json`,JSON.stringify({version:JSON.parse(fs.readFileSync('electron-prototype/package.json')).version,defaultId:catalog.defaultId,variants,files:files.map(f=>({path:`media/${f.path}`,sha256:f.sha256}))})+'\n');
console.log(JSON.stringify({demoVariants:variants.length,images:files.length,sourcePixelsChanged:false}));
