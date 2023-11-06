package twelve

import (
	"github.com/hootuu/utils/errors"
)

type Notifier interface {
	Notify(msg *Message) *errors.Error
}
