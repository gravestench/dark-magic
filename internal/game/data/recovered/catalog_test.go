package recovered

import (
	"io/fs"
	"os"
	"strings"
	"testing"
	"testing/fstest"
)

func TestCatalogLoadsRecoveredShimTables(t *testing.T) {
	source := fstest.MapFS{
		QuestsPath:   &fstest.MapFile{Data: []byte("id\tname\tact\torder\tvisible\ticon\tquestdone\tqstr\tqsts1\tqsts1a\tqsts1b\n0\tPrologue\t0\t\t\t\t\tq0\ts0\t\t\n1\tFirst Quest\t0\t1\t1\ta1q1\t0\tq1\ts1\ts1a\ts1b\n")},
		SpeechPath:   &fstest.MapFile{Data: []byte("sound\tsoundstr\nakara_intro\tAkaraIntroGossip1\n")},
		DS1TypesPath: &fstest.MapFile{Data: []byte("Name\tDef\tLevelType\nTown\t1\t1\n")},
		ObjectsPath:  &fstest.MapFile{Data: []byte("Act\tId\tDescription\tObjectId\n1\t0\tFountain\t12\n")},
	}
	catalog := New(source)
	snapshot, err := catalog.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	quest := snapshot.QuestsByID[1]
	if quest.PrerequisiteID == nil || *quest.PrerequisiteID != 0 || len(quest.Stages) != 1 || len(quest.Stages[0].Alternates) != 2 {
		t.Fatalf("quest = %#v", quest)
	}
	if got := snapshot.SpeechByName["akara_intro"].StringKey; got != "AkaraIntroGossip1" {
		t.Fatalf("speech key = %q", got)
	}
	if snapshot.DS1TypeByDef[1].Name != "Town" || snapshot.MapObjectByActID["1:0"].ObjectID != 12 {
		t.Fatalf("map recovery = %#v, %#v", snapshot.DS1TypeByDef, snapshot.MapObjectByActID)
	}
	snapshot.Quests[1].Stages[0].Alternates[0] = "changed"
	again, err := catalog.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if again.Quests[1].Stages[0].Alternates[0] != "s1a" {
		t.Fatal("snapshot leaked mutable stage storage")
	}
}

func TestParsersRejectBrokenRelationships(t *testing.T) {
	_, err := ParseQuests(strings.NewReader("id\tname\tact\tquestdone\tqstr\n1\tQuest\t0\t9\tq1\n"))
	if err == nil || !strings.Contains(err.Error(), "unknown prerequisite") {
		t.Fatalf("prerequisite error = %v", err)
	}
	_, err = ParseSpeech(strings.NewReader("sound\tsoundstr\nvoice\tOne\nVOICE\tTwo\n"))
	if err == nil || !strings.Contains(err.Error(), "duplicate sound") {
		t.Fatalf("duplicate error = %v", err)
	}
}

func TestReferenceValidationReportsSoundAndStringJoins(t *testing.T) {
	snapshot := Snapshot{
		Speech: []Speech{{Sound: "voice", StringKey: "VoiceKey"}},
		Quests: []Quest{{TitleStringKey: "QuestKey", Stages: []QuestStage{{StringKey: "MissingKey"}}}},
	}
	issues := ValidateReferences(snapshot, map[string]struct{}{"voice": {}}, func(key string) (string, error) {
		if key == "MissingKey" {
			return key, fs.ErrNotExist
		}
		return key, nil
	})
	if len(issues) != 1 || issues[0].Kind != "string" || issues[0].Identifier != "MissingKey" {
		t.Fatalf("issues = %#v", issues)
	}
}

func TestRepositoryD2LegacyDataParses(t *testing.T) {
	source := repositoryD2Legacy(t)
	snapshot, err := New(source).Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Quests) != 31 || len(snapshot.Speech) != 185 || len(snapshot.DS1Types) != 1091 || len(snapshot.MapObjects) != 750 {
		t.Fatalf("counts = %d quests, %d speech, %d DS1 types, %d map objects", len(snapshot.Quests), len(snapshot.Speech), len(snapshot.DS1Types), len(snapshot.MapObjects))
	}
}

func repositoryD2Legacy(t *testing.T) fs.FS {
	t.Helper()
	return os.DirFS("../../../content/d2legacy")
}
