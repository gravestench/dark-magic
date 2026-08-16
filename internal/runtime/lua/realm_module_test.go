package modruntime

import (
	"context"
	"testing"
	"testing/fstest"
)

type testRealmController struct {
	calls   []string
	options map[string]any
}

func (controller *testRealmController) Connect(value string) error {
	controller.calls = append(controller.calls, "connect:"+value)
	return nil
}
func (controller *testRealmController) SetGateway(value string) error {
	controller.calls = append(controller.calls, "gateway:"+value)
	return nil
}
func (controller *testRealmController) Login(name, password string) error {
	controller.calls = append(controller.calls, "login:"+name+":"+password)
	return nil
}
func (controller *testRealmController) Logout() error {
	controller.calls = append(controller.calls, "logout")
	return nil
}
func (controller *testRealmController) Signup(name, email, password string) error {
	controller.calls = append(controller.calls, "signup:"+name+":"+email+":"+password)
	return nil
}
func (controller *testRealmController) RecoverPassword(email string) error {
	controller.calls = append(controller.calls, "recover:"+email)
	return nil
}
func (controller *testRealmController) CreateCharacter(name, class string, expansion, hardcore bool) error {
	controller.calls = append(controller.calls, "character:"+name+":"+class)
	return nil
}
func (controller *testRealmController) DeleteCharacter(id string) error {
	controller.calls = append(controller.calls, "delete:"+id)
	return nil
}
func (controller *testRealmController) SelectCharacter(id string) error {
	controller.calls = append(controller.calls, "select:"+id)
	return nil
}
func (controller *testRealmController) JoinChannel(name string) error {
	controller.calls = append(controller.calls, "channel:"+name)
	return nil
}
func (controller *testRealmController) SendMessage(value string) error {
	controller.calls = append(controller.calls, "message:"+value)
	return nil
}
func (controller *testRealmController) Refresh() error {
	controller.calls = append(controller.calls, "refresh")
	return nil
}
func (controller *testRealmController) SelectGame(reference string) error {
	controller.calls = append(controller.calls, "detail:"+reference)
	return nil
}
func (controller *testRealmController) CreateGame(options map[string]any) error {
	controller.options = options
	return nil
}
func (controller *testRealmController) JoinGame(name, password string) error {
	controller.calls = append(controller.calls, "game:"+name+":"+password)
	return nil
}
func (controller *testRealmController) Cancel() {
	controller.calls = append(controller.calls, "cancel")
}
func (controller *testRealmController) Status() map[string]any {
	return map[string]any{"phase": "lobby", "members": []any{"Hero"}}
}

func TestRealmModuleTransportsIntentsAndCopiesSafeStatus(t *testing.T) {
	controller := &testRealmController{}
	runtime := New()
	if err := runtime.RegisterModule(RealmModule(controller)); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(t.Context())
	script := `
local realm = require("engine.realm/v1")
assert(realm.set_gateway("realm.example") and realm.connect(""))
assert(realm.login("Alice", "password"))
assert(realm.signup("Bob", "bob@example.test", "password"))
assert(realm.recover_password("alice@example.test"))
assert(realm.create_character("Hero", "Amazon", true, false) and realm.select_character("hero"))
assert(realm.delete_character("retired"))
assert(realm.join_channel("Diablo II") and realm.send_message("hello") and realm.refresh())
assert(realm.select_game("Fresh"))
assert(realm.create_game({name="Fresh", difficulty="normal", maximum_players=8, expansion=true}))
assert(realm.join_game("Fresh", "") and realm.status().phase == "lobby")
assert(realm.logout())
realm.cancel()
`
	if err := runtime.Execute(context.Background(), fstest.MapFS{"test.lua": &fstest.MapFile{Data: []byte(script)}}, "test.lua"); err != nil {
		t.Fatal(err)
	}
	if len(controller.calls) != 15 || controller.options["name"] != "Fresh" || controller.options["maximum_players"] != float64(8) {
		t.Fatalf("calls=%v options=%#v", controller.calls, controller.options)
	}
}
