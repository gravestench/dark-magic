package content

import (
	"strings"
	"testing"
	"testing/fstest"
)

func TestPresentationBootstrapComesFromManifest(t *testing.T) {
	const manifest = `{"schema":"darkmagic.presentation/v1","palettes":{"loading":"palette.dat"},"screens":{"title":{"background":"title.dc6"},"game_loading":{"sheet":"loading.dc6","palette":"loading"}}}`
	source := fstest.MapFS{
		presentationManifest: {Data: []byte(manifest)},
		"title.dc6":          {Data: []byte("fixture")},
	}
	bootstrap, err := LoadPresentationBootstrap(source)
	if err != nil {
		t.Fatal(err)
	}
	if bootstrap.TitleBackground != "title.dc6" || len(bootstrap.LoadingAssets) != 2 || bootstrap.LoadingAssets[0] != "loading.dc6" || bootstrap.LoadingAssets[1] != "palette.dat" {
		t.Fatalf("bootstrap = %#v", bootstrap)
	}
	if err := ValidateClientAssets(source); err != nil {
		t.Fatal(err)
	}
	delete(source, "title.dc6")
	if err := ValidateClientAssets(source); err == nil || !strings.Contains(err.Error(), "MPQ_DIRECTORY") {
		t.Fatalf("missing asset error = %v", err)
	}
}
