package gnugrep_bre

import (
	"fmt"
	"strconv"
)

// parseInt converts a PEG match result to an integer.
// The input is a slice of any containing byte slices.
func parseInt(match any) int {
	var s string
	switch v := match.(type) {
	case []byte:
		s = string(v)
	case []any:
		for _, b := range v {
			s += string(b.([]byte))
		}
	default:
		panic(fmt.Sprintf("parseInt: unexpected type %T", match))
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		panic(fmt.Sprintf("parseInt: invalid integer %q", s))
	}
	return n
}
