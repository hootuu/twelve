package test

import (
	"fmt"
	"github.com/hootuu/twelve"
	"github.com/hootuu/utils/errors"
	"github.com/rs/xid"
	"testing"
)

func TestExpect(t *testing.T) {
	type fields struct {
		Hash   string
		Expect int
	}
	tests := []struct {
		name   string
		fields fields
		suc    bool
	}{
		{
			name: "NORMAL",
			fields: fields{
				Hash:   xid.New().String(),
				Expect: 5,
			},
			suc: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := twelve.NewExpect(tt.fields.Hash, tt.fields.Expect)
			for i := 0; i < 10; i++ {
				idx := i
				go func() {
					e.Reply(fmt.Sprintf("peer_%d", idx))
				}()
			}
			result := 0
			suc := e.Waiting(func() *errors.Error {
				result = 100
				return nil
			}, func() {
				result += 99
			})
			if result != 199 {
				fmt.Println("result: ", suc)
				t.Fail()
			}
		})
	}
}
