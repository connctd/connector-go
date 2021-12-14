// Package connector implements a SDK for connector development.
// In most cases it should be sufficient to use the default provider to implement the connector.Provider interface and use it with the default service and the connector handler.
// For an example connector using this SDK to integrate an external API, see https://github.com/connctd/giphy-connector/.
package connector

import (
	"context"

	"github.com/connctd/restapi-go"
)

// ConnectorService interface is used by the ConnectorHandler and will be called to process the validated requests used in the connector protocol.
// The SDK provides a default implementation for the ConnectorService interface that should be sufficient for most connector developments.
type ConnectorService interface {
	AddInstallation(ctx context.Context, request InstallationRequest) (*InstallationResponse, error)
	RemoveInstallation(ctx context.Context, installationId string) error

	AddInstance(ctx context.Context, request InstantiationRequest) (*InstantiationResponse, error)
	RemoveInstance(ctx context.Context, instanceId string) error

	PerformAction(ctx context.Context, request ActionRequest) (*ActionResponse, error)
}

type ThingTemplates func(request InstantiationRequest) []restapi.Thing

// Database interface is used in the default service to persist new installations, instances, configurations and external device mappings.
// The SDK provides a default implementation supporting Postgresql, Mysql and Sqlite3.
type Database interface {
	AddInstallation(ctx context.Context, installationRequest InstallationRequest) error
	AddInstallationConfiguration(ctx context.Context, installationId string, config []Configuration) error
	GetInstallations(ctx context.Context) ([]*Installation, error)
	RemoveInstallation(ctx context.Context, installationId string) error

	AddInstance(ctx context.Context, instantiationRequest InstantiationRequest) error
	AddInstanceConfiguration(ctx context.Context, instanceId string, config []Configuration) error
	GetInstance(ctx context.Context, instanceId string) (*Instance, error)
	GetInstances(ctx context.Context) ([]*Instance, error)
	GetInstanceByThingId(ctx context.Context, thingId string) (*Instance, error)
	GetInstanceConfiguration(ctx context.Context, instanceId string) ([]Configuration, error)
	GetMappingByInstanceId(ctx context.Context, instanceId string) ([]ThingMapping, error)
	RemoveInstance(ctx context.Context, instanceId string) error

	AddThingMapping(ctx context.Context, instanceID string, thingID string, externalId string) error
}
