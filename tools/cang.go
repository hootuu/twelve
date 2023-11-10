package main

import (
	"fmt"
	"github.com/hootuu/twelve"
	"github.com/hootuu/utils/logger"
	"go.uber.org/zap"
	"os"
)

func main() {
	args := os.Args[1:] // 获取除了程序名称之外的所有命令行参数
	peerID := args[0]
	name := args[1]
	path := ".twelve/" + peerID + "/"
	hash := args[2]
	fmt.Println("name: ", name)
	fmt.Println("path: ", path)
	fmt.Println("hash: ", hash)
	txCang, err := twelve.NewTxCang(name, path)
	if err != nil {
		fmt.Println("New Tx Cang Failed: ", err)
		return
	}
	var iTx twelve.ImmutableTx
	err = txCang.CCollection(hash).MustGet(hash, &iTx)
	if err != nil {
		fmt.Println("Get Tx Failed: ", err)
		return
	}
	logger.Logger.Info("iTx", zap.Any("iTx", iTx))
}
