package idgen

import (
	"log"
	"time"

	"github.com/bwmarrin/snowflake"
)

var node *snowflake.Node

func Init(startTime string, machineID int64) {
	var st time.Time
	var err error
	// 格式化 1月2号下午3时4分5秒  2006年
	st, err = time.Parse("2006-01-02", startTime)
	if err != nil {
		log.Fatalln("fails to parse time:", err)
		return
	}

	snowflake.Epoch = st.UnixNano() / 1e6
	node, err = snowflake.NewNode(machineID)
	if err != nil {
		log.Fatalln("fails to create node:", err)
		return
	}
}

func New() int64 {
	return node.Generate().Int64()
}
