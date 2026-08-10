package audio

import (
	"fmt"
	"io/fs"
	"path"
	"strconv"
	"strings"

	"github.com/gravestench/dark-magic/internal/game/data/model"
)

// Definition is a resolved Sounds.txt record and its archive asset.
type Definition struct {
	Name    string
	Path    string
	Format  string
	Data    []byte
	Options PlayOptions
}

// Catalog resolves stable Sounds.txt identifiers through layered content.
type Catalog struct {
	source  fs.FS
	records SoundRecords
}

// SoundRecords is the narrow typed game-data view required by audio. Ordering
// is significant because redirects are indices and groups are adjacent rows.
type SoundRecords interface {
	SoundRecords() ([]models.SoundEntry, error)
}

// NewCatalog connects typed sound records to the layered content filesystem.
// Resolution remains lazy so startup does not decode or retain every sound.
func NewCatalog(source fs.FS, records SoundRecords) *Catalog {
	return &Catalog{source: source, records: records}
}

// Resolve deterministically selects grouped variants using seed.
func (c *Catalog) Resolve(name string, seed uint64) (Definition, error) {
	if c == nil || c.records == nil {
		return Definition{}, fmt.Errorf("audio: no sound records")
	}
	rows, err := c.records.SoundRecords()
	if err != nil {
		return Definition{}, fmt.Errorf("audio: load sound records: %w", err)
	}
	index := -1
	for candidate, row := range rows {
		if strings.EqualFold(row.Sound, name) {
			index = candidate
			break
		}
	}
	if index < 0 {
		return Definition{}, fmt.Errorf("audio: unknown sound record %q", name)
	}
	row := rows[index]
	if redirect, err := strconv.Atoi(row.Redirect); err == nil && redirect >= 0 && redirect < len(rows) {
		row = rows[redirect]
	}
	groupSize := row.GroupSize
	if groupSize > 1 {
		end := index + groupSize
		if end > len(rows) {
			end = len(rows)
		}
		row = weighted(rows[index:end], seed)
	}
	fileName, err := c.find(row.FileName, row.IsLocal != 0, row.IsMusic != 0)
	if err != nil {
		return Definition{}, fmt.Errorf("audio: sound %q: %w", name, err)
	}
	minimum, maximum := row.VolumeMin, row.VolumeMax
	if maximum <= 0 {
		maximum = 255
	}
	if minimum < 0 {
		minimum = 0
	}
	if minimum > maximum {
		minimum = maximum
	}
	volume := minimum
	if maximum > minimum {
		volume += int(seed % uint64(maximum-minimum+1))
	}
	bus := "sfx"
	if row.IsMusic != 0 {
		bus = "music"
	} else if row.IsUI != 0 {
		bus = "ui"
	} else if row.IsAmbientScene != 0 || row.IsAmbientEvent != 0 {
		bus = "ambience"
	} else if channel := strings.ToLower(row.Channel); strings.Contains(channel, "voice") || strings.Contains(channel, "speech") || strings.Contains(channel, "vocal") {
		bus = "speech"
	}
	data, err := fs.ReadFile(c.source, fileName)
	if err != nil {
		return Definition{}, fmt.Errorf("audio: read %q: %w", fileName, err)
	}
	return Definition{
		Name: row.Sound, Path: fileName, Format: strings.ToLower(path.Ext(fileName)),
		Data:    data,
		Options: PlayOptions{Bus: bus, Volume: float32(volume) / 255, Loop: row.Loop != 0, Stream: row.Stream != 0 || bus == "music", Group: name},
	}, nil
}

func (c *Catalog) find(fileName string, local, music bool) (string, error) {
	fileName = strings.TrimPrefix(strings.ReplaceAll(fileName, "\\", "/"), "/")
	if fileName == "" {
		return "", fmt.Errorf("empty FileName")
	}
	candidates := []string{fileName}
	if local {
		candidates = append(candidates, path.Join("data/local/sfx", fileName))
	}
	if music {
		candidates = append(candidates, path.Join("data/global/music", fileName))
	}
	candidates = append(candidates, path.Join("data/global/sfx", fileName))
	for _, candidate := range candidates {
		if _, err := fs.Stat(c.source, candidate); err == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("asset %q not found", fileName)
}

func weighted(rows []models.SoundEntry, seed uint64) models.SoundEntry {
	total := 0
	for _, row := range rows {
		weight := row.GroupWeight
		if weight <= 0 {
			weight = 1
		}
		total += weight
	}
	pick := int(seed % uint64(total))
	for _, row := range rows {
		weight := row.GroupWeight
		if weight <= 0 {
			weight = 1
		}
		if pick < weight {
			return row
		}
		pick -= weight
	}
	return rows[0]
}
