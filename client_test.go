package connector

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	stdlog "log"

	"github.com/go-logr/stdr"

	"github.com/connctd/api-go"
	"github.com/connctd/restapi-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	dummyServer *httptest.Server
	testClient  Client
)

var createThingTests = []struct {
	name            string
	handler         http.HandlerFunc
	expectedError   error
	expectedThingID string
}{
	{
		name: "CreateThingFailWithNoResponseBody",
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		},
		expectedError: ErrorUnexpectedStatusCode,
	},
	{
		name: "CreateThingFailWithResponseBody",
		handler: func(w http.ResponseWriter, r *http.Request) {
			dummyError := api.NewError("Foo error", "Foo err description", http.StatusBadRequest)
			dummyError.Write(w)
		},
		expectedError: ErrorUnexpectedStatusCode,
	},
	{
		name: "CreateThingSuccess",
		handler: func(w http.ResponseWriter, r *http.Request) {
			response := AddThingResponse{
				ID: "123",
			}

			b, _ := json.Marshal(response)

			w.WriteHeader(http.StatusCreated)
			w.Write(b)
		},
		expectedThingID: "123",
	},
}

func TestCreateThing(t *testing.T) {
	for _, currTest := range createThingTests {
		t.Run(currTest.name, func(r *testing.T) {
			dummyServer := httptest.NewServer(currTest.handler)
			defer dummyServer.Close()

			url, err := url.Parse(dummyServer.URL)
			require.Nil(r, err)

			// use logger based on go's standard logger
			logger := stdr.New(stdlog.New(os.Stderr, "", stdlog.LstdFlags|stdlog.Lshortfile))
			client := NewClient(&ClientOptions{ConnctdBaseURL: url}, logger)

			id, err := client.CreateThing(context.Background(), "", restapi.Thing{Name: "DummyThing"})

			if currTest.expectedError != nil {
				assert.Equal(r, currTest.expectedError, err)
			}

			if currTest.expectedThingID != "" {
				assert.Equal(r, currTest.expectedThingID, id)
			}
		})
	}

}
