// Code generated by "stringer -type=ElementType"; DO NOT EDIT.

package pbf

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[NODE-0]
	_ = x[WAY-1]
	_ = x[RELATION-2]
}

const _ElementType_name = "NODEWAYRELATION"

var _ElementType_index = [...]uint8{0, 4, 7, 15}

func (i ElementType) String() string {
	if i < 0 || i >= ElementType(len(_ElementType_index)-1) {
		return "ElementType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _ElementType_name[_ElementType_index[i]:_ElementType_index[i+1]]
}
