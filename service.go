package connector

import (
	"context"

	"github.com/connctd/restapi-go"
)

type ConnectorService interface {
	AddInstallation(ctx context.Context, request InstallationRequest) (*InstallationResponse, error)
	RemoveInstallation(ctx context.Context, installationId string) error

	AddInstance(ctx context.Context, request InstantiationRequest) (*InstantiationResponse, error)
	RemoveInstance(ctx context.Context, instanceId string) error

	PerformAction(ctx context.Context, request ActionRequest) (*ActionResponse, error)
}

type ThingService interface {
	CreateThing(ctx context.Context, instanceId string, thing restapi.Thing) (restapi.Thing, error)
	UpdateProperty(ctx context.Context, instanceId, componentId, propertyId, value string) error
	UpdateActionStatus(ctx context.Context, instanceId string, actionRequestId string, actionResponse *ActionResponse) error
}

type ThingTemplates func(request InstantiationRequest) []restapi.Thing

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
	GetThingIDsByInstanceId(ctx context.Context, instanceId string) ([]string, error)
	RemoveInstance(ctx context.Context, instanceId string) error

	AddThingID(ctx context.Context, instanceID string, thingID string) error
}
