package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

var Umask os.FileMode = FileModeMustGetUmask()

func FileModeDecode(perm string) (os.FileMode, error) {
	modeInt, err := strconv.ParseInt(perm, 8, 64)
	if err != nil {
		err = fmt.Errorf("cannot decode %v, %v", perm, err)
	}
	return os.FileMode(modeInt), err
}

func FileModeMustDecode(perm string) os.FileMode {
	mode, err := FileModeDecode(perm)
	if err != nil {
		panic(err)
	}
	return mode
}

func FileModeGetUmask() (os.FileMode, error) {
	data, err := ioutil.ReadFile("/proc/self/status")
	if err != nil {
		return os.FileMode(0), err
	}
	for _, line := range strings.Split(string(data), "\n") {
		elems := strings.Split(line, ":")
		if elems[0] == "Umask" {
			return FileModeDecode(strings.TrimSpace(elems[1]))
		}
	}
	return os.FileMode(0), fmt.Errorf("cannot find Umask in /proc/self/status")
}
func FileModeMustGetUmask() os.FileMode {
	umask, err := FileModeGetUmask()
	if err != nil {
		panic(err)
	}
	return umask
}

func FileModeApplyUmask(mode, umask os.FileMode) os.FileMode {
	return mode &^ umask
}

func FileModeSame(mode1 string, mode2 os.FileMode, umask os.FileMode) (bool, error) {
	var m1, m2 os.FileMode

	m1, err := FileModeDecode(mode1)
	if err != nil {
		return false, err
	}

	m1 = FileModeApplyUmask(m1, umask)
	m2 = FileModeApplyUmask(m1, umask)
	return m1 == m2, nil
}
