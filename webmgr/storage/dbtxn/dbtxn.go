package dbtxn

import (
	"context"
	"errors"
	"log"

	"xorm.io/xorm"
)

type key struct {
}

func GetTxn(ctx context.Context) *xorm.Session {
	s, ok := ctx.Value(key{}).(*xorm.Session)
	if !ok {
		return nil
	}
	return s
}

type TxnMgr struct {
	Engine *xorm.Engine
}

func (t *TxnMgr) StartTxn(ctx context.Context) (context.Context, error) {
	session := t.Engine.NewSession()
	if err := session.Begin(); err != nil {
		return nil, err
	}
	newCtx := context.WithValue(ctx, key{}, session)
	return newCtx, nil
}

func (t *TxnMgr) CommitTxn(ctx context.Context) error {
	session := GetTxn(ctx)
	if session == nil {
		return errors.New("no session")
	}
	if err := session.Commit(); err != nil {
		return err
	}
	return nil
}

func (t *TxnMgr) RollbackTxn(ctx context.Context) error {
	session := GetTxn(ctx)
	if session == nil {
		return errors.New("no session")
	}
	if err := session.Rollback(); err != nil {
		return err
	}
	return nil
}

func (t *TxnMgr) FinalizeTxn(ctx context.Context) error {
	session := GetTxn(ctx)
	if session == nil {
		return errors.New("no session")
	}
	if err := session.Rollback(); err != nil {
		log.Println("rollback failed")
	}
	return session.Close()
}
