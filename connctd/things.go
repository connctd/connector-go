package connctd

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	urlConform, _ = regexp.Compile("^[a-zA-Z0-9-_]{1,200}$")
)

// Thing describes a third party device or service
type Thing struct {
	ID              string           `json:"id"`
	Name            string           `json:"name"`
	Manufacturer    string           `json:"manufacturer"`
	DisplayType     string           `json:"displayType"`
	MainComponentID string           `json:"mainComponentId"`
	Status          StatusType       `json:"status"`
	Components      []Component      `json:"components"`
	Attributes      []ThingAttribute `json:"attributes,omitempty"`
}

func (t *Thing) Verify() error {
	if t.DisplayType == "" {
		return fmt.Errorf("displayType must not be empty")
	}

	if len(t.Components) == 0 {
		return fmt.Errorf("thing has no components")
	}

	if t.MainComponentID == "" {
		return fmt.Errorf("mainComponentID must not be empty")
	}

	mainComponentFound := false

	for _, component := range t.Components {
		if err := component.Verify(); err != nil {
			return err
		}
		if component.ID == t.MainComponentID {
			mainComponentFound = true
		}
	}
	if !mainComponentFound {
		return fmt.Errorf("main component does not exist")
	}
	return nil
}

func (c *Component) Verify() error {
	if c.ID == "" {
		return fmt.Errorf("component has no valid id")
	}

	if !urlConform.MatchString(c.ID) {
		return fmt.Errorf("componentID should match \"^[a-zA-Z0-9-_]{1,200}$\"")
	}

	if c.ComponentType == "" {
		return fmt.Errorf("component has no component type")
	}

	if len(c.Properties) == 0 && len(c.Actions) == 0 {
		return fmt.Errorf("Component has no properties or actions")
	}

	existingProperties := make(map[string]bool)
	for _, property := range c.Properties {
		if err := property.Verify(); err != nil {
			return err
		}

		if _, ok := existingProperties[strings.ToLower(property.ID)]; ok {
			return fmt.Errorf("property ids have to be unique within one component")
		}

		existingProperties[strings.ToLower(property.ID)] = true
	}

	existingActions := make(map[string]bool)
	for _, action := range c.Actions {
		if err := action.Verify(); err != nil {
			return err
		}

		if _, ok := existingActions[strings.ToLower(action.ID)]; ok {
			return fmt.Errorf("action ids have to be unique within a component")
		}

		existingActions[strings.ToLower(action.ID)] = true
	}

	return nil
}

func (p *Property) Verify() error {
	if p.ID == "" {
		return fmt.Errorf("one or more property ids are missing")
	} else if !urlConform.MatchString(p.ID) {
		return fmt.Errorf("at least one property id contains invalid characters. Allowed is a-Z, 0-9, -, _")
	} else if err := verifyString(p.Name); err != nil {
		return err
	}

	return nil
}

func (a *Action) Verify() error {
	if a.ID == "" {
		return fmt.Errorf("empty action ids are not allowed")
	} else if !urlConform.MatchString(a.ID) {
		return fmt.Errorf("at least one action id contains invalid characters. Allowed is a-Z, 0-9, -, _")
	}
	return nil
}

// verifyString checks for invalid user input
func verifyString(input string) error {
	if !utf8.ValidString(input) {
		return fmt.Errorf("at least one given string contains invalid characters")
	}
	return nil
}

// ThingAttribute defines a key value pair that can be stored within a thing
type ThingAttribute struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// StatusType defines thing status
type StatusType string

// Component is part of a thing
type Component struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	ComponentType string     `json:"componentType"`
	Capabilities  []string   `json:"capabilities"`
	Properties    []Property `json:"properties,omitempty"`
	Actions       []Action   `json:"actions,omitempty"`
}

// Property is part of a thing component
type Property struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Value        string    `json:"value"`
	Unit         string    `json:"unit"`
	Type         ValueType `json:"type"`
	LastUpdate   time.Time `json:"lastUpdate"`
	PropertyType string    `json:"propertyType"`
}

// ValueType defines the type of a value
type ValueType string

// Action is part of a thing component
type Action struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Parameters []ActionParameter `json:"parameters,omitempty"`
}

// ActionParameter define which parameters a thing action requires
type ActionParameter struct {
	Name string    `json:"name"`
	Type ValueType `json:"type"`
}

// definition of value and status types
const (
	ValueTypeNumber  = ValueType("NUMBER")
	ValueTypeString  = ValueType("STRING")
	ValueTypeBoolean = ValueType("BOOLEAN")

	StatusTypeUnknown     = StatusType("UNKNOWN")
	StatusTypeAvailable   = StatusType("AVAILABLE")
	StatusTypeUnavailable = StatusType("UNAVAILABLE")
)

var (
	// AllValueTypes declares list of possible value types
	AllValueTypes = map[ValueType]struct{}{
		ValueTypeNumber:  {},
		ValueTypeString:  {},
		ValueTypeBoolean: {},
	}

	// AllStatusTypes defines list of possible status types
	AllStatusTypes = map[StatusType]struct{}{
		StatusTypeUnknown:     {},
		StatusTypeAvailable:   {},
		StatusTypeUnavailable: {},
	}
)
