/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package apis

import (
	"log"

	"github.com/gin-gonic/gin"
)

type APIError struct {
	ErrorMessage string `json:"error_message"`
}

func ErrorResponse(c *gin.Context, code int, errorMsg string, err error) {
	log.Printf("%s, %v", errorMsg, err)
	c.JSON(code, &APIError{ErrorMessage: errorMsg})
	return
}
