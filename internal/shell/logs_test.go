package shell

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestLogBufferIsBoundedAndTimelineIsChronological(t *testing.T) {
	logs := NewLogBuffer(2)
	base := time.Unix(100, 0)
	logs.Append(LogEntry{At: base, Level: "debug", Message: "discarded"})
	logs.Append(LogEntry{At: base.Add(2 * time.Second), Level: "warn", Message: "later"})
	logs.Append(LogEntry{At: base.Add(time.Second), Level: "info", Message: "middle", Attributes: map[string]any{"n": 2}})
	if got := logs.Snapshot(); len(got) != 2 {
		t.Fatalf("logs = %#v", got)
	}
	session, err := NewSession("test", "client", Policy{Name: "test"}, &fakeEvaluator{})
	if err != nil {
		t.Fatal(err)
	}
	session.AttachLogs(logs)
	session.Submit(context.Background(), "command")
	events := session.Timeline()
	if len(events) != 4 || !strings.Contains(events[0].Text, "middle") || !strings.Contains(events[1].Text, "later") {
		t.Fatalf("timeline = %#v", events)
	}
}
