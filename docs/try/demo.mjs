import {Model} from './motion.mjs';
const $=s=>document.querySelector(s), stage=$('#stage'), canvas=$('#animals'), ctx=canvas.getContext('2d');
const animal=$('#animal'), count=$('#count'), size=$('#size'), pause=$('#pause'), status=$('#status');
let catalog, model, images=new Map(), token=0, running=false, pending=true, frameRequest=0;
let paused=matchMedia('(prefers-reduced-motion: reduce)').matches, selection=[], pointer={x:0,y:0}, lastPointer=false;
function tell(text){status.textContent=text;}
function profiles(){return Array.from({length:10},(_,i)=>{const v=catalog.variants.find(v=>v.id===(selection[i]??selection[0]));return {id:v.id,walk:v.motions.walk.durations,idle:v.motions.idle.durations};});}
function schedule(){if(!frameRequest && !document.hidden && running && !paused) frameRequest=requestAnimationFrame(draw);}
function draw(now){
 frameRequest=0; if(!running||document.hidden)return;
 const bounds=stage.getBoundingClientRect(), dpr=devicePixelRatio||1;
 if(canvas.width!==Math.round(bounds.width*dpr)||canvas.height!==Math.round(bounds.height*dpr)){canvas.width=Math.round(bounds.width*dpr);canvas.height=Math.round(bounds.height*dpr);}
 if(!lastPointer)pointer={x:bounds.width*.72,y:bounds.height*.58};
 const state=model.step(now,pointer,{x:0,y:0,width:bounds.width,height:bounds.height});
 ctx.setTransform(dpr,0,0,dpr,0,0);ctx.clearRect(0,0,bounds.width,bounds.height);ctx.imageSmoothingEnabled=true;ctx.imageSmoothingQuality='high';
 for(const pet of state.pets){const v=catalog.variants.find(v=>v.id===pet.animalId),motion=v.motions[pet.action],image=images.get(motion.frames[pet.frame]),t=motion.presentation;
   ctx.save();if(pet.left){ctx.translate(pet.x+pet.width,pet.y);ctx.scale(-1,1);}else ctx.translate(pet.x,pet.y);
   ctx.drawImage(image,t.x*pet.height,t.y*pet.height,pet.width*t.scale,pet.height*t.scale);ctx.restore();
 }
 canvas.dataset.pets=String(state.pets.length);canvas.dataset.action=state.pets[0]?.action??'paused';canvas.dataset.firstX=String(state.pets[0]?.x??0);canvas.dataset.firstFrame=String(state.pets[0]?.frame??0);
 schedule();
}
async function load(){
 const id=++token; pending=true;running=false;pause.disabled=true;$('#all').disabled=true;$('#retry').hidden=true;ctx.clearRect(0,0,canvas.width,canvas.height);tell('動物を読み込んでいます…');
 try {
  const wanted=new Set(profiles().slice(0,Number(count.value)).map(p=>p.id)),next=new Map();
  await Promise.all([...wanted].flatMap(vId=>Object.values(catalog.variants.find(v=>v.id===vId).motions).flatMap(m=>m.frames.map(async src=>{
    const cached=images.get(src);if(cached){next.set(src,cached);return;}const img=new Image();img.src=src;await img.decode();next.set(src,img);
  }))));
  if(id!==token)return;images=next;
  const base=catalog.variants.find(v=>v.id===selection[0]);model=new Model({walk:base.motions.walk.durations,idle:base.motions.idle.durations},{count:Number(count.value),size:Number(size.value),paused:false});model.setProfiles(profiles());
  pending=false;running=true;pause.disabled=false;$('#all').disabled=false;pause.textContent=paused?'再開':'一時停止';
  canvas.setAttribute('aria-label',`${count.value}匹の動物がカーソルに追従する表示`);tell(paused?'停止中です。「再開」で動き始めます。':`${count.value}匹を表示中。エリア内を動かしてみてください。`);draw(performance.now());
 } catch(e){if(id!==token)return;pending=false;running=false;tell('動物を読み込めませんでした。通信を確認して、もう一度お試しください。');$('#retry').hidden=false;console.error(e);}
}
stage.addEventListener('pointermove',e=>{const r=stage.getBoundingClientRect();pointer={x:e.clientX-r.left,y:e.clientY-r.top};lastPointer=true;schedule();});
stage.addEventListener('pointerdown',e=>{stage.setPointerCapture(e.pointerId);const r=stage.getBoundingClientRect();pointer={x:e.clientX-r.left,y:e.clientY-r.top};lastPointer=true;});
animal.addEventListener('change',()=>{selection=Array(10).fill(animal.value);load();});
count.addEventListener('change',()=>load());size.addEventListener('change',()=>load());
pause.addEventListener('click',()=>{paused=!paused;pause.textContent=paused?'再開':'一時停止';if(paused){cancelAnimationFrame(frameRequest);frameRequest=0;tell('停止中です。「再開」で動き始めます。');}else{model.sampler.reset();model.members.forEach(m=>m.sampler.reset());tell(`${count.value}匹を表示中。`);schedule();}});
$('#all').addEventListener('click',()=>{const species=new Map();for(const v of catalog.variants)if(!species.has(v.species))species.set(v.species,v.id);selection=[...species.values()];count.value='10';load();});
$('#background').addEventListener('click',e=>{const dark=stage.classList.toggle('dark');e.currentTarget.setAttribute('aria-pressed',String(dark));e.currentTarget.textContent=dark?'明るい背景':'暗い背景';});
$('#retry').addEventListener('click',()=>catalog?load():init());
document.addEventListener('visibilitychange',()=>{cancelAnimationFrame(frameRequest);frameRequest=0;if(!document.hidden&&model){model.sampler.reset();schedule();}});
window.addEventListener('resize',()=>{if(running)draw(performance.now());});
async function init(){try{const response=await fetch('catalog.json');if(!response.ok)throw Error('Catalog unavailable');catalog=await response.json();animal.replaceChildren();for(const v of catalog.variants){const o=document.createElement('option');o.value=v.id;o.textContent=`${v.speciesLabel} · ${v.coatLabel}`;animal.append(o);}const requested=new URLSearchParams(location.search).get('animal'),initial=catalog.variants.some(v=>v.id===requested)?requested:catalog.defaultId;animal.value=initial;animal.disabled=false;selection=Array(10).fill(initial);await load();}catch(e){tell('一覧を読み込めませんでした。もう一度お試しください。');$('#retry').hidden=false;console.error(e);}}
init();
