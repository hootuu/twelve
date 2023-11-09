package twelve

import (
	"encoding/json"
	"fmt"
	"github.com/hootuu/tome/bk"
	"github.com/hootuu/tome/ki"
	"github.com/hootuu/tome/pr"
	"github.com/hootuu/utils/crypto"
	"github.com/hootuu/utils/errors"
	"golang.org/x/exp/rand"
	"regexp"
	"time"
)

type Type int32

const (
	RequestMessage   Type = 1
	PrepareMessage   Type = 2
	CommittedMessage Type = 3
	ConfirmedMessage Type = 4
)

func TypeVerify(t Type) *errors.Error {
	switch t {
	case RequestMessage, PrepareMessage, CommittedMessage, ConfirmedMessage:
		return nil
	}
	return errors.Verify("invalid message.type")
}

type Version int32

const (
	DefaultVersion Version = 1
)

type Peer struct {
	ID  string `json:"i"`
	PUB ki.PUB `json:"p"`
}

func PeerOf(p *pr.Peer) *Peer {
	return &Peer{
		ID:  p.ID,
		PUB: p.PUB,
	}
}

func (peer *Peer) Verify() *errors.Error {
	if len(peer.ID) == 0 {
		return errors.Verify("require peer.ID")
	}
	if len(peer.PUB) == 0 {
		return errors.Verify("require peer.PUB")
	}
	return nil
}

func (peer *Peer) S() string {
	return "id=" + peer.ID + ";pub=" + peer.PUB.S()
}

type Request struct {
	ID        string `json:"i"`
	Signature string `json:"s"`
}

func (req *Request) Verify() *errors.Error {
	matched, _ := regexp.MatchString("^[a-zA-Z0-9]{3,200}$", req.ID)
	if !matched {
		return errors.Verify(fmt.Sprintf("invalid req.id: %s", req.ID))
	}
	return nil
}

type RequestPayload struct {
	Action    string `json:"a"`
	Parameter []byte `json:"p"`

	hash string
}

func RequestPayloadOf(byteData []byte) (*RequestPayload, *errors.Error) {
	var payload RequestPayload
	err := json.Unmarshal(byteData, &payload)
	if err != nil {
		return nil, errors.Verify("parse Request Payload failed.")
	}
	if err := payload.Verify(); err != nil {
		return nil, err
	}
	return &payload, nil
}

func (p *RequestPayload) Verify() *errors.Error {
	if len(p.Action) == 0 {
		return errors.Verify("invalid request payload, require action")
	}
	if len(p.Parameter) == 0 {
		return errors.Verify("invalid request payload, require parameter")
	}
	return nil
}

func (p *RequestPayload) ToBytes() ([]byte, *errors.Error) {
	b, nErr := json.Marshal(p)
	if nErr != nil {
		return nil, errors.Sys("invalid request payload", nErr)
	}
	return b, nil
}

func (p *RequestPayload) GetHash() string {
	if len(p.hash) == 0 {
		byteData := append([]byte(p.Action), p.Parameter...)
		p.hash = crypto.SHA256Bytes(byteData)
	}
	return p.hash
}

type Payload interface {
	ToBytes() ([]byte, *errors.Error)
}

type ReplyPayload struct {
	ID string `json:"i"` // Request ID
}

func ReplyPayloadOf(byteData []byte) (*ReplyPayload, *errors.Error) {
	id := string(byteData)
	return &ReplyPayload{ID: id}, nil
}

func (p *ReplyPayload) Verify() *errors.Error {
	if len(p.ID) == 0 {
		return errors.Verify("invalid reply payload, require id")
	}
	return nil
}

func (p *ReplyPayload) ToBytes() ([]byte, *errors.Error) {
	return []byte(p.ID), nil
}

type Message struct {
	Type      Type    `json:"t"`
	Version   Version `json:"v"`
	ID        string  `json:"i"`
	Nonce     int64   `json:"n"`
	Timestamp int64   `json:"tp"`
	Payload   []byte  `json:"pd"`
	Signature string  `json:"s"`
	Peer      *Peer   `json:"pr"`

	req   *RequestPayload
	reply *ReplyPayload
}

