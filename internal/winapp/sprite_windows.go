//go:build windows

package winapp

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"unsafe"

	"mofumouse/assets"
)

type bitmapInfoHeader struct {
	Size          uint32
	Width         int32
	Height        int32
	Planes        uint16
	BitCount      uint16
	Compression   uint32
	SizeImage     uint32
	XPelsPerMeter int32
	YPelsPerMeter int32
	ClrUsed       uint32
	ClrImportant  uint32
}

type bitmapInfo struct {
	Header bitmapInfoHeader
	Colors [1]uint32
}

type dibSprite struct {
	width         int32
	height        int32
	pixels        []byte
	flippedPixels []byte
	info          bitmapInfo
}

const (
	dibRGBColors = 0
	srccopy      = 0x00CC0020
	// Color-key layered windows cannot preserve partial alpha. Dropping faint edge
	// pixels avoids dark halos around ImageGen sprites on white backgrounds.
	spriteAlphaCutoff = 96
)

func loadSprite(name string) (*dibSprite, error) {
	data, err := assets.ReadSprite(name)
	if err != nil {
		return nil, err
	}
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return newDIBSprite(img), nil
}

func newDIBSprite(img image.Image) *dibSprite {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	pixels := make([]byte, 0, width*height*4)

	for y := bounds.Max.Y - 1; y >= bounds.Min.Y; y-- {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r16, g16, b16, a16 := img.At(x, y).RGBA()
			r, g, b, a := uint8(r16>>8), uint8(g16>>8), uint8(b16>>8), uint8(a16>>8)
			r, g, b = compositeOverMagenta(r, g, b, a)
			pixels = append(pixels, b, g, r, 0)
		}
	}

	return &dibSprite{
		width:         int32(width),
		height:        int32(height),
		pixels:        pixels,
		flippedPixels: flipDIBPixels(pixels, width, height),
		info: bitmapInfo{
			Header: bitmapInfoHeader{
				Size:        uint32(unsafe.Sizeof(bitmapInfoHeader{})),
				Width:       int32(width),
				Height:      int32(height),
				Planes:      1,
				BitCount:    32,
				Compression: 0,
				SizeImage:   uint32(len(pixels)),
			},
		},
	}
}

func flipDIBPixels(pixels []byte, width, height int) []byte {
	flipped := make([]byte, len(pixels))
	stride := width * 4
	for row := 0; row < height; row++ {
		rowStart := row * stride
		for x := 0; x < width; x++ {
			src := rowStart + x*4
			dst := rowStart + (width-1-x)*4
			copy(flipped[dst:dst+4], pixels[src:src+4])
		}
	}
	return flipped
}

func compositeOverMagenta(r, g, b, a uint8) (uint8, uint8, uint8) {
	if a < spriteAlphaCutoff {
		return 255, 0, 255
	}
	return r, g, b
}

func drawSprite(hdc uintptr, sprite *dibSprite, x, y, width, height int32, flipped bool) {
	if sprite == nil || len(sprite.pixels) == 0 {
		return
	}
	procSetStretchBltMode.Call(hdc, colorOnColorStretch)
	pixels := sprite.pixels
	if flipped && len(sprite.flippedPixels) == len(sprite.pixels) {
		pixels = sprite.flippedPixels
	}
	procStretchDIBits.Call(
		hdc,
		uintptr(x), uintptr(y), uintptr(width), uintptr(height),
		0, 0, uintptr(sprite.width), uintptr(sprite.height),
		uintptr(unsafe.Pointer(&pixels[0])),
		uintptr(unsafe.Pointer(&sprite.info)),
		dibRGBColors,
		srccopy,
	)
}

func embeddedSpriteImage(name string) (image.Image, error) {
	data, err := assets.ReadSprite(name)
	if err != nil {
		return nil, err
	}
	return png.Decode(bytes.NewReader(data))
}

func resizeNearest(src image.Image, size int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, size, size))
	b := src.Bounds()
	for y := 0; y < size; y++ {
		sy := b.Min.Y + y*b.Dy()/size
		for x := 0; x < size; x++ {
			sx := b.Min.X + x*b.Dx()/size
			dst.Set(x, y, src.At(sx, sy))
		}
	}
	return dst
}

func resizeArea(src image.Image, size int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, size, size))
	bounds := src.Bounds()
	for y := 0; y < size; y++ {
		y0 := bounds.Min.Y + y*bounds.Dy()/size
		y1 := bounds.Min.Y + (y+1)*bounds.Dy()/size
		if y1 <= y0 {
			y1 = y0 + 1
		}
		if y1 > bounds.Max.Y {
			y1 = bounds.Max.Y
		}
		for x := 0; x < size; x++ {
			x0 := bounds.Min.X + x*bounds.Dx()/size
			x1 := bounds.Min.X + (x+1)*bounds.Dx()/size
			if x1 <= x0 {
				x1 = x0 + 1
			}
			if x1 > bounds.Max.X {
				x1 = bounds.Max.X
			}

			var red, green, blue, alpha, count uint32
			for sy := y0; sy < y1; sy++ {
				for sx := x0; sx < x1; sx++ {
					c := color.NRGBAModel.Convert(src.At(sx, sy)).(color.NRGBA)
					a := uint32(c.A)
					red += uint32(c.R) * a
					green += uint32(c.G) * a
					blue += uint32(c.B) * a
					alpha += a
					count++
				}
			}
			if count == 0 || alpha == 0 {
				dst.Set(x, y, color.RGBA{})
				continue
			}
			dst.Set(x, y, color.RGBA{
				R: uint8(red / alpha),
				G: uint8(green / alpha),
				B: uint8(blue / alpha),
				A: uint8(alpha / count),
			})
		}
	}
	return dst
}

func grayscaleImage(src image.Image) *image.RGBA {
	b := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, b, a := src.At(x, y).RGBA()
			lum := uint8((299*(r>>8) + 587*(g>>8) + 114*(b>>8)) / 1000)
			dst.Set(x, y, color.RGBA{R: lum, G: lum, B: lum, A: uint8(a >> 8)})
		}
	}
	return dst
}
