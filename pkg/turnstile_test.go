package pkg

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestValidateCaptcha(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	token := "1x0000000000000000000000000000000AA"

	t.Run("success", func(t *testing.T) {
		input := "1x00000000000000000000AA"

		statusCode, err := ValidateCaptcha(ctx, token, input)

		if err != nil {
			t.Fatalf("Error calling ValidateCaptcha: %s", err)
		}

		if statusCode != http.StatusOK {
			t.Errorf("Expected status code %d but got %d", http.StatusOK, statusCode)
		}
	})

	t.Run("required", func(t *testing.T) {
		input := ""

		statusCode, err := ValidateCaptcha(ctx, token, input)

		if err == nil {
			t.Fatal("Error is nil")
		}

		if statusCode != http.StatusBadRequest {
			t.Errorf("Expected status code %d but got %d", http.StatusBadRequest, statusCode)
		}
	})
}

func TestFetchRequest(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	body := fmt.Sprintf(`secret=1x0000000000000000000000000000000AA&response=1x00000000000000000000AA`)
	reader := strings.NewReader(body)

	resp, err := fetchRequest(ctx, reader)

	if err != nil {
		t.Fatalf("Error calling fetchRequest: %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d but got %d", http.StatusOK, resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Error reading response body: %s", err)
	}

	var result turnstileResponse
	err = json.Unmarshal(bodyBytes, &result)
	if err != nil {
		t.Fatalf("Error converting to json: %s", err)
	}

	if !result.Success {
		t.Fatalf("Error result is not success: %v", result)
	}
}

func TestDetermineErr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input         string
		expectedCode  int
		expectedError error
	}{
		{
			input:         missingInputResp,
			expectedCode:  http.StatusBadRequest,
			expectedError: ErrMissingInput,
		},
		{
			input:         invalidInputResp,
			expectedCode:  http.StatusNotAcceptable,
			expectedError: ErrInvalidInput,
		},
		{
			input:         badRequest,
			expectedCode:  http.StatusBadRequest,
			expectedError: ErrBadRequest,
		},
		{
			input:         timeoutOrDuplicate,
			expectedCode:  http.StatusNotAcceptable,
			expectedError: ErrTimeoutOrDuplicate,
		},
		{
			input:         internalError,
			expectedCode:  http.StatusInternalServerError,
			expectedError: ErrInternal,
		},
		{
			input:         "unknown error code",
			expectedCode:  http.StatusInternalServerError,
			expectedError: ErrInternal,
		},
	}

	for _, test := range tests {
		code, err := determineErr(test.input)

		if code != test.expectedCode {
			t.Errorf("For input '%s', expected code %d but got %d", test.input, test.expectedCode, code)
		}

		if !errors.Is(err, test.expectedError) {
			t.Errorf("For input '%s', expected error '%s' but got '%s'", test.input, test.expectedError.Error(), err.Error())
		}
	}
}
