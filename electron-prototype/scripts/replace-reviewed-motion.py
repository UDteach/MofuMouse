"""Replace one reviewed motion in a pinned catalog, preserving all other assets.

Defaults to a dry run. One uniform transform preserves the previous displayed
body area, center and ground; pixels are only decoded from the reviewed APNG.
"""
from pathlib import Path
import argparse
import copy
import hashlib
import io
import json
import re
import shutil
from datetime import datetime, timezone
import numpy as np
from PIL import Image, ImageFilter

ROOT = Path(__file__).resolve().parents[2]
LOCAL = ROOT / "electron-prototype"
MEDIA = LOCAL / "app/media"


def sha(data):
    return hashlib.sha256(data).hexdigest()


def encoded(data):
    return (json.dumps(data, ensure_ascii=False, indent=2) + "\n").encode("utf-8")


def metrics(frames):
    rows = []
    for image in frames:
        a = np.asarray(image)
        mask = a[:, :, 3] >= 128
        body = np.asarray(Image.fromarray((mask * 255).astype("uint8")).filter(ImageFilter.MinFilter(7)).filter(ImageFilter.MaxFilter(7))) > 0
        by, bx = np.where(body)
        y, x = np.where(mask)
        assert len(bx) > 0
        rows.append((body.sum(), bx.mean(), y.max() + 1))
    return np.median(np.array(rows), axis=0)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--variant", required=True)
    parser.add_argument("--manifest", required=True)
    parser.add_argument("--apply", action="store_true")
    args = parser.parse_args()
    assert re.fullmatch(r"[a-z0-9_-]+", args.variant)
    source = (ROOT / args.manifest).resolve()
    assert source.is_relative_to(ROOT)
    source_bytes = source.read_bytes()
    motion = json.loads(source_bytes)
    assert motion["status"] == "reviewed_source_ready" and motion["action"] == "walk"
    review = (ROOT / motion["visual_review"]).resolve()
    assert review.is_relative_to(ROOT) and sha(review.read_bytes()) == motion["visual_review_sha256"]
    assert len(motion["frames"]) == 30 and len({f["source_sha256"] for f in motion["frames"]}) == 30
    for frame in motion["frames"]:
        file = (ROOT / frame["source"]).resolve()
        assert file.is_relative_to(ROOT) and sha(file.read_bytes()) == frame["source_sha256"]
    manifest_path = MEDIA / "manifest.json"
    before_bytes = manifest_path.read_bytes()
    catalog = json.loads(before_bytes)
    animal = next(v for v in catalog["variants"] if v["id"] == args.variant)
    assert animal["species"] == motion["species"] and animal["coat"] == motion["coat"]
    old = copy.deepcopy(animal["motions"]["walk"])
    if old["provenance"].get("sha256") == sha(source_bytes):
        print(json.dumps({"changed": False, "reason": "same reviewed source already imported"}))
        return
    assert not animal["idleFallback"], "Static idle needs an explicit frame-reference migration"
    other_variants = copy.deepcopy([v for v in catalog["variants"] if v["id"] != args.variant])
    idle_before = copy.deepcopy(animal["motions"]["idle"])
    old_paths = {p for tier in old["tiers"].values() for p in tier}
    old_files = {f["path"]: f for f in catalog["files"] if f["path"] in old_paths}
    for file in catalog["files"]:
        assert sha((MEDIA / file["path"]).read_bytes()) == file["sha256"]
    assert len(old_files) == len(old_paths)
    staged, records, tiers, decoded = {}, [], {}, {}
    for size in (64, 96):
        export = next(e for e in motion["exports"] if e["path"].endswith(f"-{size}.png"))
        data = (ROOT / export["path"]).read_bytes()
        assert sha(data) == export["sha256"]
        animation = Image.open(io.BytesIO(data))
        assert animation.n_frames == 30 and not animation.info.get("default_image")
        frames, timings, paths = [], [], []
        for i in range(30):
            animation.seek(i)
            image = animation.convert("RGBA")
            assert image.size == (size * 3 // 2, size)
            frames.append(image.copy())
            timings.append(animation.info["duration"])
            name = f"{args.variant}/walk/{size}-{i:03d}.png"
            output = io.BytesIO(); image.save(output, format="PNG")
            staged[name] = output.getvalue(); paths.append(name)
            records.append({"path": name, "sha256": sha(staged[name]), "source": {"path": export["path"], "sha256": sha(data), "frame": i}})
        assert timings == motion["frame_durations_ms"] and sum(timings) == 1250
        tiers[str(size)] = paths; decoded[str(size)] = frames
    old_stats = metrics([Image.open(MEDIA / p).convert("RGBA") for p in old["tiers"]["96"]])
    new_stats = metrics(decoded["96"])
    previous = old.get("presentation", {"scale": 1, "x": 0, "y": 0})
    scale = float(previous["scale"] * np.sqrt(old_stats[0] / new_stats[0]))
    dx = float(old_stats[1] * previous["scale"] + previous["x"] * 96 - new_stats[1] * scale)
    dy = float(old_stats[2] * previous["scale"] + previous["y"] * 96 - new_stats[2] * scale)
    assert .5 <= scale <= 1.5
    for image in decoded["96"]:
        x0, y0, x1, y1 = image.getchannel("A").point(lambda v: 255 if v >= 16 else 0).getbbox()
        assert 0 <= x0 * scale + dx < x1 * scale + dx <= 144
        assert 0 <= y0 * scale + dy < y1 * scale + dy <= 96
    presentation = {"scale": round(scale, 6), "x": round(dx / 96, 6), "y": round(dy / 96, 6), "method": "one transform for full motion, retaining prior displayed body area, center and ground", "referenceAnimal": args.variant}
    animal["motions"]["walk"] = {"durations": motion["frame_durations_ms"], "tiers": tiers, "presentation": presentation, "provenance": {"manifest": args.manifest, "sha256": sha(source_bytes), "status": motion["status"], "visualReview": motion["visual_review"]}}
    catalog["files"] = [f for f in catalog["files"] if f["path"] not in old_paths] + records
    animal["files"] = [f for f in catalog["files"] if f["path"].startswith(args.variant + "/")]
    animal["frameCount"] = sum(len(m["durations"]) for m in animal["motions"].values())
    assert animal["motions"]["idle"] == idle_before
    assert [v for v in catalog["variants"] if v["id"] != args.variant] == other_variants
    backup = LOCAL / "qa/motion-replacements" / f"{args.variant}-{sha(source_bytes)[:12]}"
    report = {"variant": args.variant, "oldFrames": len(old["durations"]), "newFrames": 30, "durationMs": 1250, "presentation": presentation, "otherVariantsUnchanged": len(other_variants), "idleUnchanged": True, "apply": args.apply}
    if args.apply:
        assert not backup.exists(), "Keep the prior backup; do not overwrite it"
        backup.mkdir(parents=True)
        (backup / "manifest.json").write_bytes(before_bytes)
        shutil.copyfile(LOCAL / "catalog-snapshot.json", backup / "catalog-snapshot.json")
        for name in old_paths:
            dest = backup / name; dest.parent.mkdir(parents=True, exist_ok=True)
            shutil.copyfile(MEDIA / name, dest)
        assert manifest_path.read_bytes() == before_bytes, "Concurrent catalog edit"
        for name, data in staged.items():
            dest = MEDIA / name; dest.parent.mkdir(parents=True, exist_ok=True); dest.write_bytes(data)
        catalog["snapshotAt"] = datetime.now(timezone.utc).isoformat()
        catalog["inputs"].append({"path": args.manifest, "sha256": sha(source_bytes), "role": "reviewed single-motion replacement"})
        manifest_path.write_bytes(encoded(catalog))
        snapshot_path = LOCAL / "catalog-snapshot.json"
        snapshot = json.loads(snapshot_path.read_text(encoding="utf-8"))
        next(v for v in snapshot["variants"] if v["id"] == args.variant)["walkFrames"] = 30
        snapshot.update(snapshotAt=catalog["snapshotAt"], inputs=catalog["inputs"], images=len(catalog["files"]), manifestSha256=sha(manifest_path.read_bytes()))
        snapshot_path.write_bytes(encoded(snapshot))
        (backup / "replacement.json").write_bytes(encoded(report))
    print(json.dumps(report, ensure_ascii=False))


if __name__ == "__main__":
    main()
