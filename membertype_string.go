// Code generated by "stringer -type=MemberType"; DO NOT EDIT.

package pbfparser

import "fmt"

const _MemberType_name = "NODEWAYRELATION"

var _MemberType_index = [...]uint8{0, 4, 7, 15}

func (i MemberType) String() string {
	if i < 0 || i >= MemberType(len(_MemberType_index)-1) {
		return fmt.Sprintf("MemberType(%d)", i)
	}
	return _MemberType_name[_MemberType_index[i]:_MemberType_index[i+1]]
}
