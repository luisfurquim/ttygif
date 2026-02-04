package xwd

import (
	"io"
	"image"
	"errors"
	"image/color"
	"encoding/binary"
)

// XWDFileHeader type
type XWDFileHeader struct {
	HeaderSize        uint32
	FileVersion       uint32
	PixmapFormat      uint32
	PixmapDepth       uint32
	PixmapWidth       uint32
	PixmapHeight      uint32
	XOffset           uint32
	ByteOrder         uint32
	BitmapUnit        uint32
	BitmapBitOrder    uint32
	BitmapPad         uint32
	BitsPerPixel      uint32
	BytesPerLine      uint32
	VisualClass       uint32
	RedMask           uint32
	GreenMask         uint32
	BlueMask          uint32
	BitsPerRgb        uint32
	NumberOfColors    uint32
	ColorMapEntries   uint32
	WindowWidth       uint32
	WindowHeight      uint32
	WindowX           uint32
	WindowY           uint32
	WindowBorderWidth uint32
}

// XWDColorMap type
type XWDColorMap struct {
	EntryNumber uint32
	Red         uint16
	Green       uint16
	Blue        uint16
	Flags       uint8
	Padding     uint8
}

type Xwd struct {
	XWDFileHeader
	PixRows [][]byte
	Pixmap []byte
}

type Color struct{
	r, g, b uint32
}

var IncompleteBuffer error = errors.New("Incomplete buffer")

func DecodeHeader(buf []byte, hdr *XWDFileHeader) {
	hdr.HeaderSize        = binary.BigEndian.Uint32(buf[0:4])
	hdr.FileVersion       = binary.BigEndian.Uint32(buf[4:8])
	hdr.PixmapFormat      = binary.BigEndian.Uint32(buf[8:12])
	hdr.PixmapDepth       = binary.BigEndian.Uint32(buf[12:16])
	hdr.PixmapWidth       = binary.BigEndian.Uint32(buf[16:20])
	hdr.PixmapHeight      = binary.BigEndian.Uint32(buf[20:24])
	hdr.XOffset           = binary.BigEndian.Uint32(buf[24:28])
	hdr.ByteOrder         = binary.BigEndian.Uint32(buf[28:32])
	hdr.BitmapUnit        = binary.BigEndian.Uint32(buf[32:36])
	hdr.BitmapBitOrder    = binary.BigEndian.Uint32(buf[36:40])
	hdr.BitmapPad         = binary.BigEndian.Uint32(buf[40:44])
	hdr.BitsPerPixel      = binary.BigEndian.Uint32(buf[44:48])
	hdr.BytesPerLine      = binary.BigEndian.Uint32(buf[48:52])
	hdr.VisualClass       = binary.BigEndian.Uint32(buf[52:56])
	hdr.RedMask           = binary.BigEndian.Uint32(buf[56:60])
	hdr.GreenMask         = binary.BigEndian.Uint32(buf[60:64])
	hdr.BlueMask          = binary.BigEndian.Uint32(buf[64:68])
	hdr.BitsPerRgb        = binary.BigEndian.Uint32(buf[68:72])
	hdr.NumberOfColors    = binary.BigEndian.Uint32(buf[72:76])
	hdr.ColorMapEntries   = binary.BigEndian.Uint32(buf[76:80])
	hdr.WindowWidth       = binary.BigEndian.Uint32(buf[80:84])
	hdr.WindowHeight      = binary.BigEndian.Uint32(buf[84:88])
	hdr.WindowX           = binary.BigEndian.Uint32(buf[88:92])
	hdr.WindowY           = binary.BigEndian.Uint32(buf[92:96])
	hdr.WindowBorderWidth = binary.BigEndian.Uint32(buf[96:100])
}

func ParsePixmap(buffer [][]byte, Pixmap []byte, width int, height int) {
	var (
		y, start int
		linesize int
	)

	linesize = width << 2

	for y = 0; y < height; y++ {
		buffer[y] = Pixmap[start:start + linesize]
		start += linesize
	}
}


