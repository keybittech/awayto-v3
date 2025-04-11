package main

import (
	"database/sql"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func setupDb() {

	println("did setup db")
}

var selectAdminRoleIdSQL = `SELECT id FROM dbtable_schema.roles WHERE name = 'Admin'`

func BenchmarkDbDefaultExec(b *testing.B) {
	reset(b)
	for c := 0; c < b.N; c++ {
		mainApi.Handlers.Database.Client().Exec(selectAdminRoleIdSQL)
	}
}

func BenchmarkDbDefaultQuery(b *testing.B) {
	reset(b)
	for c := 0; c < b.N; c++ {
		var adminRoleId string
		mainApi.Handlers.Database.Client().QueryRow(selectAdminRoleIdSQL).Scan(&adminRoleId)
	}
}

func BenchmarkDbDefaultTx(b *testing.B) {
	reset(b)
	for c := 0; c < b.N; c++ {
		var adminRoleId string
		tx, _ := mainApi.Handlers.Database.Client().Begin()
		defer tx.Rollback()
		tx.QueryRow(selectAdminRoleIdSQL).Scan(&adminRoleId)
		tx.Commit()
	}
}

func BenchmarkDbTxExec(b *testing.B) {
	reset(b)
	for c := 0; c < b.N; c++ {
		var adminRoleId string
		mainApi.Handlers.Database.TxExec(func(tx *sql.Tx) error {
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
	participants := make(map[string]*types.SocketParticipant)
	topic := "test-topic"
	err := mainApi.Handlers.Database.TxExec(func(tx *sql.Tx) error {
		reset(b)
		for c := 0; c < b.N; c++ {
			mainApi.Handlers.Database.GetTopicMessageParticipants(tx, topic, participants)
		}
		return nil
	}, "", "", "")
	if err != nil {
		b.Fatal(err)
	}
}

func BenchmarkDbSocketGetSocketAllowances(b *testing.B) {
	topic := "test-topic"
	err := mainApi.Handlers.Database.TxExec(func(tx *sql.Tx) error {
		description, handle, _ := util.SplitColonJoined(topic)
		reset(b)
		for c := 0; c < b.N; c++ {
			mainApi.Handlers.Database.GetSocketAllowances(tx, "", description, handle)
		}
		return nil
	}, "", "", "")
	if err != nil {
		b.Fatal(err)
	}
}
