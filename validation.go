package connector

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

// SignedHeaderKey defines a header field name
type SignedHeaderKey string

var (
	// SignatureFragmentDelimiter defines how different fragments like headers and
	// body are concatenated to a signable payload. CRLF is already used as a
	// separator for http 1.1 headers and payloads (https://tools.ietf.org/html/rfc7230#page-19)
	// which means underlying libraries should already be aware of correct
	// CRLF handling (e.g. prevent CRLF injection)
	SignatureFragmentDelimiter = "\r\n"

	// SignedHeaderKeyDate stands for the date header
	SignedHeaderKeyDate = SignedHeaderKey("Date")

	// SignedHeaderKeys defines a list of headers that are used to build
	// the payload-to-be-signed. If a request does not contain all of these
	// headers it can't be signed nor validated and thus is invalid.
	// The order of keys inside this list defines how the payload-to-be-signed
	// is constructed
	SignedHeaderKeys = []SignedHeaderKey{
		SignedHeaderKeyDate,
	}

	// ErrorMissingHeader thrown if one of the signed headers is missing
	ErrorMissingHeader = errors.New("A required header is missing")
)

// SignablePayload builds the payload which can be signed
// Method\r\nHost\r\nRequestURI\r\nDate Header Value\r\nBody
// Example: GET\r\nfoo.com:8080\r\n/bar?hello=world\r\nWed, 07 Oct 2020 10:00:00 GMT\r\n{\"hello\":\"world\"}
func SignablePayload(method string, host string, url *url.URL, headers http.Header, body []byte) ([]byte, error) {
	var b bytes.Buffer

	b.WriteString(method)
	b.WriteString(SignatureFragmentDelimiter)
	b.WriteString(host)
	b.WriteString(SignatureFragmentDelimiter)
	b.WriteString(url.RequestURI())
	b.WriteString(SignatureFragmentDelimiter)

	for _, currHeader := range SignedHeaderKeys {
		value := headers.Get(string(currHeader))
		if value == "" {
			fmt.Println(currHeader)
			return []byte{}, ErrorMissingHeader
		}
		b.WriteString(value)
		b.WriteString(SignatureFragmentDelimiter)
	}

	b.Write(body)
	return b.Bytes(), nil
}
