//go:build windows

package winapp

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
)

func TrayIconBytes(active bool) []byte {
	return imageToICO(TrayIconImage(active))
}

func TrayIconImage(active bool) image.Image {
	img, err := embeddedSpriteImage("mofumouse_cursor_companion_icon.png")
	if err != nil {
		img, err = embeddedSpriteImage("degu_idle_agouti.png")
	}
	if err != nil {
		return fallbackTrayIcon(active)
	}
	resized := resizeArea(img, 16)
	if !active {
		return grayscaleImage(resized)
	}
	return resized
}

func imageToICO(src image.Image) []byte {
	const (
		width       = 16
		height      = 16
		bitmapBytes = width * height * 4
		maskBytes   = height * 4
		headerBytes = 40
		imageOffset = 6 + 16
		imageSize   = headerBytes + bitmapBytes + maskBytes
	)

	var buf bytes.Buffer
	writeLE(&buf, uint16(0))
	writeLE(&buf, uint16(1))
	writeLE(&buf, uint16(1))
	buf.WriteByte(width)
	buf.WriteByte(height)
	buf.WriteByte(0)
	buf.WriteByte(0)
	writeLE(&buf, uint16(1))
	writeLE(&buf, uint16(32))
	writeLE(&buf, uint32(imageSize))
	writeLE(&buf, uint32(imageOffset))

	writeLE(&buf, uint32(headerBytes))
	writeLE(&buf, int32(width))
	writeLE(&buf, int32(height*2))
	writeLE(&buf, uint16(1))
	writeLE(&buf, uint16(32))
	writeLE(&buf, uint32(0))
	writeLE(&buf, uint32(bitmapBytes))
	writeLE(&buf, int32(0))
	writeLE(&buf, int32(0))
	writeLE(&buf, uint32(0))
	writeLE(&buf, uint32(0))

	for y := height - 1; y >= 0; y-- {
		for x := 0; x < width; x++ {
			r, g, b, a := src.At(x, y).RGBA()
			buf.WriteByte(byte(b >> 8))
			buf.WriteByte(byte(g >> 8))
			buf.WriteByte(byte(r >> 8))
			buf.WriteByte(byte(a >> 8))
		}
	}
	buf.Write(make([]byte, maskBytes))
	return buf.Bytes()
}

func writeLE(buf *bytes.Buffer, v any) {
	if err := binary.Write(buf, binary.LittleEndian, v); err != nil {
		panic(err)
	}
}

func fallbackTrayIcon(active bool) image.Image {
	const size = 16
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	body := color.RGBA{R: 156, G: 140, B: 108, A: 255}
	if !active {
		body = color.RGBA{R: 120, G: 120, B: 120, A: 255}
	}
	fillEllipse(img, 2, 5, 12, 14, body)
	fillEllipse(img, 8, 6, 15, 13, body)
	img.Set(11, 8, color.RGBA{R: 12, G: 12, B: 12, A: 255})
	return img
}

func fillEllipse(img *image.RGBA, left, top, right, bottom int, c color.Color) {
	cx := float64(left+right) / 2
	cy := float64(top+bottom) / 2
	rx := float64(right-left) / 2
	ry := float64(bottom-top) / 2
	for y := top; y <= bottom; y++ {
		for x := left; x <= right; x++ {
			dx := (float64(x) - cx) / rx
			dy := (float64(y) - cy) / ry
			if dx*dx+dy*dy <= 1 {
				img.Set(x, y, c)
			}
		}
	}
}
