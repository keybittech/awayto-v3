package main

import (
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/interfaces"
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
		api.Handlers.Database.TxExec(func(tx interfaces.IDatabaseTx) error {
			var txErr error
			txErr = tx.QueryRow(selectAdminRoleIdSQL).Scan(&adminRoleId)
			if txErr != nil {
				return util.ErrCheck(txErr)
			}

			return nil
		}, "worker", "", "")
	}
}

func BenchmarkDbSocketGetTopicMessageParticipants(b *testing.B) {
	err := api.Handlers.Database.TxExec(func(tx interfaces.IDatabaseTx) error {
		reset(b)
		for c := 0; c < b.N; c++ {
			api.Handlers.Database.GetTopicMessageParticipants(tx, topic)
		}
		return nil
	}, "", "", "")
	if err != nil {
		b.Fatal(err)
	}
}

func BenchmarkDbSocketGetSocketAllowances(b *testing.B) {
	err := api.Handlers.Database.TxExec(func(tx interfaces.IDatabaseTx) error {
		description, handle, _ := util.SplitColonJoined(topic)
		reset(b)
		for c := 0; c < b.N; c++ {
			api.Handlers.Database.GetSocketAllowances(tx, session.UserSub, description, handle)
		}
		return nil
	}, "", "", "")
	if err != nil {
		b.Fatal(err)
	}
}
