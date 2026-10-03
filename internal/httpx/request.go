package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const maxBodySize = 1 << 20 // 1 MB

func ReadJSON(
	w http.ResponseWriter,
	r *http.Request,
	dst any,
) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(dst)

	if err != nil {
		var syntaxError *json.SyntaxError
		var typeError *json.UnmarshalTypeError
		var maxBytesError *http.MaxBytesError

		switch {
		case errors.Is(err, io.EOF):
			return errors.New("request body must not be empty")

		case errors.As(err, &syntaxError):
			return fmt.Errorf(
				"invalid JSON near position %d",
				syntaxError.Offset,
			)

		case errors.As(err, &typeError):
			return fmt.Errorf(
				"invalid value for field %q",
				typeError.Field,
			)

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			return errors.New(err.Error())

		case errors.As(err, &maxBytesError):
			return fmt.Errorf(
				"request body must not exceed %d bytes",
				maxBytesError.Limit,
			)

		default:
			return errors.New("invalid JSON body")
		}
	}

	// Ensure only one JSON object exists.
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New(
			"request body must contain a single JSON object",
		)
	}

	return nil
}
