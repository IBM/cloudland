/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package common

import (
	"log"

	"github.com/gin-gonic/gin"
)

type BaseReference struct {
	ID   string `json:"id" binding:"omitempty,uuid"`
	Name string `json:"name" binding:"omitempty,min=2,max=32"`
}

type BaseID struct {
	ID   string `json:"id" binding:"required,uuid"`
}

type APIError struct {
	//	InternalErr error
	//	ErrorCode int `json:"error_code"`
	ErrorMessage string `json:"error_message"`
}

func ErrorResponse(c *gin.Context, code int, errorMsg string, err error) {
	log.Printf("%s, %v\n", errorMsg, err)
	if err != nil {
		errorMsg = errorMsg + ": " + err.Error()
	}
	c.JSON(code, &APIError{ErrorMessage: errorMsg})
	return
}
