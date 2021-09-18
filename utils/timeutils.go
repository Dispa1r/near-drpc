package utils

import "time"

func CurrentTimestamp() int64 {
	return  time.Now().Unix()
}

func CurrentTimestampMilli() int64 {
	return  time.Now().UnixNano() /int64(time.Millisecond)
}
