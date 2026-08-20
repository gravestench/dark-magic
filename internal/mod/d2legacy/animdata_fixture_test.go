package d2legacy

import (
	"encoding/binary"
	"fmt"
)

// Read serves the one animation table used by focused timing tests. Rejecting
// every other path keeps accidental fixture dependencies visible.
func (fixtureRecords) Read(path string) ([]byte, error) {
	if path != "data/global/AnimData.d2" {
		return nil, fmt.Errorf("fixture records: binary %q is absent", path)
	}

	return attackAnimDataFixture(), nil
}

// attackAnimDataFixture returns a minimal binary animdata record whose frame
// count and speed are intentionally easy for timing assertions to audit.
func attackAnimDataFixture() []byte {
	const blocks, eventBytes = 256, 144

	records := []string{"AMA1HTH", "AMSCHTH"}

	data := make([]byte, 0, blocks*4+len(records)*160)
	for block := 0; block < blocks; block++ {
		word := make([]byte, 4)
		if block == 0 {
			binary.LittleEndian.PutUint32(word, uint32(len(records)))
		}

		data = append(data, word...)

		if block != 0 {
			continue
		}

		for _, name := range records {
			data = append(data, append([]byte(name), 0)...)

			binary.LittleEndian.PutUint32(word, 8)
			data = append(data, word...)
			half := make([]byte, 2)
			binary.LittleEndian.PutUint16(half, 128)
			data = append(data, half...)
			data = append(data, 0, 0)
			events := make([]byte, eventBytes)
			events[3] = 1
			data = append(data, events...)
		}
	}

	return data
}