func MessageOf(byteData []byte) (*Message, *errors.Error) {
	var msg Message
	nErr := json.Unmarshal(byteData, &msg)
	if nErr != nil {
		return nil, errors.Verify("invalid message byte data")
	}
	if err := TypeVerify(msg.Type); err != nil {
		return nil, err
	}
	if msg.Payload == nil {
		return nil, errors.Verify("require message.payload")
	}
	if msg.Peer == nil {
		return nil, errors.Verify("require message.peer")
	}
	if err := msg.Peer.Verify(); err != nil {
		return nil, err
	}
	var err *errors.Error
	switch msg.Type {
	case RequestMessage:
		msg.req, err = RequestPayloadOf(msg.Payload)
	case PrepareMessage, CommittedMessage, ConfirmedMessage:
		msg.reply, err = ReplyPayloadOf(msg.Payload)
	default:
		return nil, errors.Verify("invalid message.type")
	}
	if err != nil {
		return nil, err
	}
	signOk, err := msg.SignVerify()
	if err != nil {
		return nil, err
	}
	if !signOk {
		return nil, errors.Verify("invalid message.signature")
	}
	return &msg, nil
}

func NewMessage(t Type, payload Payload, peer *Peer, pri ki.PRI) (*Message, *errors.Error) {
	if err := TypeVerify(t); err != nil {
		return nil, err
	}
	if payload == nil {
		return nil, errors.Verify("require payload")
	}
	if peer == nil {
		return nil, errors.Verify("require peer")
	}
	msg := &Message{
		Type:    t,
		Version: DefaultVersion,
		Peer:    peer,
	}
	if err := msg.doBuildNonce(); err != nil {
		return nil, err
	}
	if err := msg.doBuildTimestamp(); err != nil {
		return nil, err
	}
	var err *errors.Error
	switch t {
	case RequestMessage:
		requestPayload, ok := payload.(*RequestPayload)
		if !ok {
			return nil, errors.Verify("require the type is request message, payload must be Request Payload.")
		}
		msg.req = requestPayload
		msg.Payload, err = requestPayload.ToBytes()
	case PrepareMessage, CommittedMessage, ConfirmedMessage:
		replyPayload, ok := payload.(*ReplyPayload)
		if !ok {
			return nil, errors.Verify("require the type is request message, payload must be Request Payload.")
		}
		msg.reply = replyPayload
		msg.Payload, err = replyPayload.ToBytes()
	default:
		return nil, errors.Verify("invalid message.type")
	}
	if err != nil {
		return nil, err
	}
	signBuilder := msg.doGetSignBuilder()
	msg.ID = signBuilder.Hash()
	msg.Signature, err = signBuilder.Sign(pri)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func (msg *Message) ToBytes() ([]byte, *errors.Error) {
	b, nErr := json.Marshal(msg)
	if nErr != nil {
		return nil, errors.Sys("invalid message", nErr)
	}
	return b, nil
}

func (msg *Message) SignVerify() (bool, *errors.Error) {
	signBuilder := msg.doGetSignBuilder()
	return signBuilder.Verify(msg.Peer.PUB, msg.Signature)
}

func (msg *Message) GetRequestPayload() (*RequestPayload, *errors.Error) {
	if msg.req == nil {
		var err *errors.Error
		msg.req, err = RequestPayloadOf(msg.Payload)
		if err != nil {
			return nil, err
		}
	}
	return msg.req, nil
}

func (msg *Message) GetReplyPayload() (*ReplyPayload, *errors.Error) {
	if msg.reply == nil {
		var err *errors.Error
		msg.reply, err = ReplyPayloadOf(msg.Payload)
		if err != nil {
			return nil, err
		}
	}
	return msg.reply, nil
}

func (msg *Message) doBuildNonce() *errors.Error {
	rand.Seed(uint64(time.Now().UnixNano()))
	msg.Nonce = rand.Int63()
	return nil
}

func (msg *Message) doBuildTimestamp() *errors.Error {
	msg.Timestamp = time.Now().UnixMilli()
	return nil
}

func (msg *Message) doGetSignBuilder() *bk.SignBuilder {
	return bk.NewSignBuilder().
		Add("type", fmt.Sprintf("%d", msg.Type)).
		Add("version", fmt.Sprintf("%d", msg.Version)).
		Add("nonce", fmt.Sprintf("%d", msg.Nonce)).
		Add("timestamp", fmt.Sprintf("%d", msg.Timestamp)).
		Add("payload", string(msg.Payload)).
		Add("peer", msg.Peer.S())
}
