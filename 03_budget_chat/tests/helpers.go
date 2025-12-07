package tests

import (
	"bufio"
	"fmt"
	"time"
)

func readLineWithTimeout(r *bufio.Reader, d time.Duration) (string, error) {
	ch := make(chan struct {
		s   string
		err error
	}, 1)

	go func() {
		s, err := r.ReadString('\n')
		ch <- struct {
			s   string
			err error
		}{s, err}
	}()

	select {
	case res := <-ch:
		return res.s, res.err
	case <-time.After(d):
		return "", fmt.Errorf("read timeout after %v", d)
	}
}

func writeStringWithTimeout(w *bufio.Writer, s string, d time.Duration) error {
	ch := make(chan error, 1)
	go func() {
		_, err := w.WriteString(s)
		if err == nil {
			err = w.Flush()
		}
		ch <- err
	}()
	select {
	case err := <-ch:
		return err
	case <-time.After(d):
		return fmt.Errorf("write timeout after %v", d)
	}
}
