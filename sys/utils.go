package sys

import (
	"strconv"
)

func parseBoolDef(val interface{}, def bool) bool {
	s, ok := val.(string)
	if ok {
		res, err := strconv.ParseBool(s)
		if err != nil {
			return res
		}
	}
	return def
}
