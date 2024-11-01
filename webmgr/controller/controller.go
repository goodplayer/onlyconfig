package controller

import (
	"sync/atomic"

	"github.com/go-chi/chi/v5"
	"xorm.io/xorm"

	"github.com/goodplayer/onlyconfig/webmgr/config"
	"github.com/goodplayer/onlyconfig/webmgr/domains"
)

type ControllerContainer struct {
	CfgVal *atomic.Value

	UserHandler *domains.UserHandler
}

func AddControllers(r *chi.Mux, engine *xorm.Engine) {
	//FIXME Jwt secret store preparation
	//FIXME Initializing from OnlyConfig server
	cfg := &config.WebManagerConfig{
		JwtSecrets: [][]byte{[]byte("12345678123456781234567812345678"), []byte("87654321876543218765432187654321")},
	}
	cfgVal := new(atomic.Value)
	cfgVal.Store(cfg)

	cc := &ControllerContainer{
		CfgVal: cfgVal,
	}

	cc.addIndex(r)
	cc.addUserControllers(r, engine)
	cc.addConfigureControllers(r, engine)
}
