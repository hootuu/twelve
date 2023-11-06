package twelve

import "github.com/hootuu/utils/errors"

type Queue interface {
	Head() (*Job, *errors.Error)
	Tail() (*Job, *errors.Error)
	Exists(hash string) (*Job, *errors.Error)
	MustGet(hash string) (*Job, *errors.Error)
	Append(msg *Message) (*Job, *errors.Error)
	Remove(hash string) (*Job, *errors.Error)
	Confirm(hash string) (*Job, *errors.Error)
}
