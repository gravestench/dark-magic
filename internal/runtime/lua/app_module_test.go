package modruntime

import (
	"context"
	"testing"
	"testing/fstest"
)

func TestAppModuleReportsVersionAndRequestsExit(t *testing.T) {
	exited := false
	runtime := New()
	if err := runtime.RegisterModule(AppModule("test-version", func() { exited = true })); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())
	if err := runtime.Execute(context.Background(), fstest.MapFS{"test.lua": &fstest.MapFile{Data: []byte(`local app=require("engine.app/v1"); assert(app.version()=="test-version"); app.request_exit()`)}}, "test.lua"); err != nil {
		t.Fatal(err)
	}
	if !exited {
		t.Fatal("exit was not requested")
	}
}
