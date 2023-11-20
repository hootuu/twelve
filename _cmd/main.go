package main

import (
	"fmt"
	"github.com/hootuu/tome/bk"
	"github.com/hootuu/tome/pr"
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
		chain  bk.Chain
		peers  int
		option *twelve.Option
	}
	ttfields := &fields{
		chain:  "test.chain.001",
		peers:  15,
		option: nil,
	}

	bus := &twelve.MemTwelveListenerBus{}
	readNode := twelve.NewMemTwelveNode(bus, &pr.Local{
		ID:  "node_001",
		PRI: "0xf07b88a2bba771b2b9d141589a8d179cff9ea4de257e55c833c6e7dfbe3deb27",
	})
	readTW, err := twelve.NewTwelve(ttfields.chain, readNode, ttfields.option)
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
		fmt.Println("NEW TWELVE NODe:", i)
		node := twelve.NewMemTwelveNode(bus, &pr.Local{
			ID:  fmt.Sprintf("w_node_%d", i),
			PRI: "0xf07b88a2bba771b2b9d141589a8d179cff9ea4de257e55c833c6e7dfbe3deb27",
		})

		fmt.Println("NEW TWELVE NODE OK:", i)
		wNodes = append(wNodes, node)

		fmt.Println("NEW TWELVE ...", i)
		tw, err := twelve.NewTwelve(ttfields.chain, node, ttfields.option)
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
			here, _ := readNode.Peer().Peer()
			thisPeer := twelve.PeerOf(here)
			rp := &twelve.RequestPayload{
				Action:    "incribe",
				Parameter: []byte(fmt.Sprintf("tx_%d", i)),
			}
			msg, err := twelve.NewMessage(twelve.RequestMessage, rp, thisPeer, readNode.Peer().PRI)
			if err != nil {
				fmt.Println(err)
				return
			}
			err = readNode.Notify(msg)
			if err != nil {
				fmt.Println(err)
				return
			}
			time.Sleep(1 * time.Second)
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
