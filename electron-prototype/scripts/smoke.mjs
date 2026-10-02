import { spawn } from 'node:child_process';
import fs from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const parameters = process.argv.slice(2), catalog = parameters.includes('--catalog-smoke'), follow = parameters.includes('--follow-smoke');
const target = parameters.find(a => !a.startsWith('--'));
const executable = target ?? (await import('electron')).default;
const packaged = target !== undefined;
const report = path.join(root, 'qa', `${catalog ? 'catalog' : follow ? 'follow' : 'smoke'}-${packaged ? 'packaged' : 'dev'}.json`);
const env = { ...process.env }; delete env.ELECTRON_RUN_AS_NODE;
await fs.mkdir(path.dirname(report), { recursive: true });
const args = [...(packaged ? [] : [root]), catalog ? '--catalog-smoke' : follow ? '--follow-smoke' : '--smoke', '--count=10', '--size=48', `--report=${report}`,
  ...parameters.filter(a => a.startsWith('--follow-mode=') || a === '--smoke-display-rebuild')];
const child = spawn(executable, args, { cwd: root, env, windowsHide: true, stdio: ['ignore', 'pipe', 'pipe'] });
let stderr = ''; child.stdout.on('data', b => process.stdout.write(b)); child.stderr.on('data', b => { stderr = (stderr + b).slice(-8000); });
const timeout = setTimeout(() => { child.kill(); console.error('Smoke exceeded 90 seconds'); }, 90000);
const code = await new Promise((resolve, reject) => { child.on('error', reject); child.on('exit', resolve); }); clearTimeout(timeout);
if (code !== 0) { console.error(stderr); process.exitCode = 1; }
console.log(JSON.stringify({ executable, exitCode: code, report }));
