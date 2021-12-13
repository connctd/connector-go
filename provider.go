package connector

import (
	"context"

	"github.com/connctd/restapi-go"
)

// The Provider interface is used in the default service to implement all technology specific details.
// It is the only interface most connectors should need to implement.
// A default implementation meant to be embedded in new connectors is provided by the SDK.
// It implements registration and removal of installations and instances, as well as an update channel and asynchronous action request.
// Connectors can use the update channel to push thing and action request updates to the connctd platform and should implemnet an action handler if they implement actions.
type Provider interface {
	UpdateChannel() <-chan UpdateEvent
	RegisterInstances(instances ...*Instance) error
	RemoveInstance(instanceId string) error
	RegisterInstallations(installations ...*Installation) error
	RemoveInstallation(installationId string) error
	RequestAction(ctx context.Context, instance *Instance, actionRequest ActionRequest) (restapi.ActionRequestStatus, error)
}

// UpdateEvents are pushed to the UpdateChannel.
// The default service will listen to the channel.
// If it receives an UpdateEvent with only a PropertyEventUpdate it will update the specified property with the new value.
// If it receives an ActionEvent it will update the the state of the specified action request to the state in the ActionResponse.
// If the same UpdateEvent contains a PropertyUpateEvent it will first update the property and then the action request.
// If the property update fails it will set the action request state to failed.
type UpdateEvent struct {
	ActionEvent         *ActionEvent
	PropertyUpdateEvent *PropertyUpdateEvent
}

// ActionEvent is used to propagate action request results to the service.
// See UpdateEvent for details.
type ActionEvent struct {
	InstanceId      string
	ActionResponse  *ActionResponse
	ActionRequestId string
}

// ActionEvent is used to propagate property updates to the service.
// See UpdateEvent for details.
type PropertyUpdateEvent struct {
	ThingId     string
	InstanceId  string
	ComponentId string
	PropertyId  string
	Value       string
}
