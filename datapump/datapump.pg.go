package datapump

import (
	"github.com/meidoworks/nekoq-component/configure/cfgimpl"
)

type PostgresDataPump = cfgimpl.DatabaseDataPump

func NewPostgresDataPump(connStr string) *PostgresDataPump {
	return cfgimpl.NewDatabaseDataPump(connStr)
}