// Decode reads a XWD image from r and returns it as an image.Image.
func Decode(r io.Reader) (img Xwd, err error) {
	var buf []byte

	buf = make([]byte, 100)
	_, err = r.Read(buf)
	if err != nil {
		return
	}

	DecodeHeader(buf, &img.XWDFileHeader)

	// not used
	// window name
	windowName := make([]byte, img.HeaderSize-100)
	_, err = r.Read(windowName)
	if err != nil {
		return
	}

	// not used?
/*
	colorMaps := make([]XWDColorMap, img.ColorMapEntries)
	buf = make([]byte, 12)
	for i := 0; i < int(img.ColorMapEntries); i++ {
		_, err = r.Read(buf)
		if err != nil {
			return
		}
		colorMaps[i] = XWDColorMap{
			EntryNumber: binary.BigEndian.Uint32(buf[0:4]),
			Red:         binary.BigEndian.Uint16(buf[4:6]),
			Green:       binary.BigEndian.Uint16(buf[6:8]),
			Blue:        binary.BigEndian.Uint16(buf[8:10]),
			Flags:       uint8(buf[10]),
			Padding:     uint8(buf[11]),
		}
	}
*/

	// ColorMaps not used, just skip them
	buf = make([]byte, 12 * img.ColorMapEntries)
	_, err = r.Read(buf)
	if err != nil {
		return
	}

	img.PixRows = make([][]byte, img.PixmapHeight)
	img.Pixmap = make([]byte, img.PixmapHeight * img.PixmapWidth * 4)

	_, err = r.Read(img.Pixmap)
	if err != nil {
		return
	}

	ParsePixmap(img.PixRows, img.Pixmap, int(img.PixmapWidth), int(img.PixmapHeight))

	return
}


// DecodeNoCopy parses a XWD image from buf and returns it as an image.Image.
// It is faster than Decode(r io.Reader), but holds the memory associated with the provided buffer.
// Do not change the buffer contents after calling this function.
// If in doubt, use Decode(r io.Reader) instead.
func DecodeNoCopy(buf []byte) (img Xwd, err error) {
	var start, linesize int
	
	if len(buf) <= 100 {
		err = IncompleteBuffer
		return
	}

	DecodeHeader(buf, &img.XWDFileHeader)

	start    = int(img.XWDFileHeader.HeaderSize + 12 * img.XWDFileHeader.ColorMapEntries)
	linesize = int(img.PixmapWidth << 2)

	if len(buf) <= start + linesize * int(img.PixmapHeight) {
		err = IncompleteBuffer
		return
	}

	img.PixRows = make([][]byte, img.PixmapHeight)
	img.Pixmap = buf[start:]

	ParsePixmap(img.PixRows, img.Pixmap, int(img.PixmapWidth), int(img.PixmapHeight))

	return
}


// DecodePixNoCopy parses the pixels of a XWD image buffer into the xwd.Xwd image.
// It is faster than Decode(r io.Reader), but holds the memory associated with the provided buffer.
// Do not change the buffer contents after calling this function.
// If in doubt, use Decode(r io.Reader) instead.
func DecodePixNoCopy(buf []byte, img *Xwd) (err error) {
	var linesize int
	
	linesize = int(img.PixmapWidth << 2)

	if len(buf) < linesize * int(img.PixmapHeight) {
		err = IncompleteBuffer
		return
	}

	img.PixRows = make([][]byte, img.PixmapHeight)
	img.Pixmap = buf

	ParsePixmap(img.PixRows, buf, int(img.PixmapWidth), int(img.PixmapHeight))

	return
}



func (img Xwd) Bounds() image.Rectangle {
	return image.Rectangle{
		Min: image.Point{},
		Max: image.Point{
			X: int(img.PixmapWidth),
			Y: int(img.PixmapHeight),
		},
	}
}

func (img Xwd) At(x, y int) color.Color {
	var offset int

	offset = x << 2

	return Color{
		r: uint32(img.PixRows[y][offset + 2]),
		g: uint32(img.PixRows[y][offset + 1]),
		b: uint32(img.PixRows[y][offset]),
	}
}

var xwdModel color.Model = color.ModelFunc(func(c color.Color) color.Color {
    if _, ok := c.(Color); ok {
        return c
    }
    r, g, b, _ := c.RGBA()
    return Color{
        r: uint32(r >> 8),
        g: uint32(g >> 8),
        b: uint32(b >> 8),
    }
})

func (img Xwd) ColorModel() color.Model {
	return xwdModel
}

func (c Color) RGBA() (r, g, b, a uint32) {
	var r32, g32, b32 uint32

   // Converte 0-255 para 0-65535 e retorna alpha opaco
   r32 = uint32(c.r)
   g32 = uint32(c.g)
   b32 = uint32(c.b)
    
   r = (r32 << 8) | r32
   g = (g32 << 8) | g32
   b = (b32 << 8) | b32
   a = 0xffff
   return
}


func  MkColor(r, g, b byte) color.Color {
	return Color{
		r: uint32(r),
		g: uint32(g),
		b: uint32(b),
	}
}

