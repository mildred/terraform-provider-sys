package sys

import (
	"os/exec"
	"bytes"
	"fmt"
)

func systemdDaemonReload() error {
	stderr := new(bytes.Buffer)
	cmd := exec.Command("systemctl", "daemon-reload")
	cmd.Stderr = stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Error running systemctl daemon-reload: %e\n%s", err, stderr.String())
	}
	return nil
}

func systemdIsActive(unit string) (bool, error) {
	stderr := new(bytes.Buffer)
	cmd := exec.Command("systemctl", "is-active", unit)
	cmd.Stderr = stderr
	err := cmd.Run()
	if err != nil {
		if er := err.(*exec.ExitError); er != nil {
			return false, nil
		}
		return false, fmt.Errorf("Error running systemctl is-active: %e\n%s", err, stderr.String())
	}
	return true, nil
}

func systemdIsEnabled(unit string) (bool, error) {
	stderr := new(bytes.Buffer)
	cmd := exec.Command("systemctl", "is-enabled", unit)
	cmd.Stderr = stderr
	err := cmd.Run()
	if err != nil {
		if er := err.(*exec.ExitError); er != nil {
			return false, nil
		}
		return false, fmt.Errorf("Error running systemctl is-active: %e\n%s", err, stderr.String())
	}
	return true, nil
}

func systemdStartStop(unit string, up, start, enable bool) error {
	stderr := new(bytes.Buffer)
	var cmd *exec.Cmd
	if up {
		if enable && start {
			cmd = exec.Command("systemctl", "enable", "--now", unit)
		} else if enable {
			cmd = exec.Command("systemctl", "enable", unit)
		} else if start {
			cmd = exec.Command("systemctl", "start", unit)
		}
	} else {
		if enable && start {
			cmd = exec.Command("systemctl", "disable", "--now", unit)
		} else if enable {
			cmd = exec.Command("systemctl", "disable", unit)
		} else if start {
			cmd = exec.Command("systemctl", "stop", unit)
		}
	}
	cmd.Stderr = stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Error running systemctl %v: %e\n%s", cmd.Args, err, stderr.String())
	}
	return nil
}
