package photos

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math"
	"sort"
	"strings"
	"time"
)

const (
	exifSoftware = "Homie"
	exifModel    = "Homie Display"
)

var errNotJPEG = errors.New("not a JPEG")

type exifMetadata struct {
	DrawnAt     time.Time
	Description string
	Coordinate  *Coordinate
}

const (
	typeByte     = 1
	typeASCII    = 2
	typeLong     = 4
	typeRational = 5
)

type exifEntry struct {
	tag   uint16
	typ   uint16
	count uint32
	data  []byte
}

func asciiEntry(tag uint16, value string) exifEntry {
	data := append([]byte(value), 0)
	return exifEntry{tag: tag, typ: typeASCII, count: uint32(len(data)), data: data}
}

func longEntry(tag uint16, value uint32) exifEntry {
	return exifEntry{tag: tag, typ: typeLong, count: 1, data: binary.BigEndian.AppendUint32(nil, value)}
}

func rationalsEntry(tag uint16, values ...[2]uint32) exifEntry {
	var data []byte
	for _, v := range values {
		data = binary.BigEndian.AppendUint32(data, v[0])
		data = binary.BigEndian.AppendUint32(data, v[1])
	}
	return exifEntry{tag: tag, typ: typeRational, count: uint32(len(values)), data: data}
}

func ifdSize(entries []exifEntry) int {
	size := 6 + 12*len(entries)
	for _, e := range entries {
		if len(e.data) > 4 {
			size += len(e.data) + len(e.data)%2
		}
	}
	return size
}

// encodeIFD lays out an IFD that starts base bytes into the TIFF block.
func encodeIFD(entries []exifEntry, base int) []byte {
	sort.Slice(entries, func(i, j int) bool { return entries[i].tag < entries[j].tag })
	out := binary.BigEndian.AppendUint16(nil, uint16(len(entries)))
	var extra []byte
	extraStart := base + 6 + 12*len(entries)
	for _, e := range entries {
		out = binary.BigEndian.AppendUint16(out, e.tag)
		out = binary.BigEndian.AppendUint16(out, e.typ)
		out = binary.BigEndian.AppendUint32(out, e.count)
		if len(e.data) <= 4 {
			out = append(out, e.data...)
			out = append(out, make([]byte, 4-len(e.data))...)
			continue
		}
		out = binary.BigEndian.AppendUint32(out, uint32(extraStart+len(extra)))
		extra = append(extra, e.data...)
		if len(e.data)%2 == 1 {
			extra = append(extra, 0)
		}
	}
	out = binary.BigEndian.AppendUint32(out, 0)
	return append(out, extra...)
}

func gpsRational(coordinate float64) [3][2]uint32 {
	value := math.Abs(coordinate)
	degrees := math.Floor(value)
	minutes := math.Floor((value - degrees) * 60)
	seconds := (value - degrees - minutes/60) * 3600
	return [3][2]uint32{{uint32(degrees), 1}, {uint32(minutes), 1}, {uint32(math.Round(seconds * 1_000_000)), 1_000_000}}
}

func gpsEntries(c Coordinate) []exifEntry {
	latRef, lonRef := "N", "E"
	if c.Latitude < 0 {
		latRef = "S"
	}
	if c.Longitude < 0 {
		lonRef = "W"
	}
	lat, lon := gpsRational(c.Latitude), gpsRational(c.Longitude)
	return []exifEntry{
		{tag: 0x0000, typ: typeByte, count: 4, data: []byte{2, 3, 0, 0}},
		asciiEntry(0x0001, latRef),
		rationalsEntry(0x0002, lat[0], lat[1], lat[2]),
		asciiEntry(0x0003, lonRef),
		rationalsEntry(0x0004, lon[0], lon[1], lon[2]),
	}
}

func buildExif(meta exifMetadata) []byte {
	drawn := meta.DrawnAt.UTC().Format("2006:01:02 15:04:05")
	ifd0 := []exifEntry{
		asciiEntry(0x010E, meta.Description),
		asciiEntry(0x010F, exifSoftware),
		asciiEntry(0x0110, exifModel),
		asciiEntry(0x0131, exifSoftware),
		asciiEntry(0x0132, drawn),
		longEntry(0x8769, 0),
	}
	exifIFD := []exifEntry{
		asciiEntry(0x9003, drawn),
		asciiEntry(0x9004, drawn),
		asciiEntry(0x9011, "+00:00"),
		asciiEntry(0x9012, "+00:00"),
	}
	var gps []exifEntry
	if meta.Coordinate != nil {
		gps = gpsEntries(*meta.Coordinate)
		ifd0 = append(ifd0, longEntry(0x8825, 0))
	}

	exifOffset := 8 + ifdSize(ifd0)
	gpsOffset := exifOffset + ifdSize(exifIFD)
	for i := range ifd0 {
		switch ifd0[i].tag {
		case 0x8769:
			ifd0[i] = longEntry(0x8769, uint32(exifOffset))
		case 0x8825:
			ifd0[i] = longEntry(0x8825, uint32(gpsOffset))
		}
	}

	tiff := []byte{'M', 'M', 0, 0x2A, 0, 0, 0, 8}
	tiff = append(tiff, encodeIFD(ifd0, 8)...)
	tiff = append(tiff, encodeIFD(exifIFD, exifOffset)...)
	if gps != nil {
		tiff = append(tiff, encodeIFD(gps, gpsOffset)...)
	}
	return tiff
}

// embedExif inserts an EXIF APP1 segment right after the SOI marker. The
// entropy-coded image data is copied untouched, so there is no re-encode. Any
// EXIF segment already in the file is dropped.
func embedExif(jpeg []byte, meta exifMetadata) ([]byte, error) {
	if len(jpeg) < 4 || jpeg[0] != 0xFF || jpeg[1] != 0xD8 {
		return nil, errNotJPEG
	}
	payload := append([]byte("Exif\x00\x00"), buildExif(meta)...)
	if len(payload)+2 > math.MaxUint16 {
		return nil, errors.New("exif metadata too large")
	}
	segment := []byte{0xFF, 0xE1}
	segment = binary.BigEndian.AppendUint16(segment, uint16(len(payload)+2))
	segment = append(segment, payload...)

	out := make([]byte, 0, len(jpeg)+len(segment))
	out = append(out, jpeg[:2]...)
	out = append(out, segment...)

	pos := 2
	for pos+4 <= len(jpeg) && jpeg[pos] == 0xFF {
		marker := jpeg[pos+1]
		if marker == 0xFF {
			out = append(out, jpeg[pos])
			pos++
			continue
		}
		if marker == 0xDA || marker == 0xD9 {
			break
		}
		length := int(binary.BigEndian.Uint16(jpeg[pos+2:]))
		end := pos + 2 + length
		if length < 2 || end > len(jpeg) {
			return nil, errors.New("malformed JPEG segment")
		}
		if !(marker == 0xE1 && bytes.HasPrefix(jpeg[pos+4:end], []byte("Exif\x00\x00"))) {
			out = append(out, jpeg[pos:end]...)
		}
		pos = end
	}
	return append(out, jpeg[pos:]...), nil
}

func asciiOnly(s string) string {
	return strings.Map(func(r rune) rune {
		if r < 0x20 || r > 0x7E {
			return -1
		}
		return r
	}, s)
}

func drawingDescription(drawnAt time.Time, place string) string {
	description := "Drawn on Homie Display on " + drawnAt.UTC().Format("2 Jan 2006")
	if place = strings.TrimSpace(asciiOnly(place)); place != "" {
		description += " in " + place
	}
	return description
}
