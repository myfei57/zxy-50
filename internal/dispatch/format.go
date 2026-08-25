package dispatch

import "strconv"

func strconvInt(value int64) string {
	return strconv.FormatInt(value, 10)
}

func strconvFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', 2, 64)
}
