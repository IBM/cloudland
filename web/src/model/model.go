/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package model

import (
	"time"

	"github.com/google/uuid"

	"web/src/utils/log"
)

var logger = log.MustGetLogger("model")

type Model struct {
	ID        int64 `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time    `gorm:"index"`
	UUID      string        `gorm:"type:varchar(64)","index"`
	Creater   int64         `gorm:"default:1"` /* The user ID of the resource */
	OwnerInfo *Organization `gorm:"PRELOAD:false","foreignkey:Owner"`
}

func (m *Model) BeforeCreate() (err error) {
	if m.UUID == "" {
		m.UUID = uuid.New().String()
		logger.Debugf("Create a new model with uuid: %s", m.UUID)
	}
	return
}
