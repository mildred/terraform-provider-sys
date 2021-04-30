// +build !windows

package file_getter

import (
	"os"
)

var ErrUnauthorized = os.ErrPermission
var SymlinkAny = os.Symlink
