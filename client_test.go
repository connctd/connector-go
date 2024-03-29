package connector

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/connctd/connector-go/connctd"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var createThingTests = []struct {
	name            string
	handler         http.HandlerFunc
	expectedError   error
	expectedThingID string
}{
	{
		name: "Create thing fails on bad response code",
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		},
		expectedError: ErrorUnexpectedStatusCode,
	},
	{
		name: "Create thing fails on error response body",
		handler: func(w http.ResponseWriter, r *http.Request) {
			dummyError := NewError("Foo error", "Foo err description", http.StatusBadRequest)
			dummyError.Write(w)
		},
		expectedError: ErrorUnexpectedStatusCode,
	},
	{
		name: "Create thing fails on bad response",
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`foo`))
		},
		expectedError: ErrorUnexpectedResponse,
	},
	{
		name: "Create thing successful",
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

			url, err := url.Parse(dummyServer.URL + "/")
			require.Nil(r, err)

			client, err := NewClient(&ClientOptions{ConnctdBaseURL: url}, DefaultLogger)
			require.Nil(r, err)

			thing, err := client.CreateThing(context.Background(), "", connctd.Thing{Name: "DummyThing"})

			if currTest.expectedError != nil {
				assert.Equal(r, currTest.expectedError, err)
			}

			if currTest.expectedThingID != "" {
				assert.Equal(r, currTest.expectedThingID, thing.ID)
			}
		})
	}
}

var deleteThingTests = []struct {
	name          string
	handler       http.HandlerFunc
	expectedError error
}{
	{
		name: "Delete thing fails on bad response code",
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		},
		expectedError: ErrorUnexpectedStatusCode,
	},
	{
		name: "Delete thing successful",
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		},
	},
}

func TestDeleteThing(t *testing.T) {
	for _, currTest := range deleteThingTests {
		t.Run(currTest.name, func(r *testing.T) {
			dummyServer := httptest.NewServer(currTest.handler)
			defer dummyServer.Close()

			url, err := url.Parse(dummyServer.URL + "/")
			require.Nil(r, err)

			client, err := NewClient(&ClientOptions{ConnctdBaseURL: url}, DefaultLogger)
			require.Nil(r, err)

			err = client.DeleteThing(context.Background(), "", "fooid")

			if currTest.expectedError != nil {
				assert.Equal(r, currTest.expectedError, err)
			}
		})
	}
}

func TestParametersValidation(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	u, err := url.Parse("https://foobar")
	require.Nil(err)

	badOptions := ClientOptions{
		ConnctdBaseURL: u,
	}
	_, err = NewClient(&badOptions, DefaultLogger)
	assert.Equal(err, ErrorInvalidBaseURL)

	_, err = NewClient(DefaultOptions(), DefaultLogger)
	assert.Nil(err)
}

var updateThingPropetyValueTests = []struct {
	name          string
	handler       http.HandlerFunc
	expectedError error
}{
	{
		name: "Update thing property fails on invalid response status",
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		},
		expectedError: ErrorUnexpectedStatusCode,
	},
	{
		name: "Update thing property value successful",
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		},
		expectedError: nil,
	},
}

func TestUpdateThingPropertyValue(t *testing.T) {
	for _, currTest := range updateThingPropetyValueTests {
		t.Run(currTest.name, func(r *testing.T) {
			dummyServer := httptest.NewServer(currTest.handler)
			defer dummyServer.Close()

			url, err := url.Parse(dummyServer.URL + "/")
			require.Nil(r, err)

			client, err := NewClient(&ClientOptions{ConnctdBaseURL: url}, DefaultLogger)
			require.Nil(r, err)

			err = client.UpdateThingPropertyValue(context.Background(), "", "fooThingID", "fooComponentID", "fooPropertyID", "foo", time.Now())
			assert.Equal(r, currTest.expectedError, err)
		})
	}

}

