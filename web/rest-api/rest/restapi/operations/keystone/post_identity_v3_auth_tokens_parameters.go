// Code generated by go-swagger; DO NOT EDIT.

package keystone

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"io"
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"

	models "github.com/IBM/cloudland/web/rest-api/rest/models"
)

// NewPostIdentityV3AuthTokensParams creates a new PostIdentityV3AuthTokensParams object
// no default values defined in spec.
func NewPostIdentityV3AuthTokensParams() PostIdentityV3AuthTokensParams {

	return PostIdentityV3AuthTokensParams{}
}

// PostIdentityV3AuthTokensParams contains all the bound params for the post identity v3 auth tokens operation
// typically these are obtained from a http.Request
//
// swagger:parameters PostIdentityV3AuthTokens
type PostIdentityV3AuthTokensParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*
	  Required: true
	  In: body
	*/
	Body *models.PostIdentityV3AuthTokensParamsBody
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewPostIdentityV3AuthTokensParams() beforehand.
func (o *PostIdentityV3AuthTokensParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	if runtime.HasBody(r) {
		defer r.Body.Close()
		var body models.PostIdentityV3AuthTokensParamsBody
		if err := route.Consumer.Consume(r.Body, &body); err != nil {
			if err == io.EOF {
				res = append(res, errors.Required("body", "body"))
			} else {
				res = append(res, errors.NewParseError("body", "body", "", err))
			}
		} else {
			// validate body object
			if err := body.Validate(route.Formats); err != nil {
				res = append(res, err)
			}

			if len(res) == 0 {
				o.Body = &body
			}
		}
	} else {
		res = append(res, errors.Required("body", "body"))
	}
	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
