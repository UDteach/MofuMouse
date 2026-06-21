from __future__ import annotations

import argparse
from pathlib import Path

from PIL import Image


DEFAULT_SIZES = (16, 24, 32, 48, 64, 128, 256)


def make_square_icon(src: Path) -> Image.Image:
    img = Image.open(src).convert("RGBA")
    bbox = img.getchannel("A").getbbox()
    if bbox is None:
        raise ValueError(f"no visible pixels in {src}")

    cropped = img.crop(bbox)
    side = max(cropped.width, cropped.height)
    canvas = Image.new("RGBA", (side, side), (0, 0, 0, 0))
    canvas.alpha_composite(cropped, ((side - cropped.width) // 2, (side - cropped.height) // 2))
    return canvas


def write_ico(master: Image.Image, path: Path, sizes: tuple[int, ...]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    master.save(path, format="ICO", sizes=[(size, size) for size in sizes])


def main() -> None:
    parser = argparse.ArgumentParser(description="Generate Windows ICO assets from the MofuMouse ImageGen icon.")
    parser.add_argument("--source", type=Path, default=Path("assets/source/imagegen-cursor-companion-icon-v1.png"))
    parser.add_argument("--icon", type=Path, default=Path("assets/icons/mofumouse.ico"))
    parser.add_argument("--wails-icon", type=Path, default=Path("build/windows/icon.ico"))
    parser.add_argument("--sizes", type=int, nargs="+", default=list(DEFAULT_SIZES))
    args = parser.parse_args()

    sizes = tuple(sorted(set(args.sizes)))
    if not sizes or any(size <= 0 or size > 256 for size in sizes):
        raise ValueError("icon sizes must be between 1 and 256")

    master = make_square_icon(args.source)
    write_ico(master, args.icon, sizes)
    write_ico(master, args.wails_icon, sizes)
    print(args.icon)
    print(args.wails_icon)


if __name__ == "__main__":
    main()
