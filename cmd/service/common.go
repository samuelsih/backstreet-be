package service

type CommonResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (c *CommonResponse) Set(code int, msg string) {
	c.Code = code
	c.Message = msg
}
