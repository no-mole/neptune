package snowflake

import (
	"math/rand"
	"strconv"

	"github.com/bwmarrin/snowflake"
)

var node *snowflake.Node

func init() {
	temp, err := snowflake.NewNode(int64(rand.Intn(1023)))
	if err != nil {
		panic(err.Error())
	}
	node = temp
}

func GenStringBase64() string {
	return node.Generate().Base64()
}

func GenInt64String() string {
	return strconv.Itoa(int(node.Generate().Int64()))
}
