"""Snapshot completed animal assets without mutating the production ledgers."""
from __future__ import annotations
import hashlib
import io
import json
from pathlib import Path
from datetime import datetime, timezone
from PIL import Image

ROOT = Path(__file__).resolve().parents[2]
LOCAL = ROOT / 'electron-prototype'
OUTPUT = LOCAL / 'app/media'
COATS = {'agouti': 'アグーチ', 'blue': 'ブルー', 'sand': 'サンド', 'standard_gray': 'スタンダードグレー', 'beige': 'ベージュ', 'ebony': 'エボニー', 'golden_syrian': 'ゴールデン', 'cream': 'クリーム', 'black': 'ブラック', 'sable': 'セーブル', 'cinnamon': 'シナモン', 'dove': 'ダヴ', 'golden': 'ゴールデン', 'tan': 'タン', 'ruby_eyed_white': 'ルビーアイホワイト', 'salt_pepper': 'ソルト＆ペッパー', 'striped': 'ノーマル', 'white_black_eyed': 'ブラックアイホワイト'}
def digest(data): return hashlib.sha256(data).hexdigest()
def read(relative, expected=None):
    file = (ROOT / relative).resolve()
    if not file.is_relative_to(ROOT): raise ValueError('Source outside project')
    data = file.read_bytes()
    if expected and digest(data) != expected: raise ValueError(f'Source hash changed: {relative}')
    return data
def load(relative, expected=None): return json.loads(read(relative, expected).decode('utf-8'))
files, variants, inputs = [], [], []
def save_frame(variant, action, size, index, frame, source):
    assert frame.size == (int(size * 1.5), size), (source, frame.size)
    assert frame.getextrema()[3][0] == 0 and frame.getextrema()[3][1] > 0, source
    name = f'{variant}/{action}/{size}-{index:03d}.png'
    dest = OUTPUT / name
    dest.parent.mkdir(parents=True, exist_ok=True)
    frame.save(dest)
    files.append({'path': name, 'sha256': digest(dest.read_bytes()), 'source': source})
    return name
def decode(variant, action, exports, provenance):
    result = {'tiers': {}, 'provenance': provenance}
    for size in [64, 96]:
        candidates = [x for x in exports if x['path'].endswith(f'-{size}.png')]
        assert len(candidates) == 1, (variant, action, size)
        item = candidates[0]
        data = read(item['path'], item['sha256'])
        image = Image.open(io.BytesIO(data))
        assert image.is_animated and not image.info.get('default_image'), item['path']
        durations, paths = [], []
        for index in range(image.n_frames):
            image.seek(index)
            duration = image.info['duration']
            assert duration > 0
            durations.append(duration)
            paths.append(save_frame(variant, action, size, index, image.convert('RGBA'), {'path': item['path'], 'sha256': digest(data), 'frame': index}))
        assert image.n_frames == item['frame_count']
        assert abs(sum(durations) - item['cycle_duration_ms']) < 0.01
        if 'durations' in result: assert result['durations'] == durations
        result['durations'] = durations
        result['tiers'][str(size)] = paths
    return result
def base_degu_walk(variant):
    source = 'assets/prototype/degu-flow30/manifest.json'
    manifest = load(source)
    result = {'durations': manifest['durations_ms'], 'tiers': {}, 'provenance': {'manifest': source, 'sha256': digest(read(source)), 'status': 'previously_running'}}
    for size in [64, 96]:
        paths = []
        for index, file in enumerate(x for x in manifest['runtime_files'] if x['file'].startswith(f'{size}-')):
            original = f'assets/prototype/degu-flow30/{file["file"]}'
            data = read(original, file['sha256'])
            paths.append(save_frame(variant, 'walk', size, index, Image.open(io.BytesIO(data)).convert('RGBA'), {'path': original, 'sha256': digest(data)}))
        assert len(paths) == len(result['durations'])
        result['tiers'][str(size)] = paths
    return result
