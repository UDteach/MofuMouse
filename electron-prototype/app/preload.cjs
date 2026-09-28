const { contextBridge, ipcRenderer } = require('electron');
contextBridge.exposeInMainWorld('mofu', {
  onInit: handler => ipcRenderer.on('mofu:init', (_event, data) => handler(data)),
  onFrame: handler => ipcRenderer.on('mofu:frame', (_event, data) => handler(data)),
  ready: () => ipcRenderer.send('mofu:ready'),
  loaded: info => ipcRenderer.send('mofu:loaded', info),
  stats: info => ipcRenderer.send('mofu:stats', info),
  failed: message => ipcRenderer.send('mofu:failed', String(message).slice(0, 400))
});
