package test

import (
	"fmt"
	"github.com/hootuu/tome/bk"
	"github.com/hootuu/tome/pr"
	"github.com/hootuu/twelve"
	"github.com/hootuu/utils/types/pagination"
	"testing"
	"time"
)

func TestTwelve(t *testing.T) {
	type fields struct {
		chain  bk.Chain
		peers  int
		option *twelve.Option
	}
	tests := []struct {
		name   string
		fields fields
		suc    bool
	}{
		{
			name: "NORMAL",
			fields: fields{
				chain:  "test.chain.001",
				peers:  15,
				option: nil,
			},
			suc: true,
		},
	}
	bus := &twelve.MemTwelveListenerBus{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			readNode := twelve.NewMemTwelveNode(bus, &pr.Local{
				ID:  "node_001",
				PRI: "0xf07b88a2bba771b2b9d141589a8d179cff9ea4de257e55c833c6e7dfbe3deb27",
			})
			readTW, err := twelve.NewTwelve(tt.fields.chain, readNode, tt.fields.option)
			if err != nil {
				fmt.Println(err)
				t.Fail()
				return
			}
			readTW.Start()
			fmt.Println("END: readTW.Start()")

			var wNodes []twelve.ITwelveNode
			wNodes = append(wNodes, readNode)
			for i := 0; i < tt.fields.peers; i++ {
				fmt.Println("NEW TWELVE NODe:", i)
				node := twelve.NewMemTwelveNode(bus, &pr.Local{
					ID:  fmt.Sprintf("w_node_%d", i),
					PRI: "0xf07b88a2bba771b2b9d141589a8d179cff9ea4de257e55c833c6e7dfbe3deb27",
				})

				fmt.Println("NEW TWELVE NODE OK:", i)
				wNodes = append(wNodes, node)

				fmt.Println("NEW TWELVE ...", i)
				tw, err := twelve.NewTwelve(tt.fields.chain, node, tt.fields.option)
				if err != nil {
					fmt.Println(err)
					t.Fail()
					return
				}
				tw.Start()
			}

			fmt.Println("END: ALL W NODE.Start()")
			go func() {

				for i := 0; i < 50000; i++ {
					here, _ := readNode.Peer().Peer()
					thisPeer := twelve.PeerOf(here)
					rp := &twelve.RequestPayload{
						Action:    "incribe",
						Parameter: []byte(fmt.Sprintf("tx_%d", i)),
					}
					msg, err := twelve.NewMessage(twelve.RequestMessage, rp, thisPeer, readNode.Peer().PRI)
					if err != nil {
						fmt.Println(err)
						t.Fail()
						return
					}
					err = readNode.Notify(msg)
					if err != nil {
						fmt.Println(err)
						t.Fail()
						return
					}
					time.Sleep(300 * time.Millisecond)
				}

			}()
			for {
				arr, lst, err := readTW.ImmutableList("", pagination.Limit(100))
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Println("lst: ", lst, " arr.length: ", len(arr))
				fmt.Println("=====================================================")
				for i, t := range arr {
					fmt.Println(i, " ==> ", t.Hash)
				}
				fmt.Println("=====================================================")

				arr2, lst2, err := readTW.List("", pagination.Limit(100))
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Println("lst: ", lst2, " arr.length: ", len(arr2))
				fmt.Println("###=====================================================")
				for i, t := range arr2 {
					fmt.Println(i, " ###==> ", t.Hash, " state: ", t.State)
				}
				fmt.Println("###=====================================================")
				//if len(arr) == 10 {
				//	fmt.Println("OK")
				//	return
				//}
				time.Sleep(10 * time.Second)
			}

		})
	}
}
