package sys

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	hclog "github.com/hashicorp/go-hclog"
)

func systemdDaemonReload(log hclog.Logger) error {
	stderr := new(bytes.Buffer)
	log.Trace("systemctl daemon-reload")
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

func systemdCommand(log hclog.Logger, unit string, action string, now bool) error {
	stderr := new(bytes.Buffer)
	var cmd *exec.Cmd
	if now {
		log.Trace("systemctl %s --now %s", action, unit)
		cmd = exec.Command("systemctl", action, "--now", unit)
	} else {
		log.Trace("systemctl %s %s", action, unit)
		cmd = exec.Command("systemctl", action, unit)
	}
	cmd.Stderr = stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Error running systemctl %v: %e\n%s", cmd.Args, err, stderr.String())
	}
	return nil
}

func systemdEnable(log hclog.Logger, unit string, now bool) error {
	return systemdCommand(log, unit, "enable", now)
}

func systemdDisable(log hclog.Logger, unit string, now bool) error {
	return systemdCommand(log, unit, "disable", now)
}

func systemdStart(log hclog.Logger, unit string) error {
	return systemdCommand(log, unit, "start", false)
}

func systemdStop(log hclog.Logger, unit string) error {
	return systemdCommand(log, unit, "stop", false)
}
func systemdStartStopEnableDisable(log hclog.Logger, unit string, start, stop, enable, disable bool) error {
	if (start && stop) || (enable && disable) {
		return fmt.Errorf("Internal error, requesting conflicting orders start=%b stop=%b enable=%b disable=%b", start, stop, enable, disable)
	}

	if start && enable {
		return systemdEnable(log, unit, true)
	} else if start && disable {
		err := systemdDisable(log, unit, false)
		if err != nil {
			return err
		}
		return systemdStart(log, unit)
	} else if start {
		return systemdStart(log, unit)
	} else if stop && enable {
		err := systemdEnable(log, unit, false)
		if err != nil {
			return err
		}
		return systemdStop(log, unit)
	} else if stop && disable {
		return systemdDisable(log, unit, true)
	} else if stop {
		return systemdStop(log, unit)
	} else if enable {
		return systemdEnable(log, unit, false)
	} else if disable {
		return systemdDisable(log, unit, false)
	}
	return nil
}

