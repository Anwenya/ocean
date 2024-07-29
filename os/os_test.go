package os

import (
	"os/exec"
	"testing"
	"time"
)

func TestGetCurrentProcessID(t *testing.T) {
	t.Log(GetCurrentProcessID())
}

func TestGetCurrentGoroutineID(t *testing.T) {
	t.Log(GetCurrentGoroutineID())
}

func TestGetCurrentProcessAndGoroutineIDStr(t *testing.T) {
	t.Log(GetCurrentProcessAndGoroutineIDStr())
}

func TestNetworkStatus(t *testing.T) {
	cmd := exec.Command("ping", "www.baidu.com", "-n", "1", "-w", "3")
	t.Log("network status start", time.Now().Unix())
	t.Log(cmd.String())
	err := cmd.Run()
	t.Log("network status end :", time.Now().Unix())
	if err != nil {
		t.Log("network status false :", err.Error())
	} else {
		t.Log("network status true")
	}
}
