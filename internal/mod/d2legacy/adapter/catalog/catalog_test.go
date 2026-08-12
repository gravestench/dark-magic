package catalog

import (
	"context"
	"fmt"
	"os"
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/mod/d2legacy/data/recovered"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
)

type testLocale map[string]string

func (locale testLocale) Text(key string) (string, error) {
	value, found := locale[key]
	if !found {
		return key, fmt.Errorf("missing %s", key)
	}
	return value, nil
}

func TestModulesExposeImmutableRecoveredRelationships(t *testing.T) {
	source := fstest.MapFS{
		recovered.QuestsPath:   &fstest.MapFile{Data: []byte("id\tname\tact\torder\tvisible\ticon\tquestdone\tqstr\tqsts1\n0\tPrologue\t0\t\t\t\t\tq0\ts0\n1\tDen\t0\t1\t1\ta1q1\t0\tq1\ts1\n")},
		recovered.SpeechPath:   &fstest.MapFile{Data: []byte("sound\tsoundstr\nakara_intro\tAkaraIntroGossip1\n")},
		recovered.DS1TypesPath: &fstest.MapFile{Data: []byte("Name\tDef\tLevelType\nTown\t1\t1\n")},
		recovered.ObjectsPath:  &fstest.MapFile{Data: []byte("Act\tId\tDescription\tObjectId\n1\t0\tFountain\t12\n")},
	}
	runtime := modruntime.New()
	catalog := recovered.New(source)
	if err := runtime.RegisterModule(QuestModule(catalog, testLocale{"AkaraIntroGossip1": "90\nWelcome, traveler.\nStay awhile."})); err != nil {
		t.Fatal(err)
	}
	if err := runtime.RegisterModule(MapModule(catalog)); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = runtime.Stop(context.Background()) })
	if err := runtime.Execute(context.Background(), os.DirFS("."), "testdata/catalog_test.lua"); err != nil {
		t.Fatal(err)
	}
}

func TestParseDialogueTextRejectsMalformedPayloads(t *testing.T) {
	if rate, body, err := ParseDialogueText("120\r\nLine one\r\nLine two"); err != nil || rate != 2 || body != "Line one\nLine two" {
		t.Fatalf("dialog = %v, %q, %v", rate, body, err)
	}
	for _, value := range []string{"missing header", "fast\nText", "-1\nText", "60\n"} {
		if _, _, err := ParseDialogueText(value); err == nil {
			t.Errorf("ParseDialogueText(%q) succeeded", value)
		}
	}
}
