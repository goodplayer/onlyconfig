package domain_tests

import (
	"xorm.io/xorm"

	"github.com/goodplayer/onlyconfig/webmgr/domains"
	"github.com/goodplayer/onlyconfig/webmgr/storage/postgres"
)

var engine *xorm.Engine

var userStore domains.UserStore

func init() {
	ee, err := xorm.NewEngine("pgx", "postgres://admin:admin@192.168.1.111:15432/onlyconfig?sslmode=disable")
	if err != nil {
		panic(err)
	}
	ee.ShowSQL(true)
	engine = ee

	userStore = &postgres.UserStoreImpl{}
}
