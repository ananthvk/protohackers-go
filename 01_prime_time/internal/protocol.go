package internal

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

type Request struct {
	Number int64
}
type Response struct {
	Method string `json:"method"`
	Prime  bool   `json:"prime"`
}

func ParseRequest(line []byte) (Request, error) {
	decoder := json.NewDecoder(bytes.NewReader(line))
	decoder.UseNumber()

	primeRequest := struct {
		Method *string         `json:"method"`
		Number json.RawMessage `json:"number"`
	}{}

	err := decoder.Decode(&primeRequest)
	if err != nil {
		return Request{}, err
	}

	// Check if there is any trailing data by calling decode again
	err = decoder.Decode(&struct{}{})
	if err != io.EOF {
		return Request{}, errors.New("request contains extra data after JSON object")
	}

	// Check if fields are missing
	if primeRequest.Method == nil {
		return Request{}, errors.New("required field 'method' is missing")
	}
	if primeRequest.Number == nil {
		return Request{}, errors.New("required field 'number' is missing")
	}

	// Check if method is 'isPrime'
	if *primeRequest.Method != "isPrime" {
		return Request{}, errors.New("method is not 'isPrime'")
	}

	// Edge case, json.Number accepts numeric values represented as strings]
	// but the specification disallows it
	if len(primeRequest.Number) == 0 || primeRequest.Number[0] == '"' {
		return Request{}, errors.New("'number' must be a numeric JSON value")
	}

	var num json.Number
	if err := json.Unmarshal(primeRequest.Number, &num); err != nil {
		return Request{}, errors.New("invalid number")
	}

	i, err := num.Int64()
	if err != nil {
		// If it's not an integer, try parsing it as a float
		f, err := num.Float64()
		i = int64(f)

		// If both of them failed, return an error
		if err != nil {
			return Request{}, errors.New("invalid number")
		}
	}
	return Request{Number: i}, nil
}
func SendResponse(w io.Writer, response Response) error {
	return json.NewEncoder(w).Encode(response)
}
