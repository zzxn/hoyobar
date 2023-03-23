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
	st, err = time.ParseInLocation("2006-01-02", startTime, time.UTC)
	if err != nil {
		log.Fatalln("fails to parse time:", err)
		return
	}

	snowflake.Epoch = st.UnixMilli()
	node, err = snowflake.NewNode(machineID)
	if err != nil {
		log.Fatalln("fails to create node:", err)
		return
	}
}

func New() int64 {
	return node.Generate().Int64()
}
