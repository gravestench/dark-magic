package content

import (
	"reflect"
	"testing"
	"testing/fstest"
)

func TestResolvePresentationProfile(t *testing.T) {
	source := fstest.MapFS{presentationManifest: &fstest.MapFile{Data: []byte(`{
      "supported_profiles":[
        {"id":"wide","resolution":{"width":800,"height":600}},
        {"id":"classic","resolution":{"width":640,"height":480},"screens":["game_world"]}
      ]}`)}}
	profile, err := ResolvePresentationProfile(source, "classic")
	if err != nil {
		t.Fatal(err)
	}
	if profile.ID != "classic" || profile.Width != 640 || profile.Height != 480 || !reflect.DeepEqual(profile.ScreenIDs, []string{"game_world"}) {
		t.Fatalf("profile = %#v", profile)
	}
	if _, err := ResolvePresentationProfile(source, "missing"); err == nil {
		t.Fatal("unsupported profile was accepted")
	}
}