base_path = 'assets/manifest/animal-cursor-ten-species.json'
coat_path = 'assets/manifest/animal-cursor-coat-ten-species.json'
base_bytes, coat_bytes = read(base_path), read(coat_path)
base, coats = json.loads(base_bytes), json.loads(coat_bytes)
inputs = [{'path': base_path, 'sha256': digest(base_bytes)}, {'path': coat_path, 'sha256': digest(coat_bytes)}]
species_labels = {s['species']: s['label'] for s in base['species']}
for species in base['species']:
    variant = {'id': f'{species["species"]}-{species["base_coat"]}', 'species': species['species'], 'speciesLabel': species['label'], 'coat': species['base_coat'], 'coatLabel': COATS[species['base_coat']], 'motions': {}, 'idleFallback': False}
    for action in ['walk', 'idle']:
        row = species['motions'][action]
        assert row['status'] == 'ready'
        manifest = load(row['manifest'], row.get('manifest_sha256'))
        if species['species'] == 'degu' and action == 'walk':
            variant['motions'][action] = base_degu_walk(variant['id'])
        else:
            variant['motions'][action] = decode(variant['id'], action, manifest['exports'], {'manifest': row['manifest'], 'sha256': digest(read(row['manifest'])), 'status': row['status']})
    variants.append(variant)
completed = [x for x in coats['items'] if x['status'] in ['done', 'done_reused']]
for row in completed:
    variant_id = f'{row["species"]}-{row["coat"]}'
    variant = next((v for v in variants if v['id'] == variant_id), None)
    if variant is None:
        variant = {'id': variant_id, 'species': row['species'], 'speciesLabel': species_labels[row['species']], 'coat': row['coat'], 'coatLabel': COATS[row['coat']], 'motions': {}, 'idleFallback': False}
        variants.append(variant)
    if row['status'] == 'done':
        review = row['visual_review']
        assert review['status'] == 'pass', row['id']
        read(review['path'], review['sha256'])
    variant['motions'][row['motion']] = decode(variant_id, row['motion'], row['exports'], {'ledger': coat_path, 'ledgerSha256': digest(coat_bytes), 'item': row['id'], 'status': row['status'], 'visualReview': row.get('visual_review')})
for variant in variants:
    assert 'walk' in variant['motions'], variant['id']
    if 'idle' not in variant['motions']:
        walk = variant['motions']['walk']
        variant['motions']['idle'] = {'durations': [1000], 'tiers': {tier: [paths[0]] for tier, paths in walk['tiers'].items()}, 'provenance': {'fallback': 'hold same-coat walk frame 0; no generated idle'}}
        variant['idleFallback'] = True
    variant['files'] = [f for f in files if f['path'].startswith(variant['id'] + '/')]
    variant['frameCount'] = sum(len(m['durations']) for m in variant['motions'].values())
manifest = {'schema': 2, 'defaultId': 'degu-agouti', 'snapshotAt': datetime.now(timezone.utc).isoformat(), 'inputs': inputs, 'variants': variants, 'files': files}
OUTPUT.mkdir(parents=True, exist_ok=True)
(OUTPUT / 'manifest.json').write_text(json.dumps(manifest, ensure_ascii=False, indent=2) + '\n', encoding='utf-8')
summary = {'snapshotAt': manifest['snapshotAt'], 'inputs': inputs, 'species': len(species_labels), 'variants': [{'id': v['id'], 'label': f'{v["speciesLabel"]} / {v["coatLabel"]}', 'walkFrames': len(v['motions']['walk']['durations']), 'idleFrames': len(v['motions']['idle']['durations']), 'idleFallback': v['idleFallback']} for v in variants], 'images': len(files), 'manifestSha256': digest((OUTPUT / 'manifest.json').read_bytes())}
(LOCAL / 'catalog-snapshot.json').write_text(json.dumps(summary, ensure_ascii=False, indent=2) + '\n', encoding='utf-8')
print(json.dumps({'species': len(species_labels), 'variants': len(variants), 'images': len(files), 'fallbacks': [v['id'] for v in variants if v['idleFallback']]}, ensure_ascii=False))
