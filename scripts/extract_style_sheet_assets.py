from __future__ import annotations

import argparse
from collections import deque
from dataclasses import dataclass
from pathlib import Path

from PIL import Image


@dataclass(frozen=True)
class AssetSpec:
    asset_id: str
    pose: str
    coat: str
    box: tuple[int, int, int, int]


ASSETS = [
    AssetSpec("degu_idle_agouti", "idle", "agouti", (48, 36, 225, 272)),
    AssetSpec("degu_walk_agouti", "walk", "agouti", (284, 72, 584, 258)),
    AssetSpec("degu_run_agouti", "run", "agouti", (616, 78, 922, 264)),
    AssetSpec("degu_point_agouti", "point", "agouti", (990, 34, 1210, 260)),
    AssetSpec("degu_front_agouti", "front", "agouti", (1260, 38, 1420, 278)),
    AssetSpec("degu_sleepy_agouti", "sleepy", "agouti", (1450, 70, 1690, 262)),
    AssetSpec("degu_sniff_agouti", "sniff", "agouti", (30, 350, 320, 585)),
    AssetSpec("degu_groom_agouti", "groom", "agouti", (390, 322, 570, 590)),
    AssetSpec("degu_dig_agouti", "dig", "agouti", (600, 352, 880, 590)),
    AssetSpec("degu_roll_agouti", "roll", "agouti", (884, 340, 1195, 580)),
    AssetSpec("degu_nibble_agouti", "nibble", "agouti", (1230, 335, 1415, 600)),
    AssetSpec("degu_alert_agouti", "alert", "agouti", (1468, 330, 1660, 605)),
    AssetSpec("degu_idle_gray", "idle", "gray", (330, 600, 525, 862)),
    AssetSpec("degu_idle_dark", "idle", "dark", (610, 600, 828, 862)),
    AssetSpec("degu_idle_cream", "idle", "cream", (910, 600, 1100, 862)),
    AssetSpec("degu_idle_white", "idle", "white", (1192, 600, 1400, 862)),
    AssetSpec("degu_idle_pied", "idle", "pied", (1462, 600, 1685, 862)),
]


def is_background(pixel: tuple[int, int, int, int]) -> bool:
    r, g, b, a = pixel
    if a == 0:
        return True
    return min(r, g, b) >= 225 and max(r, g, b) - min(r, g, b) <= 10


def remove_edge_background(img: Image.Image) -> Image.Image:
    img = img.convert("RGBA")
    width, height = img.size
    pixels = img.load()
    seen: set[tuple[int, int]] = set()
    queue: deque[tuple[int, int]] = deque()

    for x in range(width):
        for y in (0, height - 1):
            if is_background(pixels[x, y]):
                seen.add((x, y))
                queue.append((x, y))
    for y in range(height):
        for x in (0, width - 1):
            if (x, y) not in seen and is_background(pixels[x, y]):
                seen.add((x, y))
                queue.append((x, y))

    while queue:
        x, y = queue.popleft()
        for nx, ny in ((x - 1, y), (x + 1, y), (x, y - 1), (x, y + 1)):
            if nx < 0 or ny < 0 or nx >= width or ny >= height or (nx, ny) in seen:
                continue
            if is_background(pixels[nx, ny]):
                seen.add((nx, ny))
                queue.append((nx, ny))

    for x, y in seen:
        r, g, b, _ = pixels[x, y]
        pixels[x, y] = (r, g, b, 0)
    return img


def centered_sprite(src: Image.Image, size: int) -> Image.Image:
    alpha = src.getchannel("A")
    bbox = alpha.getbbox()
    if bbox is None:
        raise ValueError("no foreground detected")

    cropped = src.crop(bbox)
    cropped.thumbnail((size, size), Image.Resampling.LANCZOS)
    out = Image.new("RGBA", (size, size), (0, 0, 0, 0))
    x = (size - cropped.width) // 2
    y = (size - cropped.height) // 2
    out.alpha_composite(cropped, (x, y))
    return out


def make_contact_sheet(sprite_paths: list[Path], out_path: Path, thumb_size: int) -> None:
    columns = 6
    label_h = 18
    rows = (len(sprite_paths) + columns - 1) // columns
    sheet = Image.new("RGBA", (columns * thumb_size, rows * (thumb_size + label_h)), (245, 245, 245, 255))
    for i, path in enumerate(sprite_paths):
        img = Image.open(path).convert("RGBA")
        row, col = divmod(i, columns)
        x = col * thumb_size
        y = row * (thumb_size + label_h)
        sheet.alpha_composite(img.resize((thumb_size, thumb_size), Image.Resampling.LANCZOS), (x, y))
    out_path.parent.mkdir(parents=True, exist_ok=True)
    sheet.save(out_path)


def main() -> None:
    parser = argparse.ArgumentParser(description="Extract MofuMouse sprites from the style reference sheet.")
    parser.add_argument("sheet", type=Path)
    parser.add_argument("--source-dir", type=Path, default=Path("assets/source"))
    parser.add_argument("--sprite-dir", type=Path, default=Path("assets/sprites"))
    parser.add_argument("--size", type=int, default=128)
    parser.add_argument("--contact", type=Path)
    args = parser.parse_args()

    sheet = Image.open(args.sheet).convert("RGBA")
    args.source_dir.mkdir(parents=True, exist_ok=True)
    args.sprite_dir.mkdir(parents=True, exist_ok=True)

    sprites: list[Path] = []
    for spec in ASSETS:
        crop = remove_edge_background(sheet.crop(spec.box))
        source_path = args.source_dir / f"style-reference-v1-{spec.pose}-{spec.coat}.png"
        sprite_path = args.sprite_dir / f"{spec.asset_id}.png"
        crop.save(source_path)
        centered_sprite(crop, args.size).save(sprite_path)
        sprites.append(sprite_path)

    if args.contact:
        make_contact_sheet(sprites, args.contact, args.size)


if __name__ == "__main__":
    main()
