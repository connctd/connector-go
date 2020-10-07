package connector

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

// SignedHeaderKey defines a header field name
type signedHeaderKey string

const (
	// signatureFragmentDelimiter defines how different fragments like headers and
	// body are concatenated to a signable payload. CRLF is already used as a
	// separator for http 1.1 headers and payloads (https://tools.ietf.org/html/rfc7230#page-19)
	// which means underlying libraries should already be aware of correct
	// CRLF handling (e.g. prevent CRLF injection)
	signatureFragmentDelimiter = "\r\n"

	// separates keys from values in constructed payload
	keyValueSeparator = ":"

	// signedHeaderKeyDate stands for the date header
	signedHeaderKeyDate = signedHeaderKey("Date")
)

// signedHeaderKeys defines a list of headers that are used to build
// the payload-to-be-signed. If a request does not contain all of these
// headers it can't be signed nor validated and thus is invalid.
// The order of keys inside this list defines how the payload-to-be-signed
// is constructed
var signedHeaderKeys = []signedHeaderKey{
	signedHeaderKeyDate,
}

// SignablePayload builds the payload which can be signed
// Method\r\nHost\r\nRequestURI\r\nDate Header Value\r\nBody
// Example: (method):GET\r\n(url):https://foo.com:8080/bar?hello=world\r\n(Date):Wed, 07 Oct 2020 10:00:00 GMT\r\n(body):{\"hello\":\"world\"}\r\n
func SignablePayload(method string, host string, url *url.URL, headers http.Header, body []byte) ([]byte, error) {
	var b bytes.Buffer

	// write method
	b.WriteString("(method)")
	b.WriteString(keyValueSeparator)
	b.WriteString(method)
	b.WriteString(signatureFragmentDelimiter)

	// write url
	b.WriteString("(url)")
	b.WriteString(keyValueSeparator)
	b.WriteString(url.String())
	b.WriteString(signatureFragmentDelimiter)

	// write all required headers
	for _, currHeader := range signedHeaderKeys {
		value := headers.Get(string(currHeader))
		if value == "" {
			fmt.Println(currHeader)
			return []byte{}, ErrorMissingHeader
		}

		b.WriteString("(" + string(currHeader) + ")")
		b.WriteString(keyValueSeparator)
		b.WriteString(value)
		b.WriteString(signatureFragmentDelimiter)
	}

	// write body
	b.WriteString("(body)")
	b.WriteString(keyValueSeparator)
	b.Write(body)
	b.WriteString(signatureFragmentDelimiter)

	return b.Bytes(), nil
}

var (
	// ErrorMissingHeader thrown if one of the signed headers is missing
	ErrorMissingHeader = errors.New("A required header is missing")
)
