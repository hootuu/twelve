package main

import (
	"fmt"
	"github.com/hootuu/tome/bk/bid"
	"github.com/hootuu/tome/kt"
	"github.com/hootuu/tome/nd"
	"github.com/hootuu/tome/vn"
	"github.com/hootuu/twelve"
	"github.com/hootuu/utils/logger"
	"go.uber.org/zap"
	"time"
)

func main() {
	//for i := 0; i < 10000000; i++ {
	//	hash := "16ac36943ddfe46e25f11dcbe5b9ed78eefc72b41a0b27f5500ce5da974b0a29"
	//	hashInt := crc32.ChecksumIEEE([]byte(hash))
	//	idx := int(hashInt) % 100
	//	if idx != 93 {
	//		fmt.Println("ERROR != 93", idx)
	//	}
	//	time.Sleep(10 * time.Millisecond)
	//}
	//
	////sys.Info("CCollection ", hash, " ", idx)
	//i := 0
	//i += 1
	//if i > 0 {
	//	return
	//}
	type fields struct {
		vn     vn.ID
		chain  kt.Chain
		peers  int
		option *twelve.Option
	}
	ttfields := &fields{
		vn:     "testVN",
		chain:  "test.chain.001",
		peers:  15,
		option: nil,
	}

	bus := &twelve.MemTwelveListenerBus{}
	rn, _ := nd.NewNode("node_001", "0xf07b88a2bba771b2b9d141589a8d179cff9ea4de257e55c833c6e7dfbe3deb27")
	readNode := twelve.NewMemTwelveNode(bus, rn)
	readTW, err := twelve.NewTwelve(ttfields.vn, ttfields.chain, readNode, ttfields.option)
	if err != nil {
		fmt.Println(err)
		return
	}
	readTW.Start()
	fmt.Println("END: readTW.Start()")

	var wNodes []twelve.ITwelveNode
	wNodes = append(wNodes, readNode)
	//var oneWTW *twelve.Twelve
	for i := 0; i < ttfields.peers; i++ {
		fmt.Println("NEW TWELVE NODE:", i)
		nn, _ := nd.NewNode(nd.ID(fmt.Sprintf("w_node_%d", i)), "0xf07b88a2bba771b2b9d141589a8d179cff9ea4de257e55c833c6e7dfbe3deb27")
		node := twelve.NewMemTwelveNode(bus, nn)

		fmt.Println("NEW TWELVE NODE OK:", i)
		wNodes = append(wNodes, node)

		fmt.Println("NEW TWELVE ...", i)
		tw, err := twelve.NewTwelve(ttfields.vn, ttfields.chain, node, ttfields.option)
		if err != nil {
			fmt.Println(err)
			return
		}
		tw.Start()
		//oneWTW = tw
	}

	fmt.Println("END: ALL W NODE.Start()")
	go func() {

		for i := 0; i < 5000000; i++ {
			letter := twelve.NewLetter(
				ttfields.vn,
				ttfields.chain,
				bid.BID(fmt.Sprintf("tx_%d", i)),
				twelve.RequestArrow,
				readNode.Node().ID)
			err := letter.Sign(readNode.Node().PRI)
			if err != nil {
				fmt.Println(err)
				return
			}
			err = readNode.Notify(letter)
			if err != nil {
				fmt.Println(err)
				return
			}
			time.Sleep(50 * time.Millisecond)
		}

	}()
	chainLogger := logger.GetLogger("chainx")
	//queueLogger := logger.GetLogger("queue")
	for {
		chainLogger.Info("chainx new dump")
		//arr, lst, err := readTW.ImmutableList("", pagination.Limit(1000))
		//if err != nil {
		//	fmt.Println(err)
		//	time.Sleep(20 * time.Second)
		//	continue
		//}
		//chainLogger.Info("chainx summary ", zap.String("lst", lst),
		//	zap.Int("len", len(arr)))
		//
		//fmt.Println("lst: ", lst, " arr.length: ", len(arr))
		//fmt.Println("=====================================================")
		tx, err := readTW.Tail()
		if err != nil {
			chainLogger.Error("read tail err", zap.Error(err))
			time.Sleep(20 * time.Second)
			continue
		}
		chainLogger.Info("chainx tail ", zap.String("tail", tx.Hash),
			zap.String("pre", tx.Pre), zap.Int64("height", tx.Height))
		//for i, t := range arr {
		//	//fmt.Println(i, " ==> ", t.Hash)
		//	chainLogger.Info("chainx items ", zap.String("hash", t.Hash),
		//		zap.Int("idx", i), zap.Any("msg", t.Message), zap.String("pre", t.Pre))
		//}
		//fmt.Println("=====================================================")
		//
		//arr2, lst2, err := readTW.List("", pagination.Limit(1000))
		//if err != nil {
		//	fmt.Println(err)
		//	time.Sleep(20 * time.Second)
		//	continue
		//}
		//
		//sys.Error("chain.lst: ", lst, " chain..length: ", len(arr))
		//sys.Error("queue.lst: ", lst2, " queue..length: ", len(arr2))
		//
		//queueLogger.Info("queue queue summary ", zap.String("lst", lst2),
		//	zap.Int("len", len(arr2)))
		//fmt.Println("lst: ", lst2, " arr.length: ", len(arr2))
		//fmt.Println("###=====================================================")
		//for i, t := range arr2 {
		//	fmt.Println(i, " ###==> ", t.Hash, " state: ", t.State)
		//	queueLogger.Info("queue items ", zap.String("hash", t.Hash),
		//		zap.Int8("state", int8(t.State)),
		//		zap.Int("idx", i), zap.Any("msg", t.Message))
		//}
		//fmt.Println("###=====================================================")
		//if len(arr) == 10 {
		//	fmt.Println("OK")
		//	return
		//}
		time.Sleep(10 * time.Second)
	}
}
