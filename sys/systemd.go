package sys

import (
	"os/exec"
	"bytes"
	"strings"
	"bufio"
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

func systemdIsExists(unit string) (bool, error) {
	stderr := new(bytes.Buffer)
	stdout := new(bytes.Buffer)
	cmd := exec.Command("systemctl", "list-unit-files", unit)
	cmd.Stderr = stderr
	cmd.Stdout = stdout
	err := cmd.Run()
	if err != nil {
		return false, fmt.Errorf("Error running systemctl is-active: %e\n%s", err, stderr.String())
	}

	r := bufio.NewScanner(stdout)
	r.Scan() // discard first line
	r.Scan() // read second line
	line := r.Text()
	return strings.HasPrefix(line, unit), nil
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

func systemdUpdnStartEnable(unit string, up, start, enable bool) error {
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

func systemdStartStopEnableDisable(unit string, start, stop, enable, disable bool) error {
	if (start && stop) || (enable && disable) {
		return fmt.Errorf("Internal error, requesting conflicting orders start=%b stop=%b enable=%b disable=%b", start, stop, enable, disable)
	}

        if enable && start {
                return systemdUpdnStartEnable(unit, true, true, true)
        } else if enable && stop {
                err := systemdUpdnStartEnable(unit, true, false, true)
                if err != nil {
                        return err
                }
                return systemdUpdnStartEnable(unit, false, true, false)
        } else if enable {
                return systemdUpdnStartEnable(unit, true, true, true)
        } else if disable && stop {
                return systemdUpdnStartEnable(unit, false, true, true)
        } else if disable && start {
                err := systemdUpdnStartEnable(unit, false, false, true)
                if err != nil {
                        return err
                }
                return systemdUpdnStartEnable(unit, true, true, false)
        } else if disable {
                return systemdUpdnStartEnable(unit, false, false, true)
        } else if start {
                return systemdUpdnStartEnable(unit, true, true, false)
        } else if stop {
                return systemdUpdnStartEnable(unit, false, true, false)
        }
        return nil
}
