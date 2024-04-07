package lua

import _ "embed"

//go:embed lock/lock.lua
var LuaLock string

//go:embed lock/unlock.lua
var LuaUnlock string

//go:embed lock/lease.lua
var LuaLease string
