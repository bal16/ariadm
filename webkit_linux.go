//go:build linux

package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// cleanupWebviewProcesses terminates WebKitGTK helper processes that survive
// window destruction, so background mode actually frees their memory.
func cleanupWebviewProcesses() {
	killWebkitChildren()
	time.Sleep(200 * time.Millisecond)
}

// killWebkitChildren terminates WebKitGTK helper processes that survive
// window destruction.
func killWebkitChildren() {
	pid := os.Getpid()
	// Collect all descendant PIDs.
	var collect func(int) []int
	collect = func(parent int) []int {
		var out []int
		entries, _ := os.ReadDir("/proc")
		for _, e := range entries {
			id, err := strconv.Atoi(e.Name())
			if err != nil {
				continue
			}
			data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", id))
			if err != nil {
				continue
			}
			fields := strings.Fields(string(data))
			if len(fields) < 4 {
				continue
			}
			ppid, err := strconv.Atoi(fields[3])
			if err != nil || ppid != parent {
				continue
			}
			out = append(out, id)
			out = append(out, collect(id)...)
		}
		return out
	}

	for _, id := range collect(pid) {
		comm, _ := os.ReadFile(fmt.Sprintf("/proc/%d/comm", id))
		name := strings.TrimSpace(string(comm))
		if strings.Contains(name, "WebKit") {
			if process, err := os.FindProcess(id); err == nil {
				_ = process.Kill()
			}
		}
	}
}
