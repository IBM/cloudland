// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// ListFlavorsOKBody list flavors o k body
//
// swagger:model listFlavorsOKBody
type ListFlavorsOKBody struct {

	// flavors
	// Required: true
	Flavors Flavors `json:"flavors"`
}

// Validate validates this list flavors o k body
func (m *ListFlavorsOKBody) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateFlavors(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *ListFlavorsOKBody) validateFlavors(formats strfmt.Registry) error {

	if err := validate.Required("flavors", "body", m.Flavors); err != nil {
		return err
	}

	if err := m.Flavors.Validate(formats); err != nil {
		if ve, ok := err.(*errors.Validation); ok {
			return ve.ValidateName("flavors")
		}
		return err
	}

	return nil
}

// ContextValidate validate this list flavors o k body based on the context it is used
func (m *ListFlavorsOKBody) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if err := m.contextValidateFlavors(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *ListFlavorsOKBody) contextValidateFlavors(ctx context.Context, formats strfmt.Registry) error {

	if err := m.Flavors.ContextValidate(ctx, formats); err != nil {
		if ve, ok := err.(*errors.Validation); ok {
			return ve.ValidateName("flavors")
		}
		return err
	}

	return nil
}

// MarshalBinary interface implementation
func (m *ListFlavorsOKBody) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *ListFlavorsOKBody) UnmarshalBinary(b []byte) error {
	var res ListFlavorsOKBody
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
