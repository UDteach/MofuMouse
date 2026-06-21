from __future__ import annotations

import argparse
import json
from pathlib import Path

from PIL import Image

from process_imagegen_asset import flood_background


DEFAULT_SIZES = (32, 48, 64, 96)


def normalize_asset_path(path: Path) -> str:
    return path.as_posix()


def prepare_master(src: Path) -> Image.Image:
    img = Image.open(src).convert("RGBA")
    bg = flood_background(img)
    pixels = img.load()
    for x, y in bg:
        r, g, b, _ = pixels[x, y]
        pixels[x, y] = (r, g, b, 0)

    alpha = img.getchannel("A")
    bbox = alpha.getbbox()
    if bbox is None:
        raise ValueError(f"no foreground detected: {src}")
    return img.crop(bbox)


def write_tier(master: Image.Image, dst: Path, size: int) -> None:
    resized = master.copy()
    resized.thumbnail((size, size), Image.Resampling.LANCZOS)
    out = Image.new("RGBA", (size, size), (0, 0, 0, 0))
    x = (size - resized.width) // 2
    y = (size - resized.height) // 2
    out.alpha_composite(resized, (x, y))
    dst.parent.mkdir(parents=True, exist_ok=True)
    out.save(dst)


def generate_tiers(assets_dir: Path, sizes: list[int], ids: set[str] | None = None) -> None:
    manifest_path = assets_dir / "manifest" / "assets.json"
    manifest = json.loads(manifest_path.read_text(encoding="utf-8"))
    manifest["runtime_sizes"] = sizes
    manifest["asset_policy"] = (
        "ImageGen-only production assets; source images are treated as masters; "
        "runtime tiers are exported at 32/48/64/96px where available."
    )

    seen_ids: set[str] = set()
    for sprite in manifest.get("sprites", []):
        sprite_id = sprite["id"]
        if ids is not None and sprite_id not in ids:
            continue
        seen_ids.add(sprite_id)
        source = assets_dir / sprite["source"]
        if not source.exists():
            raise FileNotFoundError(f"{sprite_id} source not found: {source}")

        filename = Path(sprite["path"]).name
        master = prepare_master(source)
        root_output = assets_dir / "sprites" / filename
        write_tier(master, root_output, 32)

        tiers: dict[str, str] = {}
        for size in sizes:
            tier_output = assets_dir / "sprites" / str(size) / filename
            write_tier(master, tier_output, size)
            tiers[str(size)] = normalize_asset_path(Path("sprites") / str(size) / filename)

        sprite["path"] = normalize_asset_path(Path("sprites") / filename)
        sprite["tiers"] = tiers

    if ids is not None:
        missing_ids = sorted(ids - seen_ids)
        if missing_ids:
            raise ValueError(f"sprite id not found in manifest: {', '.join(missing_ids)}")

    manifest_path.write_text(
        json.dumps(manifest, ensure_ascii=False, indent=2) + "\n",
        encoding="utf-8",
    )


def main() -> None:
    parser = argparse.ArgumentParser(description="Generate MofuMouse runtime sprite tiers from ImageGen masters.")
    parser.add_argument("--assets-dir", type=Path, default=Path("assets"))
    parser.add_argument("--sizes", type=int, nargs="+", default=list(DEFAULT_SIZES))
    parser.add_argument("--ids", nargs="+", help="Optional manifest sprite IDs to regenerate.")
    args = parser.parse_args()

    sizes = sorted(set(args.sizes))
    if any(size <= 0 for size in sizes):
        raise ValueError("sizes must be positive")
    generate_tiers(args.assets_dir, sizes, set(args.ids) if args.ids else None)


if __name__ == "__main__":
    main()
