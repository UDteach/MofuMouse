"""Set one uniform presentation transform per motion. Never changes sprite RGB/PNG."""
from pathlib import Path
import hashlib, json
import numpy as np
from PIL import Image, ImageDraw, ImageFont
root=Path(__file__).resolve().parents[1]
manifest=root/'app/media/manifest.json'
catalog=json.loads(manifest.read_text(encoding='utf-8'))
measurement=json.loads((root/'qa/appearance/measurements.json').read_text(encoding='utf-8'))
metrics={v['id']:v for v in measurement['variants']}
baseline={}
for variant in catalog['variants']: baseline.setdefault(variant['species'],variant['id'])
report=[]
for animal in catalog['variants']:
    target=metrics[baseline[animal['species']]]['motions']['walk']
    center_x=float(np.median([f['bodyCenter'][0] for f in target['framesDetail']]))
    floor=float(np.median([f['fullBox'][3] for f in target['framesDetail']]))
    for action,motion in animal['motions'].items():
        data=metrics[animal['id']]['motions'][action]
        scale=float(np.sqrt(target['bodyAreaMedian']/data['bodyAreaMedian']))
        assert .65 <= scale <= 1.4, (animal['id'],action,scale)
        cx=float(np.median([f['bodyCenter'][0] for f in data['framesDetail']]))
        bottom=float(np.median([f['fullBox'][3] for f in data['framesDetail']]))
        dx=center_x-scale*cx;dy=floor-scale*bottom
        boxes=np.array([f['fullBox'] for f in data['framesDetail']])*scale+[dx,dy,dx,dy]
        bounds=[float(boxes[:,0].min()),float(boxes[:,1].min()),float(boxes[:,2].max()),float(boxes[:,3].max())]
        # Keep the entire sequence's alpha128 silhouette in the logical rectangle.
        # A single shared offset may be adjusted; no per-frame corrections.
        if bounds[0]<1:dx+=1-bounds[0]
        if bounds[2]>143:dx-=bounds[2]-143
        if bounds[1]<1:dy+=1-bounds[1]
        if bounds[3]>95:dy-=bounds[3]-95
        boxes=np.array([f['fullBox'] for f in data['framesDetail']])*scale+[dx,dy,dx,dy]
        assert boxes[:,0].min()>=0 and boxes[:,1].min()>=0 and boxes[:,2].max()<=144 and boxes[:,3].max()<=96,(animal['id'],action)
        motion['presentation']={'scale':round(scale,6),'x':round(dx/96,6),'y':round(dy/96,6),'method':'one transform for whole motion; baseline walk body area, center and ground','referenceAnimal':baseline[animal['species']]}
        report.append({'id':animal['id'],'action':action,'scale':round(scale,4),'targetBodyArea':target['bodyAreaMedian'],'scaledMedianBodyArea':round(data['bodyAreaMedian']*scale*scale,3),'withinSequenceScaleRatio':round(data['withinScaleRatio'],3),'bodyRgbSpan':np.round(data['bodyRgbSpan'],1).tolist(),'needsSourceReview':data['withinScaleRatio']>1.08 or max(data['bodyRgbSpan'])>12})
manifest.write_text(json.dumps(catalog,ensure_ascii=False,indent=2)+'\n',encoding='utf-8')
(root/'qa/appearance/presentation.json').write_text(json.dumps({'pngsModified':False,'reference':'per-species baseline walk','motions':report},ensure_ascii=False,indent=2)+'\n',encoding='utf-8')
snapshot_path=root/'catalog-snapshot.json';snapshot=json.loads(snapshot_path.read_text(encoding='utf-8'));snapshot['manifestSha256']=hashlib.sha256(manifest.read_bytes()).hexdigest();snapshot['presentation']='per-motion body/ground alignment; source-frame variation is not hidden';snapshot_path.write_text(json.dumps(snapshot,ensure_ascii=False,indent=2)+'\n',encoding='utf-8')
print(json.dumps({'motionsAligned':len(report),'pngsModified':False,'sourceReview':[f'{x["id"]}/{x["action"]}' for x in report if x['needsSourceReview']]}))
