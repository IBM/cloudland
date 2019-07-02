/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package dbs

import (
	context "context"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var (
	locker          = sync.Mutex{}
	dbm             *gorm.DB
	testMode        = strings.HasSuffix(os.Args[0], ".test") || os.Getenv("db.testing") != ""
	objects         = []interface{}{}
	migrationErrors = map[string]string{}
	upgradeErrors   = map[string]string{}
	grades          = map[string]func(*gorm.DB) error{}
	needToUpgrade   = false
	needToMigrate   = false
	OpenDB          = openDB
)

func openDB() (db *gorm.DB) {
	dbType := cfg.GetType()
	dbUrl := cfg.GetUri()
	dbDebug := cfg.GetDebug()

	if testMode {
		dbType = "sqlite3"
		dbUrl = "file::memory:?cache=shared"
	}
	if dbType == "" {
		dbType = "sqlite3"
		dbUrl = "cland.db"
	}
	var err error
	if db, err = gorm.Open(dbType, dbUrl); err != nil {
		panic(err)
	}

	if testMode || dbDebug {
		db.LogMode(true)
	}

	// SetMaxIdleConns sets the maximum number of connections
	// in the idle connection pool.
	idle := cfg.GetIdle()
	db.DB().SetMaxIdleConns(idle)

	// SetMaxOpenConns sets the maximum number of open connections
	// to the database.
	open := cfg.GetOpen()
	db.DB().SetMaxOpenConns(open)

	// SetConnMaxLifetime set max connection lifetime(in minite)
	lifetime := cfg.GetLifetime()
	db.DB().SetConnMaxLifetime(time.Minute * time.Duration(lifetime))
	return db
}

func newDB() *gorm.DB {
	locker.Lock()
	defer locker.Unlock()
	if dbm == nil {
		dbm = OpenDB()
	}
	doAutoMigrate(dbm)
	doAutoUpgrade(dbm)
	return dbm
}

func AutoMigrate(values ...interface{}) {
	locker.Lock()
	defer locker.Unlock()
	objects = append(objects, values...)
	needToMigrate = true
}

func doAutoMigrate(db *gorm.DB) {
	logger, _ := startLogging(context.Background(), "doAutoMigrate")
	defer logger.Finish()
	if needToMigrate {
		names := tableNames(db)
		for i := 0; i < len(objects); i++ {
			obj := objects[i]
			name := names[i]
			err := db.AutoMigrate(obj).Error
			if err != nil {
				logger.Error(err)
				msg := err.Error()
				if s, ok := migrationErrors[name]; ok {
					migrationErrors[name] = fmt.Sprintf("%s\n%s", s, msg)
				} else {
					migrationErrors[name] = msg
				}
			}
		}
		needToMigrate = false
	}
}

func AutoUpgrade(name string, grade func(*gorm.DB) error) {
	locker.Lock()
	defer locker.Unlock()
	grades[name] = grade
	needToUpgrade = true
}

func doAutoUpgrade(db *gorm.DB) (err error) {
	logger, _ := startLogging(context.Background(), "doAutoUpgrade")
	defer logger.Finish()
	if !needToUpgrade || len(grades) == 0 {
		return
	}
	names := []string{}
	for name, _ := range grades {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		grade := grades[name]
		if grade == nil { // skip nil
			continue
		}
		if err = grade(db); err != nil {
			logger.Error(err)
			upgradeErrors[name] = err.Error()
			continue
		}
	}
	needToUpgrade = false
	return
}

func DB() *gorm.DB {
	if dbm != nil {
		if needToMigrate || needToUpgrade {
			locker.Lock()
			if needToMigrate {
				doAutoMigrate(dbm)
			}
			if needToUpgrade {
				doAutoUpgrade(dbm)
			}
			locker.Unlock()
		}
		return dbm
	}
	return newDB()
}

func SetDB(db *gorm.DB) {
	locker.Lock()
	defer locker.Unlock()
	if db == dbm {
		return
	}
	needToMigrate = true
	needToUpgrade = true
	dbm = db
}

func TableNames() (names []string) {
	if dbm == nil {
		return
	}
	names = tableNames(dbm)
	return
}

func tableNames(db *gorm.DB) (names []string) {
	for i := 0; i < len(objects); i++ {
		obj := objects[i]
		names = append(names, db.NewScope(obj).TableName())
	}
	return
}
