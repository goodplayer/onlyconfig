package domain_tests

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/goodplayer/onlyconfig/webmgr/config"
	"github.com/goodplayer/onlyconfig/webmgr/domains"
	"github.com/goodplayer/onlyconfig/webmgr/storage/dbtxn"
)

func TestLoginUser(t *testing.T) {
	txnMgr := &dbtxn.TxnMgr{
		Engine: engine,
	}

	uh := &domains.UserHandler{
		UserStore: userStore,
	}
	ctx, err := txnMgr.StartTxn(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer func(txnMgr *dbtxn.TxnMgr, ctx context.Context) {
		err := txnMgr.FinalizeTxn(ctx)
		if err != nil {
			t.Fatal(err)
		}
	}(txnMgr, ctx)

	u, err := uh.LoginUser(ctx, "admin", "admin")
	if err != nil {
		t.Fatal(err)
	}
	if err := txnMgr.CommitTxn(ctx); err != nil {
		t.Fatal(err)
	}
	t.Log(*u)
}

func TestUserJwt(t *testing.T) {
	cfg := &config.WebManagerConfig{
		JwtSecrets: [][]byte{[]byte("12345678123456781234567812345678"), []byte("87654321876543218765432187654321")},
	}
	cfgVal := new(atomic.Value)
	cfgVal.Store(cfg)

	uh := &domains.UserHandler{
		Config:    cfgVal,
		UserStore: userStore,
	}

	u := &domains.User{
		UserId:   "123",
		UserName: "example_user",
		Password: "(none)",
		Name:     "username",
		Email:    "example@example.com",
	}
	token, err := u.GenerateJwtToken(cfg.JwtSecrets[1])
	if err != nil {
		t.Fatal(err)
	}
	if ok, err := uh.ValidateUserJwtToken(token); err != nil {
		t.Fatal(err)
	} else if !ok {
		t.Fatal(errors.New("token invalid"))
	}

	token2, err := u.GenerateJwtToken([]byte("11111111111111111111111111111111"))
	if err != nil {
		t.Fatal(err)
	}
	if ok, err := uh.ValidateUserJwtToken(token2); err == nil && ok {
		t.Fatal(errors.New("token2 should not be valid"))
	} else {
		t.Log(ok)
		t.Log(err)
	}
}
