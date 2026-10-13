package apk

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"unicode/utf16"
)

// AXML Chunk Constants
const (
	chunkTypeXML            uint16 = 0x0003
	chunkTypeStringPool     uint16 = 0x0001
	chunkTypeResourceMap    uint16 = 0x0180
	chunkTypeStartNamespace uint16 = 0x0100
	chunkTypeEndNamespace   uint16 = 0x0101
	chunkTypeStartElement   uint16 = 0x0102
	chunkTypeEndElement     uint16 = 0x0103
	chunkTypeCDATA          uint16 = 0x0104
)

// ComponentInfo describes an activity, service, or receiver in the manifest.
type ComponentInfo struct {
	Name     string `json:"name"`
	Exported bool   `json:"exported"`
}

// ManifestData represents parsed AndroidManifest.xml attributes.
type ManifestData struct {
	Package          string          `json:"package"`
	VersionCode      int64           `json:"versionCode"`
	VersionName      string          `json:"versionName"`
	Permissions      []string        `json:"permissions"`
	Activities       []ComponentInfo `json:"activities"`
	Services         []ComponentInfo `json:"services"`
	Receivers        []ComponentInfo `json:"receivers"`
	HasOverlay       bool            `json:"hasOverlay"`
	HasAccessibility bool            `json:"hasAccessibility"`
}

// ParseAXML decodes an Android binary XML stream into ManifestData.
func ParseAXML(r io.Reader) (*ManifestData, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read axml: %w", err)
	}

	if len(data) < 8 {
		return nil, errors.New("axml data too short")
	}

	chunkType := binary.LittleEndian.Uint16(data[0:2])
	if chunkType != chunkTypeXML {
		return nil, fmt.Errorf("invalid axml header type: 0x%04x", chunkType)
	}

	manifest := &ManifestData{
		Permissions: []string{},
		Activities:  []ComponentInfo{},
		Services:    []ComponentInfo{},
		Receivers:   []ComponentInfo{},
	}

	offset := int(binary.LittleEndian.Uint16(data[2:4])) // header size (usually 8)
	var stringPool []string

	for offset < len(data) {
		if offset+8 > len(data) {
			break
		}

		cType := binary.LittleEndian.Uint16(data[offset : offset+2])
		cHeaderSize := binary.LittleEndian.Uint16(data[offset+2 : offset+4])
		cSize := binary.LittleEndian.Uint32(data[offset+4 : offset+8])
		if cSize == 0 || int(cSize) > len(data)-offset {
			break
		}

		chunkData := data[offset : offset+int(cSize)]

		switch cType {
		case chunkTypeStringPool:
			pool, err := parseStringPool(chunkData, cHeaderSize)
			if err == nil {
				stringPool = pool
			}
		case chunkTypeStartElement:
			if stringPool != nil {
				parseStartElement(chunkData, stringPool, manifest)
			}
		}

		offset += int(cSize)
	}

	return manifest, nil
}

func parseStringPool(chunk []byte, headerSize uint16) ([]string, error) {
	if len(chunk) < 28 {
		return nil, errors.New("string pool chunk too short")
	}

	stringCount := binary.LittleEndian.Uint32(chunk[8:12])
	flags := binary.LittleEndian.Uint32(chunk[16:20])
	stringsStart := binary.LittleEndian.Uint32(chunk[20:24])
	isUTF8 := (flags & (1 << 8)) != 0

	offsets := make([]uint32, stringCount)
	offsetStart := int(headerSize)

	for i := 0; i < int(stringCount); i++ {
		pos := offsetStart + (i * 4)
		if pos+4 > len(chunk) {
			break
		}
		offsets[i] = binary.LittleEndian.Uint32(chunk[pos : pos+4])
	}

	pool := make([]string, stringCount)
	dataStart := int(stringsStart)

	for i, off := range offsets {
		strPos := dataStart + int(off)
		if strPos >= len(chunk) {
			continue
		}

		if isUTF8 {
			// UTF-8 string: skip length bytes (1 or 2 bytes for char len, 1 or 2 for byte len)
			idx := strPos
			// Skip char length
			if idx < len(chunk) && chunk[idx]&0x80 != 0 {
				idx += 2
			} else {
				idx += 1
			}
			// Skip byte length
			var byteLen int
			if idx < len(chunk) && chunk[idx]&0x80 != 0 {
				if idx+1 < len(chunk) {
					byteLen = (int(chunk[idx]&0x7F) << 8) | int(chunk[idx+1])
					idx += 2
				}
			} else if idx < len(chunk) {
				byteLen = int(chunk[idx])
				idx += 1
			}

			if idx+byteLen <= len(chunk) {
				pool[i] = string(chunk[idx : idx+byteLen])
			}
		} else {
			// UTF-16LE string: 2 bytes length (or 4 bytes if high bit set)
			idx := strPos
			var charLen int
			if idx+2 <= len(chunk) {
				l1 := binary.LittleEndian.Uint16(chunk[idx : idx+2])
				idx += 2
				if l1&0x8000 != 0 && idx+2 <= len(chunk) {
					l2 := binary.LittleEndian.Uint16(chunk[idx : idx+2])
					idx += 2
					charLen = (int(l1&0x7FFF) << 16) | int(l2)
				} else {
					charLen = int(l1)
				}
			}

			byteLen := charLen * 2
			if idx+byteLen <= len(chunk) {
				u16 := make([]uint16, charLen)
				for j := 0; j < charLen; j++ {
					u16[j] = binary.LittleEndian.Uint16(chunk[idx+j*2 : idx+j*2+2])
				}
				pool[i] = string(utf16.Decode(u16))
			}
		}
	}

	return pool, nil
}

