package assetdecode

import (
	"encoding/binary"
	"fmt"
	"unicode/utf16"
	"unicode/utf8"
)

// Text decodes UTF-8 or BOM-marked UTF-16 text assets used by localized UI
// resources such as ExpansionCredits.txt. UTF-16 requires a BOM so byte order
// is never guessed from localized content.
func Text(data []byte) (string, error) {
	if hasUTF16ByteOrderMark(data) {
		return decodeUTF16Text(data)
	}

	if !utf8.Valid(data) {
		return "", fmt.Errorf("text: asset is neither UTF-8 nor BOM-marked UTF-16")
	}

	return string(data), nil
}

// hasUTF16ByteOrderMark recognizes both supported byte orders without reading
// beyond short inputs that must instead be validated as UTF-8.
func hasUTF16ByteOrderMark(data []byte) bool {
	return len(data) >= 2 &&
		(data[0] == 0xff && data[1] == 0xfe || data[0] == 0xfe && data[1] == 0xff)
}

// decodeUTF16Text consumes the byte-order mark and rejects incomplete code
// units so malformed assets cannot silently lose their trailing byte.
func decodeUTF16Text(data []byte) (string, error) {
	if (len(data)-2)%2 != 0 {
		return "", fmt.Errorf("text: odd UTF-16 byte count %d", len(data)-2)
	}

	var order binary.ByteOrder = binary.LittleEndian
	if data[0] == 0xfe {
		order = binary.BigEndian
	}

	units := make([]uint16, (len(data)-2)/2)
	for index := range units {
		units[index] = order.Uint16(data[2+index*2:])
	}

	return string(utf16.Decode(units)), nil
}
