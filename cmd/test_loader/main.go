package main

import (
	"github.com/gravestench/runtime"
	"github.com/gravestench/runtime/examples/services/config_file"
	"github.com/gravestench/runtime/examples/services/web_router"

	"github.com/gravestench/dark-magic/pkg/services/d2_asset_loader"
	"github.com/gravestench/dark-magic/pkg/services/dc6_loader"
	"github.com/gravestench/dark-magic/pkg/services/dcc_loader"
	"github.com/gravestench/dark-magic/pkg/services/ds1_loader"
	"github.com/gravestench/dark-magic/pkg/services/dt1_loader"
	"github.com/gravestench/dark-magic/pkg/services/font_table_loader"
	"github.com/gravestench/dark-magic/pkg/services/gpl_loader"
	"github.com/gravestench/dark-magic/pkg/services/lua"
	"github.com/gravestench/dark-magic/pkg/services/mpq_loader"
	"github.com/gravestench/dark-magic/pkg/services/pl2_loader"
	"github.com/gravestench/dark-magic/pkg/services/record_manager"
	"github.com/gravestench/dark-magic/pkg/services/tbl_loader"
	"github.com/gravestench/dark-magic/pkg/services/tsv_loader"
	"github.com/gravestench/dark-magic/pkg/services/wav_loader"
	"github.com/gravestench/dark-magic/pkg/services/web_server"
)

const (
	projectName      = "Dark Magic Runtime"
	projectConfigDir = "~/.config/dark-magic"
)

func main() {
	rt := runtime.New(projectName)

	// utility services
	rt.Add(&lua.Service{})
	rt.Add(&web_server.Service{})
	rt.Add(&web_router.Service{})
	rt.Add(&config_file.Service{RootDirectory: projectConfigDir})

	// d2 file loaders
	rt.Add(&font_table_loader.Service{})
	rt.Add(&dc6_loader.Service{})
	rt.Add(&dcc_loader.Service{})
	rt.Add(&ds1_loader.Service{})
	rt.Add(&dt1_loader.Service{})
	rt.Add(&gpl_loader.Service{})
	rt.Add(&pl2_loader.Service{})
	rt.Add(&tbl_loader.Service{})
	rt.Add(&tsv_loader.Service{})
	rt.Add(&wav_loader.Service{})
	rt.Add(&mpq_loader.Service{})

	// high level d2 services
	rt.Add(&d2_asset_loader.Service{})
	rt.Add(&record_manager.Service{})

	rt.Run()
}
