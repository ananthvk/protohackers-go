package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"log/slog"
	"math"
	"net"
	"strings"

	"github.com/ananthvk/protohackers-go/internal/tcp"
)

func isPrime(n int64) bool {
	if n <= 1 {
		return false
	}
	if n == 2 || n == 3 {
		return true
	}
	if n%2 == 0 {
		return false
	}
	// All primes are of the form 6k +- 1 (except 2 & 3)
	remainder := n % 6
	if !(remainder == 1 || remainder == 5) {
		return false
	}

	m := int64(math.Sqrt(float64(n)))
	for i := int64(3); i <= m; i += 2 {
		if n%i == 0 {
			return false
		}
	}
	return true
}

func decodeJSONRequest(reader *bufio.Reader) (int64, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return 0, err
	}
	decoder := json.NewDecoder(strings.NewReader(line))
	decoder.UseNumber()

	primeRequest := struct {
		Method *string          `json:"method"`
		Number *json.RawMessage `json:"number"`
	}{}
	err = decoder.Decode(&primeRequest)
	if err != nil {
		return 0, err
	}
	// Check for trailing data by calling decode again
	err = decoder.Decode(&struct{}{})
	if err != io.EOF {
		return 0, errors.New("body contains extra data after JSON value")
	}
	if primeRequest.Method == nil && primeRequest.Number == nil {
		return 0, errors.New("required fields 'method' and 'number' are missing")
	}
	if primeRequest.Method == nil {
		return 0, errors.New("required field 'method' is missing")
	}
	if primeRequest.Number == nil {
		return 0, errors.New("required field 'number' is missing")
	}
	if *primeRequest.Method != "isPrime" {
		return 0, errors.New("method field is not 'isPrime'")
	}

	// Check if it's an integer, if it's a floating point number, return 0
	if len(*primeRequest.Number) == 0 {
		return 0, errors.New("field 'number' is empty")
	}
	if (*primeRequest.Number)[0] == '"' {
		return 0, errors.New("field 'number' is a string, expected a number")
	}

	var number json.Number
	err = json.Unmarshal(*primeRequest.Number, &number)
	if err != nil {
		return 0, errors.New("field 'number' is invalid")
	}

	num, err := number.Int64()
	if err != nil {
		// Check if it's a float
		_, err := number.Float64()
		if err == nil {
			return 0, nil
		}
	}
	return num, err
}

func handlePrimeTask(conn net.Conn) error {
	type errorResponse struct {
		Error string `json:"error"`
	}
	type Response struct {
		Method string `json:"method"`
		Prime  bool   `json:"prime"`
	}

	numberOfRequests := 0
	// Log the number of requests made by the client when it exits
	defer func() {
		slog.Info("number of requests handled", "address", conn.RemoteAddr(), "num_requests", numberOfRequests)
	}()

	reader := bufio.NewReader(conn)

	for {
		number, err := decodeJSONRequest(reader)
		numberOfRequests++
		if err != nil {
			// Send back an error response
			errorResponse := errorResponse{Error: err.Error()}
			e := json.NewEncoder(conn).Encode(errorResponse)
			if e != nil {
				slog.Error("error while sending error message to client", "error", err)
			}
			return err
		}
		isPrime := isPrime(number)
		response := Response{Method: "isPrime", Prime: isPrime}
		e := json.NewEncoder(conn).Encode(response)
		if e != nil {
			slog.Error("error while sending response to client", "error", err)
			return err
		}
	}
}

func main() {
	server := tcp.NewServer(context.Background())
	server.AddToFlags()
	flag.Parse()
	server.LoadFromFlags()
	server.SetClientHandler(handlePrimeTask)
	server.ListenAndServe()
}
