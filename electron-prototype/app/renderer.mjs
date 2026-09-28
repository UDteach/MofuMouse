const canvas = document.querySelector('#animals');
const context = canvas.getContext('2d', { alpha: true });
let images = new Map(), presentations = new Map(), selection = null, loadToken = 0;
let current = null, ready = false, scheduled = false, display = null, lastSignature = '', painted = 0;
const seen = new Set();
let maxPets = 0, lastStats = 0;
const sizes = new Set();
function draw(now) {
  scheduled = false;
  if (!ready || !current || !display) return;
  const dpr = window.devicePixelRatio || 1;
  const width = Math.round(innerWidth * dpr), height = Math.round(innerHeight * dpr);
  if (canvas.width !== width || canvas.height !== height) { canvas.width = width; canvas.height = height; lastSignature = ''; }
  const signature = JSON.stringify(current.pets.map(p => [p.animalId, p.x, p.y, p.height, p.action, p.frame, p.left]));
  if (signature !== lastSignature) {
    context.setTransform(1, 0, 0, 1, 0, 0); context.clearRect(0, 0, width, height);
    context.setTransform(dpr, 0, 0, dpr, 0, 0);
    context.imageSmoothingEnabled = true; context.imageSmoothingQuality = 'high';
    for (const pet of current.pets) {
      const tier = pet.height <= 32 || pet.height === 64 ? 64 : 96;
      const image = images.get(`${pet.animalId}/${pet.action}/${tier}/${pet.frame}`);
      if (!image) throw new Error(`Missing ${pet.action} frame ${pet.frame}`);
      const x = pet.x - display.x, y = pet.y - display.y;
      context.save();
      if (pet.left) { context.translate(x + pet.width, y); context.scale(-1, 1); }
      else context.translate(x, y);
      const transform = presentations.get(`${pet.animalId}/${pet.action}`) ?? { scale: 1, x: 0, y: 0 };
      context.drawImage(image, transform.x * pet.height, transform.y * pet.height, pet.width * transform.scale, pet.height * transform.scale); context.restore();
      seen.add(`${pet.action}:${pet.frame}`); sizes.add(pet.height);
    }
    maxPets = Math.max(maxPets, current.pets.length); painted++; lastSignature = signature;
  }
  if (now - lastStats >= 500) {
    lastStats = now;
    window.mofu.stats({ selection, painted, seen: [...seen], sizes: [...sizes], maxPets, pets: current.pets.length, petAnimals: current.pets.map(p => p.animalId), canvas: [width, height], dpr });
  }
}
function schedule() { if (!scheduled) { scheduled = true; requestAnimationFrame(draw); } }
window.mofu.onInit(async ({ animals, bounds, selection: nextSelection }) => {
  const token = ++loadToken;
  ready = false; current = null; lastSignature = ''; seen.clear();
  context.setTransform(1, 0, 0, 1, 0, 0); context.clearRect(0, 0, canvas.width, canvas.height);
  try {
    const nextImages = new Map(), nextPresentations = new Map();
    for (const animal of animals) for (const [action, motion] of Object.entries(animal.motions)) nextPresentations.set(`${animal.id}/${action}`, motion.presentation ?? { scale: 1, x: 0, y: 0 });
    await Promise.all(animals.flatMap(animal => Object.entries(animal.motions).flatMap(([action, motion]) => Object.entries(motion.tiers).flatMap(([tier, files]) => files.map(async (file, index) => {
      const image = new Image(); image.src = `media/${file}`; await image.decode(); nextImages.set(`${animal.id}/${action}/${tier}/${index}`, image);
    })))));
    if (token !== loadToken) return;
    images = nextImages; presentations = nextPresentations; display = bounds; selection = nextSelection;
    ready = true; window.mofu.loaded({ decoded: images.size, selection }); schedule();
  } catch (e) { if (token === loadToken) window.mofu.failed(e.message); }
});
window.mofu.onFrame(payload => { if (!ready || payload.selection !== selection) return; current = payload; display = payload.bounds; schedule(); });
window.addEventListener('resize', () => { lastSignature = ''; schedule(); });
window.addEventListener('error', e => window.mofu.failed(e.message));
window.addEventListener('unhandledrejection', e => window.mofu.failed(e.reason?.message ?? e.reason));
window.mofu.ready();
