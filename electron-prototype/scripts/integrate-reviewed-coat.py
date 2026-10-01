#!/usr/bin/env python3
"""Targeted, receipt-bound integration for one reviewed coat.

This module deliberately does not know about the production coat ledger.  A
parent review receipt is the authority for one complete walk+idle import or an
explicit replace-idle import that preserves the existing walk verbatim.  The
stage command validates that receipt and materialises an isolated catalog; the
apply command is the only command that switches the live catalog.
"""

from __future__ import annotations

import argparse
import copy
import datetime as _datetime
import hashlib
import io
import json
import math
import os
import re
import shutil
import sys
import uuid
from contextlib import contextmanager
from pathlib import Path, PurePath
from typing import Any

from PIL import Image


SCHEMA = 1
CATALOG_SCHEMA = 2
MEDIA_REL = Path("electron-prototype/app/media")
MANIFEST_REL = MEDIA_REL / "manifest.json"
SNAPSHOT_REL = Path("electron-prototype/catalog-snapshot.json")
STAGE_ROOT_REL = Path("electron-prototype/qa/coat-stages")
LOCK_REL = Path(".codex/locks/integrate-reviewed-coat.lock")
SHA_RE = re.compile(r"^[0-9a-f]{64}$")
ID_RE = re.compile(r"^[a-z0-9_-]+$")
MEDIA_RE = re.compile(r"^(?P<variant>[a-z0-9_-]+)/(walk|idle)/(?P<tier>64|96)-(?P<index>[0-9]{3,})\.png$")
EPSILON = 0.01


class IntegrationError(ValueError):
    """A receipt, catalog, stage, or transaction validation failure."""


def _fail(message: str) -> None:
    raise IntegrationError(message)


