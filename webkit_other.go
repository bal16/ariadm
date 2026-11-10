//go:build !linux

package main

// cleanupWebviewProcesses is a no-op outside Linux, where destroying the
// window also releases its webview processes.
func cleanupWebviewProcesses() {}
