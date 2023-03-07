package pkg

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"strings"
)

const urlDst = `https://challenges.cloudflare.com/turnstile/v0/siteverify`

var client = &http.Client{}

const (
	missingInputResp   = "missing-input-response"
	invalidInputResp   = "invalid-input-response"
	badRequest         = "bad-request"
	timeoutOrDuplicate = "timeout-or-duplicate"
	internalError      = "internal-error"
)

var (
	ErrMissingInput       = errors.New("captcha is required")
	ErrInvalidInput       = errors.New("invalid captcha")
	ErrBadRequest         = errors.New("malformed captcha")
	ErrTimeoutOrDuplicate = errors.New("duplicate captcha")
	ErrInternal           = errors.New("internal captcha validation error")
)

type turnstileResponse struct {
	Success    bool     `json:"success" form:"success"`
	ErrorCodes []string `json:"error-codes" form:"error-codes"`
}

func ValidateCaptcha(ctx context.Context, token string, input string) (int, error) {
	args := strings.NewReader(fmt.Sprintf(`secret=%s&response=%s`, token, input))

	resp, err := fetchRequest(ctx, args)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Warn().Err(err).Msgf(`err closing turnstile resp body: %v`, err)
		}
	}()

	var result turnstileResponse

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Debug().Err(err)
		return http.StatusInternalServerError, ErrInternal
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Debug().Err(err)
		return http.StatusInternalServerError, ErrInternal
	}

	if !result.Success {
		return determineErr(result.ErrorCodes[0])
	}

	return http.StatusOK, nil
}

func fetchRequest(ctx context.Context, args io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlDst, args)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		log.Debug().Err(err)
		return nil, ErrInternal
	}

	return resp, nil
}

func determineErr(errorCodes string) (int, error) {
	switch errorCodes {
	case missingInputResp:
		return http.StatusBadRequest, ErrMissingInput
	case invalidInputResp:
		return http.StatusNotAcceptable, ErrInvalidInput
	case badRequest:
		return http.StatusBadRequest, ErrBadRequest
	case timeoutOrDuplicate:
		return http.StatusNotAcceptable, ErrTimeoutOrDuplicate
	case internalError:
		return http.StatusInternalServerError, ErrInternal
	default:
		log.Debug().Msg(errorCodes)
		return http.StatusInternalServerError, ErrInternal
	}
}
