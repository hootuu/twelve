package pbft

type About string
type Payload string
type Peer struct {
	ID  string `json:"id"`
	PUB string `json:"pub"`
}
type Result string

type Request struct {
	ID        string `json:"id"`
	Signature string `json:"signature"`
}

type RequestMessage struct {
	ID        string  `json:"id"`
	Nonce     int64   `json:"nonce"`
	Signature string  `json:"signature"`
	Timestamp int64   `json:"timestamp"`
	Peer      *Peer   `json:"peer"`
	About     About   `json:"about"`
	Payload   Payload `json:"payload"`
}

type ReplyMessage struct {
	ID        string  `json:"id"`
	Nonce     int64   `json:"nonce"`
	Signature string  `json:"signature"`
	Timestamp int64   `json:"timestamp"`
	Peer      *Peer   `json:"peer"`
	RequestID string  `json:"request_id"`
	About     About   `json:"about"`
	Payload   Payload `json:"payload"`
	Result    Result  `json:"result"`
}
