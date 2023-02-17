package service

import (
	"backstreetlinkv2/cmd/helper"
	"github.com/rs/zerolog/log"
	"strings"
)

type CommonResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (c *CommonResponse) Set(code int, msg string) {
	c.Code = code
	c.Message = msg
}

func (c *CommonResponse) SetOK() {
	c.Code = 200
	c.Message = "OK"
}

func (c *CommonResponse) SetErr(err error) {
	ops := helper.Ops(err)
	c.Code = helper.GetKind(err)
	c.Message = err.Error()

	log.Debug().
		Int("status", c.Code).
		Str("trace", strings.Join(ops, " - ")).
		Errs("errs", helper.TraceErr(err)).
		Msg(err.Error())

}
