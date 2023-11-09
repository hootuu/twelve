package twelve

import (
	"fmt"
	"github.com/hootuu/cang"
	"github.com/hootuu/utils/errors"
	"hash/crc32"
)

const TxCangCount = 100

type TxCang struct {
	cang  *cang.Cang
	qColl [TxCangCount]*cang.Collection // queue collection
	cColl [TxCangCount]*cang.Collection // chain collection
}

func NewTxCang(name string, path string) (*TxCang, *errors.Error) {
	txCang := &TxCang{}
	var err *errors.Error
	txCang.cang, err = cang.NewCang(name, path)
	if err != nil {
		return nil, err
	}
	for i := 0; i < TxCangCount; i++ {
		txCang.qColl[i], err = txCang.cang.Collection(fmt.Sprintf("Q_%s_%d", name, i+1))
		if err != nil {
			return nil, err
		}
		txCang.cColl[i], err = txCang.cang.Collection(fmt.Sprintf("C_%s_%d", name, i+1))
		if err != nil {
			return nil, err
		}
	}
	return txCang, nil
}

// QCollection Get Queue Collection
func (tc *TxCang) QCollection(hash string) *cang.Collection {
	hashInt := crc32.ChecksumIEEE([]byte(hash))
	idx := int(hashInt) % TxCangCount
	return tc.qColl[idx]
}

// CCollection Get Chain Collection
func (tc *TxCang) CCollection(hash string) *cang.Collection {
	hashInt := crc32.ChecksumIEEE([]byte(hash))
	idx := int(hashInt) % TxCangCount
	return tc.cColl[idx]
}
