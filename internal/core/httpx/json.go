package httpx

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func DecodeJSON(r *http.Request, out any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(out)
}

func DecodeJSONAllowEmpty(r *http.Request, out any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(out)
	if errors.Is(err, context.Canceled) {
		return err
	}
	if errors.Is(err, io.EOF) {
		return nil
	}
	return err
}
