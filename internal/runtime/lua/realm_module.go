package modruntime

import (
	"errors"

	lua "github.com/yuin/gopher-lua"
)

// RealmController is the presentation-safe asynchronous Realm boundary. Login
// credentials cross it only as transient call arguments; copied status never
// contains passwords, bearer tokens, or transport handles.
type RealmController interface {
	Connect(string) error
	SetGateway(string) error
	Login(string, string) error
	Logout() error
	Signup(string, string, string) error
	RecoverPassword(string) error
	CreateCharacter(string, string, bool, bool) error
	DeleteCharacter(string) error
	SelectCharacter(string) error
	JoinChannel(string) error
	SendMessage(string) error
	Refresh() error
	SelectGame(string) error
	CreateGame(map[string]any) error
	JoinGame(string, string) error
	Cancel()
	Status() map[string]any
}

// RealmModule defines the realm module Lua boundary in one place so scripts receive a stable command and error
// contract.
func RealmModule(controller RealmController) Module {
	return Module{
		Name: "engine.realm/v1",
		Help: documentedModule(
			"Operate explicit Realm account, character, and lobby flows.",
			map[string]CommandHelp{
				"connect": commandHelp(
					"engine.realm.connect(endpoint)",
					"Prepare a TLS realm endpoint without opening gameplay transport.",
				),
				"set_gateway": commandHelp(
					"engine.realm.set_gateway(endpoint)",
					"Validate and persist the selected realm gateway.",
				),
				"login": commandHelp(
					"engine.realm.login(name, password)",
					"Log into the prepared Realm with explicit account credentials.",
				),
				"logout": commandHelp(
					"engine.realm.logout()",
					"Log out and remove the selected character from live Realm presence.",
				),
				"signup": commandHelp(
					"engine.realm.signup(name, email, password)",
					"Create an unverified Realm account and send its verification email.",
				),
				"recover_password": commandHelp(
					"engine.realm.recover_password(email)",
					"Send a browser password-reset link without revealing whether the email exists.",
				),
				"create_character": commandHelp(
					"engine.realm.create_character(name, class, expansion, hardcore)",
					"Create and select a realm-owned character.",
				),
				"delete_character": commandHelp(
					"engine.realm.delete_character(id)",
					"Delete an idle account-owned character.",
				),
				"select_character": commandHelp(
					"engine.realm.select_character(id)",
					"Select an account-owned character.",
				),
				"join_channel": commandHelp(
					"engine.realm.join_channel(name)",
					"Enter a public realm channel with the selected character.",
				),
				"send_message": commandHelp(
					"engine.realm.send_message(text)",
					"Send public channel chat or a slash command.",
				),
				"refresh": commandHelp(
					"engine.realm.refresh()",
					"Refresh channel members, chat, and public games.",
				),
				"select_game": commandHelp(
					"engine.realm.select_game(name_or_id)",
					"Load public details for a selected realm game.",
				),
				"create_game": commandHelp(
					"engine.realm.create_game(options)",
					"Create a uniquely named realm game.",
				),
				"join_game": commandHelp(
					"engine.realm.join_game(name_or_id, password)",
					"Resolve a named realm game for admission.",
				),
				"cancel": commandHelp(
					"engine.realm.cancel()",
					"Cancel the current realm request.",
				),
				"status": commandHelp("engine.realm.status()", "Return copied safe realm state."),
			},
		),
		Loader: func(state *lua.LState) int {
			accepted := func(operation func() error) int {
				state.Push(lua.LBool(controller != nil && operation() == nil))
				return 1
			}
			module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
				"connect": func(state *lua.LState) int {
					return accepted(
						func() error { return controller.Connect(state.CheckString(1)) },
					)
				},
				"set_gateway": func(state *lua.LState) int {
					return accepted(
						func() error { return controller.SetGateway(state.CheckString(1)) },
					)
				},
				"login": func(state *lua.LState) int {
					name, password := state.CheckString(1), state.CheckString(2)
					return accepted(func() error { return controller.Login(name, password) })
				},
				"logout": func(*lua.LState) int { return accepted(controller.Logout) },
				"signup": func(state *lua.LState) int {
					name, email, password := state.CheckString(
						1,
					), state.CheckString(
						2,
					), state.CheckString(
						3,
					)

					return accepted(
						func() error { return controller.Signup(name, email, password) },
					)
				},
				"recover_password": func(state *lua.LState) int {
					email := state.CheckString(1)
					return accepted(func() error { return controller.RecoverPassword(email) })
				},
				"create_character": func(state *lua.LState) int {
					name, class := state.CheckString(1), state.CheckString(2)
					expansion, hardcore := state.OptBool(3, true), state.OptBool(4, false)

					return accepted(
						func() error { return controller.CreateCharacter(name, class, expansion, hardcore) },
					)
				},
				"delete_character": func(state *lua.LState) int {
					id := state.CheckString(1)
					return accepted(func() error { return controller.DeleteCharacter(id) })
				},
				"select_character": func(state *lua.LState) int {
					id := state.CheckString(1)
					return accepted(func() error { return controller.SelectCharacter(id) })
				},
				"join_channel": func(state *lua.LState) int {
					channel := state.OptString(1, "Diablo II")
					return accepted(func() error { return controller.JoinChannel(channel) })
				},
				"send_message": func(state *lua.LState) int {
					message := state.CheckString(1)
					return accepted(func() error { return controller.SendMessage(message) })
				},
				"refresh": func(*lua.LState) int { return accepted(controller.Refresh) },
				"select_game": func(state *lua.LState) int {
					reference := state.CheckString(1)
					return accepted(func() error { return controller.SelectGame(reference) })
				},
				"create_game": func(state *lua.LState) int {
					options, err := luaTableToStringMap(state.CheckTable(1))
					if err != nil {
						state.RaiseError("realm create game options: %v", err)
						return 0
					}

					return accepted(func() error { return controller.CreateGame(options) })
				},
				"join_game": func(state *lua.LState) int {
					reference, password := state.CheckString(1), state.OptString(2, "")

					return accepted(
						func() error { return controller.JoinGame(reference, password) },
					)
				},
				"cancel": func(*lua.LState) int {
					if controller != nil {
						controller.Cancel()
					}

					return 0
				},
				"status": func(state *lua.LState) int {
					value := map[string]any{}
					if controller != nil {
						value = controller.Status()
					}

					copied, err := goToLua(state, value, 0)
					if err != nil {
						state.RaiseError("copy realm status: %v", err)
						return 0
					}

					state.Push(copied)

					return 1
				},
			})
			module.RawSetString("api", lua.LNumber(1))
			state.Push(module)

			return 1
		},
	}
}

// luaTableToStringMap owns the lua table to string map step at this boundary, keeping its side effects and failure
// point explicit to callers.
func luaTableToStringMap(table *lua.LTable) (map[string]any, error) {
	result := make(map[string]any)

	var conversionErr error

	table.ForEach(func(key, value lua.LValue) {
		if conversionErr != nil {
			return
		}

		name, ok := key.(lua.LString)
		if !ok {
			conversionErr = errors.New("keys must be strings")
			return
		}

		switch typed := value.(type) {
		case lua.LString:
			result[string(name)] = string(typed)
		case lua.LNumber:
			result[string(name)] = float64(typed)
		case lua.LBool:
			result[string(name)] = bool(typed)
		default:
			conversionErr = errors.New("values must be strings, numbers, or booleans")
		}
	})

	return result, conversionErr
}
