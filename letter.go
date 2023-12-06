package twelve

import (
	"encoding/json"
	"github.com/hootuu/tome/bk/bid"
	"github.com/hootuu/tome/ki"
	"github.com/hootuu/tome/kt"
	"github.com/hootuu/tome/nd"
	"github.com/hootuu/tome/vn"
	"github.com/hootuu/utils/errors"
)

const (
	ErrInvalidLetter = "Invalid Letter"
)

type Arrow int32

const (
	RequestArrow   Arrow = 1
	PrepareArrow   Arrow = 2
	CommittedArrow Arrow = 3
	ConfirmedArrow Arrow = 4
)

func ArrowVerify(t Arrow) *errors.Error {
	switch t {
	case RequestArrow, PrepareArrow, CommittedArrow, ConfirmedArrow:
		return nil
	}
	return errors.Verify("invalid letter.arrow")
}

const (
	LetterType    kt.Type = "HOTU.12.LETTER"
	LetterVersion         = kt.DefaultVersion
)

type Letter struct {
	Arrow      Arrow         `json:"a"`
	Type       kt.Type       `json:"t"`
	Version    kt.Version    `json:"v"`
	Vn         vn.ID         `json:"vn"`
	Chain      kt.Chain      `json:"c"`
	Invariable bid.BID       `json:"i"`
	From       nd.ID         `json:"f"`
	Signature  *kt.Signature `json:"s"`
}

func NewLetter(
	vnID vn.ID,
	chain kt.Chain,
	invID bid.BID,
	arrow Arrow,
	from nd.ID,
) *Letter {
	return &Letter{
		Arrow:      arrow,
		Type:       LetterType,
		Version:    LetterVersion,
		Vn:         vnID,
		Chain:      chain,
		Invariable: invID,
		From:       from,
		Signature:  nil,
	}
}

func (l *Letter) ToBytes() ([]byte, *errors.Error) {
	data, nErr := json.Marshal(l)
	if nErr != nil {
		return nil, errors.Sys("invalid letter, can not marshal: " + nErr.Error())
	}
	return data, nil
}

func LetterOfBytes(data []byte) (*Letter, *errors.Error) {
	var letter Letter
	nErr := json.Unmarshal(data, &letter)
	if nErr != nil {
		return nil, errors.Sys("invalid letter bytes, can not unmarshal: " + nErr.Error())
	}
	return &letter, nil
}

func (l *Letter) GetType() kt.Type {
	return l.Type
}

func (l *Letter) GetVersion() kt.Version {
	return l.Version
}

func (l *Letter) GetVN() vn.ID {
	return l.Vn
}

func (l *Letter) GetSignature() *kt.Signature {
	return l.Signature
}

func (l *Letter) SetSignature(signature *kt.Signature) {
	l.Signature = signature
}

func (l *Letter) Signing() *kt.Signing {
	return kt.NewSigning().
		Add("chain", l.Chain.S()).
		Add("from", l.From.S())
}

func (l *Letter) Sign(pri ki.PRI) *errors.Error {
	return kt.InvariableSign(l, pri)
}

func (l *Letter) Verify() *errors.Error {
	return kt.InvariableVerify(l)
}
