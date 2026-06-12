package vm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"starliner.app/runner/internal/infrastructure/firecracker/config"
)

func PrintDiagnostics(vmDir, guestIP string) {
	fmt.Fprintf(os.Stderr, "\n--- VM diagnostics (%s) ---\n", vmDir)
	for _, name := range []string{config.LogFileName, config.SerialLogName, config.FileName} {
		tailFile(filepath.Join(vmDir, name))
	}
	fmt.Fprintf(os.Stderr, "Check from host: ping %s && nc -zv %s 1234\n", guestIP, guestIP)
}

func tailFile(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n[%s] (not found)\n", filepath.Base(path))
		return
	}

	text := strings.TrimSpace(string(data))
	if text == "" {
		fmt.Fprintf(os.Stderr, "\n[%s] (empty)\n", filepath.Base(path))
		return
	}

	lines := strings.Split(text, "\n")
	const maxLines = 40
	if len(lines) > maxLines {
		lines = lines[len(lines)-maxLines:]
	}

	fmt.Fprintf(os.Stderr, "\n[%s] (last %d lines)\n%s\n", filepath.Base(path), len(lines), strings.Join(lines, "\n"))
}
