from __future__ import annotations

import argparse
from collections import deque
from pathlib import Path

from PIL import Image


def is_background(pixel: tuple[int, int, int, int]) -> bool:
    r, g, b, _ = pixel
    return min(r, g, b) >= 215 and max(r, g, b) - min(r, g, b) <= 42


def flood_background(img: Image.Image) -> set[tuple[int, int]]:
    width, height = img.size
    pixels = img.load()
    seen: set[tuple[int, int]] = set()
    queue: deque[tuple[int, int]] = deque()

    for x in range(width):
        for y in (0, height - 1):
            if is_background(pixels[x, y]):
                queue.append((x, y))
                seen.add((x, y))
    for y in range(height):
        for x in (0, width - 1):
            if (x, y) not in seen and is_background(pixels[x, y]):
                queue.append((x, y))
                seen.add((x, y))

    while queue:
        x, y = queue.popleft()
        for nx, ny in ((x - 1, y), (x + 1, y), (x, y - 1), (x, y + 1)):
            if nx < 0 or ny < 0 or nx >= width or ny >= height or (nx, ny) in seen:
                continue
            if is_background(pixels[nx, ny]):
                seen.add((nx, ny))
                queue.append((nx, ny))
    return seen


def process(src: Path, dst: Path, size: int) -> None:
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

    cropped = img.crop(bbox)
    cropped.thumbnail((size, size), Image.Resampling.LANCZOS)
    out = Image.new("RGBA", (size, size), (0, 0, 0, 0))
    x = (size - cropped.width) // 2
    y = (size - cropped.height) // 2
    out.alpha_composite(cropped, (x, y))
    dst.parent.mkdir(parents=True, exist_ok=True)
    out.save(dst)


def main() -> None:
    parser = argparse.ArgumentParser(description="Prepare one ImageGen MofuMouse asset.")
    parser.add_argument("src", type=Path)
    parser.add_argument("dst", type=Path)
    parser.add_argument("--size", type=int, default=128)
    args = parser.parse_args()
    process(args.src, args.dst, args.size)


if __name__ == "__main__":
    main()
