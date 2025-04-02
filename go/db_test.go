package main

import (
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func setupDb() {
	println("did setup db")
}

var selectAdminRoleIdSQL = `SELECT id FROM dbtable_schema.roles WHERE name = 'Admin'`

func BenchmarkDbDefaultExec(b *testing.B) {
	reset(b)
	for c := 0; c < b.N; c++ {
		api.Handlers.Database.Client().Exec(selectAdminRoleIdSQL)
	}
}

func BenchmarkDbDefaultQuery(b *testing.B) {
	reset(b)
	for c := 0; c < b.N; c++ {
		var adminRoleId string
		api.Handlers.Database.Client().QueryRow(selectAdminRoleIdSQL).Scan(&adminRoleId)
	}
}

func BenchmarkDbDefaultTx(b *testing.B) {
	reset(b)
	for c := 0; c < b.N; c++ {
		var adminRoleId string
		tx, _ := api.Handlers.Database.Client().Begin()
		defer tx.Rollback()
		tx.QueryRow(selectAdminRoleIdSQL).Scan(&adminRoleId)
		tx.Commit()
	}
}

func BenchmarkDbTxExec(b *testing.B) {
	reset(b)
	for c := 0; c < b.N; c++ {
		var adminRoleId string
		api.Handlers.Database.TxExec(func(tx clients.IDatabaseTx) error {
			var txErr error
			txErr = tx.QueryRow(selectAdminRoleIdSQL).Scan(&adminRoleId)
			if txErr != nil {
				return util.ErrCheck(txErr)
			}

			return nil
		}, "worker", "", "")
	}
}
