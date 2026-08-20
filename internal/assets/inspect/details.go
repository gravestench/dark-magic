package assetinspect

import "github.com/gravestench/dark-magic/internal/assets/decode"

// decodeDetails selects only the decoder implied by the file extension. Keeping
// unsupported formats successful preserves InspectData's metadata-only fallback.
func decodeDetails(extension string, data []byte) (any, error) {
	switch extension {
	case "bik":
		return assetdecode.BIK(data)
	case "dc6":
		return decodeDC6Details(data)
	case "dcc":
		return decodeDCCDetails(data)
	case "ds1":
		return decodeDS1Details(data)
	case "dt1":
		return decodeDT1Details(data)
	case "tbl":
		return decodeTBLDetails(data)
	case "txt", "tsv":
		return decodeTabularTextDetails(data), nil
	default:
		return nil, nil
	}
}
