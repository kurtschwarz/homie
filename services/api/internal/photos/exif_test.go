package photos

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/jpeg"
	"math"
	"testing"
	"time"
)

func testJPEG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 16, 8))
	for x := 0; x < 16; x++ {
		img.Set(x, 3, color.RGBA{R: 200, A: 255})
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

type parsedExif struct {
	ifd0, exif, gps map[uint16][]byte
}

func readIFD(t *testing.T, tiff []byte, offset int) map[uint16][]byte {
	t.Helper()
	out := map[uint16][]byte{}
	count := int(binary.BigEndian.Uint16(tiff[offset:]))
	for i := 0; i < count; i++ {
		e := tiff[offset+2+12*i:]
		tag, typ, n := binary.BigEndian.Uint16(e), binary.BigEndian.Uint16(e[2:]), int(binary.BigEndian.Uint32(e[4:]))
		size := n
		switch typ {
		case typeLong:
			size = 4 * n
		case typeRational:
			size = 8 * n
		}
		if size <= 4 {
			out[tag] = e[8 : 8+size]
		} else {
			start := int(binary.BigEndian.Uint32(e[8:]))
			out[tag] = tiff[start : start+size]
		}
	}
	return out
}

func parseExif(t *testing.T, data []byte) parsedExif {
	t.Helper()
	if data[2] != 0xFF || data[3] != 0xE1 {
		t.Fatal("APP1 must follow SOI")
	}
	length := int(binary.BigEndian.Uint16(data[4:]))
	payload := data[6 : 4+length]
	if !bytes.HasPrefix(payload, []byte("Exif\x00\x00")) {
		t.Fatal("missing Exif header")
	}
	tiff := payload[6:]
	p := parsedExif{ifd0: readIFD(t, tiff, int(binary.BigEndian.Uint32(tiff[4:])))}
	p.exif = readIFD(t, tiff, int(binary.BigEndian.Uint32(p.ifd0[0x8769])))
	if gps, ok := p.ifd0[0x8825]; ok {
		p.gps = readIFD(t, tiff, int(binary.BigEndian.Uint32(gps)))
	}
	return p
}

func ascii(b []byte) string { return string(bytes.TrimRight(b, "\x00")) }

func degrees(b []byte) float64 {
	var total float64
	for i, divisor := range []float64{1, 60, 3600} {
		num, den := binary.BigEndian.Uint32(b[8*i:]), binary.BigEndian.Uint32(b[8*i+4:])
		total += float64(num) / float64(den) / divisor
	}
	return total
}

func TestEmbedExifRoundTrip(t *testing.T) {
	original := testJPEG(t)
	drawnAt := time.Date(2026, 10, 28, 15, 4, 5, 0, time.UTC)
	out, err := embedExif(original, exifMetadata{
		DrawnAt: drawnAt, Description: "Drawn on Homie Display", Coordinate: &Coordinate{Latitude: 45.4215, Longitude: -75.6972},
	})
	if err != nil {
		t.Fatal(err)
	}

	p := parseExif(t, out)
	if got := ascii(p.ifd0[0x010F]); got != "Homie" {
		t.Fatalf("make = %q", got)
	}
	if got := ascii(p.ifd0[0x0131]); got != "Homie" {
		t.Fatalf("software = %q", got)
	}
	if got := ascii(p.ifd0[0x0110]); got != "Homie Display" {
		t.Fatalf("model = %q", got)
	}
	if got := ascii(p.ifd0[0x010E]); got != "Drawn on Homie Display" {
		t.Fatalf("description = %q", got)
	}
	if got := ascii(p.exif[0x9003]); got != "2026:10:28 15:04:05" {
		t.Fatalf("date original = %q", got)
	}
	if ascii(p.gps[0x0001]) != "N" || ascii(p.gps[0x0003]) != "W" {
		t.Fatalf("hemisphere refs: %q %q", p.gps[0x0001], p.gps[0x0003])
	}
	if lat := degrees(p.gps[0x0002]); math.Abs(lat-45.4215) > 1e-6 {
		t.Fatalf("latitude = %v", lat)
	}
	if lon := degrees(p.gps[0x0004]); math.Abs(lon-75.6972) > 1e-6 {
		t.Fatalf("longitude = %v", lon)
	}

	if _, err := jpeg.Decode(bytes.NewReader(out)); err != nil {
		t.Fatalf("output must stay a valid JPEG: %v", err)
	}
	segmentEnd := 4 + int(binary.BigEndian.Uint16(out[4:]))
	if !bytes.Equal(out[segmentEnd:], original[2:]) {
		t.Fatal("image bytes after the EXIF segment must be untouched")
	}
}

func TestEmbedExifWithoutCoordinate(t *testing.T) {
	out, err := embedExif(testJPEG(t), exifMetadata{DrawnAt: time.Unix(1_761_000_000, 0), Description: "no place"})
	if err != nil {
		t.Fatal(err)
	}
	p := parseExif(t, out)
	if p.gps != nil {
		t.Fatalf("unexpected GPS block: %v", p.gps)
	}
	if got := ascii(p.ifd0[0x0110]); got != "Homie Display" {
		t.Fatalf("model = %q", got)
	}
	if _, err := jpeg.Decode(bytes.NewReader(out)); err != nil {
		t.Fatal(err)
	}
}

func TestEmbedExifReplacesExistingExif(t *testing.T) {
	meta := exifMetadata{DrawnAt: time.Unix(1_761_000_000, 0), Description: "again"}
	once, err := embedExif(testJPEG(t), meta)
	if err != nil {
		t.Fatal(err)
	}
	twice, err := embedExif(once, meta)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(once, twice) {
		t.Fatal("embedding twice must not stack EXIF segments")
	}
}

func TestEmbedExifRejectsNonJPEG(t *testing.T) {
	if _, err := embedExif([]byte("not a jpeg"), exifMetadata{}); err == nil {
		t.Fatal("expected error")
	}
}