var updateInstanceStateTests = []struct {
	name          string
	handler       http.HandlerFunc
	details       json.RawMessage
	expectedError error
}{
	{
		name: "Update state fails on invalid response status",
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		},
		details:       nil,
		expectedError: ErrorUnexpectedStatusCode,
	},
	{
		name: "Update was successful",
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		},
		details:       nil,
		expectedError: nil,
	},
	{
		name: "Successful with body",
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		},
		details:       []byte(`Hello world`),
		expectedError: nil,
	},
}

func TestUpdateInstanceState(t *testing.T) {
	for _, currTest := range updateInstanceStateTests {
		t.Run(currTest.name, func(r *testing.T) {
			dummyServer := httptest.NewServer(currTest.handler)
			defer dummyServer.Close()

			url, err := url.Parse(dummyServer.URL + "/")
			require.Nil(r, err)

			client, err := NewClient(&ClientOptions{ConnctdBaseURL: url}, DefaultLogger)
			require.Nil(r, err)

			err = client.UpdateInstanceState(context.Background(), "", InstantiationStateComplete, nil)
			assert.Equal(r, currTest.expectedError, err)
		})
	}

}

var updateActionStatusTests = []struct {
	name          string
	handler       http.HandlerFunc
	expectedError error
}{
	{
		name: "Update action status fails on bad response code",
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		},
		expectedError: ErrorUnexpectedStatusCode,
	},
	{
		name: "Update action request was successful",
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		},
	},
}

func TestUpdateActionStatus(t *testing.T) {
	for _, currTest := range updateActionStatusTests {
		t.Run(currTest.name, func(r *testing.T) {
			dummyServer := httptest.NewServer(currTest.handler)
			defer dummyServer.Close()

			url, err := url.Parse(dummyServer.URL + "/")
			require.Nil(r, err)

			client, err := NewClient(&ClientOptions{ConnctdBaseURL: url}, DefaultLogger)
			require.Nil(r, err)

			err = client.UpdateActionStatus(context.Background(), "", "fooid", ActionRequestStatusCompleted, "")

			if currTest.expectedError != nil {
				assert.Equal(r, currTest.expectedError, err)
			}
		})
	}
}

var updateThingStatusTests = []struct {
	name          string
	handler       http.HandlerFunc
	expectedError error
}{
	{
		name: "Update thing status fails on bad response code",
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		},
		expectedError: ErrorUnexpectedStatusCode,
	},
	{
		name: "Update thing status was successful",
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		},
	},
}

func TestUpdateThingStatus(t *testing.T) {
	for _, currTest := range updateThingStatusTests {
		t.Run(currTest.name, func(r *testing.T) {
			dummyServer := httptest.NewServer(currTest.handler)
			defer dummyServer.Close()

			url, err := url.Parse(dummyServer.URL + "/")
			require.Nil(r, err)

			client, err := NewClient(&ClientOptions{ConnctdBaseURL: url}, DefaultLogger)
			require.Nil(r, err)

			err = client.UpdateThingStatus(context.Background(), "", "foothingid", connctd.StatusTypeAvailable)

			if currTest.expectedError != nil {
				assert.Equal(r, currTest.expectedError, err)
			}
		})
	}
}

var updateInstallationStateTests = []struct {
	name          string
	handler       http.HandlerFunc
	expectedError error
}{
	{
		name: "Update installation state fails on bad response code",
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		},
		expectedError: ErrorUnexpectedStatusCode,
	},
	{
		name: "Update installation state was successful",
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		},
	},
}

func TestUpdateInstallationState(t *testing.T) {
	for _, currTest := range updateInstallationStateTests {
		t.Run(currTest.name, func(r *testing.T) {
			dummyServer := httptest.NewServer(currTest.handler)
			defer dummyServer.Close()

			url, err := url.Parse(dummyServer.URL + "/")
			require.Nil(r, err)

			client, err := NewClient(&ClientOptions{ConnctdBaseURL: url}, DefaultLogger)
			require.Nil(r, err)

			err = client.UpdateInstallationState(context.Background(), "footoken", InstallationStateOngoing, nil)

			if currTest.expectedError != nil {
				assert.Equal(r, currTest.expectedError, err)
			}
		})
	}
}
