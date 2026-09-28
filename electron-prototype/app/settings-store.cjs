const fs = require('node:fs');
const path = require('node:path');
function readSettings(file) { return JSON.parse(fs.readFileSync(file, 'utf8')); }
function writeSettings(file, settings) {
  fs.mkdirSync(path.dirname(file), { recursive: true });
  fs.writeFileSync(`${file}.tmp`, `${JSON.stringify(settings, null, 2)}\n`, 'utf8');
  fs.renameSync(`${file}.tmp`, file);
}
module.exports = { readSettings, writeSettings };
