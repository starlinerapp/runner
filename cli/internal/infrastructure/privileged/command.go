package privileged

import (
	"os"
	"os/exec"
)

func Run(name string, args ...string) error {
	cmd := Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func Command(name string, args ...string) *exec.Cmd {
	if os.Geteuid() == 0 {
		return exec.Command(name, args...)
	}
	return exec.Command("sudo", append([]string{name}, args...)...)
}
