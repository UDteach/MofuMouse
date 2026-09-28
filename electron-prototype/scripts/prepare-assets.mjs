import path from 'node:path';
import { fileURLToPath } from 'node:url';
import catalog from '../app/catalog.cjs';
const local = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
// Import pins reviewed assets. Packaging never rereads an active worker ledger.
export async function prepareAssets() {
  const data = catalog.loadCatalog(path.join(local, 'app/media'));
  console.log(JSON.stringify({ assets: data.files.length, variants: data.variants.length, species: new Set(data.variants.map(v => v.species)).size, sourceImagesChanged: false }));
  return data;
}
if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) await prepareAssets();
