package d2legacy

import (
	"encoding/binary"
	"fmt"
)

func (fixtureRecords) Read(path string) ([]byte, error) {
	if path != "data/global/AnimData.d2" {
		return nil, fmt.Errorf("fixture records: binary %q is absent", path)
	}
	return attackAnimDataFixture(), nil
}

func attackAnimDataFixture() []byte {
	const blocks, eventBytes = 256, 144
	data := make([]byte, 0, blocks*4+160)
	for block := 0; block < blocks; block++ {
		word := make([]byte, 4)
		if block == 0 {
			binary.LittleEndian.PutUint32(word, 1)
		}
		data = append(data, word...)
		if block != 0 {
			continue
		}
		data = append(data, []byte{'A', 'M', 'A', '1', 'H', 'T', 'H', 0}...)
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
	return data
}