func parseStartElement(chunk []byte, pool []string, manifest *ManifestData) {
	if len(chunk) < 36 {
		return
	}

	nameIdx := binary.LittleEndian.Uint32(chunk[20:24])
	if int(nameIdx) >= len(pool) {
		return
	}
	elemName := pool[nameIdx]

	attrStart := binary.LittleEndian.Uint16(chunk[24:26])
	attrSize := binary.LittleEndian.Uint16(chunk[26:28])
	attrCount := binary.LittleEndian.Uint16(chunk[28:30])

	attrs := make(map[string]string)
	offset := int(attrStart)

	for i := 0; i < int(attrCount); i++ {
		pos := offset + (i * int(attrSize))
		if pos+20 > len(chunk) {
			break
		}

		attrNameIdx := binary.LittleEndian.Uint32(chunk[pos+4 : pos+8])
		rawValIdx := binary.LittleEndian.Uint32(chunk[pos+8 : pos+12])
		valType := chunk[pos+15]
		valData := binary.LittleEndian.Uint32(chunk[pos+16 : pos+20])

		var attrName, attrVal string
		if int(attrNameIdx) < len(pool) {
			attrName = pool[attrNameIdx]
		}

		if rawValIdx != 0xFFFFFFFF && int(rawValIdx) < len(pool) {
			attrVal = pool[rawValIdx]
		} else {
			// Interpret primitive typed value
			switch valType {
			case 0x03: // TYPE_STRING
				if int(valData) < len(pool) {
					attrVal = pool[valData]
				}
			case 0x10: // TYPE_INT_DEC
				attrVal = fmt.Sprintf("%d", int32(valData))
			case 0x11: // TYPE_INT_HEX
				attrVal = fmt.Sprintf("0x%x", valData)
			case 0x12: // TYPE_INT_BOOLEAN
				if valData != 0 {
					attrVal = "true"
				} else {
					attrVal = "false"
				}
			}
		}

		if attrName != "" {
			attrs[attrName] = attrVal
		}
	}

	switch elemName {
	case "manifest":
		if pkg, ok := attrs["package"]; ok {
			manifest.Package = pkg
		}
		if vc, ok := attrs["versionCode"]; ok {
			var code int64
			_, _ = fmt.Sscanf(vc, "%d", &code)
			manifest.VersionCode = code
		}
		if vn, ok := attrs["versionName"]; ok {
			manifest.VersionName = vn
		}
	case "uses-permission":
		if name, ok := attrs["name"]; ok {
			manifest.Permissions = append(manifest.Permissions, name)
			if name == "android.permission.SYSTEM_ALERT_WINDOW" {
				manifest.HasOverlay = true
			}
		}
	case "activity":
		if name, ok := attrs["name"]; ok {
			comp := ComponentInfo{Name: name, Exported: attrs["exported"] == "true"}
			manifest.Activities = append(manifest.Activities, comp)
		}
	case "service":
		if name, ok := attrs["name"]; ok {
			comp := ComponentInfo{Name: name, Exported: attrs["exported"] == "true"}
			manifest.Services = append(manifest.Services, comp)
			if attrs["permission"] == "android.permission.BIND_ACCESSIBILITY_SERVICE" {
				manifest.HasAccessibility = true
			}
		}
	case "receiver":
		if name, ok := attrs["name"]; ok {
			comp := ComponentInfo{Name: name, Exported: attrs["exported"] == "true"}
			manifest.Receivers = append(manifest.Receivers, comp)
		}
	}
}

