package main

import "strconv"

func timeFormat(duration int) string {
	return spaceSep(strconv.Itoa(duration))
}

func spaceSep(d string) string {
	l := len(d)
	if l < 4 {
		return d
	} else {
		return spaceSep(d[0:l-3]) + " " + d[l-3:l]
	}
}
