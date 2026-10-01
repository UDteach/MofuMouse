const { app, BrowserWindow, Menu, Tray, nativeImage, screen, ipcMain, globalShortcut, dialog } = require('electron');
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const { pathToFileURL } = require('node:url');
const { loadCatalog, resolveAnimalIds, animalMenu } = require('./catalog.cjs');
const { readSettings, writeSettings } = require('./settings-store.cjs');
const { captureCurrent } = require('./capture-current.cjs');
const option = (key, fallback) => process.argv.find(s => s.startsWith(`--${key}=`))?.slice(key.length + 3) ?? fallback;
const catalogSmoke = process.argv.includes('--catalog-smoke');
const smoke = process.argv.includes('--smoke') || catalogSmoke;
const smokeDisplayRebuild = smoke && process.argv.includes('--smoke-display-rebuild');
const reportFile = option('report', path.join(process.cwd(), 'qa/smoke.json'));
const started = performance.now();
app.setName('MofuMouse');
if (process.platform === 'win32') app.setAppUserModelId('com.kdevelopk.mofumouse.electron.prototype');
app.setPath('userData', smoke ? path.join(path.dirname(reportFile), 'profile') : path.join(app.getPath('appData'), 'MofuMouseElectronPrototype'));
const report = { platform: process.platform, architecture: process.arch, electron: process.versions.electron, renderer: 'Electron Canvas / original PNG sequence', startedAt: new Date().toISOString(), errors: [], overlays: [], samples: [], rendererStats: {}, captures: [], settingsChanges: [], sourceVerification: null, focusEvents: 0 };
let model, media, settings, tray, trayMenu, timer, currentDisplay, finishing = false, validate;
let selectedId, petAnimalIds = [], selection = 0, catalogCase = null;
const selectedAnimal = () => media.variants.find(v => v.id === selectedId);
const animalById = id => media.variants.find(v => v.id === id);
const activeAnimals = () => [...new Set(petAnimalIds.slice(0, settings.count))].map(animalById);
const profiles = () => petAnimalIds.map(id => { const a = animalById(id); return { id, walk: a.motions.walk.durations, idle: a.motions.idle.durations }; });
const newModel = () => { const m = new Model({ walk: selectedAnimal().motions.walk.durations, idle: selectedAnimal().motions.idle.durations }, settings); m.setProfiles(profiles()); return m; };
report.catalogSelections = [];
let latest = null;
let scenarioStarted = null;
let nativeProbeWritten = false;
const overlays = new Map(), ready = new Set(), loaded = new Set();
const configFile = path.join(app.getPath('userData'), 'settings.json');
const hash = bytes => crypto.createHash('sha256').update(bytes).digest('hex');
function verifyMedia() {
  const dir = path.join(__dirname, 'media');
  const data = loadCatalog(dir);
  report.sourceVerification = { images: data.files.length, hashMatch: true, animals: data.variants.length, species: new Set(data.variants.map(v => v.species)).size, snapshotAt: data.snapshotAt };
  return data;
}
function saveSettings() {
  if (smoke) return;
  writeSettings(configFile, { ...settings, animalId: selectedId, animalIds: petAnimalIds });
}
function changeSettings(patch) {
  const oldAnimals = activeAnimals().map(a => a.id).join(',');
  const next = validate({ ...settings, ...patch });
  model.configure(next); settings = next; saveSettings(); buildMenu();
  report.settingsChanges.push({ atMs: Math.round(performance.now() - started), ...settings });
  if (settings.paused) for (const win of overlays.values()) win.hide();
  if (oldAnimals !== activeAnimals().map(a => a.id).join(',')) reloadSelection();
  tick();
}
function sendSelection(win) { win.webContents.send('mofu:init', { animals: activeAnimals(), bounds: win.getBounds(), selection }); }
function reloadSelection() {
  selection++; model.setProfiles(profiles()); loaded.clear();
  for (const win of overlays.values()) { win.hide(); if (ready.has(win.id)) sendSelection(win); }
}
function changeAnimal(id, slot = null) {
  if (!media.variants.some(v => v.id === id)) throw new Error('Unknown animal');
  if (slot !== null && (!Number.isInteger(slot) || slot < 0 || slot >= 10)) throw new Error('Invalid animal slot');
  if (slot === null) petAnimalIds.fill(id); else petAnimalIds[slot] = id;
  selectedId = petAnimalIds[0]; reloadSelection();
  saveSettings(); buildMenu(); tick();
}
function buildMenu() {
  const animal = selectedAnimal();
  const mixed = activeAnimals().length > 1;
  trayMenu = Menu.buildFromTemplate([
    { label: 'MofuMouse · 先行版', enabled: false },
    { label: mixed ? `${settings.count}匹の組み合わせ` : `${animal.speciesLabel} / ${animal.coatLabel}`, enabled: false },
    { label: '動物・毛色（全員）', submenu: animalMenu(media, mixed ? null : selectedId, id => changeAnimal(id)) },
    { label: '1匹ずつ選ぶ', submenu: petAnimalIds.slice(0, settings.count).map((id, slot) => {
      const a = animalById(id);
      return { label: `${slot + 1}匹目：${a.speciesLabel} / ${a.coatLabel}`, submenu: animalMenu(media, id, next => changeAnimal(next, slot), `pet-${slot + 1}`) };
    }) },
    { type: 'separator' },
    { label: '表示する数', submenu: Array.from({ length: 10 }, (_, i) => ({ id: `count-${i + 1}`, label: `${i + 1}匹`, type: 'radio', checked: settings.count === i + 1, click: () => changeSettings({ count: i + 1 }) })) },
    { label: '大きさ', submenu: [32, 48, 64, 96].map(size => ({ id: `size-${size}`, label: `${size}px${size === 48 ? '（標準）' : ''}`, type: 'radio', checked: settings.size === size, click: () => changeSettings({ size }) })) },
    { id: 'pause', label: settings.paused ? '表示を再開' : '一時停止', click: () => changeSettings({ paused: !settings.paused }) },
    { label: '使い方', click: () => dialog.showMessageBox({ type: 'info', title: 'MofuMouse', message: '動物がカーソルに付いてきます。', detail: `「1匹ずつ選ぶ」から順番ごとに指定できます。\n「動物・毛色（全員）」ではまとめて変更します。\n終了: ${process.platform === 'darwin' ? 'Command' : 'Ctrl'} + Alt + Shift + Q\n\n${media.variants.length}種類の姿を収録しています。待機が未完成の色は同じ色の静止画で休みます。` }) },
    { type: 'separator' },
    { label: '終了', accelerator: 'CommandOrControl+Alt+Shift+Q', click: () => app.quit() }
  ]);
  tray.setContextMenu(trayMenu);
  tray.setToolTip(`MofuMouse · ${mixed ? '組み合わせ' : animal.speciesLabel + ' / ' + animal.coatLabel} · ${settings.count}匹 · ${settings.size}px${settings.paused ? ' · 一時停止' : ''}`);
}
function trusted(event) { return [...overlays.values()].some(w => !w.isDestroyed() && w.webContents.id === event.sender.id); }
function ensureOverlay(display) {
  if (overlays.has(display.id)) return overlays.get(display.id);
  const win = new BrowserWindow({ ...display.bounds, show: false, frame: false, transparent: true, backgroundColor: '#00000000',
    focusable: false, skipTaskbar: true, hasShadow: false, resizable: false, fullscreenable: false,
    webPreferences: { preload: path.join(__dirname, 'preload.cjs'), contextIsolation: true, nodeIntegration: false, sandbox: true, backgroundThrottling: false } });
  win.setIgnoreMouseEvents(true, { forward: true }); win.setAlwaysOnTop(true, 'floating');
  if (process.platform === 'darwin') win.setVisibleOnAllWorkspaces(true, { visibleOnFullScreen: true });
  win.webContents.setWindowOpenHandler(() => ({ action: 'deny' }));
  win.webContents.on('will-navigate', e => e.preventDefault());
  const isCurrent = () => !win.isDestroyed() && overlays.get(display.id) === win;
  win.webContents.on('render-process-gone', (_event, details) => { if (isCurrent()) fail(`Renderer stopped: ${details.reason}`); });
  win.webContents.on('did-fail-load', (_event, code, description) => { if (isCurrent()) fail(`Page load failed ${code}: ${description}`); });
  win.on('focus', () => report.focusEvents++);
  win.on('closed', () => { if (overlays.get(display.id) === win) overlays.delete(display.id); ready.delete(win.id); loaded.delete(win.id); });
  const handle = win.getNativeWindowHandle();
  report.overlays.push({ id: display.id, windowId: win.id, bounds: display.bounds, workArea: display.workArea, scaleFactor: display.scaleFactor, nativeHandle: handle.length === 8 ? handle.readBigUInt64LE().toString() : handle.readUInt32LE().toString(), transparent: true, clickThroughRequested: true, focusable: win.isFocusable() });
  overlays.set(display.id, win);
  win.loadFile(path.join(__dirname, 'overlay.html')).catch(e => { if (isCurrent()) fail(e.message); });
  return win;
}
function fail(message) {
  report.errors.push(String(message)); console.error(message);
  if (smoke) void finishSmoke(false); else { dialog.showErrorBox('MofuMouse', String(message)); app.quit(); }
}
function diagnosticCursor(t, area) {
  const left = area.x + area.width * 0.27, top = area.y + area.height * 0.44;
  const span = Math.min(400, area.width * 0.40);
  if (t < 2) return { x: Math.round(left + t * 55), y: Math.round(top) };
  if (t < 4) return { x: Math.round(left + 110 + span * Math.sin((t - 2) * Math.PI)), y: Math.round(top + 90 * Math.sin((t - 2) * Math.PI * 2)) };
  return { x: Math.round(left + 110), y: Math.round(top) };
}
let sampledAt = 0;
function tick() {
  if (!model || finishing) return;
  const now = performance.now(), t = scenarioStarted === null ? 0 : (now - scenarioStarted) / 1000;
  let cursor = screen.getCursorScreenPoint();
  if (smoke) cursor = diagnosticCursor(t, screen.getPrimaryDisplay().workArea);
  const display = screen.getDisplayNearestPoint(cursor);
  currentDisplay = display;
  const win = ensureOverlay(display);
  for (const [id, other] of overlays) if (id !== display.id && other.isVisible()) other.hide();
  if (settings.paused) { if (win.isVisible()) win.hide(); return; }
  latest = { ...model.step(now, cursor, display.workArea), bounds: display.bounds, selection };
  if (catalogCase) latest.pets = latest.pets.map(p => ({ ...p, action: catalogCase.action, frame: catalogCase.frame % animalById(p.animalId).motions[catalogCase.action].durations.length }));
  if (loaded.has(win.id)) {
    win.webContents.send('mofu:frame', latest);
    if (!win.isVisible()) win.showInactive();
  }
  if (smoke && scenarioStarted !== null && now - sampledAt >= 250) {
    sampledAt = now;
    report.samples.push({ t: +t.toFixed(2), speed: +latest.speed.toFixed(1), rate: +latest.rate.toFixed(3), settings: { ...settings }, pets: latest.pets.map(p => ({ x: p.x, y: p.y, action: p.action, frame: p.frame })) });
  }
}
async function capture(name) {
  const requestedSelection = selection;
  const { image, snapshot } = await captureCurrent({
    snapshot: () => {
      if (finishing || selection !== requestedSelection) throw new Error(`Capture selection changed: ${name}`);
      const win = currentDisplay && overlays.get(currentDisplay.id);
      if (!win || win.isDestroyed() || !loaded.has(win.id) || !latest?.pets.length) return null;
      const stats = report.rendererStats[win.id];
      if (stats?.selection !== selection || !stats.painted) return null;
      if (catalogCase && !latest.pets.every(p => stats.seen?.includes(`${p.action}:${p.frame}`))) return null;
      const origin = currentDisplay.bounds;
      const left = Math.max(0, Math.floor(Math.min(...latest.pets.map(p => p.x)) - origin.x - 20));
      const top = Math.max(0, Math.floor(Math.min(...latest.pets.map(p => p.y)) - origin.y - 20));
      const right = Math.min(origin.width, Math.ceil(Math.max(...latest.pets.map(p => p.x + p.width)) - origin.x + 20));
      const bottom = Math.min(origin.height, Math.ceil(Math.max(...latest.pets.map(p => p.y + p.height)) - origin.y + 20));
      if (right <= left || bottom <= top) throw new Error(`Empty capture bounds: ${name}`);
      if (catalogSmoke && smokeDisplayRebuild && !report.captureRebuildInjected) {
        report.captureRebuildInjected = { retiredWindowId: win.id };
        queueMicrotask(() => screen.emit('display-metrics-changed', {}, currentDisplay, ['bounds']));
      }
      return { window: win, displayId: currentDisplay.id, selection, animalId: selectedId,
        rect: { x: left, y: top, width: right - left, height: bottom - top } };
    },
    isCurrent: shot => !shot.window.isDestroyed() && overlays.get(shot.displayId) === shot.window && shot.selection === selection,
    onRetired: (shot, error) => {
      report.captureRetirements ??= [];
      report.captureRetirements.push({ name, windowId: shot.window.id, selection: shot.selection, error: error?.message ?? null });
    }
  });
  const file = path.join(path.dirname(reportFile), `${path.basename(reportFile, '.json')}-${name}.png`);
  fs.mkdirSync(path.dirname(file), { recursive: true }); fs.writeFileSync(file, image.toPNG());
  report.captures.push({ name, animalId: snapshot.animalId, windowId: snapshot.window.id, selection: snapshot.selection, path: file, dimensions: image.getSize(), sha256: hash(image.toPNG()) });
}
function smokeSchedule() {
  const menu = id => { const item = trayMenu.getMenuItemById(id); if (!item?.click) return fail(`Missing tray command ${id}`); item.click(); };
  const later = (ms, fn) => setTimeout(() => { if (!finishing) Promise.resolve().then(fn).catch(e => fail(e.message)); }, ms);
  if (smokeDisplayRebuild) later(1000, () => screen.emit('display-metrics-changed', {}, screen.getPrimaryDisplay(), ['workArea']));
  later(3000, () => capture('walking'));
  later(8000, () => capture('idle'));
  later(10000, () => menu('size-32')); later(10500, () => capture('size32'));
  later(11000, () => menu('size-64')); later(11500, () => capture('size64'));
  later(12000, () => menu('size-96')); later(12500, () => capture('size96'));
  later(13000, () => menu('count-1')); later(13500, () => capture('single'));
  later(14000, () => { menu('size-48'); menu('count-10'); }); later(14500, () => capture('ten'));
  later(15000, () => menu('pause')); later(15750, () => { report.pauseHidden = [...overlays.values()].every(w => !w.isVisible()); });
  later(16000, () => menu('pause'));
  later(19000, () => finishSmoke(true));
}
async function finishSmoke(requestedPass) {
  if (finishing) return;
  if (requestedPass && !catalogSmoke) {
    const observed = new Set(Object.values(report.rendererStats).flatMap(s => s.seen ?? []));
    const complete = ['walk', 'idle'].every(action => selectedAnimal().motions[action].durations.every((_, i) => observed.has(`${action}:${i}`)));
    report.coverageWaitStartedAtMs ??= performance.now();
    report.coverageWaitMs = Math.round(performance.now() - report.coverageWaitStartedAtMs);
    // A loaded machine can miss a displayed frame in one cycle. Observe another
    // natural cycle, bounded to eight seconds; never fabricate render coverage.
    if (!complete && report.coverageWaitMs < 8000) { setTimeout(() => void finishSmoke(true), 250); return; }
  }
  finishing = true; clearInterval(timer);
  report.processes = app.getAppMetrics().map(p => ({ type: p.type, cpu: p.cpu, memory: p.memory }));
  report.finishedAt = new Date().toISOString();
  const stats = Object.values(report.rendererStats), frames = new Set(stats.flatMap(s => s.seen ?? []));
  report.checks = {
    sources: report.sourceVerification?.hashMatch === true,
    decoded: stats.some(s => s.decoded === selectedAnimal().frameCount * 2),
    allWalkFrames: selectedAnimal().motions.walk.durations.every((_, i) => frames.has(`walk:${i}`)),
    allIdleFrames: selectedAnimal().motions.idle.durations.every((_, i) => frames.has(`idle:${i}`)),
    tenPets: stats.some(s => s.maxPets === 10),
    allSizes: [32, 48, 64, 96].every(size => stats.some(s => s.sizes?.includes(size))),
    settingsCommands: report.settingsChanges.length >= 8,
    pauseHidesWindows: report.pauseHidden === true,
    distinctTen: report.samples.some(s => s.pets.length === 10 && new Set(s.pets.map(p => `${p.x}:${p.y}`)).size === 10),
    cursorSpeedVaries: report.samples.some(s => s.speed > 600) && report.samples.some(s => s.speed > 0 && s.speed < 200),
    noFocusEvents: report.focusEvents === 0,
    noRendererErrors: report.errors.length === 0,
    screenshots: report.captures.length >= 7,
    walkingCaptureDiffers: report.captures.find(c => c.name === 'walking')?.sha256 !== report.captures.find(c => c.name === 'idle')?.sha256
  };
  if (catalogSmoke) report.checks = report.catalogChecks ?? { completedCatalogTest: false };
  if (smokeDisplayRebuild) report.checks.displayRebuildReprobed = (report.nativeProbeWindows?.length ?? 0) >= 2;
  report.pass = requestedPass && Object.values(report.checks).every(Boolean);
  fs.mkdirSync(path.dirname(reportFile), { recursive: true });
  fs.writeFileSync(reportFile, `${JSON.stringify(report, null, 2)}\n`, 'utf8');
  console.log(JSON.stringify({ report: reportFile, pass: report.pass, checks: report.checks }));
  app.exit(report.pass ? 0 : 1);
}
async function runCatalogSmoke() {
  const delay = ms => new Promise(resolve => setTimeout(resolve, ms));
  const until = async predicate => {
    const deadline = performance.now() + 10000;
    while (!predicate()) { if (finishing) throw new Error('Catalog test stopped'); if (performance.now() > deadline) throw new Error(`Catalog load/paint timeout: ${selectedId}`); await delay(25); }
  };
  const menu = id => { const item = trayMenu.getMenuItemById(id); if (!item) throw new Error(`Missing menu ${id}`); item.click(); };
  menu('count-3'); menu('size-64');
  report.catalogCases = [];
  report.menu = trayMenu.items.find(i => i.label === '動物・毛色（全員）').submenu.items.map(group => ({ label: group.label, items: group.submenu.items.map(i => ({ id: i.id, label: i.label })) }));
  for (const animal of media.variants) {
    catalogCase = { action: 'walk', frame: Math.floor(animal.motions.walk.durations.length / 2) };
    menu(`animal-${animal.id}`);
    await until(() => loaded.size > 0);
    for (const action of ['walk', 'idle']) {
      const frame = action === 'walk' ? Math.floor(animal.motions.walk.durations.length / 2) : animal.motions.idle.durations.length - 1;
      catalogCase = { action, frame }; tick();
      await until(() => Object.values(report.rendererStats).some(s => s.selection === selection && s.seen?.includes(`${action}:${frame}`)));
      await capture(`${animal.id}-${action}`);
    }
    const stats = Object.values(report.rendererStats).find(s => s.selection === selection);
    report.catalogCases.push({ id: animal.id, decoded: stats?.decoded, expectedDecoded: animal.frameCount * 2, settingsRetained: settings.count === 3 && settings.size === 64, checked: trayMenu.getMenuItemById(`animal-${animal.id}`).checked });
  }
  // Three independent slots, followed by ten different species. Exercise the
  // exact native-menu callbacks and wait for the corresponding renderer revision.
  catalogCase = { action: 'walk', frame: 0 };
  const mix = ['degu-agouti', 'chinchilla-beige', 'hamster-dove'];
  for (let i = 0; i < mix.length; i++) menu(`pet-${i + 1}-${mix[i]}`);
  await until(() => loaded.size > 0);
  await until(() => Object.values(report.rendererStats).some(s => s.selection === selection && JSON.stringify(s.petAnimals) === JSON.stringify(mix)));
  await capture('mixed-three-walk');
  catalogCase = { action: 'idle', frame: 0 }; tick(); await delay(600); await capture('mixed-three-idle');
  report.mixedThree = [...petAnimalIds.slice(0, 3)];
  menu('count-1'); await until(() => loaded.size > 0);
  menu('count-3'); await until(() => loaded.size > 0);
  report.hiddenSlotsRetained = JSON.stringify(petAnimalIds.slice(0, 3)) === JSON.stringify(mix);
  menu('count-10');
  const ten = [...new Map(media.variants.map(a => [a.species, media.variants.find(v => v.species === a.species).id])).values()];
  for (let i = 0; i < ten.length; i++) menu(`pet-${i + 1}-${ten[i]}`);
  await until(() => loaded.size > 0);
  await until(() => Object.values(report.rendererStats).some(s => s.selection === selection && new Set(s.petAnimals).size === 10));
  await capture('mixed-ten-idle');
  report.mixedTen = [...petAnimalIds];
  report.mixedDecoded = Object.values(report.rendererStats).find(s => s.selection === selection)?.decoded;
  report.mixedExpectedDecoded = activeAnimals().reduce((sum, a) => sum + a.frameCount * 2, 0);
  // Deliberately replace a still-decoding selection several times. Only the last
  // response may make the overlay visible or become its image cache.
  for (const animal of media.variants.slice(0, 5)) menu(`animal-${animal.id}`);
  const lastId = selectedId, lastSelection = selection;
  await until(() => loaded.size > 0);
  await delay(600);
  report.rapidSelection = report.catalogSelections.at(-1)?.selection === lastSelection && report.catalogSelections.at(-1)?.animalId === lastId;
  menu('pause'); menu(`animal-${media.defaultId}`); await until(() => loaded.size > 0); await delay(100);
  report.pausedSelectionHidden = [...overlays.values()].every(w => !w.isVisible());
  menu('pause'); catalogCase = null; tick(); await delay(150);
  report.catalogChecks = {
    sources: report.sourceVerification.hashMatch,
    everyAnimal: report.catalogCases.length === media.variants.length,
    everyImageDecoded: report.catalogCases.every(c => c.decoded === c.expectedDecoded),
    settingsRetained: report.catalogCases.every(c => c.settingsRetained),
    menuChecks: report.catalogCases.every(c => c.checked),
    walkAndIdleCaptured: report.captures.length === media.variants.length * 2 + 3,
    mixedThree: JSON.stringify(report.mixedThree) === JSON.stringify(['degu-agouti', 'chinchilla-beige', 'hamster-dove']),
    mixedTen: new Set(report.mixedTen).size === 10,
    mixedDecoded: report.mixedDecoded === report.mixedExpectedDecoded,
    hiddenSlotsRetained: report.hiddenSlotsRetained,
    rapidSelection: report.rapidSelection,
    pausedSelectionHidden: report.pausedSelectionHidden,
    resumedVisible: [...overlays.values()].some(w => w.isVisible()),
    noRendererErrors: report.errors.length === 0,
    noFocusEvents: report.focusEvents === 0
  };
  if (smokeDisplayRebuild) {
    report.catalogChecks.captureRebuildRecovered = Boolean(report.captureRebuildInjected) &&
      report.captureRetirements?.some(r => r.windowId === report.captureRebuildInjected.retiredWindowId &&
        report.captures.some(c => c.name === r.name && c.windowId !== r.windowId)) === true;
  }
  await finishSmoke(true);
}
async function start() {
  ({ validateSettings: validate, Model } = await import(pathToFileURL(path.join(__dirname, 'motion.mjs')).href));
  media = verifyMedia();
  let saved = { count: 10, size: 48, paused: false };
  let savedAnimals = {};
  if (!smoke && fs.existsSync(configFile)) { try { const raw = readSettings(configFile); saved = validate(raw); savedAnimals = raw; } catch { /* Invalid saved settings use defaults. */ } }
  petAnimalIds = resolveAnimalIds(media, savedAnimals);
  if (option('animal', null)) petAnimalIds.fill(option('animal'));
  selectedId = petAnimalIds[0];
  if (!media.variants.some(v => v.id === selectedId)) throw new Error('Unknown requested animal');
  settings = validate({ ...saved, count: Number(option('count', saved.count)), size: Number(option('size', saved.size)), paused: false });
  model = newModel();
  const icon = nativeImage.createFromPath(path.join(__dirname, 'media', selectedAnimal().motions.idle.tiers[96][0])).resize({ width: 24, height: 16 });
  tray = new Tray(icon); buildMenu(); tray.on('double-click', () => changeSettings({ paused: !settings.paused }));
  if (process.platform === 'darwin') app.dock.hide();
  ipcMain.on('mofu:ready', event => {
    if (!trusted(event)) return;
    const win = BrowserWindow.fromWebContents(event.sender);
    ready.add(win.id);
    sendSelection(win);
  });
  ipcMain.on('mofu:loaded', (event, info) => {
    if (!trusted(event) || info.selection !== selection) return;
    const win = BrowserWindow.fromWebContents(event.sender); loaded.add(win.id);
    report.rendererStats[win.id] = { decoded: info.decoded, animalId: selectedId };
    report.catalogSelections.push({ animalId: selectedId, animalIds: petAnimalIds.slice(0, settings.count), selection, decoded: info.decoded });
    if (smoke && scenarioStarted === null) {
      // Begin diagnostic motion only after PNG decode. Startup/GPU initialization
      // must not consume the moving portion and accidentally test only idle.
      scenarioStarted = performance.now(); sampledAt = 0; report.samples = [];
      model = newModel();
      report.motionStartedAt = new Date().toISOString();
      if (catalogSmoke) void runCatalogSmoke().catch(e => fail(e.stack || e.message)); else smokeSchedule();
    }
    tick();
  });
  ipcMain.on('mofu:stats', (event, info) => {
    if (!trusted(event) || info.selection !== selection) return;
    const win = BrowserWindow.fromWebContents(event.sender);
    report.rendererStats[win.id] = { ...report.rendererStats[win.id], ...info };
    if (smoke && !nativeProbeWritten && latest?.pets.length && process.platform === 'win32') {
      const pet = latest.pets[0];
      const point = screen.dipToScreenPoint({ x: Math.round(pet.x + pet.width / 2), y: Math.round(pet.y + pet.height / 2) });
      const probe = { processId: process.pid, nativeHandle: win.getNativeWindowHandle().readBigUInt64LE().toString(), point, createdAt: new Date().toISOString() };
      fs.mkdirSync(path.dirname(reportFile), { recursive: true });
      fs.writeFileSync(`${reportFile}.live.json`, `${JSON.stringify(probe, null, 2)}\n`, 'utf8');
      report.nativeProbeWindows ??= [];
      report.nativeProbeWindows.push(win.id);
      nativeProbeWritten = true;
    }
  });
  ipcMain.on('mofu:failed', (event, message) => { if (trusted(event)) fail(message); });
  const rebuild = () => {
    if (smoke) {
      nativeProbeWritten = false;
      fs.rmSync(`${reportFile}.live.json`, { force: true });
    }
    const previous = [...overlays.values()];
    overlays.clear(); loaded.clear(); ready.clear();
    for (const win of previous) win.destroy();
    model.resetLayout = true; tick();
  };
  screen.on('display-added', rebuild); screen.on('display-removed', rebuild); screen.on('display-metrics-changed', rebuild);
  report.exitShortcutRegistered = globalShortcut.register('CommandOrControl+Alt+Shift+Q', () => app.quit());
  Menu.setApplicationMenu(null);
  tick(); timer = setInterval(tick, 16);
}
let Model;
if (!app.requestSingleInstanceLock()) app.quit();
else {
  app.on('second-instance', () => { if (settings?.paused) changeSettings({ paused: false }); });
  app.on('window-all-closed', () => {});
  app.on('will-quit', () => { clearInterval(timer); globalShortcut.unregisterAll(); });
  app.whenReady().then(start).catch(e => fail(e.stack || e.message));
}
