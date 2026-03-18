package cli

import (
	"encoding/json"
	"fmt"
	"io"
)

type Issue struct {
	Code   string `json:"code"`
	Detail string `json:"detail"`
}

type Response struct {
	OK       bool    `json:"ok"`
	Code     string  `json:"code"`
	Message  string  `json:"message"`
	Data     any     `json:"data,omitempty"`
	Warnings []Issue `json:"warnings,omitempty"`
	Errors   []Issue `json:"errors,omitempty"`
}

func writeResponse(out io.Writer, asJSON bool, response Response) error {
	if asJSON {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(response)
	}

	if _, err := fmt.Fprintln(out, response.Message); err != nil {
		return err
	}
	if response.Data == nil {
		return nil
	}

	data, err := json.MarshalIndent(response.Data, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "%s\n", data)
	return err
}
