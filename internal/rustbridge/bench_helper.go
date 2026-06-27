package rustbridge

import "time"

func timeNowNs() int64 {
	return time.Now().UnixNano()
}
