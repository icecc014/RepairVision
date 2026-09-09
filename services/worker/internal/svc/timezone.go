package svc

import "time"

func init() {
	time.Local = time.FixedZone("Asia/Shanghai", 8*60*60)
}
