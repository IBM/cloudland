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

// PostIdentityV3AuthTokensParamsBodyAuthIdentityPasswordUserDomain post identity v3 auth tokens params body auth identity password user domain
//
// swagger:model postIdentityV3AuthTokensParamsBodyAuthIdentityPasswordUserDomain
type PostIdentityV3AuthTokensParamsBodyAuthIdentityPasswordUserDomain struct {

	// id
	// Example: default
	// Pattern: ^[A-Za-z][-A-Za-z0-9_]*$
	ID string `json:"id,omitempty"`
}

// Validate validates this post identity v3 auth tokens params body auth identity password user domain
func (m *PostIdentityV3AuthTokensParamsBodyAuthIdentityPasswordUserDomain) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateID(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *PostIdentityV3AuthTokensParamsBodyAuthIdentityPasswordUserDomain) validateID(formats strfmt.Registry) error {
	if swag.IsZero(m.ID) { // not required
		return nil
	}

	if err := validate.Pattern("id", "body", m.ID, `^[A-Za-z][-A-Za-z0-9_]*$`); err != nil {
		return err
	}

	return nil
}

// ContextValidate validates this post identity v3 auth tokens params body auth identity password user domain based on context it is used
func (m *PostIdentityV3AuthTokensParamsBodyAuthIdentityPasswordUserDomain) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *PostIdentityV3AuthTokensParamsBodyAuthIdentityPasswordUserDomain) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *PostIdentityV3AuthTokensParamsBodyAuthIdentityPasswordUserDomain) UnmarshalBinary(b []byte) error {
	var res PostIdentityV3AuthTokensParamsBodyAuthIdentityPasswordUserDomain
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
