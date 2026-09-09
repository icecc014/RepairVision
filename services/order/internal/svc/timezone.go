package svc

import "time"

func init() {
	// MySQL 与业务均按 Asia/Shanghai；DATETIME 经 loc=Local 读取，time.Now() 写入需同一时区。
	time.Local = time.FixedZone("Asia/Shanghai", 8*60*60)
}
