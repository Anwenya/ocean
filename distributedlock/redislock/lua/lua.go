package lua

import _ "embed"

//go:embed lease.lua
var LuaCheckAndExpireDistributionLock string

//go:embed del.lua
var LuaCheckAndDeleteDistributionLock string
