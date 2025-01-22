package proxy

import (
	"bytes"
	"fmt"
	"golang.org/x/exp/slices"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type HttpReader struct {
	io.Reader
}

func NewHttpReader(req *http.Request) *HttpReader {
	var buf bytes.Buffer
	path := req.URL.Path
	if req.URL.RawQuery != "" {
		path += "?" + req.URL.RawQuery
	}
	_, _ = fmt.Fprintf(&buf, "%s %s %s\r\n", req.Method, path, req.Proto)

	hasHost := false
	for k, vals := range req.Header {
		if strings.EqualFold(k, "Host") {
			hasHost = true
		}
		for _, val := range vals {
			_, _ = fmt.Fprintf(&buf, "%s: %s\r\n", k, val)
		}
	}

	if !hasHost && req.Host != "" {
		_, _ = fmt.Fprintf(&buf, "Host: %s\r\n", req.Host)
	}

	_, _ = fmt.Fprintf(&buf, "\r\n")

	// is not chunked
	if !slices.ContainsFunc(req.TransferEncoding, func(te string) bool {
		return strings.EqualFold(te, "chunked")
	}) {
		r := io.MultiReader(bytes.NewReader(buf.Bytes()), req.Body)
		return &HttpReader{r}
	}

	pr, pw := io.Pipe()

	// chunked
	go func() {
		defer pw.Close()
		// chunk
		chunk := make([]byte, 4096)
		for {
			n, err := req.Body.Read(chunk)
			if n > 0 {
				// write chunk size (十六进制) + CRLF

				chunkSize := strconv.FormatInt(int64(n), 16) + "\r\n"
				if _, err = pw.Write([]byte(chunkSize)); err != nil {
					return
				}
				// write chunk data + CRLF
				if _, err = pw.Write(chunk[:n]); err != nil {
					return
				}
				if _, err = pw.Write([]byte("\r\n")); err != nil {
					return
				}
			}
			if err != nil {
				if err == io.EOF {
					// write ending 0\r\n\r\n
					_, _ = pw.Write([]byte("0\r\n\r\n"))
				}
				return
			}
		}
	}()
	r := io.MultiReader(bytes.NewReader(buf.Bytes()), pr)
	return &HttpReader{r}
}

func (h *HttpReader) Read(p []byte) (int, error) {
	return h.Reader.Read(p)
}
