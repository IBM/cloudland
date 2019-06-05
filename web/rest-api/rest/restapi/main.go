package main

import (
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

func main() {

	schemaLoader := gojsonschema.NewStringLoader(getSchema())
	documentLoader := gojsonschema.NewStringLoader(getDoc())

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		panic(err.Error())
	}

	if result.Valid() {
		fmt.Printf("The document is valid\n")
	} else {
		fmt.Printf("The document is not valid. see errors :\n")
		for _, desc := range result.Errors() {
			fmt.Printf("- %s\n", desc)
		}
	}
}

func getSchema() string {
	return `{
	  "definitions": {},
	  "$schema": "http://json-schema.org/draft-06/schema#",
	  "$id": "http://example.com/root.json",
	  "type": "object",
	  "title": "The Root Schema",
	  "required": [
	    "auth"
	  ],
	  "properties": {
	    "auth": {
	      "$id": "#/properties/auth",
	      "type": "object",
	      "title": "The Auth Schema",
	      "required": [
	        "identity"
	      ],
	      "properties": {
	        "identity": {
	          "$id": "#/properties/auth/properties/identity",
	          "type": "object",
	          "title": "The Identity Schema",
	          "required": [
	            "methods",
	            "password"
	          ],
	          "properties": {
	            "methods": {
	              "$id": "#/properties/auth/properties/identity/properties/methods",
	              "type": "array",
	              "title": "The Methods Schema",
	              "items": {
	                "$id": "#/properties/auth/properties/identity/properties/methods/items",
	                "type": "string",
	                "title": "The Items Schema",
	                "default": "",
	                "examples": [
	                  "password"
	                ],
	                "pattern": "^(.*)$"
	              }
	            },
	            "password": {
	              "$id": "#/properties/auth/properties/identity/properties/password",
	              "type": "object",
	              "title": "The Password Schema",
	              "required": [
	                "user"
	              ],
	              "properties": {
	                "user": {
	                  "$id": "#/properties/auth/properties/identity/properties/password/properties/user",
	                  "type": "object",
	                  "title": "The User Schema",
	                  "required": [
	                    "name",
	                    "domain"
	                  ],
	                  "properties": {
	                    "name": {
	                      "$id": "#/properties/auth/properties/identity/properties/password/properties/user/properties/name",
	                      "type": "string",
	                      "title": "The Name Schema",
	                      "default": "",
	                      "examples": [
	                        "admin"
	                      ],
	                      "pattern": "^(.*)$"
	                    },
	                    "domain": {
	                      "$id": "#/properties/auth/properties/identity/properties/password/properties/user/properties/domain",
	                      "type": "object",
	                      "title": "The Domain Schema",
	                      "required": [
	                        "id",
	                        "name",
	                        "password"
	                      ],
	                      "properties": {
	                        "id": {
	                          "$id": "#/properties/auth/properties/identity/properties/password/properties/user/properties/domain/properties/id",
	                          "type": "string",
	                          "title": "The Id Schema",
	                          "default": "",
	                          "examples": [
	                            "default"
	                          ],
	                          "pattern": "^(.*)$"
	                        },
	                        "name": {
	                          "$id": "#/properties/auth/properties/identity/properties/password/properties/user/properties/domain/properties/name",
	                          "type": "string",
	                          "title": "The Name Schema",
	                          "default": "",
	                          "examples": [
	                            "default"
	                          ],
	                          "pattern": "^(.*)$"
	                        },
	                        "password": {
	                          "$id": "#/properties/auth/properties/identity/properties/password/properties/user/properties/domain/properties/password",
	                          "type": "string",
	                          "title": "The Password Schema",
	                          "default": "",
	                          "examples": [
	                            "adminpwd"
	                          ],
	                          "pattern": "^(.*)$"
	                        }
	                      }
	                    }
	                  }
	                }
	              }
	            }
	          }
	        }
	      }
	    }
	  }
	}`
}

func getDoc() string {
	return `{
	    "auth": {
	      "identity": {
	      "methods": [
	        "password"
	      ],
	      "password": {
	        "user": {
	          "name": "admin",
	          "domain": {
	            "id": "default",
	            "name": "default"
	          }
	        }
	      }
	    }
	  }
	}`
}
