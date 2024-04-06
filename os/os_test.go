package os

import "testing"

func TestGetCurrentProcessID(t *testing.T) {
	t.Log(GetCurrentProcessID())
}

func TestGetCurrentGoroutineID(t *testing.T) {
	t.Log(GetCurrentGoroutineID())
}

func TestGetCurrentProcessAndGoroutineIDStr(t *testing.T) {
	t.Log(GetCurrentProcessAndGoroutineIDStr())
}
