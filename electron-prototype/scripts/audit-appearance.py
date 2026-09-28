"""Read-only silhouette/color diagnostics and comparison boards; never alters sprites."""
from pathlib import Path
import json
import numpy as np
from PIL import Image, ImageDraw, ImageFont, ImageFilter

ROOT = Path(__file__).resolve().parents[1]
MEDIA = ROOT / 'app/media'
OUT = ROOT / 'qa/appearance'
OUT.mkdir(parents=True, exist_ok=True)
catalog = json.loads((MEDIA / 'manifest.json').read_text(encoding='utf-8'))
font = ImageFont.truetype('C:/Windows/Fonts/meiryo.ttc', 15)
def largest_opened(mask):
    opened = np.asarray(Image.fromarray(mask.astype('uint8') * 255).filter(ImageFilter.MinFilter(7)).filter(ImageFilter.MaxFilter(7))) > 0
    seen = np.zeros_like(opened)
    best = []
    for y, x in zip(*np.where(opened)):
        if seen[y, x]: continue
        todo = [(int(y), int(x))]; seen[y, x] = True; group = []
        while todo:
            yy, xx = todo.pop(); group.append((yy, xx))
            for y2, x2 in [(yy-1,xx),(yy+1,xx),(yy,xx-1),(yy,xx+1)]:
                if 0 <= y2 < opened.shape[0] and 0 <= x2 < opened.shape[1] and opened[y2,x2] and not seen[y2,x2]:
                    seen[y2,x2] = True; todo.append((y2,x2))
        if len(group) > len(best): best = group
    out = np.zeros_like(opened)
    for y, x in best: out[y,x] = True
    return out & mask
records, rows = [], []
for animal in catalog['variants']:
    record = {'id': animal['id'], 'species': animal['species'], 'motions': {}}
    row = Image.new('RGB', (1230, 144), '#e6e6e6')
    draw = ImageDraw.Draw(row)
    draw.text((5, 4), f'{animal["speciesLabel"]} / {animal["coatLabel"]}', font=font, fill='black')
    for action_index, action in enumerate(['walk', 'idle']):
        paths = animal['motions'][action]['tiers']['96']
        stats = []
        for p in paths:
            image = Image.open(MEDIA / p).convert('RGBA')
            a = np.asarray(image)
            mask = a[..., 3] >= 128
            body = largest_opened(mask)
            ys, xs = np.where(body)
            fy, fx = np.where(mask)
            rgb = a[..., :3][body].mean(axis=0)
            stats.append({'bodyArea': int(body.sum()), 'bodyBox': [int(xs.min()), int(ys.min()), int(xs.max()+1), int(ys.max()+1)], 'fullBox': [int(fx.min()), int(fy.min()), int(fx.max()+1), int(fy.max()+1)], 'bodyCenter': [float(xs.mean()), float(ys.mean())], 'bodyRgb': rgb.tolist()})
        areas = [s['bodyArea'] for s in stats]
        colors = np.array([s['bodyRgb'] for s in stats])
        record['motions'][action] = {'frames': len(paths), 'bodyAreaMedian': float(np.median(areas)), 'bodyAreaMin': min(areas), 'bodyAreaMax': max(areas), 'withinScaleRatio': float(np.sqrt(max(areas)/min(areas))), 'bodyRgbMean': colors.mean(axis=0).tolist(), 'bodyRgbSpan': (colors.max(axis=0)-colors.min(axis=0)).tolist(), 'framesDetail': stats}
        for j, ix in enumerate([0, len(paths)//3, len(paths)*2//3, len(paths)-1]):
            im = Image.open(MEDIA / paths[ix]).convert('RGBA')
            x = 5 + (action_index * 4 + j) * 152
            draw.rectangle((x, 29, x+144, 126), fill='white' if action_index == 0 else '#242a32')
            row.paste(im, (x, 30), im)
            draw.text((x, 126), f'{action} {ix+1}', font=font, fill='black')
    record['idleWalkScaleRatio'] = float(np.sqrt(record['motions']['idle']['bodyAreaMedian']/record['motions']['walk']['bodyAreaMedian']))
    records.append(record); rows.append(row)
for n in range(0, len(rows), 7):
    out = Image.new('RGB', (1230, len(rows[n:n+7])*144))
    for i, row in enumerate(rows[n:n+7]): out.paste(row, (0, i*144))
    out.save(OUT/f'appearance-{n//7+1}.png')
(OUT/'measurements.json').write_text(json.dumps({'method': 'alpha128 opening square7 largest connected component; diagnostic only, not anatomy truth', 'variants': records}, indent=2)+'\n', encoding='utf-8')
for r in records:
    print(json.dumps({'id':r['id'], 'idleWalkScale':round(r['idleWalkScaleRatio'],3), 'walkWithinScale':round(r['motions']['walk']['withinScaleRatio'],3), 'idleWithinScale':round(r['motions']['idle']['withinScaleRatio'],3), 'walkColorSpan':np.round(r['motions']['walk']['bodyRgbSpan'],1).tolist(), 'idleColorSpan':np.round(r['motions']['idle']['bodyRgbSpan'],1).tolist()}))
