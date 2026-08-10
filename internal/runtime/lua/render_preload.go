package modruntime

import (
	"fmt"
	"io/fs"
	"runtime"
	"sync"

	lua "github.com/yuin/gopher-lua"
)

// AssetPreloadRequest describes CPU-side work that is safe to perform away
// from the renderer thread. Texture creation is deliberately not represented:
// the renderer adapter remains the sole owner of native graphics resources.
type AssetPreloadRequest struct {
	Kind      string
	Path      string
	Secondary string
	Palette   string
	Table     string
	Sheet     string
	Transform string
	Tiles     []string
	Direction int
	Frame     int
	Anchor    string
}

// AssetPreloadStatus is an immutable snapshot suitable for loading screens,
// diagnostics, or deciding whether a scene transition can proceed.
type AssetPreloadStatus struct {
	Total, Completed, Failed int
	Done                     bool
	Errors                   []string
}

type assetPreloadJob struct {
	total, completed, failed int
	errors                   []string
}

type assetPreloader struct {
	assets fs.FS
	cache  *renderAssetCache

	mu     sync.Mutex
	nextID uint64
	jobs   map[uint64]*assetPreloadJob
	work   chan struct{}
}

func newAssetPreloader(assets fs.FS, cache *renderAssetCache) *assetPreloader {
	workers := runtime.GOMAXPROCS(0)
	if workers > 4 {
		workers = 4
	}
	if workers < 1 {
		workers = 1
	}
	return &assetPreloader{assets: assets, cache: cache, jobs: make(map[uint64]*assetPreloadJob), work: make(chan struct{}, workers)}
}

func (p *assetPreloader) Start(requests []AssetPreloadRequest) uint64 {
	p.mu.Lock()
	p.nextID++
	id := p.nextID
	p.jobs[id] = &assetPreloadJob{total: len(requests)}
	p.mu.Unlock()
	if len(requests) == 0 {
		return id
	}

	queue := make(chan AssetPreloadRequest)
	workerCount := min(len(requests), cap(p.work))
	for worker := 0; worker < workerCount; worker++ {
		go func() {
			for request := range queue {
				p.work <- struct{}{}
				err := p.load(request)
				<-p.work

				p.mu.Lock()
				job := p.jobs[id]
				job.completed++
				if err != nil {
					job.failed++
					job.errors = append(job.errors, fmt.Sprintf("%s %q: %v", request.Kind, request.Path, err))
				}
				p.mu.Unlock()
			}
		}()
	}
	go func() {
		defer close(queue)
		for _, request := range requests {
			queue <- request
		}
	}()
	return id
}

func (p *assetPreloader) Status(id uint64) (AssetPreloadStatus, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	job, ok := p.jobs[id]
	if !ok {
		return AssetPreloadStatus{}, false
	}
	return AssetPreloadStatus{
		Total: job.total, Completed: job.completed, Failed: job.failed,
		Done: job.completed == job.total, Errors: append([]string(nil), job.errors...),
	}, true
}

func (p *assetPreloader) load(request AssetPreloadRequest) error {
	switch request.Kind {
	case "image":
		_, err := p.cache.loadImage(p.assets, request.Path)
		return err
	case "dc6":
		_, err := p.cache.loadDC6(p.assets, request.Path, request.Palette)
		return err
	case "dc6_frame":
		_, err := p.cache.loadDC6Frame(p.assets, request.Path, request.Palette, request.Direction, request.Frame)
		return err
	case "dc6_combined":
		_, err := p.cache.loadDC6Combined(p.assets, request.Path, request.Palette, request.Direction)
		return err
	case "dc6_animation":
		_, err := p.cache.loadDC6Animation(p.assets, request.Path, request.Palette, request.Direction, request.Anchor)
		return err
	case "dc6_composite":
		first, err := p.cache.loadDC6(p.assets, request.Path, request.Palette)
		if err != nil {
			return err
		}
		second, err := p.cache.loadDC6(p.assets, request.Secondary, request.Palette)
		if err != nil {
			return err
		}
		bounds := dc6AnimationBounds(first, request.Direction).Union(dc6AnimationBounds(second, request.Direction))
		if _, err := p.cache.loadDC6Animation(p.assets, request.Path, request.Palette, request.Direction, request.Anchor, bounds); err != nil {
			return err
		}
		_, err = p.cache.loadDC6Animation(p.assets, request.Secondary, request.Palette, request.Direction, request.Anchor, bounds)
		return err
	case "dcc":
		_, err := p.cache.loadDCC(p.assets, request.Path, request.Palette)
		return err
	case "cof":
		_, err := p.cache.loadCOF(p.assets, request.Path)
		return err
	case "font":
		_, err := p.cache.loadFont(p.assets, request.Table, request.Sheet, request.Palette, request.Transform)
		return err
	case "ds1":
		_, err := p.cache.loadDS1(p.assets, request.Path, request.Tiles, request.Palette)
		return err
	default:
		return fmt.Errorf("unsupported asset kind %q", request.Kind)
	}
}

func luaPreloadRequests(state *lua.LState, index int) ([]AssetPreloadRequest, error) {
	table := state.CheckTable(index)
	if table.Len() == 0 {
		return nil, fmt.Errorf("at least one preload request is required")
	}
	requests := make([]AssetPreloadRequest, 0, table.Len())
	for item := 1; item <= table.Len(); item++ {
		definition, ok := table.RawGetInt(item).(*lua.LTable)
		if !ok {
			return nil, fmt.Errorf("request %d must be a table", item)
		}
		stringField := func(name string) string {
			value, _ := definition.RawGetString(name).(lua.LString)
			return string(value)
		}
		request := AssetPreloadRequest{
			Kind: stringField("kind"), Path: stringField("path"), Secondary: stringField("overlay"), Palette: stringField("palette"),
			Table: stringField("table"), Sheet: stringField("sheet"), Transform: stringField("transform"),
		}
		if value, ok := definition.RawGetString("direction").(lua.LNumber); ok {
			request.Direction = int(value)
		}
		if value, ok := definition.RawGetString("frame").(lua.LNumber); ok {
			request.Frame = int(value)
		}
		request.Anchor = stringField("anchor")
		if request.Anchor == "" {
			request.Anchor = "offsets"
		}
		if tileTable, ok := definition.RawGetString("tiles").(*lua.LTable); ok {
			for tile := 1; tile <= tileTable.Len(); tile++ {
				value, ok := tileTable.RawGetInt(tile).(lua.LString)
				if !ok {
					return nil, fmt.Errorf("request %d tile %d must be a path", item, tile)
				}
				request.Tiles = append(request.Tiles, string(value))
			}
		}
		if request.Kind == "" {
			return nil, fmt.Errorf("request %d kind is required", item)
		}
		if request.Kind == "font" {
			if request.Table == "" || request.Sheet == "" {
				return nil, fmt.Errorf("request %d font table and sheet are required", item)
			}
		} else if request.Path == "" {
			return nil, fmt.Errorf("request %d path is required", item)
		}
		if request.Kind == "dc6_composite" && request.Secondary == "" {
			return nil, fmt.Errorf("request %d composite overlay is required", item)
		}
		requests = append(requests, request)
	}
	return requests, nil
}
