package twelve

import (
	"encoding/json"
	"github.com/hootuu/tome/bk/bid"
	"github.com/hootuu/tome/ki"
	"github.com/hootuu/tome/kt"
	"github.com/hootuu/tome/nd"
	"github.com/hootuu/utils/errors"
	"go.uber.org/zap"
)

const (
	ErrInvalidLetter = "Invalid Letter"
)

type Arrow int32

const (
	RequestArrow    Arrow = 1
	PrepareArrow    Arrow = 2
	CommittedArrow  Arrow = 3
	ConfirmedArrow  Arrow = 4
	InvariableArrow Arrow = 999999999
)

func (a Arrow) S() string {
	str := "?"
	switch a {
	case RequestArrow:
		str = "REQUEST"
	case PrepareArrow:
		str = "PREPARE"
	case CommittedArrow:
		str = "COMMITTED"
	case ConfirmedArrow:
		str = "CONFIRMED"
	case InvariableArrow:
		str = "INVARIABLE"
	}
	return str
}

func ArrowVerify(t Arrow) *errors.Error {
	switch t {
	case RequestArrow, PrepareArrow, CommittedArrow, ConfirmedArrow, InvariableArrow:
		return nil
	}
	return errors.Verify("invalid letter.arrow")
}

type ConsensusPayload struct {
	Tx         kt.Hash `json:"tx"`
	Invariable bid.BID `json:"i"`
	Pre        kt.KID  `json:"pre"`
}

type Lock = kt.KID

const (
	UnLock Lock = "?"
)

const (
	LetterType    kt.Type = "HOTU.12.LETTER"
	LetterVersion         = kt.DefaultVersion
)

type Letter struct {
	Hash      kt.Hash       `json:"h"`
	Arrow     Arrow         `json:"a"`
	Type      kt.Type       `json:"t"`
	Version   kt.Version    `json:"v"`
	Chain     *kt.Chain     `json:"c"`
	Payload   []byte        `json:"p"`
	From      nd.ID         `json:"f"`
	Signature *kt.Signature `json:"s"`
}

func NewLetter(chain *kt.Chain, arrow Arrow) *Letter {
	l := &Letter{
		Arrow:   arrow,
		Type:    LetterType,
		Version: LetterVersion,
		Chain:   chain,
		From:    nd.Here().ID,
	}
	return l
}

func (l *Letter) WithTx(tx *kt.Tx) *errors.Error {
	var nErr error
	l.Payload, nErr = json.Marshal(tx)
	if nErr != nil {
		gLogger.Error("WithTx.json.Marshal(tx) Error", zap.Error(nErr))
		return errors.Sys("WithTx.json.Marshal(tx) Error: " + nErr.Error())
	}
	return nil
}

func (l *Letter) WithConsensus(cp *ConsensusPayload) *errors.Error {
	var nErr error
	l.Payload, nErr = json.Marshal(cp)
	if nErr != nil {
		gLogger.Error("WithTx.json.Marshal(cp) Error", zap.Error(nErr))
		return errors.Sys("WithTx.json.Marshal(cp) Error: " + nErr.Error())
	}
	return nil
}

func (l *Letter) Sign(pri ki.PRI) *errors.Error {
	var err *errors.Error
	l.Signature = kt.NewSignature()
	l.Hash, err = l.Signature.Sign(pri, l.doInjectSigning)
	if err != nil {
		return err
	}
	return nil
}

func (l *Letter) doInjectSigning(builder *kt.Signing) {
	builder.Add("Chain", l.Chain.S())
	builder.Add("Arrow", l.Arrow.S())
	builder.Add("From", l.From.S())
	builder.Add("Payload", string(l.Payload))
}

func (l *Letter) GetTx() (*kt.Tx, *errors.Error) {
	var txM kt.Tx
	nErr := json.Unmarshal(l.Payload, &txM)
	if nErr != nil {
		gLogger.Error("json.Unmarshal(l.Payload, &txM) err", zap.Error(nErr))
		return nil, errors.Sys("Letter.GetTx Error: " + nErr.Error())
	}
	return &txM, nil
}

func (l *Letter) GetConsensus() (*ConsensusPayload, *errors.Error) {
	var cpM ConsensusPayload
	nErr := json.Unmarshal(l.Payload, &cpM)
	if nErr != nil {
		gLogger.Error("json.Unmarshal(l.Payload, &cpM) err", zap.Error(nErr))
		return nil, errors.Sys("Letter.GetConsensus Error: " + nErr.Error())
	}
	return &cpM, nil
}

//
//func NewLetter(
//	vnID vn.ID,
//	chain kt.Chain,
//	invID bid.BID,
//	lock Lock,
//	arrow Arrow,
//	from nd.ID,
//) *Letter {
//	return &Letter{
//		Arrow:      arrow,
//		Type:       LetterType,
//		Version:    LetterVersion,
//		Vn:         vnID,
//		Chain:      chain,
//		Invariable: invID,
//		Lock:       lock,
//		From:       from,
//		Signature:  nil,
//	}
//}

func (l *Letter) ToBytes() ([]byte, *errors.Error) {
	data, nErr := json.Marshal(l)
	if nErr != nil {
		return nil, errors.Sys("invalid letter, can not marshal: " + nErr.Error())
	}
	return data, nil
}

//
//func LetterOfBytes(data []byte) (*Letter, *errors.Error) {
//	var letter Letter
//	nErr := json.Unmarshal(data, &letter)
//	if nErr != nil {
//		return nil, errors.Sys("invalid letter bytes, can not unmarshal: " + nErr.Error())
//	}
//	return &letter, nil
//}
//
//func (l *Letter) GetType() kt.Type {
//	return l.Type
//}
//
//func (l *Letter) GetVersion() kt.Version {
//	return l.Version
//}
//
//func (l *Letter) GetVN() vn.ID {
//	return l.Vn
//}
//
//func (l *Letter) GetSignature() *kt.Signature {
//	return l.Signature
//}
//
//func (l *Letter) SetSignature(signature *kt.Signature) {
//	l.Signature = signature
//}

//func (l *Letter) Signing() *kt.Signing {
//	return kt.NewSigning().
//		Add("chain", l.Chain.S()).
//		Add("from", l.From.S())
//}

func (l *Letter) Verify() *errors.Error {
	if l.Signature == nil {
		return errors.Verify("Require Signature")
	}
	hash, err := l.Signature.Verify(l.doInjectSigning)
	if err != nil {
		return err
	}
	if l.Hash != hash {
		return errors.Verify("Invalid Letter.Hash")
	}
	return nil
}
