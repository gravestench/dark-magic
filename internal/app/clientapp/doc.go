// Package clientapp assembles the interactive Dark Magic client and owns the
// lifetime ordering between native backends, Lua components, sessions, and UI.
//
// A composition root is where independent parts are connected. This package
// describes those connections and their trust/ownership boundaries. Reusable
// game rules remain in domain packages so presentation cannot become authority.
package clientapp
