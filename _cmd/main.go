package main

import (
	"fmt"
	"github.com/hootuu/tome/bk"
	"github.com/hootuu/tome/pr"
	"github.com/hootuu/twelve"
	"github.com/hootuu/utils/logger"
	"github.com/hootuu/utils/types/pagination"
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
			time.Sleep(800 * time.Millisecond)
		}

	}()
	chainLogger := logger.GetLogger("chainx")
	for {
		arr, lst, err := readTW.ImmutableList("", pagination.Limit(100))
		if err != nil {
			fmt.Println(err)
			time.Sleep(20 * time.Second)
			continue
		}
		chainLogger.Info("chainx summary ", zap.String("lst", lst),
			zap.Int("len", len(arr)))

		fmt.Println("lst: ", lst, " arr.length: ", len(arr))
		fmt.Println("=====================================================")

		for i, t := range arr {
			fmt.Println(i, " ==> ", t.Hash)
			chainLogger.Info("chainx items ", zap.String("hash", t.Hash),
				zap.Int("idx", i))
		}
		fmt.Println("=====================================================")

		arr2, lst2, err := readTW.List("", pagination.Limit(100))
		if err != nil {
			fmt.Println(err)
			time.Sleep(20 * time.Second)
			continue
		}

		chainLogger.Info("chainx queue summary ", zap.String("lst", lst2),
			zap.Int("len", len(arr2)))
		fmt.Println("lst: ", lst2, " arr.length: ", len(arr2))
		fmt.Println("###=====================================================")
		for i, t := range arr2 {
			fmt.Println(i, " ###==> ", t.Hash, " state: ", t.State)
			chainLogger.Info("chainx items ", zap.String("hash", t.Hash),
				zap.Int("idx", i))
		}
		fmt.Println("###=====================================================")
		//if len(arr) == 10 {
		//	fmt.Println("OK")
		//	return
		//}
		time.Sleep(10 * time.Second)
	}
}