def sha256_bytes(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def sha256_file(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as handle:
        for block in iter(lambda: handle.read(1024 * 1024), b""):
            digest.update(block)
    return digest.hexdigest()


def canonical_json_bytes(value: Any) -> bytes:
    try:
        text = json.dumps(
            value,
            ensure_ascii=False,
            sort_keys=True,
            separators=(",", ":"),
            allow_nan=False,
        )
    except (TypeError, ValueError) as exc:
        _fail(f"cannot canonicalize JSON: {exc}")
    return text.encode("utf-8")


def canonical_payload_bytes(receipt: dict[str, Any]) -> bytes:
    if not isinstance(receipt, dict):
        _fail("receipt must be an object")
    payload = copy.deepcopy(receipt)
    payload.pop("visual_review", None)
    return canonical_json_bytes(payload)


def receipt_payload_hash(receipt: dict[str, Any]) -> str:
    return sha256_bytes(canonical_payload_bytes(receipt))


# Friendly aliases for callers that want to bind a review before invoking the
# CLI.  Keeping these names here also makes the hash rule easy to discover.
payload_hash = receipt_payload_hash
canonical_receipt_payload_hash = receipt_payload_hash


def _default_root() -> Path:
    return Path(__file__).resolve().parents[2]


def _root_path(root: Path | str | None) -> Path:
    path = Path(root) if root is not None else _default_root()
    try:
        return path.resolve(strict=True)
    except FileNotFoundError:
        _fail(f"repository root does not exist: {path}")


def _repo_relative(value: Any, root: Path, *, label: str, must_exist: bool = True) -> Path:
    """Resolve a repo-relative path and reject traversal and symlink escapes."""

    if not isinstance(value, str) or not value.strip():
        _fail(f"{label} must be a non-empty repository-relative path")
    if "\x00" in value:
        _fail(f"{label} contains NUL")
    # Paths in manifests are POSIX-style even on Windows.  Reject all absolute
    # spellings before Path normalisation can reinterpret them.
    if value.startswith(("/", "\\")) or re.match(r"^[A-Za-z]:[\\/]", value):
        _fail(f"{label} must be repository-relative: {value}")
    parts = PurePath(value).parts
    if ".." in parts:
        _fail(f"{label} escapes repository: {value}")
    base = root.resolve(strict=True)
    candidate = (base / Path(value)).resolve(strict=False)
    try:
        candidate.relative_to(base)
    except ValueError:
        _fail(f"{label} escapes repository: {value}")
    if must_exist and not candidate.exists():
        _fail(f"{label} does not exist: {value}")
    return candidate


def _relative_string(path: Path, root: Path) -> str:
    try:
        return path.resolve(strict=False).relative_to(root.resolve(strict=True)).as_posix()
    except ValueError:
        _fail(f"path is outside repository: {path}")


def _ensure_inside(path: Path, root: Path, label: str) -> Path:
    resolved = path.resolve(strict=True)
    try:
        resolved.relative_to(root.resolve(strict=True))
    except ValueError:
        _fail(f"{label} escapes its owner: {path}")
    return resolved


def _check_sha(value: Any, label: str) -> str:
    if not isinstance(value, str) or SHA_RE.fullmatch(value) is None:
        _fail(f"{label} must be a lowercase SHA-256")
    return value


def _descriptor(value: Any, label: str) -> dict[str, str]:
    if not isinstance(value, dict) or set(value) != {"path", "sha256"}:
        _fail(f"{label} must be {{path, sha256}}")
    if not isinstance(value["path"], str):
        _fail(f"{label}.path must be a string")
    return {"path": value["path"], "sha256": _check_sha(value["sha256"], f"{label}.sha256")}


def _read_descriptor(root: Path, value: Any, label: str) -> tuple[dict[str, str], bytes, Path]:
    descriptor = _descriptor(value, label)
    path = _repo_relative(descriptor["path"], root, label=f"{label}.path")
    if not path.is_file():
        _fail(f"{label}.path is not a file: {descriptor['path']}")
    data = path.read_bytes()
    actual = sha256_bytes(data)
    if actual != descriptor["sha256"]:
        _fail(f"{label} hash changed: {descriptor['path']}")
    return descriptor, data, path


def _json_bytes(data: bytes, label: str) -> dict[str, Any]:
    try:
        value = json.loads(data.decode("utf-8"))
    except (UnicodeDecodeError, json.JSONDecodeError) as exc:
        _fail(f"{label} is not UTF-8 JSON: {exc}")
    if not isinstance(value, dict):
        _fail(f"{label} must contain a JSON object")
    return value


def _read_json_descriptor(root: Path, value: Any, label: str) -> tuple[dict[str, str], dict[str, Any], Path]:
    descriptor, data, path = _read_descriptor(root, value, label)
    return descriptor, _json_bytes(data, label), path


def _finite_number(value: Any, label: str) -> float:
    if isinstance(value, bool) or not isinstance(value, (int, float)) or not math.isfinite(float(value)):
        _fail(f"{label} must be finite")
    return float(value)


def _positive_duration_list(value: Any, label: str) -> list[float | int]:
    if not isinstance(value, list) or not value:
        _fail(f"{label} must be a non-empty array")
    result: list[float | int] = []
    for index, item in enumerate(value):
        number = _finite_number(item, f"{label}[{index}]")
        if number <= 0:
            _fail(f"{label}[{index}] must be positive")
        result.append(item)
    return result


def _numbers_equal(left: Any, right: Any, label: str) -> None:
    if not isinstance(left, list) or not isinstance(right, list) or len(left) != len(right):
        _fail(f"{label} length differs")
    for index, (a, b) in enumerate(zip(left, right)):
        if not math.isclose(_finite_number(a, f"{label}[{index}]") , _finite_number(b, f"{label}[{index}]") , abs_tol=EPSILON):
            _fail(f"{label}[{index}] differs")


def _sum_close(values: list[float | int], expected: Any, label: str) -> None:
    target = _finite_number(expected, label)
    if not math.isclose(sum(float(v) for v in values), target, abs_tol=EPSILON):
        _fail(f"{label} does not equal duration sum")


def _load_json_file(path: Path, label: str) -> dict[str, Any]:
    return _json_bytes(path.read_bytes(), label)


def _png_rgba(data: bytes, label: str, *, expected_size: tuple[int, int] | None = None) -> tuple[Image.Image, list[tuple[int, int, int, int]], list[float]]:
    try:
        animation = Image.open(io.BytesIO(data))
    except Exception as exc:  # Pillow raises several format-specific errors.
        _fail(f"{label} is not a readable PNG: {exc}")
    if animation.format != "PNG":
        _fail(f"{label} is not PNG")
    if not getattr(animation, "is_animated", False) or getattr(animation, "n_frames", 1) < 1:
        _fail(f"{label} must be an animated PNG")
    if animation.info.get("default_image"):
        _fail(f"{label} contains a default poster frame")
    if expected_size is not None and animation.size != expected_size:
        _fail(f"{label} dimensions are {animation.size}, expected {expected_size}")
    frames: list[tuple[int, int, int, int]] = []
    durations: list[float] = []
    first: Image.Image | None = None
    for index in range(animation.n_frames):
        try:
            animation.seek(index)
            duration = animation.info.get("duration")
            if duration is None:
                _fail(f"{label} frame {index} has no duration")
            durations.append(_finite_number(duration, f"{label} duration {index}"))
            frame = animation.convert("RGBA")
            if first is None:
                first = frame.copy()
            alpha = frame.getchannel("A")
            minimum, maximum = alpha.getextrema()
            if minimum != 0 or maximum < 16:
                _fail(f"{label} frame {index} has unusable transparency")
            bbox = alpha.point(lambda value: 255 if value >= 16 else 0).getbbox()
            if bbox is None:
                _fail(f"{label} frame {index} has no visible animal")
            if bbox[0] <= 0 or bbox[1] <= 0 or bbox[2] >= frame.width or bbox[3] >= frame.height:
                _fail(f"{label} frame {index} is cropped at the canvas edge")
            frames.append(bbox)
        except IntegrationError:
            raise
        except Exception as exc:
            _fail(f"{label} frame {index} cannot be decoded: {exc}")
    assert first is not None
    return first, frames, durations


def _decode_apng_frames(data: bytes, label: str, tier: str) -> tuple[list[bytes], list[tuple[int, int, int, int]], list[float]]:
    expected_size = (96, 64) if tier == "64" else (144, 96)
    try:
        animation = Image.open(io.BytesIO(data))
    except Exception as exc:
        _fail(f"{label} is not a readable PNG: {exc}")
    if animation.format != "PNG" or not getattr(animation, "is_animated", False):
        _fail(f"{label} must be an animated PNG")
    if animation.info.get("default_image"):
        _fail(f"{label} contains a default poster frame")
    if animation.size != expected_size:
        _fail(f"{label} dimensions are {animation.size}, expected {expected_size}")
    png_frames: list[bytes] = []
    boxes: list[tuple[int, int, int, int]] = []
    durations: list[float] = []
    for index in range(animation.n_frames):
        animation.seek(index)
        duration = animation.info.get("duration")
        durations.append(_finite_number(duration, f"{label} duration {index}"))
        image = animation.convert("RGBA")
        alpha = image.getchannel("A")
        minimum, maximum = alpha.getextrema()
        if minimum != 0 or maximum < 16:
            _fail(f"{label} frame {index} has unusable transparency")
        bbox = alpha.point(lambda value: 255 if value >= 16 else 0).getbbox()
        if bbox is None:
            _fail(f"{label} frame {index} has no visible animal")
        if bbox[0] <= 0 or bbox[1] <= 0 or bbox[2] >= image.width or bbox[3] >= image.height:
            _fail(f"{label} frame {index} is cropped at the canvas edge")
        boxes.append(bbox)
        output = io.BytesIO()
        image.save(output, format="PNG", optimize=False)
        png_frames.append(output.getvalue())
    return png_frames, boxes, durations


def _validate_presentation(value: Any, label: str, *, strict: bool = True) -> dict[str, float | int]:
    if not isinstance(value, dict) or not {"scale", "x", "y"}.issubset(value):
        _fail(f"{label} must contain exactly scale, x, y")
    if strict and set(value) != {"scale", "x", "y"}:
        _fail(f"{label} must contain exactly scale, x, y")
    scale = _finite_number(value["scale"], f"{label}.scale")
    x = _finite_number(value["x"], f"{label}.x")
    y = _finite_number(value["y"], f"{label}.y")
    if scale < 0.5 or scale > 1.5:
        _fail(f"{label}.scale is outside .5..1.5")
    return {"scale": value["scale"], "x": value["x"], "y": value["y"]}


def _check_transformed_boxes(boxes: list[tuple[int, int, int, int]], presentation: dict[str, float | int], tier: str, label: str) -> None:
    height = 64 if tier == "64" else 96
    width = 96 if tier == "64" else 144
    scale = float(presentation["scale"])
    # Presentation offsets are normalised to the canvas height, matching the
    # existing manifest's presentation records.
    offset_x = float(presentation["x"]) * height
    offset_y = float(presentation["y"]) * height
    for index, (x0, y0, x1, y1) in enumerate(boxes):
        transformed = (x0 * scale + offset_x, y0 * scale + offset_y, x1 * scale + offset_x, y1 * scale + offset_y)
        if transformed[0] < 0 or transformed[1] < 0 or transformed[2] > width or transformed[3] > height:
            _fail(f"{label} frame {index} leaves canvas after presentation transform")


def _validate_catalog(root: Path, catalog: dict[str, Any], media_dir: Path, *, check_files: bool = True) -> dict[str, Any]:
    if not isinstance(catalog, dict) or catalog.get("schema") != CATALOG_SCHEMA:
        _fail("catalog must use schema 2")
    variants = catalog.get("variants")
    files = catalog.get("files")
    if not isinstance(variants, list) or not variants or not isinstance(files, list) or not files:
        _fail("catalog variants/files are missing")
    _ensure_inside(media_dir, root, "catalog media directory")
    if not isinstance(catalog.get("defaultId"), str):
        _fail("catalog defaultId is missing")
    file_map: dict[str, dict[str, Any]] = {}
    for index, record in enumerate(files):
        if not isinstance(record, dict) or not isinstance(record.get("path"), str):
            _fail(f"catalog file {index} is malformed")
        path = record["path"]
        match = MEDIA_RE.fullmatch(path)
        if match is None or path in file_map:
            _fail(f"catalog file path is invalid or duplicated: {path}")
        _check_sha(record.get("sha256"), f"catalog file {path}.sha256")
        media_path = _repo_relative(path, media_dir, label=f"catalog file {path}", must_exist=check_files)
        if check_files and sha256_file(media_path) != record["sha256"]:
            _fail(f"catalog image hash changed: {path}")
        file_map[path] = record
    ids: set[str] = set()
    for variant in variants:
        if not isinstance(variant, dict) or not isinstance(variant.get("id"), str):
            _fail("catalog variant is malformed")
        variant_id = variant["id"]
        if ID_RE.fullmatch(variant_id) is None or variant_id in ids:
            _fail(f"catalog variant id is invalid or duplicated: {variant_id}")
        ids.add(variant_id)
        if not isinstance(variant.get("species"), str) or not isinstance(variant.get("coat"), str):
            _fail(f"catalog variant metadata is missing: {variant_id}")
        motions = variant.get("motions")
        if not isinstance(motions, dict):
            _fail(f"catalog variant motions are missing: {variant_id}")
        used: set[str] = set()
        for action in ("walk", "idle"):
            motion = motions.get(action)
            if not isinstance(motion, dict):
                _fail(f"catalog variant motion is missing: {variant_id}/{action}")
            durations = _positive_duration_list(motion.get("durations"), f"catalog {variant_id}/{action} durations")
            tiers = motion.get("tiers")
            if not isinstance(tiers, dict):
                _fail(f"catalog tiers are missing: {variant_id}/{action}")
            fallback = action == "idle" and variant.get("idleFallback") is True
            for tier in ("64", "96"):
                paths = tiers.get(tier)
                if not isinstance(paths, list) or len(paths) != len(durations):
                    _fail(f"catalog tier length differs: {variant_id}/{action}/{tier}")
                for path in paths:
                    if not isinstance(path, str) or path not in file_map or not path.startswith(variant_id + "/"):
                        _fail(f"catalog tier points outside variant: {variant_id}/{action}/{tier}")
                    if fallback:
                        if len(paths) != 1 or paths[0] != motions["walk"]["tiers"][tier][0]:
                            _fail(f"catalog fallback idle is malformed: {variant_id}")
                    elif not path.startswith(f"{variant_id}/{action}/{tier}-"):
                        _fail(f"catalog tier path has wrong action: {path}")
                    used.add(path)
            presentation = motion.get("presentation")
            if presentation is not None:
                _validate_presentation(presentation, f"catalog {variant_id}/{action} presentation", strict=False)
        indexed = variant.get("files")
        if not isinstance(indexed, list) or len(indexed) != len(used):
            _fail(f"catalog variant file index is malformed: {variant_id}")
        indexed_map: dict[str, str] = {}
        for record in indexed:
            if not isinstance(record, dict) or record.get("path") not in used:
                _fail(f"catalog variant file index points outside used files: {variant_id}")
            indexed_map[record["path"]] = record.get("sha256")
        if set(indexed_map) != used:
            _fail(f"catalog variant file index differs: {variant_id}")
        for path, digest in indexed_map.items():
            if digest != file_map[path]["sha256"]:
                _fail(f"catalog variant file hash differs: {path}")
        if variant.get("frameCount") != sum(len(motions[action]["durations"]) for action in ("walk", "idle")):
            _fail(f"catalog frameCount differs: {variant_id}")
    if catalog["defaultId"] not in ids:
        _fail("catalog defaultId is not present")
    # No active image may be orphaned.  Old target images can remain physically
    # on disk after a replacement, but they must no longer be listed in the
    # active catalog.
    used_all = {path for variant in variants for path in [record["path"] for record in variant["files"]]}
    if used_all != set(file_map):
        _fail("catalog contains orphaned active file records")
    return file_map


def _validate_snapshot(snapshot: dict[str, Any], catalog: dict[str, Any], catalog_hash: str) -> None:
    if not isinstance(snapshot, dict) or snapshot.get("manifestSha256") != catalog_hash:
        _fail("snapshot does not bind the catalog hash")
    ids = [variant.get("id") for variant in catalog.get("variants", [])]
    snapshot_ids = [variant.get("id") for variant in snapshot.get("variants", [])]
    if ids != snapshot_ids:
        _fail("snapshot variant rows differ from catalog")
    if snapshot.get("images") != len(catalog.get("files", [])):
        _fail("snapshot image count differs from catalog")


def _validate_live_baseline(root: Path, expected_catalog_hash: str, expected_snapshot_hash: str) -> tuple[dict[str, Any], dict[str, Any], dict[str, Any]]:
    manifest_path = root / MANIFEST_REL
    snapshot_path = root / SNAPSHOT_REL
    if not manifest_path.is_file() or not snapshot_path.is_file():
        _fail("live catalog or snapshot is missing")
    catalog_bytes = manifest_path.read_bytes()
    snapshot_bytes = snapshot_path.read_bytes()
    if sha256_bytes(catalog_bytes) != expected_catalog_hash:
        _fail("live catalog baseline hash differs")
    if sha256_bytes(snapshot_bytes) != expected_snapshot_hash:
        _fail("live snapshot baseline hash differs")
    catalog = _json_bytes(catalog_bytes, "live catalog")
    snapshot = _json_bytes(snapshot_bytes, "live snapshot")
    file_map = _validate_catalog(root, catalog, root / MEDIA_REL)
    _validate_snapshot(snapshot, catalog, expected_catalog_hash)
    return catalog, snapshot, file_map


def _reference_frames(root: Path, action: str, variant: dict[str, Any], motion: dict[str, Any]) -> dict[str, Any]:
    reference_desc, reference, _ = _read_json_descriptor(root, motion["base_reference"], f"{action}.base_reference")
    source_desc, source_manifest, _ = _read_json_descriptor(root, motion["source_manifest"], f"{action}.source_manifest")
    if reference.get("status") != "ready":
        _fail(f"{action} base reference is not ready")
    if reference.get("species") != variant["species"] or reference.get("motion") != action:
        _fail(f"{action} base reference species/motion differs")
    if reference.get("source_manifest") != source_desc["path"] or reference.get("source_manifest_sha256") != source_desc["sha256"]:
        _fail(f"{action} base reference source manifest binding differs")
    if source_manifest.get("species") != variant["species"] or source_manifest.get("action") != action:
        _fail(f"{action} source manifest species/motion differs")
    reference_frames = reference.get("frames")
    source_frames = source_manifest.get("frames")
    if not isinstance(reference_frames, list) or not isinstance(source_frames, list):
        _fail(f"{action} reference/source frames are missing")
    reference_source_count = reference.get("source_frame_count")
    if isinstance(reference_source_count, bool) or not isinstance(reference_source_count, int) or reference_source_count != len(source_frames):
        _fail(f"{action} reference source_frame_count differs")
    if "frame_count" in source_manifest:
        source_count = source_manifest.get("frame_count")
        if isinstance(source_count, bool) or not isinstance(source_count, int) or source_count != len(source_frames):
            _fail(f"{action} source manifest frame_count differs")
    count = reference.get("output_frame_count")
    if count != len(reference_frames) or not isinstance(count, int):
        _fail(f"{action} reference output_frame_count differs")
    durations = _positive_duration_list(reference.get("frame_durations_ms"), f"{action} base durations")
    if len(durations) != count:
        _fail(f"{action} reference duration count differs")
    _sum_close(durations, reference.get("cycle_duration_ms"), f"{action} reference cycle")
    # Full30 references repeat video metadata at the reference level.  The
    # sampled idle handoffs intentionally keep it only in their source
    # manifest, so derive the same binding there when it is absent.
    video = reference.get("video", source_manifest.get("video"))
    video_sha = _check_sha(reference.get("video_sha256", source_manifest.get("video_sha256")), f"{action} reference video_sha256")
    if not isinstance(video, str):
        _fail(f"{action} reference video path is missing")
    video_path = _repo_relative(video, root, label=f"{action} reference video")
    if sha256_file(video_path) != video_sha:
        _fail(f"{action} reference video hash changed")
    source_video = source_manifest.get("video")
    source_video_sha = source_manifest.get("video_sha256")
    if source_video != video or source_video_sha != video_sha:
        _fail(f"{action} source manifest video binding differs")
    video_fps = _finite_number(reference.get("video_fps", source_manifest.get("video_fps")), f"{action} reference video_fps")
    source_fps = _finite_number(source_manifest.get("video_fps"), f"{action} source manifest video_fps")
    if not math.isclose(video_fps, 24.0, abs_tol=EPSILON) or not math.isclose(source_fps, 24.0, abs_tol=EPSILON):
        _fail(f"{action} source video must be 24 fps")
    selected = reference.get("selected_indices_zero_based")
    if selected is None:
        source_indices = list(range(count))
    else:
        if not isinstance(selected, list) or len(selected) != count:
            _fail(f"{action} selected source indices differ")
        source_indices = []
        for index, source_index in enumerate(selected):
            if isinstance(source_index, bool) or not isinstance(source_index, int) or source_index < 0:
                _fail(f"{action} selected source index {index} is invalid")
            source_indices.append(source_index)
    if any(index >= len(source_frames) for index in source_indices):
        _fail(f"{action} selected source frame is outside source manifest")
    ids: list[Any] = []
    source_hashes: list[str] = []
    video_indices: list[int] = []
    for index, (frame, source_index) in enumerate(zip(reference_frames, source_indices)):
        if not isinstance(frame, dict):
            _fail(f"{action} reference frame {index} is malformed")
        frame_id = frame.get("id")
        if isinstance(frame_id, bool) or not isinstance(frame_id, int) or frame_id <= 0:
            _fail(f"{action} reference frame id is invalid at {index}")
        frame_sha = _check_sha(frame.get("source_sha256"), f"{action} reference frame {index}.source_sha256")
        frame_source = frame.get("source")
        if not isinstance(frame_source, str):
            _fail(f"{action} reference frame {index}.source is missing")
        source_path = _repo_relative(frame_source, root, label=f"{action} reference frame {index}.source")
        if sha256_file(source_path) != frame_sha:
            _fail(f"{action} reference frame {index} source hash changed")
        source_frame = source_frames[source_index]
        if not isinstance(source_frame, dict) or source_frame.get("source") != frame_source or source_frame.get("source_sha256") != frame_sha:
            _fail(f"{action} source manifest frame binding differs at {index}")
        if source_frame.get("id") != frame_id:
            _fail(f"{action} source manifest frame id differs at {index}")
        if source_frame.get("video_sha256") != video_sha or frame.get("video_sha256") != video_sha:
            _fail(f"{action} frame video hash differs at {index}")
        video_index = frame.get("video_frame_zero_based")
        if isinstance(video_index, bool) or not isinstance(video_index, int) or video_index < 0:
            _fail(f"{action} frame video index is invalid at {index}")
        reference_time = _finite_number(frame.get("reference_time_s"), f"{action} reference_time_s[{index}]")
        if not math.isclose(reference_time, video_index / video_fps, abs_tol=EPSILON):
            _fail(f"{action} reference_time_s does not bind video index at {index}")
        source_video_index = source_frame.get("video_frame_zero_based")
        if isinstance(source_video_index, bool) or not isinstance(source_video_index, int) or source_video_index != video_index:
            _fail(f"{action} source manifest video index differs at {index}")
        source_reference_time = _finite_number(source_frame.get("reference_time_s"), f"{action} source reference_time_s[{index}]")
        if not math.isclose(source_reference_time, reference_time, abs_tol=EPSILON):
            _fail(f"{action} source manifest reference_time_s differs at {index}")
        ids.append(frame_id)
        source_hashes.append(frame_sha)
        video_indices.append(video_index)
    if len(set(ids)) != count or len(set(source_hashes)) != count:
        _fail(f"{action} reference source poses are duplicated")
    if action == "idle" and any(later <= earlier for earlier, later in zip(video_indices, video_indices[1:])):
        # Distinct PNGs and IDs must still describe distinct chronological
        # video frames. Sparse sampled idle poses are valid; repeated or
        # backwards video positions are not. Times are bound to these indices
        # above, so this also preserves chronological source timing.
        _fail("idle source-video frames must be unique and strictly increasing")
    if action == "walk":
        if count != 30:
            _fail("walk base reference must contain exactly 30 frames")
        if ids != list(range(1, 31)):
            _fail("walk base reference frame ids are not ordered")
        if any(video_indices[index] != video_indices[0] + index for index in range(count)):
            _fail("walk base reference video frames are not consecutive")
        expected_cycle = 556 if variant["species"] == "degu" else 1250
        if not math.isclose(float(reference["cycle_duration_ms"]), expected_cycle, abs_tol=EPSILON):
            _fail(f"walk base reference must use {expected_cycle} ms playback")
        # The reference playback is deliberately separate from source 24 fps;
        # require an original 30-frame/24-fps span rather than treating 556 ms
        # degu playback as source-video timing.
        if not math.isclose((video_indices[-1] - video_indices[0] + 1) / video_fps, 30 / 24, abs_tol=EPSILON):
            _fail("walk source timing is not a 30-frame 24 fps sequence")
    return {
        "reference_descriptor": reference_desc,
        "reference": reference,
        "source_descriptor": source_desc,
        "source_manifest": source_manifest,
        "reference_frames": reference_frames,
        "source_indices": source_indices,
        "durations": durations,
        "video_sha256": video_sha,
    }


def _validate_motion(root: Path, action: str, variant: dict[str, Any], motion: dict[str, Any]) -> dict[str, Any]:
    if not isinstance(motion, dict):
        _fail(f"{action} receipt motion is missing")
    expected = {"base_reference", "source_manifest", "frames", "durations_ms", "exports", "presentation"}
    if set(motion) != expected:
        _fail(f"{action} receipt motion fields differ")
    base = _reference_frames(root, action, variant, motion)
    frames = motion["frames"]
    if not isinstance(frames, list) or len(frames) != len(base["reference_frames"]):
        _fail(f"{action} receipt frame count differs from base reference")
    if action == "walk" and len(frames) != 30:
        _fail("walk receipt must contain exactly 30 frames")
    if action == "idle" and len(frames) < 2:
        _fail("idle receipt must contain at least two real frames")
    durations = _positive_duration_list(motion["durations_ms"], f"{action} receipt durations")
    _numbers_equal(durations, base["durations"], f"{action} receipt durations")
    presentation = _validate_presentation(motion["presentation"], f"{action} receipt presentation")
    export_descriptors = motion["exports"]
    if not isinstance(export_descriptors, dict) or set(export_descriptors) != {"64", "96"}:
        _fail(f"{action} receipt exports must contain 64 and 96")
    edited_hashes: list[str] = []
    frame_records: list[dict[str, Any]] = []
    for index, (frame, reference_frame) in enumerate(zip(frames, base["reference_frames"])):
        if not isinstance(frame, dict):
            _fail(f"{action} receipt frame {index} is malformed")
        if set(frame) != {"path", "sha256", "base_source_sha256", "provenance"}:
            _fail(f"{action} receipt frame {index} fields differ")
        edited_desc, edited_bytes, edited_path = _read_descriptor(root, {"path": frame["path"], "sha256": frame["sha256"]}, f"{action} edited frame {index}")
        base_sha = _check_sha(frame["base_source_sha256"], f"{action} frame {index}.base_source_sha256")
        if base_sha != reference_frame["source_sha256"]:
            _fail(f"{action} frame {index} is not bound to same-index base source")
        provenance, _, _ = _read_descriptor(root, frame["provenance"], f"{action} frame {index}.provenance")
        try:
            source_image = Image.open(io.BytesIO(edited_bytes))
            if source_image.format != "PNG" or getattr(source_image, "is_animated", False):
                _fail(f"{action} edited frame {index} must be an individual PNG")
            rgba = source_image.convert("RGBA")
            minimum, maximum = rgba.getchannel("A").getextrema()
            if minimum != 0 or maximum < 16:
                _fail(f"{action} edited frame {index} has unusable transparency")
            if rgba.getchannel("A").point(lambda value: 255 if value >= 16 else 0).getbbox() is None:
                _fail(f"{action} edited frame {index} has no visible animal")
        except IntegrationError:
            raise
        except Exception as exc:
            _fail(f"{action} edited frame {index} cannot be decoded: {exc}")
        edited_hashes.append(edited_desc["sha256"])
        frame_records.append({
            "index": index,
            "source": edited_desc,
            "base_source_sha256": base_sha,
            "provenance": provenance,
            "base_frame_id": reference_frame["id"],
            "base_source_frame_index": reference_frame["video_frame_zero_based"],
            "base_cell": reference_frame.get("cell"),
        })
    if len(set(edited_hashes)) != len(edited_hashes):
        _fail(f"{action} edited source poses are duplicated")
    if action == "idle" and [record["base_source_sha256"] for record in frame_records] != [frame["source_sha256"] for frame in base["reference_frames"]]:
        _fail("idle edited frames are not ordered as the base reference")
    decoded_by_tier: dict[str, list[bytes]] = {}
    boxes_by_tier: dict[str, list[tuple[int, int, int, int]]] = {}
    durations_by_tier: dict[str, list[float]] = {}
    for tier in ("64", "96"):
        export_desc, export_bytes, _ = _read_descriptor(root, export_descriptors[tier], f"{action} export {tier}")
        decoded, boxes, export_durations = _decode_apng_frames(export_bytes, f"{action} export {tier}", tier)
        if len(decoded) != len(frames):
            _fail(f"{action} export {tier} frame count differs")
        _numbers_equal(export_durations, durations, f"{action} export {tier} durations")
        _check_transformed_boxes(boxes, presentation, tier, f"{action} export {tier}")
        decoded_by_tier[tier] = decoded
        boxes_by_tier[tier] = boxes
        durations_by_tier[tier] = export_durations
        export_descriptors[tier] = export_desc
    _numbers_equal(durations_by_tier["64"], durations_by_tier["96"], f"{action} tier durations")
    return {
        "base": base,
        "frames": frame_records,
        "durations": durations,
        "presentation": presentation,
        "exports": export_descriptors,
        "decoded": decoded_by_tier,
        "boxes": boxes_by_tier,
    }


def _validate_review(root: Path, receipt: dict[str, Any], payload_digest: str) -> dict[str, Any]:
    review_desc, review, _ = _read_json_descriptor(root, receipt["visual_review"], "visual_review")
    if review.get("status") != "pass":
        _fail("visual review status is not pass")
    variant_id = receipt["variant"]["id"]
    if review.get("variant_id") != variant_id or review.get("scope") != "source_export_presentation":
        _fail("visual review variant or scope differs")
    if review.get("receipt_payload_sha256") != payload_digest:
        _fail("visual review is not bound to receipt payload")
    evidence = review.get("evidence")
    if not isinstance(evidence, list) or not evidence:
        _fail("visual review evidence must be a non-empty descriptor list")
    evidence_records = []
    for index, descriptor in enumerate(evidence):
        normalized, _, _ = _read_descriptor(root, descriptor, f"visual_review.evidence[{index}]")
        evidence_records.append(normalized)
    return {"descriptor": review_desc, "review": review, "evidence": evidence_records}


def _validate_receipt(root: Path, receipt: dict[str, Any]) -> dict[str, Any]:
    if receipt.get("schema") != SCHEMA or receipt.get("operation") not in {"add", "replace", "replace-idle"}:
        _fail("receipt schema or operation is invalid")
    idle_only = receipt["operation"] == "replace-idle"
    if idle_only and set(receipt) != {"schema", "operation", "expected_catalog_sha256", "expected_snapshot_sha256", "variant", "motions", "visual_review"}:
        _fail("replace-idle receipt fields differ")
    for key in ("expected_catalog_sha256", "expected_snapshot_sha256"):
        _check_sha(receipt.get(key), f"receipt.{key}")
    variant = receipt.get("variant")
    if not isinstance(variant, dict) or set(variant) != {"id", "species", "speciesLabel", "coat", "coatLabel"}:
        _fail("receipt.variant fields differ")
    for key in ("id", "species", "speciesLabel", "coat", "coatLabel"):
        if not isinstance(variant[key], str) or not variant[key]:
            _fail(f"receipt.variant.{key} is missing")
    if ID_RE.fullmatch(variant["id"]) is None or variant["id"] != f"{variant['species']}-{variant['coat']}":
        _fail("receipt variant id does not match species-coat")
    motions = receipt.get("motions")
    actions = ("idle",) if idle_only else ("walk", "idle")
    if not isinstance(motions, dict) or set(motions) != set(actions):
        _fail("replace-idle receipt must contain only idle" if idle_only else "receipt must contain walk and idle")
    if "visual_review" not in receipt:
        _fail("receipt.visual_review is missing")
    payload_digest = receipt_payload_hash(receipt)
    review = _validate_review(root, receipt, payload_digest)
    motion_results = {action: _validate_motion(root, action, variant, motions[action]) for action in actions}
    return {
        "receipt": receipt,
        "payload_hash": payload_digest,
        "variant": copy.deepcopy(variant),
        "motions": motion_results,
        "review": review,
    }


def _load_receipt(root: Path, receipt_path: str | Path) -> tuple[dict[str, Any], str, str]:
    path = _repo_relative(str(receipt_path), root, label="receipt", must_exist=True)
    data = path.read_bytes()
    receipt = _json_bytes(data, "receipt")
    return receipt, sha256_bytes(data), _relative_string(path, root)


def _allocate_paths(root: Path, catalog: dict[str, Any], variant_id: str, action: str, tier: str, count: int) -> list[str]:
    media = root / MEDIA_REL
    highest = -1
    directory = media / variant_id / action
    if directory.is_dir():
        for path in directory.iterdir():
            match = re.fullmatch(rf"{re.escape(tier)}-(\d+)\.png", path.name)
            if match:
                highest = max(highest, int(match.group(1)))
    for record in catalog.get("files", []):
        path = record.get("path", "")
        match = re.fullmatch(rf"{re.escape(variant_id)}/{action}/{re.escape(tier)}-(\d+)\.png", path)
        if match:
            highest = max(highest, int(match.group(1)))
    result = []
    for index in range(highest + 1, highest + 1 + count):
        path = f"{variant_id}/{action}/{tier}-{index:03d}.png"
        if path in result or (media / path).exists():
            _fail(f"cannot allocate new image path: {path}")
        result.append(path)
    return result


def _motion_provenance(root: Path, receipt_rel: str, receipt_hash: str, result: dict[str, Any], action: str) -> dict[str, Any]:
    base = result["base"]
    return {
        "receipt": {"path": receipt_rel, "sha256": receipt_hash},
        "base_reference": base["reference_descriptor"],
        "source_manifest": base["source_descriptor"],
        "visual_review": result.get("review_descriptor"),
        "action": action,
    }


def _build_motion_with_paths(variant_id: str, action: str, result: dict[str, Any], receipt_rel: str, receipt_hash: str, paths_by_tier: dict[str, list[str]]) -> tuple[dict[str, Any], list[dict[str, Any]], dict[str, bytes]]:
    count = len(result["frames"])
    records: list[dict[str, Any]] = []
    image_bytes_by_path: dict[str, bytes] = {}
    for tier in ("64", "96"):
        paths = paths_by_tier[tier]
        if not isinstance(paths, list) or len(paths) != count:
            _fail(f"planned {action}/{tier} path count differs")
        for index, (path, image_bytes, frame) in enumerate(zip(paths, result["decoded"][tier], result["frames"])):
            image_bytes_by_path[path] = image_bytes
            records.append({
                "path": path,
                "sha256": sha256_bytes(image_bytes),
                "source": {
                    "receipt": {"path": receipt_rel, "sha256": receipt_hash},
                    "export": result["exports"][tier],
                    "frame": index,
                    "base_frame_id": frame["base_frame_id"],
                    "base_source_frame_index": frame["base_source_frame_index"],
                    "base_source_sha256": frame["base_source_sha256"],
                    "edited_source": frame["source"],
                    "provenance": frame["provenance"],
                    "cell": frame["base_cell"],
                },
            })
    motion = {
        "durations": result["durations"],
        "tiers": paths_by_tier,
        "presentation": result["presentation"],
        "provenance": {
            "receipt": {"path": receipt_rel, "sha256": receipt_hash},
            "base_reference": result["base"]["reference_descriptor"],
            "source_manifest": result["base"]["source_descriptor"],
            "exports": result["exports"],
            "action": action,
        },
    }
    return motion, records, image_bytes_by_path


def _build_motion_and_records(root: Path, variant_id: str, action: str, result: dict[str, Any], receipt_rel: str, receipt_hash: str, stage_media: Path, catalog: dict[str, Any]) -> tuple[dict[str, Any], list[dict[str, Any]], list[str]]:
    count = len(result["frames"])
    paths_by_tier = {tier: _allocate_paths(root, catalog, variant_id, action, tier, count) for tier in ("64", "96")}
    motion, records, image_bytes_by_path = _build_motion_with_paths(variant_id, action, result, receipt_rel, receipt_hash, paths_by_tier)
    for path, image_bytes in image_bytes_by_path.items():
        destination = stage_media / path
        destination.parent.mkdir(parents=True, exist_ok=True)
        destination.write_bytes(image_bytes)
    return motion, records, list(image_bytes_by_path)


def _new_snapshot(old: dict[str, Any], catalog: dict[str, Any], catalog_hash: str, receipt_rel: str, receipt_hash: str, *, operation: str, variant_id: str) -> dict[str, Any]:
    snapshot = copy.deepcopy(old)
    snapshot["snapshotAt"] = _datetime.datetime.now(_datetime.timezone.utc).isoformat()
    snapshot["inputs"] = copy.deepcopy(old.get("inputs", []))
    snapshot["inputs"].append({"path": receipt_rel, "sha256": receipt_hash, "role": "targeted reviewed coat integration"})
    target = next(item for item in catalog["variants"] if item["id"] == variant_id)
    target_row = {
        "id": target["id"],
        "label": f"{target['speciesLabel']} / {target['coatLabel']}",
        "walkFrames": len(target["motions"]["walk"]["durations"]),
        "idleFrames": len(target["motions"]["idle"]["durations"]),
        "idleFallback": target.get("idleFallback", False),
    }
    old_rows = copy.deepcopy(old.get("variants", []))
    if operation == "add":
        if any(row.get("id") == variant_id for row in old_rows):
            _fail(f"snapshot already contains target variant: {variant_id}")
        snapshot["variants"] = old_rows + [target_row]
    elif operation in {"replace", "replace-idle"}:
        replaced = False
        snapshot["variants"] = []
        for row in old_rows:
            if row.get("id") == variant_id:
                # An idle-only transaction cannot rewrite walk metadata or
                # any extra fields retained by an existing snapshot row.
                if operation == "replace-idle":
                    row.update({key: target_row[key] for key in ("idleFrames", "idleFallback")})
                    snapshot["variants"].append(row)
                else:
                    snapshot["variants"].append(target_row)
                replaced = True
            else:
                snapshot["variants"].append(row)
        if not replaced:
            _fail(f"snapshot target row is missing: {variant_id}")
    else:
        _fail(f"unsupported snapshot operation: {operation}")
    snapshot["images"] = len(catalog["files"])
    snapshot["manifestSha256"] = catalog_hash
    return snapshot


def _write_json(path: Path, value: Any) -> bytes:
    data = (json.dumps(value, ensure_ascii=False, indent=2) + "\n").encode("utf-8")
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_bytes(data)
    return data


def _atomic_write(path: Path, data: bytes) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    temporary = path.with_name(f".{path.name}.{uuid.uuid4().hex}.tmp")
    with temporary.open("wb") as handle:
        handle.write(data)
        handle.flush()
        os.fsync(handle.fileno())
    os.replace(temporary, path)


def _stage_path(root: Path, output: str | Path, *, must_exist: bool = False) -> Path:
    value = str(output)
    path = _repo_relative(value, root, label="stage output", must_exist=False)
    stage_root = (root / STAGE_ROOT_REL).resolve(strict=False)
    try:
        path.relative_to(stage_root)
    except ValueError:
        _fail("stage output must be under electron-prototype/qa/coat-stages")
    if must_exist and not path.is_dir():
        _fail(f"stage directory does not exist: {value}")
    return path


def _replaced_paths(existing: dict[str, Any] | None, *, idle_only: bool) -> set[str]:
    if existing is None:
        return set()
    paths = {record["path"] for record in existing["files"]}
    if idle_only:
        # A fallback idle points at walk PNGs.  They remain active, including
        # their original records, even when that fallback is removed.
        walk_paths = {path for tier_paths in existing["motions"]["walk"]["tiers"].values() for path in tier_paths}
        paths -= walk_paths
    return paths


def stage_receipt(root: Path | str, receipt_path: str | Path, output: str | Path) -> dict[str, Any]:
    root_path = _root_path(root)
    (root_path / STAGE_ROOT_REL).mkdir(parents=True, exist_ok=True)
    stage = _stage_path(root_path, output, must_exist=False)
    if stage.exists():
        _fail(f"stage directory already exists: {output}")
    receipt, receipt_hash, receipt_rel = _load_receipt(root_path, receipt_path)
    expected_catalog = _check_sha(receipt.get("expected_catalog_sha256"), "receipt.expected_catalog_sha256")
    expected_snapshot = _check_sha(receipt.get("expected_snapshot_sha256"), "receipt.expected_snapshot_sha256")
    catalog, snapshot, file_map = _validate_live_baseline(root_path, expected_catalog, expected_snapshot)
    validated = _validate_receipt(root_path, receipt)
    idle_only = receipt["operation"] == "replace-idle"
    variant = validated["variant"]
    existing = next((item for item in catalog["variants"] if item["id"] == variant["id"]), None)
    if receipt["operation"] == "add":
        if existing is not None:
            _fail(f"add requires an absent variant id: {variant['id']}")
    else:
        if existing is None:
            _fail(f"replace requires an existing variant id: {variant['id']}")
        for key in ("species", "speciesLabel", "coat", "coatLabel"):
            if existing.get(key) != variant[key]:
                _fail(f"replace variant metadata differs: {key}")
    species = next((item for item in catalog["variants"] if item.get("species") == variant["species"]), None)
    if species is None:
        _fail(f"species is not present in current catalog: {variant['species']}")
    if species.get("speciesLabel") != variant["speciesLabel"]:
        _fail("variant speciesLabel differs from current catalog species")
    stage.mkdir(parents=True)
    stage_media = stage / "media"
    baseline_dir = stage / "baseline"
    baseline_dir.mkdir(parents=True)
    baseline_dir.joinpath("manifest.json").write_bytes((root_path / MANIFEST_REL).read_bytes())
    baseline_dir.joinpath("catalog-snapshot.json").write_bytes((root_path / SNAPSHOT_REL).read_bytes())
    for record in catalog["files"]:
        source = root_path / MEDIA_REL / record["path"]
        destination = stage_media / record["path"]
        destination.parent.mkdir(parents=True, exist_ok=True)
        shutil.copyfile(source, destination)
    receipt_copy = stage / "receipt.json"
    receipt_copy.write_bytes((root_path / receipt_rel).read_bytes())
    target_old_paths = _replaced_paths(existing, idle_only=idle_only)
    target_records: list[dict[str, Any]] = []
    target_paths: list[str] = []
    result_motions: dict[str, dict[str, Any]] = {}
    for action in validated["motions"]:
        result = validated["motions"][action]
        motion, records, paths = _build_motion_and_records(root_path, variant["id"], action, result, receipt_rel, receipt_hash, stage_media, catalog)
        result_motions[action] = motion
        target_records.extend(records)
        target_paths.extend(paths)
    if existing is None:
        new_variant = {
            **variant,
            "motions": result_motions,
            "idleFallback": False,
            "files": copy.deepcopy(target_records),
            "frameCount": sum(len(result_motions[action]["durations"]) for action in ("walk", "idle")),
        }
        new_catalog = copy.deepcopy(catalog)
        new_catalog["variants"].append(new_variant)
    else:
        new_catalog = copy.deepcopy(catalog)
        replacement = next(item for item in new_catalog["variants"] if item["id"] == variant["id"])
        if idle_only:
            replacement["motions"]["idle"] = result_motions["idle"]
            replacement["files"] = [record for record in replacement["files"] if record["path"] not in target_old_paths] + copy.deepcopy(target_records)
        else:
            replacement["motions"] = result_motions
            replacement["files"] = copy.deepcopy(target_records)
        replacement["idleFallback"] = False
        replacement["frameCount"] = sum(len(replacement["motions"][action]["durations"]) for action in ("walk", "idle"))
    active_without_target = [record for record in catalog["files"] if record["path"] not in target_old_paths]
    new_catalog["files"] = active_without_target + target_records
    new_catalog["snapshotAt"] = _datetime.datetime.now(_datetime.timezone.utc).isoformat()
    _validate_catalog(root_path, new_catalog, stage_media)
    manifest_bytes = (json.dumps(new_catalog, ensure_ascii=False, indent=2) + "\n").encode("utf-8")
    manifest_path = stage / "manifest.json"
    manifest_path.write_bytes(manifest_bytes)
    manifest_hash = sha256_bytes(manifest_bytes)
    new_snapshot = _new_snapshot(snapshot, new_catalog, manifest_hash, receipt_rel, receipt_hash, operation=receipt["operation"], variant_id=variant["id"])
    _validate_snapshot(new_snapshot, new_catalog, manifest_hash)
    snapshot_bytes = _write_json(stage / "catalog-snapshot.json", new_snapshot)
    snapshot_hash = sha256_bytes(snapshot_bytes)
    baseline_files = [{"path": record["path"], "sha256": record["sha256"]} for record in catalog["files"]]
    new_files = [{"path": record["path"], "sha256": record["sha256"], "stage_path": f"media/{record['path']}"} for record in target_records]
    staged_files = [{"path": record["path"], "sha256": record["sha256"], "stage_path": f"media/{record['path']}"} for record in catalog["files"]] + new_files
    plan = {
        "schema": SCHEMA,
        "operation": receipt["operation"],
        "variant": variant,
        "receipt": {"path": receipt_rel, "sha256": receipt_hash},
        "receipt_payload_sha256": validated["payload_hash"],
        "baseline": {
            "catalog": {"path": MANIFEST_REL.as_posix(), "sha256": expected_catalog},
            "snapshot": {"path": SNAPSHOT_REL.as_posix(), "sha256": expected_snapshot},
            "stage_manifest": {"path": "baseline/manifest.json", "sha256": expected_catalog},
            "stage_snapshot": {"path": "baseline/catalog-snapshot.json", "sha256": expected_snapshot},
            "files": baseline_files,
        },
        "staged": {
            "manifest": {"path": "manifest.json", "sha256": manifest_hash},
            "snapshot": {"path": "catalog-snapshot.json", "sha256": snapshot_hash},
            "files": staged_files,
        },
        "new_files": new_files,
        "old_target_paths": sorted(target_old_paths),
        "target_paths": target_paths,
        "created_at": _datetime.datetime.now(_datetime.timezone.utc).isoformat(),
    }
    _write_json(stage / "plan.json", plan)
    evidence = {
        "schema": SCHEMA,
        "operation": receipt["operation"],
        "variant": variant,
        "receipt_sha256": receipt_hash,
        "receipt_payload_sha256": validated["payload_hash"],
        "baseline_catalog_sha256": expected_catalog,
        "baseline_snapshot_sha256": expected_snapshot,
        "new_catalog_sha256": manifest_hash,
        "new_snapshot_sha256": snapshot_hash,
        "staged_file_count": len(staged_files),
        "new_file_count": len(new_files),
        "source_review_verified": True,
        "runtime_or_package_acceptance": False,
    }
    evidence_bytes = _write_json(stage / "evidence.json", evidence)
    # Keep the evidence hash discoverable without a self-referential plan.
    plan["evidence"] = {"path": "evidence.json", "sha256": sha256_bytes(evidence_bytes)}
    _write_json(stage / "plan.json", plan)
    return {"stage": _relative_string(stage, root_path), "plan": plan, "evidence": evidence}


def _load_stage(root: Path | str, stage_path: str | Path) -> tuple[Path, dict[str, Any], dict[str, Any], bytes, str]:
    root_path = _root_path(root)
    stage = _stage_path(root_path, stage_path, must_exist=True)
    plan_path = stage / "plan.json"
    if not plan_path.is_file():
        _fail("stage plan is missing")
    plan = _load_json_file(plan_path, "stage plan")
    if plan.get("schema") != SCHEMA:
        _fail("stage plan schema differs")
    def stage_file(value: Any, label: str) -> Path:
        if not isinstance(value, str) or not value or value.startswith(("/", "\\")) or re.match(r"^[A-Za-z]:[\\/]", value) or ".." in PurePath(value).parts:
            _fail(f"{label} escapes stage: {value}")
        candidate = (stage / Path(value)).resolve(strict=False)
        try:
            candidate.relative_to(stage.resolve(strict=True))
        except ValueError:
            _fail(f"{label} escapes stage: {value}")
        return candidate

    manifest_stage_path = stage_file(plan.get("staged", {}).get("manifest", {}).get("path"), "stage manifest path")
    snapshot_stage_path = stage_file(plan.get("staged", {}).get("snapshot", {}).get("path"), "stage snapshot path")
    manifest_bytes = manifest_stage_path.read_bytes()
    snapshot_bytes = snapshot_stage_path.read_bytes()
    manifest_hash = sha256_bytes(manifest_bytes)
    snapshot_hash = sha256_bytes(snapshot_bytes)
    if manifest_hash != plan["staged"]["manifest"]["sha256"] or snapshot_hash != plan["staged"]["snapshot"]["sha256"]:
        _fail("stage catalog or snapshot hash differs")
    staged_catalog = _json_bytes(manifest_bytes, "staged catalog")
    staged_snapshot = _json_bytes(snapshot_bytes, "staged snapshot")
    _ensure_inside(stage / "media", stage, "stage media directory")
    _validate_catalog(root_path, staged_catalog, stage / "media")
    _validate_snapshot(staged_snapshot, staged_catalog, manifest_hash)
    baseline_manifest_path = stage_file(plan.get("baseline", {}).get("stage_manifest", {}).get("path"), "stage baseline manifest path")
    baseline_snapshot_path = stage_file(plan.get("baseline", {}).get("stage_snapshot", {}).get("path"), "stage baseline snapshot path")
    if not baseline_manifest_path.is_file() or sha256_file(baseline_manifest_path) != plan["baseline"]["stage_manifest"]["sha256"]:
        _fail("stage baseline manifest hash differs")
    if not baseline_snapshot_path.is_file() or sha256_file(baseline_snapshot_path) != plan["baseline"]["stage_snapshot"]["sha256"]:
        _fail("stage baseline snapshot hash differs")
    staged_records = plan.get("staged", {}).get("files")
    if not isinstance(staged_records, list):
        _fail("stage file plan is missing")
    staged_plan_map: dict[str, str] = {}
    for index, record in enumerate(staged_records):
        if not isinstance(record, dict) or not isinstance(record.get("path"), str) or not isinstance(record.get("stage_path"), str):
            _fail(f"stage file plan record {index} is malformed")
        media_match = MEDIA_RE.fullmatch(record["path"])
        if media_match is None or record["stage_path"] != f"media/{record['path']}" or record["path"] in staged_plan_map:
            _fail(f"stage file plan record {index} has an invalid path")
        _check_sha(record.get("sha256"), f"stage file plan {record['path']}.sha256")
        path = stage_file(record["stage_path"], f"stage file plan {record['path']}")
        if not path.is_file() or sha256_file(path) != record["sha256"]:
            _fail(f"staged file hash differs: {record.get('path')}")
        staged_plan_map[record["path"]] = record["sha256"]
    staged_catalog_map = {record["path"]: record["sha256"] for record in staged_catalog["files"]}
    if any(path not in staged_plan_map or staged_plan_map[path] != digest for path, digest in staged_catalog_map.items()):
        _fail("stage file plan does not cover staged catalog files")
    new_records = plan.get("new_files")
    if not isinstance(new_records, list) or not new_records:
        _fail("stage new file plan is missing")
    new_map: dict[str, str] = {}
    for index, record in enumerate(new_records):
        if not isinstance(record, dict) or not isinstance(record.get("path"), str) or record.get("stage_path") != f"media/{record['path']}":
            _fail(f"stage new file plan record {index} is malformed")
        if MEDIA_RE.fullmatch(record["path"]) is None or record["path"] in new_map:
            _fail(f"stage new file plan record {index} has an invalid path")
        _check_sha(record.get("sha256"), f"stage new file plan {record['path']}.sha256")
        if record["path"] not in staged_catalog_map or staged_catalog_map[record["path"]] != record["sha256"]:
            _fail(f"stage new file is not an active catalog file: {record['path']}")
        if record["path"] in {item["path"] for item in plan.get("baseline", {}).get("files", [])}:
            _fail(f"stage new file was active in baseline: {record['path']}")
        stage_file(record["stage_path"], f"stage new file plan {record['path']}")
        new_map[record["path"]] = record["sha256"]
    if set(plan.get("target_paths", [])) != set(new_map):
        _fail("stage target path plan differs from new files")
    evidence = plan.get("evidence")
    if not isinstance(evidence, dict) or not isinstance(evidence.get("path"), str):
        _fail("stage evidence binding is missing")
    evidence_path = stage_file(evidence["path"], "stage evidence path")
    if not evidence_path.is_file() or sha256_file(evidence_path) != evidence.get("sha256"):
        _fail("stage evidence hash differs")
    receipt_path = stage / "receipt.json"
    if not receipt_path.is_file() or sha256_file(receipt_path) != plan["receipt"]["sha256"]:
        _fail("staged receipt hash differs")
    return stage, plan, staged_catalog, manifest_bytes, snapshot_hash


def _verify_plan_baseline(root: Path, plan: dict[str, Any]) -> tuple[dict[str, Any], dict[str, Any], dict[str, Any]]:
    baseline = plan.get("baseline", {})
    catalog_desc = baseline.get("catalog", {})
    snapshot_desc = baseline.get("snapshot", {})
    catalog, snapshot, file_map = _validate_live_baseline(root, catalog_desc.get("sha256"), snapshot_desc.get("sha256"))
    expected_files = {record["path"]: record["sha256"] for record in baseline.get("files", [])}
    if expected_files != {path: record["sha256"] for path, record in file_map.items()}:
        _fail("live catalog active file baseline differs")
    return catalog, snapshot, file_map


def _derive_expected_stage(root: Path, stage: Path, plan: dict[str, Any], receipt: dict[str, Any], validated: dict[str, Any], catalog: dict[str, Any], snapshot: dict[str, Any]) -> None:
    """Rebuild the staged target from the receipt before any live commit.

    The stage plan is evidence, not authority.  Re-deriving the motion records
    and image bytes here prevents a plan/catalog pair from being edited into an
    import that was never covered by the reviewed receipt.
    """
    variant = validated["variant"]
    variant_id = variant["id"]
    idle_only = receipt["operation"] == "replace-idle"
    target_records_plan = plan.get("new_files")
    paths_by_action_tier: dict[str, dict[str, list[str]]] = {action: {tier: [] for tier in ("64", "96")} for action in validated["motions"]}
    for record in target_records_plan:
        match = MEDIA_RE.fullmatch(record["path"])
        if match is None or match.group("variant") != variant_id:
            _fail("stage target path does not match receipt variant")
        action = record["path"].split("/")[1]
        tier = match.group("tier")
        if action not in paths_by_action_tier:
            _fail("stage target path has an invalid action")
        paths_by_action_tier[action][tier].append(record["path"])
    target_records: list[dict[str, Any]] = []
    target_bytes: dict[str, bytes] = {}
    motions: dict[str, dict[str, Any]] = {}
    for action in validated["motions"]:
        result = validated["motions"][action]
        for tier in ("64", "96"):
            if len(paths_by_action_tier[action][tier]) != len(result["frames"]):
                _fail(f"stage target {action}/{tier} allocation count differs from receipt")
        motion, records, image_bytes = _build_motion_with_paths(variant_id, action, result, plan["receipt"]["path"], plan["receipt"]["sha256"], paths_by_action_tier[action])
        motions[action] = motion
        target_records.extend(records)
        target_bytes.update(image_bytes)
    planned_new_map = {record["path"]: record["sha256"] for record in target_records_plan}
    expected_new_map = {record["path"]: record["sha256"] for record in target_records}
    if planned_new_map != expected_new_map:
        _fail("stage new file hashes do not match receipt-derived images")
    staged_catalog = _json_bytes((stage / "manifest.json").read_bytes(), "staged catalog")
    staged_files_map = {record["path"]: record for record in staged_catalog["files"]}
    for record in target_records:
        if staged_files_map.get(record["path"]) != record:
            _fail(f"staged target file provenance differs from receipt: {record['path']}")
        staged_path = stage / "media" / record["path"]
        if not staged_path.is_file() or staged_path.read_bytes() != target_bytes[record["path"]]:
            _fail(f"staged target image bytes differ from receipt: {record['path']}")
    existing = next((item for item in catalog["variants"] if item["id"] == variant_id), None)
    target_old_paths = _replaced_paths(existing, idle_only=idle_only)
    if plan.get("old_target_paths") != sorted(target_old_paths):
        _fail("stage old target paths differ from receipt-derived paths")
    active_without_target = [record for record in catalog["files"] if record["path"] not in target_old_paths]
    expected_catalog = copy.deepcopy(catalog)
    if receipt["operation"] == "add":
        if existing is not None:
            _fail("receipt add target is already present during apply")
        expected_catalog["variants"].append({
            **variant,
            "motions": motions,
            "idleFallback": False,
            "files": copy.deepcopy(target_records),
            "frameCount": sum(len(motions[action]["durations"]) for action in ("walk", "idle")),
        })
    else:
        if existing is None:
            _fail("receipt replace target disappeared before apply")
        replacement = next(item for item in expected_catalog["variants"] if item["id"] == variant_id)
        if idle_only:
            for key in ("species", "speciesLabel", "coat", "coatLabel"):
                if existing.get(key) != variant[key]:
                    _fail(f"replace-idle variant metadata differs: {key}")
            replacement["motions"]["idle"] = motions["idle"]
            replacement["files"] = [record for record in replacement["files"] if record["path"] not in target_old_paths] + copy.deepcopy(target_records)
        else:
            replacement["motions"] = motions
            replacement["files"] = copy.deepcopy(target_records)
        replacement["idleFallback"] = False
        replacement["frameCount"] = sum(len(replacement["motions"][action]["durations"]) for action in ("walk", "idle"))
    expected_catalog["files"] = active_without_target + target_records
    staged_catalog["snapshotAt"] = expected_catalog.get("snapshotAt")
    expected_catalog["snapshotAt"] = staged_catalog.get("snapshotAt")
    if staged_catalog != expected_catalog:
        _fail("staged catalog differs from receipt-derived catalog")
    staged_snapshot = _json_bytes((stage / "catalog-snapshot.json").read_bytes(), "staged snapshot")
    expected_snapshot = _new_snapshot(snapshot, expected_catalog, plan["staged"]["manifest"]["sha256"], plan["receipt"]["path"], plan["receipt"]["sha256"], operation=receipt["operation"], variant_id=variant_id)
    expected_snapshot["snapshotAt"] = staged_snapshot.get("snapshotAt")
    if staged_snapshot != expected_snapshot:
        _fail("staged snapshot differs from receipt-derived snapshot")


def _acquire_file_lock(root: Path):
    lock_path = root / LOCK_REL
    lock_path.parent.mkdir(parents=True, exist_ok=True)
    handle = lock_path.open("a+b")
    try:
        if os.name == "nt":
            import msvcrt
            handle.seek(0)
            if handle.read(1) == b"":
                handle.seek(0)
                handle.write(b"0")
                handle.flush()
            handle.seek(0)
            msvcrt.locking(handle.fileno(), msvcrt.LK_NBLCK, 1)
        else:
            import fcntl
            fcntl.flock(handle.fileno(), fcntl.LOCK_EX | fcntl.LOCK_NB)
        return handle
    except (OSError, BlockingIOError) as exc:
        handle.close()
        _fail(f"integration lock is held: {exc}")


def _release_file_lock(handle) -> None:
    try:
        if os.name == "nt":
            import msvcrt
            handle.seek(0)
            msvcrt.locking(handle.fileno(), msvcrt.LK_UNLCK, 1)
        else:
            import fcntl
            fcntl.flock(handle.fileno(), fcntl.LOCK_UN)
    finally:
        handle.close()


@contextmanager
def project_lock(root: Path):
    handle = _acquire_file_lock(root)
    try:
        yield
    finally:
        _release_file_lock(handle)


def _journal_path(stage: Path) -> Path:
    return stage / "journal.json"


def _write_journal(stage: Path, journal: dict[str, Any]) -> None:
    _atomic_write(_journal_path(stage), (json.dumps(journal, ensure_ascii=False, indent=2) + "\n").encode("utf-8"))


def _assert_media_destination(destination: Path, media_root: Path) -> None:
    """Reject a destination whose existing parent or target escapes media."""

    owner = media_root.resolve(strict=True)
    for candidate in (destination.parent, destination):
        resolved = candidate.resolve(strict=False)
        try:
            resolved.relative_to(owner)
        except ValueError:
            _fail(f"destination escapes media root: {destination}")


def _copy_exclusive_or_verify(source: Path, destination: Path, expected_hash: str, *, media_root: Path) -> None:
    _assert_media_destination(destination, media_root)
    destination.parent.mkdir(parents=True, exist_ok=True)
    # Recheck after mkdir so a junction/symlink already present in the path
    # cannot be followed by the exclusive open below.
    _assert_media_destination(destination, media_root)
    if destination.exists():
        if not destination.is_file() or sha256_file(destination) != expected_hash:
            _fail(f"destination conflict: {destination}")
        return
    try:
        with source.open("rb") as source_handle, destination.open("xb") as destination_handle:
            shutil.copyfileobj(source_handle, destination_handle)
            destination_handle.flush()
            os.fsync(destination_handle.fileno())
    except FileExistsError:
        if not destination.is_file() or sha256_file(destination) != expected_hash:
            _fail(f"destination conflict: {destination}")
    if sha256_file(destination) != expected_hash:
        _fail(f"copied file hash differs: {destination}")


def apply_stage(root: Path | str, stage_path: str | Path) -> dict[str, Any]:
    root_path = _root_path(root)
    stage, plan, staged_catalog, staged_manifest_bytes, staged_snapshot_hash = _load_stage(root_path, stage_path)
    receipt, receipt_hash, receipt_rel = _load_receipt(root_path, plan["receipt"]["path"])
    if receipt_hash != plan["receipt"]["sha256"] or receipt_rel != plan["receipt"]["path"]:
        _fail("receipt binding differs from stage plan")
    if receipt_payload_hash(receipt) != plan["receipt_payload_sha256"]:
        _fail("receipt payload binding differs from stage plan")
    validated_receipt = _validate_receipt(root_path, receipt)
    if validated_receipt["receipt"].get("operation") != plan.get("operation") or validated_receipt["variant"] != plan.get("variant"):
        _fail("receipt operation or variant differs from stage plan")
    with project_lock(root_path):
        live_catalog_bytes = (root_path / MANIFEST_REL).read_bytes()
        live_catalog_hash = sha256_bytes(live_catalog_bytes)
        live_snapshot_path = root_path / SNAPSHOT_REL
        live_snapshot_hash = sha256_file(live_snapshot_path)
        old_hash = plan["baseline"]["catalog"]["sha256"]
        new_hash = plan["staged"]["manifest"]["sha256"]
        old_snapshot_hash = plan["baseline"]["snapshot"]["sha256"]
        new_snapshot_hash = plan["staged"]["snapshot"]["sha256"]
        if live_catalog_hash == new_hash:
            live_catalog = _json_bytes(live_catalog_bytes, "live applied catalog")
            _validate_catalog(root_path, live_catalog, root_path / MEDIA_REL)
            baseline_catalog = _load_json_file(stage / "baseline/manifest.json", "staged baseline catalog")
            baseline_snapshot = _load_json_file(stage / "baseline/catalog-snapshot.json", "staged baseline snapshot")
            _derive_expected_stage(root_path, stage, plan, receipt, validated_receipt, baseline_catalog, baseline_snapshot)
            if live_snapshot_hash == new_snapshot_hash:
                _validate_snapshot(_json_bytes(live_snapshot_path.read_bytes(), "live applied snapshot"), live_catalog, new_hash)
                return {"changed": False, "status": "already_applied", "stage": _relative_string(stage, root_path)}
            if live_snapshot_hash != old_snapshot_hash:
                _fail("live snapshot hash is unknown after catalog switch")
            _validate_snapshot(_json_bytes(live_snapshot_path.read_bytes(), "live old snapshot"), baseline_catalog, old_hash)
            _atomic_write(live_snapshot_path, (stage / "catalog-snapshot.json").read_bytes())
            return {"changed": True, "status": "snapshot_finished", "stage": _relative_string(stage, root_path)}
        if live_catalog_hash != old_hash or live_snapshot_hash != old_snapshot_hash:
            _fail("live catalog/snapshot baseline differs")
        catalog, snapshot, file_map = _verify_plan_baseline(root_path, plan)
        _derive_expected_stage(root_path, stage, plan, receipt, validated_receipt, catalog, snapshot)
        backup = stage / "backup"
        backup.mkdir(exist_ok=True)
        old_manifest_path = backup / "manifest.json"
        old_snapshot_path = backup / "catalog-snapshot.json"
        if old_manifest_path.exists() and sha256_file(old_manifest_path) != old_hash:
            _fail("immutable old manifest backup conflicts")
        if old_snapshot_path.exists() and sha256_file(old_snapshot_path) != old_snapshot_hash:
            _fail("immutable old snapshot backup conflicts")
        if not old_manifest_path.exists():
            old_manifest_path.write_bytes(live_catalog_bytes)
        if not old_snapshot_path.exists():
            old_snapshot_path.write_bytes(live_snapshot_path.read_bytes())
        journal = {
            "schema": SCHEMA,
            "stage": _relative_string(stage, root_path),
            "operation": plan["operation"],
            "variant": plan["variant"],
            "old_catalog_sha256": old_hash,
            "old_snapshot_sha256": old_snapshot_hash,
            "new_catalog_sha256": new_hash,
            "new_snapshot_sha256": new_snapshot_hash,
            "status": "prepared",
        }
        if _journal_path(stage).exists():
            existing_journal = _load_json_file(_journal_path(stage), "journal")
            if existing_journal != journal and existing_journal.get("status") not in {"images_copied", "catalog_switched", "completed"}:
                _fail("journal differs from stage plan")
            if existing_journal.get("status") in {"images_copied", "catalog_switched", "completed"}:
                journal = existing_journal
        else:
            _write_journal(stage, journal)
        for record in plan["new_files"]:
            source = stage / "media" / record["path"]
            destination = root_path / MEDIA_REL / record["path"]
            if record["path"] in file_map:
                _fail(f"new destination is already an active catalog file: {record['path']}")
            _copy_exclusive_or_verify(source, destination, record["sha256"], media_root=root_path / MEDIA_REL)
        journal["status"] = "images_copied"
        _write_journal(stage, journal)
        if sha256_file(root_path / MANIFEST_REL) != old_hash or sha256_file(root_path / SNAPSHOT_REL) != old_snapshot_hash:
            _fail("live catalog changed before atomic switch")
        temporary_manifest = root_path / MEDIA_REL / f".manifest.integrate.{uuid.uuid4().hex}.tmp"
        temporary_manifest.write_bytes(staged_manifest_bytes)
        with temporary_manifest.open("ab") as handle:
            handle.flush()
            os.fsync(handle.fileno())
        # Recheck immediately before the catalog commit point.
        if sha256_file(root_path / MANIFEST_REL) != old_hash:
            temporary_manifest.unlink(missing_ok=True)
            _fail("live catalog changed at atomic switch boundary")
        os.replace(temporary_manifest, root_path / MANIFEST_REL)
        journal["status"] = "catalog_switched"
        _write_journal(stage, journal)
        _atomic_write(root_path / SNAPSHOT_REL, (stage / "catalog-snapshot.json").read_bytes())
        journal["status"] = "completed"
        _write_journal(stage, journal)
        return {"changed": True, "status": "applied", "stage": _relative_string(stage, root_path), "catalog_sha256": new_hash, "snapshot_sha256": new_snapshot_hash}


def recover_stage(root: Path | str, stage_path: str | Path) -> dict[str, Any]:
    root_path = _root_path(root)
    stage = _stage_path(root_path, stage_path, must_exist=True)
    _, plan, _, _, _ = _load_stage(root_path, stage_path)
    journal_path = _journal_path(stage)
    if not journal_path.is_file():
        _fail("exact recovery journal is missing")
    journal = _load_json_file(journal_path, "journal")
    if journal.get("schema") != SCHEMA:
        _fail("journal schema differs")
    plan = _load_json_file(stage / "plan.json", "stage plan")
    if journal.get("stage") != _relative_string(stage, root_path) or journal.get("operation") != plan.get("operation") or journal.get("variant") != plan.get("variant"):
        _fail("journal does not bind exact stage")
    for key in ("old_catalog_sha256", "old_snapshot_sha256", "new_catalog_sha256", "new_snapshot_sha256"):
        _check_sha(journal.get(key), f"journal.{key}")
    if (
        journal.get("old_catalog_sha256") != plan["baseline"]["catalog"]["sha256"]
        or journal.get("old_snapshot_sha256") != plan["baseline"]["snapshot"]["sha256"]
        or journal.get("new_catalog_sha256") != plan["staged"]["manifest"]["sha256"]
        or journal.get("new_snapshot_sha256") != plan["staged"]["snapshot"]["sha256"]
    ):
        _fail("journal does not bind stage plan")
    receipt, receipt_hash, receipt_rel = _load_receipt(root_path, plan["receipt"]["path"])
    if receipt_hash != plan["receipt"]["sha256"] or receipt_rel != plan["receipt"]["path"]:
        _fail("receipt binding differs from stage plan")
    if receipt_payload_hash(receipt) != plan["receipt_payload_sha256"]:
        _fail("receipt payload binding differs from stage plan")
    validated_receipt = _validate_receipt(root_path, receipt)
    if validated_receipt["receipt"].get("operation") != plan.get("operation") or validated_receipt["variant"] != plan.get("variant"):
        _fail("receipt operation or variant differs from stage plan")
    baseline_catalog = _load_json_file(stage / "baseline/manifest.json", "staged baseline catalog")
    baseline_snapshot = _load_json_file(stage / "baseline/catalog-snapshot.json", "staged baseline snapshot")
    _derive_expected_stage(root_path, stage, plan, receipt, validated_receipt, baseline_catalog, baseline_snapshot)
    old_manifest = stage / "backup/manifest.json"
    old_snapshot = stage / "backup/catalog-snapshot.json"
    staged_snapshot = stage / "catalog-snapshot.json"
    if not old_manifest.is_file() or sha256_file(old_manifest) != journal["old_catalog_sha256"]:
        _fail("immutable old manifest backup is missing or changed")
    if not old_snapshot.is_file() or sha256_file(old_snapshot) != journal["old_snapshot_sha256"]:
        _fail("immutable old snapshot backup is missing or changed")
    if not staged_snapshot.is_file() or sha256_file(staged_snapshot) != journal["new_snapshot_sha256"]:
        _fail("staged snapshot is missing or changed")
    with project_lock(root_path):
        live_catalog_path = root_path / MANIFEST_REL
        live_snapshot_path = root_path / SNAPSHOT_REL
        live_catalog_hash = sha256_file(live_catalog_path)
        live_snapshot_hash = sha256_file(live_snapshot_path)
        old_hash = journal["old_catalog_sha256"]
        new_hash = journal["new_catalog_sha256"]
        old_snapshot_hash = journal["old_snapshot_sha256"]
        new_snapshot_hash = journal["new_snapshot_sha256"]
        if live_catalog_hash == old_hash:
            old_catalog = _json_bytes(live_catalog_path.read_bytes(), "live old catalog")
            _validate_catalog(root_path, old_catalog, root_path / MEDIA_REL)
            if live_snapshot_hash not in {old_snapshot_hash, new_snapshot_hash}:
                _fail("live snapshot hash is unknown for old catalog recovery")
            if live_snapshot_hash == new_snapshot_hash:
                _atomic_write(live_snapshot_path, old_snapshot.read_bytes())
                journal["status"] = "recovered_old"
                _write_journal(stage, journal)
                return {"changed": True, "status": "recovered_old", "stage": _relative_string(stage, root_path)}
            return {"changed": False, "status": "already_old", "stage": _relative_string(stage, root_path)}
        if live_catalog_hash == new_hash:
            new_catalog = _json_bytes(live_catalog_path.read_bytes(), "live new catalog")
            _validate_catalog(root_path, new_catalog, root_path / MEDIA_REL)
            if live_snapshot_hash not in {old_snapshot_hash, new_snapshot_hash}:
                _fail("live snapshot hash is unknown for new catalog recovery")
            if live_snapshot_hash == old_snapshot_hash:
                _atomic_write(live_snapshot_path, staged_snapshot.read_bytes())
                journal["status"] = "recovered_new"
                _write_journal(stage, journal)
                return {"changed": True, "status": "recovered_new", "stage": _relative_string(stage, root_path)}
            return {"changed": False, "status": "already_new", "stage": _relative_string(stage, root_path)}
        _fail("live catalog hash is unknown for recovery")


# Importable names for the parent orchestration/tests.
stage = stage_receipt
apply = apply_stage
recover = recover_stage


def _cli_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--root", default=None, help=argparse.SUPPRESS)
    sub = parser.add_subparsers(dest="command", required=True)
    payload = sub.add_parser("payload-hash")
    payload.add_argument("--receipt", required=True)
    stage = sub.add_parser("stage")
    stage.add_argument("--receipt", required=True)
    stage.add_argument("--output", required=True)
    apply = sub.add_parser("apply")
    apply.add_argument("--stage", required=True)
    recover = sub.add_parser("recover")
    recover.add_argument("--stage", required=True)
    return parser


def main(argv: list[str] | None = None) -> int:
    args = _cli_parser().parse_args(argv)
    try:
        root = _root_path(args.root)
        if args.command == "payload-hash":
            receipt, _, _ = _load_receipt(root, args.receipt)
            print(receipt_payload_hash(receipt))
        elif args.command == "stage":
            result = stage_receipt(root, args.receipt, args.output)
            print(json.dumps(result["evidence"], ensure_ascii=False, sort_keys=True))
        elif args.command == "apply":
            print(json.dumps(apply_stage(root, args.stage), ensure_ascii=False, sort_keys=True))
        elif args.command == "recover":
            print(json.dumps(recover_stage(root, args.stage), ensure_ascii=False, sort_keys=True))
        return 0
    except (IntegrationError, OSError, KeyError, TypeError) as exc:
        print(f"integrate-reviewed-coat: {exc}", file=sys.stderr)
        return 2


if __name__ == "__main__":
    raise SystemExit(main())
