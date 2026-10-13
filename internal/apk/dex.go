package apk

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strings"
)

// DEXHeader represents the 112-byte header of an Android Dalvik Executable.
type DEXHeader struct {
	Magic         [8]byte
	Checksum      uint32
	Signature     [20]byte
	FileSize      uint32
	HeaderSize    uint32
	EndianTag     uint32
	LinkSize      uint32
	LinkOff       uint32
	MapOff        uint32
	StringIDsSize uint32
	StringIDsOff  uint32
}

var knownAdSDKSignatures = map[string]string{
	"com/google/android/gms/ads":    "Google Mobile Ads (AdMob)",
	"com/unity3d/ads":               "Unity Ads",
	"com/applovin":                  "AppLovin MAX",
	"com/ironsource":                "ironSource Ads",
	"com/mbridge/msdk":              "Mintegral SDK",
	"com/vungle":                    "Vungle (Liftoff)",
	"com/inmobi":                    "InMobi SDK",
	"com/bytedance/sdk/openadsdk":   "Pangle (ByteDance)",
	"com/adcolony":                  "AdColony",
	"com/chartboost":                "Chartboost",
	"com/fyber":                     "Digital Turbine (Fyber)",
}

// DEXAnalysis holds scanned DEX strings and detected ad networks.
type DEXAnalysis struct {
	DetectedAdSDKs map[string]string `json:"detectedAdSdks"` // signature -> human label
	TotalStrings   int               `json:"totalStrings"`
}

// ParseDEX parses a DEX stream and identifies ad networks and embedded namespaces.
func ParseDEX(r io.Reader) (*DEXAnalysis, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read dex: %w", err)
	}

	if len(data) < 112 {
		return nil, errors.New("dex file smaller than standard 112-byte header")
	}

	// Validate magic
	magic := string(data[:4])
	if magic != "dex\n" {
		return nil, fmt.Errorf("invalid dex magic: %q", magic)
	}

	var h DEXHeader
	buf := bytes.NewReader(data[:112])
	if err := binary.Read(buf, binary.LittleEndian, &h); err != nil {
		return nil, fmt.Errorf("read dex header: %w", err)
	}

	analysis := &DEXAnalysis{
		DetectedAdSDKs: make(map[string]string),
		TotalStrings:   int(h.StringIDsSize),
	}

	// Iterate string ID offsets
	strCount := int(h.StringIDsSize)
	baseOff := int(h.StringIDsOff)

	for i := 0; i < strCount; i++ {
		offPos := baseOff + (i * 4)
		if offPos+4 > len(data) {
			break
		}
		strDataOff := int(binary.LittleEndian.Uint32(data[offPos : offPos+4]))
		if strDataOff >= len(data) {
			continue
		}

		// Read MUTF-8 length (ULEB128)
		idx := strDataOff
		_, bytesRead := readULEB128(data[idx:])
		idx += bytesRead

		// Read null-terminated string
		end := idx
		for end < len(data) && data[end] != 0 {
			end++
		}
		if end > idx {
			strVal := string(data[idx:end])
			for sig, label := range knownAdSDKSignatures {
				if strings.Contains(strVal, sig) {
					analysis.DetectedAdSDKs[sig] = label
				}
			}
		}
	}

	return analysis, nil
}

func readULEB128(data []byte) (uint32, int) {
	var result uint32
	var shift uint
	var count int

	for i, b := range data {
		count++
		result |= uint32(b&0x7F) << shift
		if b&0x80 == 0 {
			break
		}
		shift += 7
		if i >= 4 {
			break
		}
	}
	return result, count
}

// GenerateSyntheticDEX creates a minimal valid DEX binary buffer containing string entries.
func GenerateSyntheticDEX(strList []string) []byte {
	var stringsData bytes.Buffer
	var stringOffsets []uint32

	headerSize := uint32(112)
	stringIDsOff := headerSize
	stringIDsSize := uint32(len(strList))
	stringsDataStart := stringIDsOff + (stringIDsSize * 4)

	for _, s := range strList {
		off := stringsDataStart + uint32(stringsData.Len())
		stringOffsets = append(stringOffsets, off)

		// ULEB128 length of string
		lenBytes := encodeULEB128(uint32(len(s)))
		stringsData.Write(lenBytes)
		stringsData.WriteString(s)
		stringsData.WriteByte(0) // null terminator
	}

	totalSize := stringsDataStart + uint32(stringsData.Len())

	var out bytes.Buffer
	// Magic
	out.WriteString("dex\n035\x00")
	// Checksum (dummy)
	_ = binary.Write(&out, binary.LittleEndian, uint32(0))
	// Signature (20 bytes dummy)
	out.Write(make([]byte, 20))
	// FileSize
	_ = binary.Write(&out, binary.LittleEndian, totalSize)
	// HeaderSize
	_ = binary.Write(&out, binary.LittleEndian, headerSize)
	// EndianTag
	_ = binary.Write(&out, binary.LittleEndian, uint32(0x12345678))
	// LinkSize, LinkOff, MapOff
	_ = binary.Write(&out, binary.LittleEndian, uint32(0))
	_ = binary.Write(&out, binary.LittleEndian, uint32(0))
	_ = binary.Write(&out, binary.LittleEndian, uint32(0))
	// StringIDsSize, StringIDsOff
	_ = binary.Write(&out, binary.LittleEndian, stringIDsSize)
	_ = binary.Write(&out, binary.LittleEndian, stringIDsOff)
	// Other fields in 112 byte header (type_ids, proto_ids, field_ids, method_ids, class_defs, data)
	// 112 - 64 = 48 bytes of zeros
	out.Write(make([]byte, 48))

	// Write string offsets
	for _, off := range stringOffsets {
		_ = binary.Write(&out, binary.LittleEndian, off)
	}

	// Write string data
	out.Write(stringsData.Bytes())

	return out.Bytes()
}

func encodeULEB128(val uint32) []byte {
	var res []byte
	for {
		b := byte(val & 0x7F)
		val >>= 7
		if val != 0 {
			b |= 0x80
		}
		res = append(res, b)
		if val == 0 {
			break
		}
	}
	return res
}
