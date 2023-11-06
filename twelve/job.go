package twelve

type State int8

const (
	Committed State = 1
	Pending   State = 2
	Confirmed State = 9
)

type Job struct {
	Hash    string
	State   State
	Message *Message
	Pre     string
	Nxt     string
}
