package main

import (
	"strings"
)

/*
 *
 * author Guo Zhiqiang
 * datetime 2019/8/3 16:27
 */
func main() {
	//s := "varchar(32)"
	s := "varchar"
	n := strings.Index(s, "(")
	if n > 0 {
		s = s[0:n]
	}
	n = strings.Index(s, " ")
	if n > 0 {
		s = s[0:n]
	}

}
