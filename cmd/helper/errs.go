package helper

import (
	"net/http"
)

var _ error = (*Error)(nil)

type (
	Op   = string
	Kind = int
	Err  = error
)

// common error status code
const (
	KindNotFound   = http.StatusNotFound
	KindBadRequest = http.StatusBadRequest
	KindUnexpected = http.StatusInternalServerError
)

type Error struct {
	Op        Op
	Kind      Kind
	Err       Err
	ClientMsg string
}

func (e Error) Error() string {
	return e.ClientMsg
}

func E(op Op, kind Kind, err Err, clientMsg string) error {
	return &Error{
		Op:        op,
		Kind:      kind,
		Err:       err,
		ClientMsg: clientMsg,
	}
}

func Ops(err error) []Op {
	var res []Op

	for {
		subErr, ok := err.(*Error)
		if !ok {
			return res
		}

		res = append(res, subErr.Op)

		err = subErr.Err
	}
}

func GetKind(err error) Kind {
	for {
		e, ok := err.(*Error)
		if !ok {
			return KindUnexpected
		}

		if e.Kind != 0 {
			return e.Kind
		}

		err = e
	}
}
