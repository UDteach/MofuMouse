"""Bundle only the portable prototype and its hash-verified source PNGs."""
from __future__ import annotations
import hashlib
import json
from pathlib import Path
import stat
import zipfile

local = Path(__file__).resolve().parents[1]
project = local.parent
output = local / "release" / "MofuMouseElectron-Mac-Source.zip"
files: dict[str, bytes] = {}
def add(file: Path, relative: str) -> None:
    data = file.read_bytes()
    if file.suffix == ".command":
        data = data.replace(b"\r\n", b"\n")
    files[relative] = data

for name in ["README.md", "package.json", "package-lock.json", "build-mac.command", ".gitignore", "catalog-snapshot.json", "electron-builder.yml"]:
    add(local / name, f"electron-prototype/{name}")
for folder, patterns in [("app", ["*.cjs", "*.mjs", "*.html", "*.css"]), ("scripts", ["*.mjs"]), ("test", ["*.mjs"])]:
    for pattern in patterns:
        for file in sorted((local / folder).glob(pattern)):
            add(file, f"electron-prototype/{folder}/{file.name}")
for file in sorted((local / 'build').iterdir()):
    if file.is_file():
        add(file, f'electron-prototype/build/{file.name}')
add(project / 'scripts/publish-preview.mjs', 'scripts/publish-preview.mjs')
images = 0
source = local / 'app/media'
catalog = json.loads((source / 'manifest.json').read_text(encoding='utf-8'))
add(source / 'manifest.json', 'electron-prototype/app/media/manifest.json')
for entry in catalog['files']:
    file = source / entry['path']
    assert hashlib.sha256(file.read_bytes()).hexdigest() == entry['sha256'], file.name
    add(file, f'electron-prototype/app/media/{entry["path"]}')
    images += 1
manifest = {
    "platforms": ["darwin-arm64", "darwin-x64"],
    "kind": "source-kit", "macRuntimeVerified": False,
    "sourcePngs": images,
    "files": [{"path": name, "sha256": hashlib.sha256(data).hexdigest()} for name, data in sorted(files.items())],
}
files["BUNDLE-MANIFEST.json"] = (json.dumps(manifest, ensure_ascii=False, indent=2) + "\n").encode("utf-8")
files["START-HERE.txt"] = ("MofuMouse Electron - Mac source kit\n\nInstall Node.js 22.12 or newer.\nOpen Terminal in electron-prototype, then run:\n  bash build-mac.command\n\nSee electron-prototype/README.md. This is source, not a prebuilt .app.\nmacOS runtime has not been verified from the Windows build host.\n").encode("utf-8")
output.parent.mkdir(parents=True, exist_ok=True)
with zipfile.ZipFile(output, "w", compression=zipfile.ZIP_DEFLATED, compresslevel=6) as archive:
    for name, data in sorted(files.items()):
        info = zipfile.ZipInfo(f"MofuMouseElectron-Mac-Source/{name}")
        info.create_system = 3
        info.external_attr = (stat.S_IFREG | (0o755 if name.endswith(".command") else 0o644)) << 16
        info.compress_type = zipfile.ZIP_DEFLATED
        archive.writestr(info, data)
with zipfile.ZipFile(output) as archive:
    assert archive.testzip() is None
    for name, data in files.items():
        assert archive.read(f"MofuMouseElectron-Mac-Source/{name}") == data
print(json.dumps({"archive": str(output), "files": len(files), "sourcePngs": images, "bytes": output.stat().st_size, "sha256": hashlib.sha256(output.read_bytes()).hexdigest(), "macRuntimeVerified": False}, ensure_ascii=False))
