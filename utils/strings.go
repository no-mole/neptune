package utils

import (
	"fmt"
	"strconv"
)

//string 转int，去掉前后的字符，返回中间的int，如果中间也有字符，则报错
func AToAbsI(s string) (int64, error) {
	b := []byte(s)
	for len(b) > 0 {
		if b[0] < '0' || b[0] > '9' {
			b = b[1:]
		} else {
			break
		}
	}
	for len(b) > 0 {
		last := len(b) - 1
		if b[last] < '0' || b[last] > '9' {
			b = b[:last]
		} else {
			break
		}
	}
	if len(b) == 0 {
		return 0, fmt.Errorf("can not covert string [%s] to int", s)
	}
	n, err := strconv.Atoi(string(b))
	return int64(n), err
}
