package pbft

import (
	"github.com/hootuu/twelve/twelve"
	"github.com/hootuu/utils/errors"
)

type Negotiable interface {
	OnRequest(msg *twelve.Message) *errors.Error
	OnPrepare(msg *twelve.Message) *errors.Error
	OnCommit(msg *twelve.Message) *errors.Error
	OnConfirm(msg *twelve.Message) *errors.Error
	Notify(msg *twelve.Message) *errors.Error
}
