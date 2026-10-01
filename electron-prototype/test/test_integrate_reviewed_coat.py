from __future__ import annotations

import copy
import hashlib
import importlib.util
import io
import json
import os
import subprocess
import tempfile
import unittest
from pathlib import Path

from PIL import Image, ImageDraw


MODULE_PATH = Path(__file__).resolve().parents[1] / "scripts" / "integrate-reviewed-coat.py"
SPEC = importlib.util.spec_from_file_location("integrate_reviewed_coat", MODULE_PATH)
assert SPEC and SPEC.loader
integration = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(integration)


def digest(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def write_bytes(root: Path, relative: str, data: bytes) -> dict[str, str]:
    path = root / relative
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_bytes(data)
    return {"path": relative, "sha256": digest(data)}


def png_bytes(size: tuple[int, int], index: int, color: tuple[int, int, int]) -> bytes:
    image = Image.new("RGBA", size, (0, 0, 0, 0))
    draw = ImageDraw.Draw(image)
    x = 8 + index % max(1, size[0] // 8)
    y = 8 + index % max(1, size[1] // 10)
    draw.rectangle((x, y, min(size[0] - 8, x + 20), min(size[1] - 8, y + 18)), fill=(*color, 255))
    output = io.BytesIO()
    image.save(output, format="PNG", optimize=False)
    return output.getvalue()


def apng_bytes(size: tuple[int, int], durations: list[int], color_seed: int) -> bytes:
    frames = [Image.open(io.BytesIO(png_bytes(size, color_seed + i, ((color_seed + i * 17) % 220 + 20, 50, 140)))).convert("RGBA") for i in range(len(durations))]
    output = io.BytesIO()
    frames[0].save(output, format="PNG", save_all=True, append_images=frames[1:], duration=durations, loop=0, disposal=2, optimize=False)
    return output.getvalue()


def edge_apng_bytes(size: tuple[int, int], durations: list[int], *, opaque: bool = False) -> bytes:
    frames = []
    for index in range(len(durations)):
        if opaque:
            image = Image.new("RGBA", size, (100, 100, 100, 255))
        else:
            image = Image.new("RGBA", size, (0, 0, 0, 0))
            ImageDraw.Draw(image).rectangle((0, 8, 20, 24), fill=(100, 100, 100, 255))
        frames.append(image)
    output = io.BytesIO()
    frames[0].save(output, format="PNG", save_all=True, append_images=frames[1:], duration=durations, loop=0, disposal=2, optimize=False)
    return output.getvalue()


class Fixture:
    def __init__(self, species: str = "guinea_pig", coat: str = "cream", *, degu: bool = False, degu_coat: str = "blue", idle_durations: list[int] | None = None):
        self.temp = tempfile.TemporaryDirectory()
        self.root = Path(self.temp.name)
        self.species = "degu" if degu else species
        self.coat = degu_coat if degu else coat
        self.variant_id = f"{self.species}-{self.coat}"
        self.walk_durations = [19 if i % 2 == 0 else 18 for i in range(30)] if degu else [42 if i % 3 else 41 for i in range(30)]
        if degu:
            self.walk_durations[-2:] = [19, 19]
        self.idle_durations = idle_durations or [300, 200]
        self._write_catalog()
        self.receipt_path = self._write_receipt()

    def close(self):
        self.temp.cleanup()

    def _write_catalog(self):
        media = self.root / "electron-prototype/app/media"
        media.mkdir(parents=True)
        variants = []
        files = []
        variants.append(self._make_variant(media, files, "guinea_pig-golden", "guinea_pig", "golden", "モルモット", "ゴールデン", walk_count=2, idle_count=2, walk_duration=[50, 50], idle_duration=[100, 100]))
        degu_coat = self.coat if self.species == "degu" else "blue"
        degu_id = f"degu-{degu_coat}"
        variants.append(self._make_variant(media, files, degu_id, "degu", degu_coat, "デグー", "サンド" if degu_coat == "sand" else "ブルー", walk_count=8, idle_count=1, walk_duration=[74, 74, 74, 56, 74, 74, 74, 56], idle_duration=[1000], fallback=True))
        variants.append(self._make_variant(media, files, "rabbit-ruby_eyed_white", "rabbit", "ruby_eyed_white", "うさぎ", "ルビーアイホワイト", walk_count=2, idle_count=2, walk_duration=[50, 50], idle_duration=[100, 100]))
        catalog = {"schema": 2, "defaultId": degu_id, "snapshotAt": "fixture", "inputs": [], "variants": variants, "files": files}
        manifest = (json.dumps(catalog, ensure_ascii=False, indent=2) + "\n").encode("utf-8")
        write_bytes(self.root, "electron-prototype/app/media/manifest.json", manifest)
        self.catalog_hash = digest(manifest)
        snapshot = {
            "snapshotAt": "fixture",
            "inputs": [],
            "species": 3,
            "variants": [{"id": v["id"], "label": f"{v['speciesLabel']} / {v['coatLabel']}", "walkFrames": len(v["motions"]["walk"]["durations"]), "idleFrames": len(v["motions"]["idle"]["durations"]), "idleFallback": v["idleFallback"]} for v in variants],
            "images": len(files),
            "manifestSha256": self.catalog_hash,
            "presentation": "fixture",
        }
        snapshot_bytes = (json.dumps(snapshot, ensure_ascii=False, indent=2) + "\n").encode("utf-8")
        write_bytes(self.root, "electron-prototype/catalog-snapshot.json", snapshot_bytes)
        self.snapshot_hash = digest(snapshot_bytes)

    def _make_variant(self, media: Path, files: list[dict], variant_id: str, species: str, coat: str, species_label: str, coat_label: str, *, walk_count: int, idle_count: int, walk_duration: list[int], idle_duration: list[int], fallback: bool = False) -> dict:
        tiers = {"64": [], "96": []}
        idle_tiers = {"64": [], "96": []}
        for action, count, duration, target in [("walk", walk_count, walk_duration, tiers), ("idle", idle_count, idle_duration, idle_tiers)]:
            for tier in ("64", "96"):
                for i in range(count):
                    relative = f"{variant_id}/{action}/{tier}-{i:03d}.png"
                    data = png_bytes((96, 64) if tier == "64" else (144, 96), i + len(files), (30 + i * 3, 70, 100))
                    written = write_bytes(self.root, f"electron-prototype/app/media/{relative}", data)
                    record = {"path": relative, "sha256": written["sha256"], "source": {"fixture": True}}
                    files.append(record)
                    target[tier].append(relative)
        if fallback:
            idle_tiers = {tier: [tiers[tier][0]] for tier in ("64", "96")}
            # Remove the unused physical idle records from the active index and
            # leave them on disk as harmless historical files.
            files[:] = [record for record in files if not (record["path"].startswith(variant_id + "/idle/") and record["path"] not in idle_tiers["64"] + idle_tiers["96"])]
        motions = {
            "walk": {"durations": walk_duration, "tiers": tiers, "presentation": {"scale": 1, "x": 0, "y": 0}},
            "idle": {"durations": idle_duration, "tiers": idle_tiers, "presentation": {"scale": 1, "x": 0, "y": 0}},
        }
        used = {path for motion in motions.values() for paths in motion["tiers"].values() for path in paths}
        indexed = [record for record in files if record["path"] in used]
        return {"id": variant_id, "species": species, "speciesLabel": species_label, "coat": coat, "coatLabel": coat_label, "motions": motions, "idleFallback": fallback, "files": indexed, "frameCount": sum(len(m["durations"]) for m in motions.values())}

    def _write_receipt(self, *, operation: str = "add", variant_id: str | None = None, walk_count: int | None = None, changed_review: bool = False, receipt_name: str = "receipt.json") -> Path:
        variant_id = variant_id or self.variant_id
        species = "degu" if variant_id.startswith("degu-") else "guinea_pig"
        coat = variant_id.split("-", 1)[1]
        species_label = "デグー" if species == "degu" else "モルモット"
        coat_label = ("サンド" if coat == "sand" else "ブルー") if species == "degu" else "クリーム"
        walk_count = walk_count or 30
        walk_duration = self.walk_durations[:walk_count]
        source_root = f"fixtures/{variant_id}/walk"
        idle_root = f"fixtures/{variant_id}/idle"
        walk_base_frames = []
        walk_source_frames = []
        walk_source_manifest_frames = []
        for i in range(30):
            base_path = f"{source_root}/base-{i:03d}.png"
            base_data = png_bytes((96, 64), i, (40 + i, 80, 120))
            base_desc = write_bytes(self.root, base_path, base_data)
            walk_base_frames.append({"id": i + 1, "source": base_path, "source_sha256": base_desc["sha256"], "video_sha256": "VIDEO_SHA_PLACEHOLDER", "video_frame_zero_based": i, "reference_time_s": i / 24})
            edited_path = f"{source_root}/edited-{i:03d}.png"
            edited_desc = write_bytes(self.root, edited_path, png_bytes((96, 64), i + 50, (80 + i, 130, 190)))
            provenance = write_bytes(self.root, f"{source_root}/provenance.txt", b"fixture provenance\n")
            walk_source_frames.append({"path": edited_path, "sha256": edited_desc["sha256"], "base_source_sha256": base_desc["sha256"], "provenance": provenance})
            walk_source_manifest_frames.append({"id": i + 1, "source": base_path, "source_sha256": base_desc["sha256"], "video_sha256": "VIDEO_SHA_PLACEHOLDER", "video_frame_zero_based": i, "reference_time_s": i / 24})
        video_desc = write_bytes(self.root, f"{source_root}/video.mp4", b"fixture video bytes")
        video_sha = video_desc["sha256"]
        for item in walk_base_frames + walk_source_manifest_frames:
            item["video_sha256"] = video_sha
        walk_source_manifest = {"schema_version": 1, "species": species, "coat": "agouti" if species == "degu" else "golden", "action": "walk", "video": video_desc["path"], "video_sha256": video_sha, "video_fps": 24, "frame_count": 30, "frame_durations_ms": walk_duration, "cycle_duration_ms": sum(walk_duration), "frames": walk_source_manifest_frames}
        walk_source_path = f"{source_root}/source-manifest.json"
        walk_source_desc = write_bytes(self.root, walk_source_path, (json.dumps(walk_source_manifest, indent=2) + "\n").encode("utf-8"))
        walk_reference = {"status": "ready", "species": species, "motion": "walk", "source_manifest": walk_source_path, "source_manifest_sha256": walk_source_desc["sha256"], "source_frame_count": 30, "target_real_poses": 30, "video": video_desc["path"], "video_sha256": video_sha, "video_fps": 24, "output_frame_count": 30, "frame_durations_ms": walk_duration, "cycle_duration_ms": sum(walk_duration), "frames": walk_base_frames}
        walk_ref_path = f"{source_root}/reference.json"
        walk_ref_desc = write_bytes(self.root, walk_ref_path, (json.dumps(walk_reference, indent=2) + "\n").encode("utf-8"))
        idle_base_frames = []
        idle_source_frames = []
        idle_source_manifest_frames = []
        for i in range(len(self.idle_durations)):
            base_path = f"{idle_root}/base-{i:03d}.png"
            base_data = png_bytes((96, 64), i + 200, (40, 100, 120))
            base_desc = write_bytes(self.root, base_path, base_data)
            idle_base_frames.append({"id": i + 1, "source": base_path, "source_sha256": base_desc["sha256"], "video_sha256": "IDLE_VIDEO_SHA_PLACEHOLDER", "video_frame_zero_based": i, "reference_time_s": i / 24})
            edited_path = f"{idle_root}/edited-{i:03d}.png"
            edited_desc = write_bytes(self.root, edited_path, png_bytes((96, 64), i + 240, (100, 150, 180)))
            provenance = write_bytes(self.root, f"{idle_root}/provenance.txt", b"fixture idle provenance\n")
            idle_source_frames.append({"path": edited_path, "sha256": edited_desc["sha256"], "base_source_sha256": base_desc["sha256"], "provenance": provenance})
            idle_source_manifest_frames.append({"id": i + 1, "source": base_path, "source_sha256": base_desc["sha256"], "video_sha256": "IDLE_VIDEO_SHA_PLACEHOLDER", "video_frame_zero_based": i, "reference_time_s": i / 24})
        idle_video_desc = write_bytes(self.root, f"{idle_root}/video.mp4", b"fixture idle video bytes")
        for item in idle_base_frames + idle_source_manifest_frames:
            item["video_sha256"] = idle_video_desc["sha256"]
        idle_source_manifest = {"schema_version": 1, "species": species, "coat": "agouti" if species == "degu" else "golden", "action": "idle", "video": idle_video_desc["path"], "video_sha256": idle_video_desc["sha256"], "video_fps": 24, "frame_count": len(self.idle_durations), "frame_durations_ms": self.idle_durations, "cycle_duration_ms": sum(self.idle_durations), "frames": idle_source_manifest_frames}
        idle_source_path = f"{idle_root}/source-manifest.json"
        idle_source_desc = write_bytes(self.root, idle_source_path, (json.dumps(idle_source_manifest, indent=2) + "\n").encode("utf-8"))
        idle_reference = {"status": "ready", "species": species, "motion": "idle", "source_manifest": idle_source_path, "source_manifest_sha256": idle_source_desc["sha256"], "source_frame_count": len(self.idle_durations), "output_frame_count": len(self.idle_durations), "video": idle_video_desc["path"], "video_sha256": idle_video_desc["sha256"], "video_fps": 24, "frame_durations_ms": self.idle_durations, "cycle_duration_ms": sum(self.idle_durations), "frames": idle_base_frames}
        idle_ref_path = f"{idle_root}/reference.json"
        idle_ref_desc = write_bytes(self.root, idle_ref_path, (json.dumps(idle_reference, indent=2) + "\n").encode("utf-8"))
        receipt = {
            "schema": 1,
            "operation": operation,
            "expected_catalog_sha256": self.catalog_hash,
            "expected_snapshot_sha256": self.snapshot_hash,
            "variant": {"id": variant_id, "species": species, "speciesLabel": species_label, "coat": coat, "coatLabel": coat_label},
            "motions": {
                "walk": self._motion_receipt(walk_ref_desc, walk_source_desc, walk_source_frames, walk_duration, walk_source_path, 1250 if not degu_variant(variant_id) else 556, f"{source_root}/walk-64.png", f"{source_root}/walk-96.png", seed=10),
                "idle": self._motion_receipt(idle_ref_desc, idle_source_desc, idle_source_frames, self.idle_durations, idle_source_path, sum(self.idle_durations), f"{idle_root}/idle-64.png", f"{idle_root}/idle-96.png", seed=100),
            },
        }
        review_evidence = write_bytes(self.root, f"fixtures/{variant_id}/review-evidence.txt", b"parent visual review fixture\n")
        review_path = f"fixtures/{variant_id}/review.json"
        review = {"status": "pass", "variant_id": variant_id, "scope": "source_export_presentation", "receipt_payload_sha256": integration.receipt_payload_hash({**receipt, "visual_review": {"path": review_path, "sha256": "0" * 64}}), "evidence": [review_evidence]}
        review_data = (json.dumps(review, ensure_ascii=False, indent=2) + "\n").encode("utf-8")
        review_desc = write_bytes(self.root, review_path, review_data)
        receipt["visual_review"] = review_desc
        if changed_review:
            review["receipt_payload_sha256"] = "0" * 64
            review_data = (json.dumps(review, ensure_ascii=False, indent=2) + "\n").encode("utf-8")
            review_desc = write_bytes(self.root, review_path, review_data)
            receipt["visual_review"] = review_desc
        receipt_data = (json.dumps(receipt, ensure_ascii=False, indent=2) + "\n").encode("utf-8")
        receipt_path = write_bytes(self.root, f"fixtures/{receipt_name}", receipt_data)
        return self.root / receipt_path["path"]

    def _motion_receipt(self, ref_desc, source_desc, frames, durations, source_path, cycle, export64, export96, *, seed: int):
        write_bytes(self.root, export64, apng_bytes((96, 64), durations, seed))
        write_bytes(self.root, export96, apng_bytes((144, 96), durations, seed + 1))
        return {"base_reference": ref_desc, "source_manifest": source_desc, "frames": frames, "durations_ms": durations, "exports": {"64": write_bytes(self.root, export64, apng_bytes((96, 64), durations, seed)), "96": write_bytes(self.root, export96, apng_bytes((144, 96), durations, seed + 1))}, "presentation": {"scale": 1, "x": 0, "y": 0}}


def degu_variant(variant_id: str) -> bool:
    return variant_id.startswith("degu-")


class IntegrateReviewedCoatTests(unittest.TestCase):
    def setUp(self):
        self.fixture = Fixture()

    def tearDown(self):
        self.fixture.close()

    def stage(self, receipt: Path | None = None, name: str = "stage"):
        selected = receipt or self.fixture.receipt_path
        relative = selected.resolve().relative_to(self.fixture.root.resolve()).as_posix()
        return integration.stage_receipt(self.fixture.root, relative, f"electron-prototype/qa/coat-stages/{name}")

    def refresh_review(self, receipt_path: Path, root: Path | None = None):
        root = root or self.fixture.root
        receipt = json.loads(receipt_path.read_text(encoding="utf-8"))
        review_path = root / receipt["visual_review"]["path"]
        review = json.loads(review_path.read_text(encoding="utf-8"))
        review["variant_id"] = receipt["variant"]["id"]
        review["receipt_payload_sha256"] = integration.receipt_payload_hash(receipt)
        review_bytes = (json.dumps(review, ensure_ascii=False, indent=2) + "\n").encode("utf-8")
        review_path.write_bytes(review_bytes)
        receipt["visual_review"]["sha256"] = digest(review_bytes)
        receipt_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")

    def test_add_stage_preserves_catalog_and_allocates_only_new_media(self):
        before_manifest = (self.fixture.root / "electron-prototype/app/media/manifest.json").read_bytes()
        before_snapshot = (self.fixture.root / "electron-prototype/catalog-snapshot.json").read_bytes()
        result = self.stage()
        plan = result["plan"]
        self.assertEqual(plan["operation"], "add")
        self.assertEqual(len(plan["new_files"]), 64)
        self.assertEqual(min(int(record["path"].rsplit("-", 1)[1][:-4]) for record in plan["new_files"]), 0)
        self.assertEqual(max(int(record["path"].rsplit("-", 1)[1][:-4]) for record in plan["new_files"]), 29)
        self.assertEqual((self.fixture.root / "electron-prototype/app/media/manifest.json").read_bytes(), before_manifest)
        self.assertEqual((self.fixture.root / "electron-prototype/catalog-snapshot.json").read_bytes(), before_snapshot)
        staged = json.loads((self.fixture.root / result["stage"] / "manifest.json").read_text(encoding="utf-8"))
        self.assertEqual([v["id"] for v in staged["variants"]], ["guinea_pig-golden", "degu-blue", "rabbit-ruby_eyed_white", "guinea_pig-cream"])
        self.assertFalse(next(v for v in staged["variants"] if v["id"] == "guinea_pig-cream")["idleFallback"])
        self.assertEqual(len(staged["files"]), len(plan["baseline"]["files"]) + 64)

    def test_append_only_allocation_uses_physical_files_above_999(self):
        orphan = self.fixture.root / "electron-prototype/app/media/guinea_pig-cream/walk/64-1000.png"
        orphan.parent.mkdir(parents=True, exist_ok=True)
        orphan.write_bytes(png_bytes((96, 64), 7, (20, 80, 150)))
        result = self.stage(name="high-index")
        walk_paths = [record["path"] for record in result["plan"]["new_files"] if "/walk/" in record["path"] and "/64-" in record["path"]]
        self.assertEqual(walk_paths[0], "guinea_pig-cream/walk/64-1001.png")

    def test_replace_degu_fallback_preserves_independent_idle_timing(self):
        degu = Fixture(degu=True)
        try:
            degu.receipt_path = degu._write_receipt(operation="replace", receipt_name="degu-receipt.json")
            result = integration.stage_receipt(degu.root, "fixtures/degu-receipt.json", "electron-prototype/qa/coat-stages/degu-replace")
            staged = json.loads((degu.root / result["stage"] / "manifest.json").read_text(encoding="utf-8"))
            target = next(v for v in staged["variants"] if v["id"] == "degu-blue")
            self.assertFalse(target["idleFallback"])
            self.assertEqual(sum(target["motions"]["walk"]["durations"]), 556)
            self.assertEqual(sum(target["motions"]["idle"]["durations"]), 500)
            self.assertEqual(len(target["motions"]["walk"]["durations"]), 30)
            self.assertEqual([v["id"] for v in staged["variants"]], ["guinea_pig-golden", "degu-blue", "rabbit-ruby_eyed_white"])
        finally:
            degu.close()

    def test_apply_and_idempotent_repeat_then_finish_snapshot_recovery(self):
        result = self.stage(name="apply")
        applied = integration.apply_stage(self.fixture.root, result["stage"])
        self.assertEqual(applied["status"], "applied")
        again = integration.apply_stage(self.fixture.root, result["stage"])
        self.assertEqual(again["status"], "already_applied")
        stage = self.fixture.root / result["stage"]
        old_snapshot = (stage / "backup/catalog-snapshot.json").read_bytes()
        (self.fixture.root / "electron-prototype/catalog-snapshot.json").write_bytes(old_snapshot)
        recovered = integration.recover_stage(self.fixture.root, result["stage"])
        self.assertEqual(recovered["status"], "recovered_new")
        self.assertEqual(integration.apply_stage(self.fixture.root, result["stage"])["status"], "already_applied")

    def test_recovery_rechecks_receipt_derived_snapshot_after_catalog_switch(self):
        result = self.stage(name="recovery-tamper")
        stage = self.fixture.root / result["stage"]
        live_snapshot = self.fixture.root / "electron-prototype/catalog-snapshot.json"
        original_atomic_write = integration._atomic_write

        def fail_snapshot_write(path, data):
            if Path(path).resolve() == live_snapshot.resolve():
                raise RuntimeError("injected snapshot interruption")
            return original_atomic_write(path, data)

        integration._atomic_write = fail_snapshot_write
        try:
            with self.assertRaisesRegex(RuntimeError, "injected snapshot interruption"):
                integration.apply_stage(self.fixture.root, result["stage"])
        finally:
            integration._atomic_write = original_atomic_write

        old_snapshot = live_snapshot.read_bytes()
        staged_snapshot_path = stage / "catalog-snapshot.json"
        staged_snapshot = json.loads(staged_snapshot_path.read_text(encoding="utf-8"))
        staged_snapshot["tamperMarker"] = "unreviewed"
        staged_snapshot_bytes = (json.dumps(staged_snapshot, ensure_ascii=False, indent=2) + "\n").encode("utf-8")
        staged_snapshot_path.write_bytes(staged_snapshot_bytes)
        tampered_snapshot_hash = digest(staged_snapshot_bytes)

        plan_path = stage / "plan.json"
        plan = json.loads(plan_path.read_text(encoding="utf-8"))
        plan["staged"]["snapshot"]["sha256"] = tampered_snapshot_hash
        plan_path.write_text(json.dumps(plan, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
        journal_path = stage / "journal.json"
        journal = json.loads(journal_path.read_text(encoding="utf-8"))
        journal["new_snapshot_sha256"] = tampered_snapshot_hash
        journal_path.write_text(json.dumps(journal, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")

        with self.assertRaisesRegex(integration.IntegrationError, "staged snapshot differs from receipt-derived snapshot"):
            integration.recover_stage(self.fixture.root, result["stage"])
        self.assertEqual(live_snapshot.read_bytes(), old_snapshot)

    def test_apply_rejects_destination_junction_escape(self):
        if os.name != "nt":
            self.skipTest("directory junctions are Windows-only")
        result = self.stage(name="destination-junction")
        media = self.fixture.root / "electron-prototype/app/media"
        target_parent = media / self.fixture.variant_id
        junction = target_parent / "walk"
        target_parent.mkdir(parents=True, exist_ok=True)
        with tempfile.TemporaryDirectory() as external_temp:
            external = Path(external_temp) / "walk-target"
            external.mkdir()
            try:
                subprocess.run(["cmd.exe", "/c", "mklink", "/J", str(junction), str(external)], check=True, capture_output=True, text=True)
            except (OSError, subprocess.CalledProcessError) as exc:
                self.skipTest(f"cannot create test junction: {exc}")
            try:
                before_manifest = (self.fixture.root / "electron-prototype/app/media/manifest.json").read_bytes()
                before_snapshot = (self.fixture.root / "electron-prototype/catalog-snapshot.json").read_bytes()
                with self.assertRaisesRegex(integration.IntegrationError, "destination escapes media root"):
                    integration.apply_stage(self.fixture.root, result["stage"])
                self.assertFalse((external / "64-000.png").exists())
                self.assertEqual((self.fixture.root / "electron-prototype/app/media/manifest.json").read_bytes(), before_manifest)
                self.assertEqual((self.fixture.root / "electron-prototype/catalog-snapshot.json").read_bytes(), before_snapshot)
            finally:
                if junction.exists() or junction.is_symlink():
                    os.rmdir(junction)

    def test_rejects_duplicate_mismatch_and_bad_payload_without_mutation(self):
        duplicate = Fixture()
        try:
            receipt = json.loads(duplicate.receipt_path.read_text(encoding="utf-8"))
            receipt["operation"] = "add"
            receipt["variant"]["id"] = "guinea_pig-golden"
            receipt["variant"]["coat"] = "golden"
            receipt["variant"]["coatLabel"] = "ゴールデン"
            path = duplicate.root / "fixtures/duplicate.json"
            path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
            self.refresh_review(path, duplicate.root)
            before = (duplicate.root / "electron-prototype/app/media/manifest.json").read_bytes()
            with self.assertRaisesRegex(integration.IntegrationError, "add requires an absent variant id"):
                integration.stage_receipt(duplicate.root, "fixtures/duplicate.json", "electron-prototype/qa/coat-stages/duplicate")
            self.assertEqual((duplicate.root / "electron-prototype/app/media/manifest.json").read_bytes(), before)
        finally:
            duplicate.close()
        bad = Fixture()
        try:
            receipt = json.loads(bad.receipt_path.read_text(encoding="utf-8"))
            review_path = bad.root / receipt["visual_review"]["path"]
            review = json.loads(review_path.read_text(encoding="utf-8"))
            review["receipt_payload_sha256"] = "0" * 64
            review_bytes = (json.dumps(review, indent=2) + "\n").encode("utf-8")
            review_path.write_bytes(review_bytes)
            receipt["visual_review"]["sha256"] = digest(review_bytes)
            bad.receipt_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
            with self.assertRaisesRegex(integration.IntegrationError, "visual review is not bound to receipt payload"):
                integration.stage_receipt(bad.root, "fixtures/receipt.json", "electron-prototype/qa/coat-stages/bad-review")
        finally:
            bad.close()

    def test_rejects_short_walk_path_escape_bad_timing_alpha_and_presentation(self):
        short = Fixture()
        try:
            receipt = json.loads(short.receipt_path.read_text(encoding="utf-8"))
            receipt["motions"]["walk"]["frames"] = receipt["motions"]["walk"]["frames"][:8]
            receipt["motions"]["walk"]["durations_ms"] = receipt["motions"]["walk"]["durations_ms"][:8]
            path = short.root / "fixtures/short.json"
            path.write_text(json.dumps(receipt, indent=2) + "\n", encoding="utf-8")
            self.refresh_review(path, short.root)
            with self.assertRaisesRegex(integration.IntegrationError, "walk receipt frame count differs"):
                integration.stage_receipt(short.root, "fixtures/short.json", "electron-prototype/qa/coat-stages/short")
        finally:
            short.close()
        escape = Fixture()
        try:
            receipt = json.loads(escape.receipt_path.read_text(encoding="utf-8"))
            receipt["motions"]["walk"]["frames"][0]["path"] = "../outside.png"
            path = escape.root / "fixtures/escape.json"
            path.write_text(json.dumps(receipt, indent=2) + "\n", encoding="utf-8")
            self.refresh_review(path, escape.root)
            with self.assertRaisesRegex(integration.IntegrationError, "escapes repository"):
                integration.stage_receipt(escape.root, "fixtures/escape.json", "electron-prototype/qa/coat-stages/escape")
        finally:
            escape.close()
        invalid = Fixture()
        try:
            receipt = json.loads(invalid.receipt_path.read_text(encoding="utf-8"))
            receipt["motions"]["walk"]["presentation"]["scale"] = 2
            path = invalid.root / "fixtures/presentation.json"
            path.write_text(json.dumps(receipt, indent=2) + "\n", encoding="utf-8")
            self.refresh_review(path, invalid.root)
            with self.assertRaisesRegex(integration.IntegrationError, r"outside \.5\.\.1\.5"):
                integration.stage_receipt(invalid.root, "fixtures/presentation.json", "electron-prototype/qa/coat-stages/presentation")
        finally:
            invalid.close()

    def test_apply_rejects_stage_path_or_catalog_tampering_before_live_write(self):
        result = self.stage(name="tamper")
        stage = self.fixture.root / result["stage"]
        plan_path = stage / "plan.json"
        plan = json.loads(plan_path.read_text(encoding="utf-8"))
        plan["new_files"][0]["stage_path"] = "../outside.png"
        plan_path.write_text(json.dumps(plan, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
        before = (self.fixture.root / "electron-prototype/app/media/manifest.json").read_bytes()
        with self.assertRaises(integration.IntegrationError):
            integration.apply_stage(self.fixture.root, result["stage"])
        self.assertEqual((self.fixture.root / "electron-prototype/app/media/manifest.json").read_bytes(), before)

        result = self.stage(name="tamper-catalog")
        stage = self.fixture.root / result["stage"]
        plan_path = stage / "plan.json"
        plan = json.loads(plan_path.read_text(encoding="utf-8"))
        manifest_path = stage / "manifest.json"
        staged = json.loads(manifest_path.read_text(encoding="utf-8"))
        target = next(item for item in staged["variants"] if item["id"] == "guinea_pig-cream")
        target["motions"]["walk"]["presentation"]["x"] = 0.1
        manifest_bytes = (json.dumps(staged, ensure_ascii=False, indent=2) + "\n").encode("utf-8")
        manifest_path.write_bytes(manifest_bytes)
        new_manifest_hash = digest(manifest_bytes)
        snapshot_path = stage / "catalog-snapshot.json"
        snapshot = json.loads(snapshot_path.read_text(encoding="utf-8"))
        snapshot["manifestSha256"] = new_manifest_hash
        snapshot_bytes = (json.dumps(snapshot, ensure_ascii=False, indent=2) + "\n").encode("utf-8")
        snapshot_path.write_bytes(snapshot_bytes)
        plan["staged"]["manifest"]["sha256"] = new_manifest_hash
        plan["staged"]["snapshot"]["sha256"] = digest(snapshot_bytes)
        plan_path.write_text(json.dumps(plan, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
        before = (self.fixture.root / "electron-prototype/app/media/manifest.json").read_bytes()
        with self.assertRaises(integration.IntegrationError):
            integration.apply_stage(self.fixture.root, result["stage"])
        self.assertEqual((self.fixture.root / "electron-prototype/app/media/manifest.json").read_bytes(), before)

    def test_noop_and_recovery_verify_active_png_hashes(self):
        result = self.stage(name="active-hash")
        integration.apply_stage(self.fixture.root, result["stage"])
        active = self.fixture.root / "electron-prototype/app/media/guinea_pig-golden/walk/64-000.png"
        active.write_bytes(b"changed active image")
        with self.assertRaises(integration.IntegrationError):
            integration.apply_stage(self.fixture.root, result["stage"])
        # Restore a clean applied state through a fresh fixture and then check
        # the recovery path independently.
        self.tearDown()
        self.fixture = Fixture()
        result = self.stage(name="recovery-hash")
        integration.apply_stage(self.fixture.root, result["stage"])
        active = self.fixture.root / "electron-prototype/app/media/guinea_pig-golden/walk/64-000.png"
        active.write_bytes(b"changed active image")
        (self.fixture.root / "electron-prototype/catalog-snapshot.json").write_bytes((self.fixture.root / result["stage"] / "backup/catalog-snapshot.json").read_bytes())
        with self.assertRaises(integration.IntegrationError):
            integration.recover_stage(self.fixture.root, result["stage"])

    def test_rejects_malformed_alpha_timing_and_cropped_exports(self):
        cases = [
            ("opaque", edge_apng_bytes((96, 64), [42] * 30, opaque=True)),
            ("cropped", edge_apng_bytes((96, 64), [42] * 30, opaque=False)),
            ("timing", apng_bytes((96, 64), [43] + [42] * 29, 300)),
        ]
        for name, export_bytes in cases:
            fixture = Fixture()
            try:
                receipt = json.loads(fixture.receipt_path.read_text(encoding="utf-8"))
                relative = f"fixtures/{name}-walk-64.png"
                descriptor = write_bytes(fixture.root, relative, export_bytes)
                receipt["motions"]["walk"]["exports"]["64"] = descriptor
                receipt_path = fixture.root / f"fixtures/{name}-receipt.json"
                receipt_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
                # Rebind the parent review after changing the receipt payload.
                self.refresh_review(receipt_path, fixture.root)
                with self.assertRaises(integration.IntegrationError):
                    integration.stage_receipt(fixture.root, f"fixtures/{name}-receipt.json", f"electron-prototype/qa/coat-stages/{name}")
            finally:
                fixture.close()


class ReplaceReviewedIdleTests(unittest.TestCase):
    """All media are synthetic TEST fixtures in a disposable root."""

    refresh_review = IntegrateReviewedCoatTests.refresh_review
    test_idle_only_recovery_rechecks_receipt_derived_snapshot = IntegrateReviewedCoatTests.test_recovery_rechecks_receipt_derived_snapshot_after_catalog_switch
    idle_durations = [1083, 334, 666, 334, 1166, 250, 125, 42]

    def setUp(self):
        self.fixture = Fixture(degu=True, degu_coat="sand", idle_durations=self.idle_durations)
        self.root = self.fixture.root
        self.manifest_path = self.root / integration.MANIFEST_REL
        self.snapshot_path = self.root / integration.SNAPSHOT_REL
        self.media = self.root / integration.MEDIA_REL
        self.receipt_path = self.fixture.receipt_path
        # Extra metadata makes accidental reconstruction of the walk, other
        # rows, or snapshot rows visible rather than comparing only counts.
        catalog = self.read(self.manifest_path)
        target = self.target(catalog)
        target["customVariantMetadata"] = {"keep": True}
        target["motions"]["walk"]["provenance"] = {"legacy": "eight actual provisional poses"}
        target["motions"]["walk"]["presentation"] = {"scale": 0.93, "x": 0.01, "y": -0.01}
        self.write(self.manifest_path, catalog)
        snapshot = self.read(self.snapshot_path)
        next(row for row in snapshot["variants"] if row["id"] == "degu-sand")["customRowMetadata"] = ["keep"]
        snapshot["inputs"] = [{"path": "fixture-history.json", "role": "preserve history"}]
        snapshot["manifestSha256"] = digest(self.manifest_path.read_bytes())
        self.write(self.snapshot_path, snapshot)
        self.settings_path = self.root / "fixture-profile/settings.json"
        write_bytes(self.root, "fixture-profile/settings.json", b'{"animals":["degu-sand","rabbit-ruby_eyed_white"],"size":96}\n')
        receipt = self.read(self.receipt_path)
        self.full_walk = receipt["motions"].pop("walk")
        receipt["operation"] = "replace-idle"
        receipt["expected_catalog_sha256"] = digest(self.manifest_path.read_bytes())
        receipt["expected_snapshot_sha256"] = digest(self.snapshot_path.read_bytes())
        # Retain original nonsequential source IDs and selected video times.
        idle = receipt["motions"]["idle"]
        reference = self.read(self.root / idle["base_reference"]["path"])
        source = self.read(self.root / idle["source_manifest"]["path"])
        for index, video_index in enumerate([0, 26, 34, 50, 58, 86, 92, 95]):
            for frame in (reference["frames"][index], source["frames"][index]):
                frame.update(id=video_index + 1, video_frame_zero_based=video_index, reference_time_s=video_index / 24)
            reference["frames"][index]["cell"] = index + 1
        idle["source_manifest"] = self.write(self.root / idle["source_manifest"]["path"], source)
        reference["source_manifest_sha256"] = idle["source_manifest"]["sha256"]
        idle["base_reference"] = self.write(self.root / idle["base_reference"]["path"], reference)
        self.save_receipt(receipt)
        self.before_catalog = self.read(self.manifest_path)
        self.before_snapshot = self.read(self.snapshot_path)
        self.before_live = self.live_bytes()

    def tearDown(self):
        self.fixture.close()

    @staticmethod
    def read(path):
        return json.loads(path.read_text(encoding="utf-8"))

    def write(self, path, value):
        return write_bytes(self.root, path.relative_to(self.root).as_posix(), (json.dumps(value, ensure_ascii=False, indent=2) + "\n").encode("utf-8"))

    @staticmethod
    def target(catalog):
        return next(row for row in catalog["variants"] if row["id"] == "degu-sand")

    def save_receipt(self, receipt, *, rebind=True):
        self.write(self.receipt_path, receipt)
        if rebind:
            self.refresh_review(self.receipt_path)

    def live_bytes(self):
        paths = list(self.media.rglob("*")) + [self.snapshot_path, self.settings_path]
        return {path.relative_to(self.root).as_posix(): path.read_bytes() for path in paths if path.is_file()}

    def assert_old_bytes_preserved(self):
        for relative, data in self.before_live.items():
            if relative not in {integration.MANIFEST_REL.as_posix(), integration.SNAPSHOT_REL.as_posix()}:
                self.assertEqual((self.root / relative).read_bytes(), data, relative)

    def stage(self, name="idle"):
        return integration.stage_receipt(self.root, "fixtures/receipt.json", f"electron-prototype/qa/coat-stages/{name}")

    def test_idle_only_stage_apply_preserves_walk8_and_every_unrelated_byte(self):
        # No new walk input is needed and old provisional walk provenance is
        # not misrepresented as a new full30 source acceptance.
        walk_source = self.root / "fixtures/degu-sand/walk"
        walk_source.rename(walk_source.with_name("unused-walk-fixture"))
        result = self.stage()
        self.assertEqual(self.live_bytes(), self.before_live)
        self.assertEqual(result["plan"]["operation"], "replace-idle")
        self.assertEqual(len(result["plan"]["new_files"]), 16)
        self.assertEqual(result["plan"]["old_target_paths"], [])
        self.assertTrue(all("/idle/" in row["path"] for row in result["plan"]["new_files"]))
        staged = self.read(self.root / result["stage"] / "manifest.json")
        target = self.target(staged)
        self.assertEqual(target["motions"]["walk"], self.target(self.before_catalog)["motions"]["walk"])
        self.assertEqual(len(target["motions"]["walk"]["durations"]), 8)
        self.assertEqual(sum(target["motions"]["walk"]["durations"]), 556)
        self.assertEqual(target["motions"]["idle"]["durations"], self.idle_durations)
        self.assertEqual(sum(target["motions"]["idle"]["durations"]), 4000)
        self.assertFalse(target["idleFallback"])
        self.assertEqual(target["frameCount"], 16)
        self.assertEqual(target["customVariantMetadata"], {"keep": True})
        self.assertEqual(target["files"][:16], self.target(self.before_catalog)["files"])
        idle64 = [row for row in target["files"] if "/idle/64-" in row["path"]]
        self.assertEqual([row["source"]["base_frame_id"] for row in idle64], [1, 27, 35, 51, 59, 87, 93, 96])
        self.assertEqual([row["source"]["base_source_frame_index"] for row in idle64], [0, 26, 34, 50, 58, 86, 92, 95])
        self.assertEqual([row["source"]["cell"] for row in idle64], list(range(1, 9)))
        self.assertEqual(staged["files"][:len(self.before_catalog["files"])], self.before_catalog["files"])
        self.assertEqual(staged["defaultId"], self.before_catalog["defaultId"])
        self.assertEqual([v["id"] for v in staged["variants"]], [v["id"] for v in self.before_catalog["variants"]])
        self.assertEqual([v for v in staged["variants"] if v["id"] != "degu-sand"], [v for v in self.before_catalog["variants"] if v["id"] != "degu-sand"])
        staged_snapshot = self.read(self.root / result["stage"] / "catalog-snapshot.json")
        expected_rows = copy.deepcopy(self.before_snapshot["variants"])
        next(row for row in expected_rows if row["id"] == "degu-sand").update(idleFrames=8, idleFallback=False)
        self.assertEqual(staged_snapshot["variants"], expected_rows)
        self.assertEqual(staged_snapshot["inputs"][:-1], self.before_snapshot["inputs"])
        self.assertEqual(integration.apply_stage(self.root, result["stage"])["status"], "applied")
        self.assert_old_bytes_preserved()
        self.assertEqual(integration.apply_stage(self.root, result["stage"])["status"], "already_applied")
        self.assertEqual(integration.recover_stage(self.root, result["stage"])["status"], "already_new")

    def test_idle_only_replaces_real_idle_append_only_and_preserves_walk(self):
        first = self.stage("first-idle")
        integration.apply_stage(self.root, first["stage"])
        old_catalog = self.read(self.manifest_path)
        old_idle = self.target(old_catalog)["motions"]["idle"]
        old_idle_paths = {path for tier in old_idle["tiers"].values() for path in tier}
        old_idle_bytes = {path: (self.media / path).read_bytes() for path in old_idle_paths}
        receipt = self.read(self.receipt_path)
        receipt["expected_catalog_sha256"] = digest(self.manifest_path.read_bytes())
        receipt["expected_snapshot_sha256"] = digest(self.snapshot_path.read_bytes())
        self.save_receipt(receipt)
        result = self.stage("second-idle")
        self.assertEqual(set(result["plan"]["old_target_paths"]), old_idle_paths)
        self.assertTrue(old_idle_paths.isdisjoint(result["plan"]["target_paths"]))
        integration.apply_stage(self.root, result["stage"])
        new_catalog = self.read(self.manifest_path)
        self.assertEqual(self.target(new_catalog)["motions"]["walk"], self.target(old_catalog)["motions"]["walk"])
        self.assertEqual(len(new_catalog["files"]), len(old_catalog["files"]))
        self.assertTrue(old_idle_paths.isdisjoint(row["path"] for row in new_catalog["files"]))
        for path, data in old_idle_bytes.items():
            self.assertEqual((self.media / path).read_bytes(), data)

    def test_idle_only_receipt_shape_target_mapping_and_review_rejections(self):
        original = self.read(self.receipt_path)
        def wrong_id(receipt):
            receipt["variant"].update(id="degu-absent", coat="absent")
        cases = [
            ("complete-replace-missing-walk", lambda r: r.update(operation="replace"), "walk and idle"),
            ("complete-add-missing-walk", lambda r: r.update(operation="add"), "walk and idle"),
            ("unexpected-walk", lambda r: r["motions"].update(walk=self.full_walk), "only idle"),
            ("top-level-walk", lambda r: r.update(walk=self.full_walk), "receipt fields differ"),
            ("wrong-id", wrong_id, "requires an existing variant"),
            ("wrong-label", lambda r: r["variant"].update(coatLabel="wrong"), "metadata differs"),
            ("wrong-species", lambda r: r["variant"].update(id="rabbit-sand", species="rabbit", speciesLabel="うさぎ"), "species/motion differs"),
            ("timing", lambda r: r["motions"]["idle"]["durations_ms"].__setitem__(0, 1084), "receipt durations"),
            ("reordered", lambda r: r["motions"]["idle"]["frames"].reverse(), "same-index base source"),
            ("base-hash", lambda r: r["motions"]["idle"]["frames"][0].update(base_source_sha256="0" * 64), "same-index base source"),
            ("source-hash", lambda r: r["motions"]["idle"]["source_manifest"].update(sha256="0" * 64), "hash changed"),
            ("provenance-hash", lambda r: r["motions"]["idle"]["frames"][0]["provenance"].update(sha256="0" * 64), "hash changed"),
            ("escape", lambda r: r["motions"]["idle"]["frames"][0].update(path="../outside.png"), "escapes repository"),
            ("cropping-transform", lambda r: r["motions"]["idle"]["presentation"].update(x=2), "leaves canvas"),
        ]
        for name, mutate, error in cases:
            with self.subTest(name=name):
                receipt = copy.deepcopy(original)
                mutate(receipt)
                self.save_receipt(receipt)
                with self.assertRaisesRegex(integration.IntegrationError, error):
                    self.stage(name)
                self.assertEqual(self.live_bytes(), self.before_live)
                self.assertFalse((self.root / integration.STAGE_ROOT_REL / name).exists())
        receipt = copy.deepcopy(original)
        frames = receipt["motions"]["idle"]["frames"]
        frames[1].update(path=frames[0]["path"], sha256=frames[0]["sha256"])
        self.save_receipt(receipt)
        with self.assertRaisesRegex(integration.IntegrationError, "edited source poses are duplicated"):
            self.stage("duplicate")
        # A changed payload cannot borrow the previous review's approval.
        self.save_receipt(original)
        receipt = self.read(self.receipt_path)
        receipt["motions"]["idle"]["presentation"]["x"] = 0.001
        self.save_receipt(receipt, rebind=False)
        with self.assertRaisesRegex(integration.IntegrationError, "not bound to receipt payload"):
            self.stage("unbound-review")
        self.assertEqual(self.live_bytes(), self.before_live)

    def test_idle_rejects_duplicate_or_backwards_source_video_indices(self):
        original = self.read(self.receipt_path)
        idle = original["motions"]["idle"]
        reference_path = self.root / idle["base_reference"]["path"]
        source_path = self.root / idle["source_manifest"]["path"]
        original_reference = self.read(reference_path)
        original_source = self.read(source_path)
        for name, video_index in (("duplicate-time", 0), ("backwards-time", 35)):
            with self.subTest(name=name):
                receipt = copy.deepcopy(original)
                reference = copy.deepcopy(original_reference)
                source = copy.deepcopy(original_source)
                # Keep IDs, source/edited PNGs and selected array positions
                # unique. Forge only the time/index binding in both manifests.
                for frame in (reference["frames"][1], source["frames"][1]):
                    frame.update(video_frame_zero_based=video_index, reference_time_s=video_index / 24)
                motion = receipt["motions"]["idle"]
                motion["source_manifest"] = self.write(source_path, source)
                reference["source_manifest_sha256"] = motion["source_manifest"]["sha256"]
                motion["base_reference"] = self.write(reference_path, reference)
                self.save_receipt(receipt)
                with self.assertRaisesRegex(integration.IntegrationError, "unique and strictly increasing"):
                    self.stage(name)
                self.assertEqual(self.live_bytes(), self.before_live)
                self.assertFalse((self.root / integration.STAGE_ROOT_REL / name).exists())

    def test_idle_only_rejects_invalid_apng_exports(self):
        original = self.read(self.receipt_path)
        cases = [
            ("opaque", edge_apng_bytes((96, 64), self.idle_durations, opaque=True)),
            ("cropped", edge_apng_bytes((96, 64), self.idle_durations)),
            ("duration", apng_bytes((96, 64), [1000] * 8, 20)),
            ("size", apng_bytes((144, 96), self.idle_durations, 20)),
            ("static", png_bytes((96, 64), 1, (30, 40, 50))),
            ("count", apng_bytes((96, 64), self.idle_durations[:-1], 20)),
        ]
        for name, data in cases:
            with self.subTest(name=name):
                receipt = copy.deepcopy(original)
                receipt["motions"]["idle"]["exports"]["64"] = write_bytes(self.root, f"fixtures/bad-{name}.png", data)
                self.save_receipt(receipt)
                with self.assertRaises(integration.IntegrationError):
                    self.stage(name)
                self.assertEqual(self.live_bytes(), self.before_live)

    def test_idle_only_apply_rejects_source_baseline_and_destination_drift(self):
        result = self.stage()
        receipt = self.read(self.receipt_path)
        source_path = self.root / receipt["motions"]["idle"]["frames"][0]["path"]
        original_source = source_path.read_bytes()
        source_path.write_bytes(b"changed source")
        with self.assertRaisesRegex(integration.IntegrationError, "hash changed"):
            integration.apply_stage(self.root, result["stage"])
        source_path.write_bytes(original_source)
        walk_path = self.media / self.target(self.before_catalog)["motions"]["walk"]["tiers"]["64"][0]
        original_walk = walk_path.read_bytes()
        walk_path.write_bytes(b"changed walk")
        with self.assertRaisesRegex(integration.IntegrationError, "catalog image hash changed"):
            integration.apply_stage(self.root, result["stage"])
        walk_path.write_bytes(original_walk)
        new_path = self.media / result["plan"]["new_files"][0]["path"]
        new_path.parent.mkdir(parents=True, exist_ok=True)
        new_path.write_bytes(b"conflicting destination")
        with self.assertRaisesRegex(integration.IntegrationError, "conflict"):
            integration.apply_stage(self.root, result["stage"])
        self.assertEqual(self.manifest_path.read_bytes(), self.before_live[integration.MANIFEST_REL.as_posix()])
        self.assertEqual(self.snapshot_path.read_bytes(), self.before_live[integration.SNAPSHOT_REL.as_posix()])
        self.assert_old_bytes_preserved()

    def test_idle_only_apply_rejects_walk_tampering_even_with_updated_plan_hashes(self):
        result = self.stage()
        stage = self.root / result["stage"]
        catalog = self.read(stage / "manifest.json")
        self.target(catalog)["motions"]["walk"]["presentation"]["x"] = 0.02
        manifest_desc = self.write(stage / "manifest.json", catalog)
        snapshot = self.read(stage / "catalog-snapshot.json")
        snapshot["manifestSha256"] = manifest_desc["sha256"]
        snapshot_desc = self.write(stage / "catalog-snapshot.json", snapshot)
        plan = self.read(stage / "plan.json")
        plan["staged"]["manifest"]["sha256"] = manifest_desc["sha256"]
        plan["staged"]["snapshot"]["sha256"] = snapshot_desc["sha256"]
        self.write(stage / "plan.json", plan)
        with self.assertRaisesRegex(integration.IntegrationError, "receipt-derived catalog"):
            integration.apply_stage(self.root, result["stage"])
        self.assertEqual(self.live_bytes(), self.before_live)

    def test_idle_only_interrupted_switch_recovery_and_verified_noop(self):
        result = self.stage()
        original_atomic = integration._atomic_write
        def interrupt_snapshot(path, data):
            if path == self.snapshot_path:
                raise RuntimeError("injected idle snapshot interruption")
            return original_atomic(path, data)
        integration._atomic_write = interrupt_snapshot
        try:
            with self.assertRaisesRegex(RuntimeError, "injected idle"):
                integration.apply_stage(self.root, result["stage"])
        finally:
            integration._atomic_write = original_atomic
        self.assertEqual(self.snapshot_path.read_bytes(), self.before_live[integration.SNAPSHOT_REL.as_posix()])
        self.assertEqual(self.target(self.read(self.manifest_path))["motions"]["walk"], self.target(self.before_catalog)["motions"]["walk"])
        self.assertEqual(integration.recover_stage(self.root, result["stage"])["status"], "recovered_new")
        self.assertEqual(integration.apply_stage(self.root, result["stage"])["status"], "already_applied")
        self.assert_old_bytes_preserved()
        walk_path = self.media / self.target(self.before_catalog)["motions"]["walk"]["tiers"]["64"][0]
        walk_path.write_bytes(b"changed retained walk PNG")
        for action in (integration.apply_stage, integration.recover_stage):
            with self.assertRaisesRegex(integration.IntegrationError, "catalog image hash changed"):
                action(self.root, result["stage"])

    def test_idle_only_precommit_interruption_recovers_old_and_can_resume(self):
        result = self.stage()
        original_copy = integration._copy_exclusive_or_verify
        def interrupt_copy(*args, **kwargs):
            original_copy(*args, **kwargs)
            raise RuntimeError("injected copy interruption")
        integration._copy_exclusive_or_verify = interrupt_copy
        try:
            with self.assertRaisesRegex(RuntimeError, "injected copy"):
                integration.apply_stage(self.root, result["stage"])
        finally:
            integration._copy_exclusive_or_verify = original_copy
        self.assertEqual(integration.recover_stage(self.root, result["stage"])["status"], "already_old")
        self.assertEqual(self.manifest_path.read_bytes(), self.before_live[integration.MANIFEST_REL.as_posix()])
        self.assertEqual(integration.apply_stage(self.root, result["stage"])["status"], "applied")
        self.assert_old_bytes_preserved()

    def test_idle_only_recovery_rejects_journal_and_backup_tampering(self):
        result = self.stage()
        integration.apply_stage(self.root, result["stage"])
        applied_live = self.live_bytes()
        stage = self.root / result["stage"]
        journal_path = stage / "journal.json"
        journal = self.read(journal_path)
        tampered = copy.deepcopy(journal)
        tampered["operation"] = "replace"
        self.write(journal_path, tampered)
        with self.assertRaisesRegex(integration.IntegrationError, "journal does not bind exact stage"):
            integration.recover_stage(self.root, result["stage"])
        self.write(journal_path, journal)
        (stage / "backup/manifest.json").write_bytes(b"changed backup")
        with self.assertRaisesRegex(integration.IntegrationError, "immutable old manifest backup"):
            integration.recover_stage(self.root, result["stage"])
        self.assertEqual(self.live_bytes(), applied_live)

    def test_idle_only_apply_rejects_destination_junction_escape(self):
        if os.name != "nt":
            self.skipTest("directory junctions are Windows-only")
        result = self.stage()
        junction = self.media / "degu-sand/idle"
        preserved = junction.with_name("unused-idle-fixture")
        # These are only orphan TEST files created by Fixture's old fallback.
        self.assertTrue(junction.resolve().is_relative_to(self.root.resolve()))
        self.assertTrue(preserved.resolve().is_relative_to(self.root.resolve()))
        junction.rename(preserved)
        with tempfile.TemporaryDirectory() as external_temp:
            external = Path(external_temp) / "idle-target"
            external.mkdir()
            try:
                subprocess.run(["cmd.exe", "/c", "mklink", "/J", str(junction), str(external)], check=True, capture_output=True, text=True)
            except (OSError, subprocess.CalledProcessError) as exc:
                self.skipTest(f"cannot create test junction: {exc}")
            try:
                with self.assertRaisesRegex(integration.IntegrationError, "destination escapes media root"):
                    integration.apply_stage(self.root, result["stage"])
                self.assertEqual(list(external.iterdir()), [])
                self.assertEqual(self.manifest_path.read_bytes(), self.before_live[integration.MANIFEST_REL.as_posix()])
                self.assertEqual(self.snapshot_path.read_bytes(), self.before_live[integration.SNAPSHOT_REL.as_posix()])
            finally:
                if junction.exists() or junction.is_symlink():
                    os.rmdir(junction)



if __name__ == "__main__":
    unittest.main()
