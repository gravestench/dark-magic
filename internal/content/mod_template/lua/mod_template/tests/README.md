# Tests

Keep focused `*_test.lua` files beside production modules. Use this directory
for shared semantic fixtures and genuinely cross-domain scenarios. A host must
provide a mod-specific suite discovery entry point, as d2legacy does in
`internal/mod/d2legacy/lua_suite_test.go`.