// GenerateSyntheticAXML produces a valid binary AndroidManifest.xml buffer for testing without external tools.
func GenerateSyntheticAXML(pkg, versionName string, versionCode int64, permissions []string) []byte {
	var poolBuf bytes.Buffer
	var offsets []uint32

	// Build strings: manifest, package, versionCode, versionName, uses-permission, name, plus custom strings
	rawStrings := []string{
		"manifest", "package", "versionCode", "versionName", "uses-permission", "name",
		pkg, fmt.Sprintf("%d", versionCode), versionName,
	}
	rawStrings = append(rawStrings, permissions...)

	for _, s := range rawStrings {
		offsets = append(offsets, uint32(poolBuf.Len()))
		u16 := utf16.Encode([]rune(s))
		_ = binary.Write(&poolBuf, binary.LittleEndian, uint16(len(u16)))
		for _, c := range u16 {
			_ = binary.Write(&poolBuf, binary.LittleEndian, c)
		}
		_ = binary.Write(&poolBuf, binary.LittleEndian, uint16(0)) // null terminator
	}

	// Pad pool to 4-byte boundary
	for poolBuf.Len()%4 != 0 {
		poolBuf.WriteByte(0)
	}

	strPoolHeaderSize := uint16(28)
	strOffsetsSize := len(offsets) * 4
	stringsStart := uint32(strPoolHeaderSize) + uint32(strOffsetsSize)
	strPoolChunkSize := stringsStart + uint32(poolBuf.Len())

	var strPoolChunk bytes.Buffer
	_ = binary.Write(&strPoolChunk, binary.LittleEndian, chunkTypeStringPool)
	_ = binary.Write(&strPoolChunk, binary.LittleEndian, strPoolHeaderSize)
	_ = binary.Write(&strPoolChunk, binary.LittleEndian, strPoolChunkSize)
	_ = binary.Write(&strPoolChunk, binary.LittleEndian, uint32(len(rawStrings))) // stringCount
	_ = binary.Write(&strPoolChunk, binary.LittleEndian, uint32(0))                // styleCount
	_ = binary.Write(&strPoolChunk, binary.LittleEndian, uint32(0))                // flags (UTF-16)
	_ = binary.Write(&strPoolChunk, binary.LittleEndian, stringsStart)
	_ = binary.Write(&strPoolChunk, binary.LittleEndian, uint32(0)) // stylesStart

	for _, off := range offsets {
		_ = binary.Write(&strPoolChunk, binary.LittleEndian, off)
	}
	strPoolChunk.Write(poolBuf.Bytes())

	// Start manifest element
	var elemManifest bytes.Buffer
	_ = binary.Write(&elemManifest, binary.LittleEndian, chunkTypeStartElement)
	_ = binary.Write(&elemManifest, binary.LittleEndian, uint16(16))               // header size
	_ = binary.Write(&elemManifest, binary.LittleEndian, uint32(36+20*3))          // chunk size (36 + 3 attrs)
	_ = binary.Write(&elemManifest, binary.LittleEndian, uint32(1))                // line number
	_ = binary.Write(&elemManifest, binary.LittleEndian, uint32(0xFFFFFFFF))       // comment
	_ = binary.Write(&elemManifest, binary.LittleEndian, uint32(0xFFFFFFFF))       // ns
	_ = binary.Write(&elemManifest, binary.LittleEndian, uint32(0))                // name ("manifest")
	_ = binary.Write(&elemManifest, binary.LittleEndian, uint16(36))               // attrStart
	_ = binary.Write(&elemManifest, binary.LittleEndian, uint16(20))               // attrSize
	_ = binary.Write(&elemManifest, binary.LittleEndian, uint16(3))                // attrCount
	_ = binary.Write(&elemManifest, binary.LittleEndian, uint16(0))                // id
	_ = binary.Write(&elemManifest, binary.LittleEndian, uint16(0))                // class
	_ = binary.Write(&elemManifest, binary.LittleEndian, uint16(0))                // style
	// Attr 1: package
	_ = binary.Write(&elemManifest, binary.LittleEndian, uint32(0xFFFFFFFF)) // ns
	_ = binary.Write(&elemManifest, binary.LittleEndian, uint32(1))          // name ("package")
	_ = binary.Write(&elemManifest, binary.LittleEndian, uint32(6))          // raw val (pkg)
	_ = binary.Write(&elemManifest, binary.LittleEndian, uint16(8))          // size
	_ = binary.Write(&elemManifest, binary.LittleEndian, uint8(0))           // res
	_ = binary.Write(&elemManifest, binary.LittleEndian, uint8(0x03))        // type string
	_ = binary.Write(&elemManifest, binary.LittleEndian, uint32(6))          // data
	// Attr 2: versionCode
	_ = binary.Write(&elemManifest, binary.LittleEndian, uint32(0xFFFFFFFF))
	_ = binary.Write(&elemManifest, binary.LittleEndian, uint32(2))
	_ = binary.Write(&elemManifest, binary.LittleEndian, uint32(7))
	_ = binary.Write(&elemManifest, binary.LittleEndian, uint16(8))
	_ = binary.Write(&elemManifest, binary.LittleEndian, uint8(0))
	_ = binary.Write(&elemManifest, binary.LittleEndian, uint8(0x10)) // int dec
	_ = binary.Write(&elemManifest, binary.LittleEndian, uint32(versionCode))
	// Attr 3: versionName
	_ = binary.Write(&elemManifest, binary.LittleEndian, uint32(0xFFFFFFFF))
	_ = binary.Write(&elemManifest, binary.LittleEndian, uint32(3))
	_ = binary.Write(&elemManifest, binary.LittleEndian, uint32(8))
	_ = binary.Write(&elemManifest, binary.LittleEndian, uint16(8))
	_ = binary.Write(&elemManifest, binary.LittleEndian, uint8(0))
	_ = binary.Write(&elemManifest, binary.LittleEndian, uint8(0x03))
	_ = binary.Write(&elemManifest, binary.LittleEndian, uint32(8))

	// Permission elements
	var permsBuf bytes.Buffer
	for idx, _ := range permissions {
		permStrIdx := uint32(9 + idx)
		_ = binary.Write(&permsBuf, binary.LittleEndian, chunkTypeStartElement)
		_ = binary.Write(&permsBuf, binary.LittleEndian, uint16(16))
		_ = binary.Write(&permsBuf, binary.LittleEndian, uint32(56)) // 36 + 20
		_ = binary.Write(&permsBuf, binary.LittleEndian, uint32(2))
		_ = binary.Write(&permsBuf, binary.LittleEndian, uint32(0xFFFFFFFF))
		_ = binary.Write(&permsBuf, binary.LittleEndian, uint32(0xFFFFFFFF))
		_ = binary.Write(&permsBuf, binary.LittleEndian, uint32(4)) // "uses-permission"
		_ = binary.Write(&permsBuf, binary.LittleEndian, uint16(36))
		_ = binary.Write(&permsBuf, binary.LittleEndian, uint16(20))
		_ = binary.Write(&permsBuf, binary.LittleEndian, uint16(1)) // 1 attr
		_ = binary.Write(&permsBuf, binary.LittleEndian, uint16(0))
		_ = binary.Write(&permsBuf, binary.LittleEndian, uint16(0))
		_ = binary.Write(&permsBuf, binary.LittleEndian, uint16(0))
		// Attr: name
		_ = binary.Write(&permsBuf, binary.LittleEndian, uint32(0xFFFFFFFF))
		_ = binary.Write(&permsBuf, binary.LittleEndian, uint32(5)) // "name"
		_ = binary.Write(&permsBuf, binary.LittleEndian, permStrIdx)
		_ = binary.Write(&permsBuf, binary.LittleEndian, uint16(8))
		_ = binary.Write(&permsBuf, binary.LittleEndian, uint8(0))
		_ = binary.Write(&permsBuf, binary.LittleEndian, uint8(0x03))
		_ = binary.Write(&permsBuf, binary.LittleEndian, permStrIdx)
	}

	totalSize := uint32(8 + strPoolChunk.Len() + elemManifest.Len() + permsBuf.Len())
	var fileBuf bytes.Buffer
	_ = binary.Write(&fileBuf, binary.LittleEndian, chunkTypeXML)
	_ = binary.Write(&fileBuf, binary.LittleEndian, uint16(8))
	_ = binary.Write(&fileBuf, binary.LittleEndian, totalSize)

	fileBuf.Write(strPoolChunk.Bytes())
	fileBuf.Write(elemManifest.Bytes())
	fileBuf.Write(permsBuf.Bytes())

	return fileBuf.Bytes()
}
