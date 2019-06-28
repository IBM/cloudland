// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// SubnetAllocationPoolsItems subnet allocation pools items
// swagger:model subnetAllocationPoolsItems
type SubnetAllocationPoolsItems struct {

	// end
	// Format: ipv4
	End strfmt.IPv4 `json:"end,omitempty"`

	// start
	// Format: ipv4
	Start strfmt.IPv4 `json:"start,omitempty"`
}

// Validate validates this subnet allocation pools items
func (m *SubnetAllocationPoolsItems) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateEnd(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateStart(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *SubnetAllocationPoolsItems) validateEnd(formats strfmt.Registry) error {

	if swag.IsZero(m.End) { // not required
		return nil
	}

	if err := validate.FormatOf("end", "body", "ipv4", m.End.String(), formats); err != nil {
		return err
	}

	return nil
}

func (m *SubnetAllocationPoolsItems) validateStart(formats strfmt.Registry) error {

	if swag.IsZero(m.Start) { // not required
		return nil
	}

	if err := validate.FormatOf("start", "body", "ipv4", m.Start.String(), formats); err != nil {
		return err
	}

	return nil
}

// MarshalBinary interface implementation
func (m *SubnetAllocationPoolsItems) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *SubnetAllocationPoolsItems) UnmarshalBinary(b []byte) error {
	var res SubnetAllocationPoolsItems
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}