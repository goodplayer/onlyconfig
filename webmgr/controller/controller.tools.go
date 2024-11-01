package controller

import (
	"context"
	"log"
	"net/http"

	"xorm.io/xorm"

	"github.com/goodplayer/onlyconfig/webmgr/storage/dbtxn"
)

type TxnStatus int

const (
	TxnStatusCommit   TxnStatus = 1
	TxnStatusRollback TxnStatus = 2
)

type RenderFn func(http.ResponseWriter, *http.Request)

type TxnController struct {
	Engine *xorm.Engine
}

func (tc *TxnController) RunInTxn(w http.ResponseWriter, r *http.Request, fn func(ctx context.Context) (RenderFn, TxnStatus)) {
	inner := func() (result RenderFn) {
		txnMgr := &dbtxn.TxnMgr{
			Engine: tc.Engine,
		}
		ctx, err := txnMgr.StartTxn(context.Background())
		if err != nil {
			log.Println("error:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer func(txnMgr *dbtxn.TxnMgr, ctx context.Context) {
			err := txnMgr.FinalizeTxn(ctx)
			if err != nil {
				log.Println("error:", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}(txnMgr, ctx)

		var txnStatus TxnStatus
		result, txnStatus = fn(ctx)

		switch txnStatus {
		case TxnStatusCommit:
			if err := txnMgr.CommitTxn(ctx); err != nil {
				log.Println("commit txn error:", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		case TxnStatusRollback:
			if err := txnMgr.RollbackTxn(ctx); err != nil {
				log.Println("rollback txn error:", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		return
	}

	if result := inner(); result != nil {
		result(w, r)
	}
}
