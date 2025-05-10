// SPDX-FileCopyrightText: © 2024 - 2025 Anthony Champagne <dev@anthonychampagne.fr>
//
// SPDX-License-Identifier: AGPL-3.0-only

package util

import (
	"encoding/json"
	"io"
	"net/http"
)

type HttpRequest struct {
	http.Request
	ResponseWriter http.ResponseWriter
}

func (z HttpRequest) WriteJson(data interface{}) error {
	return json.NewEncoder(z.ResponseWriter).Encode(data)
}

func (z HttpRequest) WriteStream(data io.ReadCloser) error {
	if _, err := CopyBuf(z.ResponseWriter, data); err != nil {
		return err
	}

	data.Close()
	return nil
}

func (z HttpRequest) DecodeJsonBody(data interface{}) error {
	return json.NewDecoder(z.Body).Decode(data)
}

func (z HttpRequest) HttpErrorStr(status int, errorStr string) error {
	z.ResponseWriter.WriteHeader(status)

	if errorStr != "" {
		z.ResponseWriter.Header().Set("Content-Type", "text/plain")
		_, err := z.ResponseWriter.Write([]byte(errorStr))
		return err
	}

	return nil
}

func (z HttpRequest) HttpErrorJson(status int, data interface{}) error {
	z.ResponseWriter.WriteHeader(status)

	return z.WriteJson(data)
}
