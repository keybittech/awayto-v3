package clients

import (
	"encoding/json"
	"slices"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (db *Database) InitDBSocketConnection(tx IDatabaseTx, userSub string, connId string) (func(), error) {
	_, err := tx.Exec(`
		INSERT INTO dbtable_schema.sock_connections (created_sub, connection_id)
		VALUES ($1::uuid, $2)
	`, userSub, connId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return func() {
		err := db.TxExec(func(itx IDatabaseTx) error {
			_, txErr := itx.Exec(`
				DELETE FROM dbtable_schema.sock_connections
				USING dbtable_schema.sock_connections sc
				LEFT OUTER JOIN dbtable_schema.topic_messages tm ON tm.connection_id = sc.connection_id
				WHERE dbtable_schema.sock_connections.id = sc.id AND tm.id IS NULL AND sc.connection_id = $1 
			`, connId)
			if txErr != nil {
				return util.ErrCheck(err)
			}
			return nil
		}, "worker", "", "")
		if err != nil {
			util.ErrorLog.Println(err)
		}
	}, nil
}

func (db *Database) GetSocketAllowances(tx IDatabaseTx, userSub string) ([]util.IdStruct, error) {

	rows, err := tx.Query(`
		SELECT b.id
		FROM dbtable_schema.bookings b
		JOIN dbtable_schema.schedule_bracket_slots sbs ON sbs.id = b.schedule_bracket_slot_id
		JOIN dbtable_schema.quotes q ON q.id = b.quote_id
		WHERE sbs.created_sub = $1 OR q.created_sub = $1
	`, userSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	defer rows.Close()

	ids := []util.IdStruct{}

	for rows.Next() {
		var r util.IdStruct
		err := rows.Scan(&r.Id)

		if err != nil {
			return nil, util.ErrCheck(err)
		}

		ids = append(ids, r)
	}

	return ids, nil
}

func (db *Database) GetTopicMessageParticipants(tx IDatabaseTx, topic string) (SocketParticipants, error) {

	err := tx.SetDbVar("sock_topic", topic)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	participants := make(SocketParticipants)

	topicRows, err := tx.Query(`
		SELECT
			created_sub as scid,
			JSONB_AGG(connection_id) as cids
		FROM dbtable_schema.topic_messages
		WHERE topic = $1
		GROUP BY created_sub
	`, topic)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	defer topicRows.Close()

	for topicRows.Next() {
		var scid string
		var cids []string
		var cidsBytes []byte

		err = topicRows.Scan(&scid, &cidsBytes)
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		err = json.Unmarshal(cidsBytes, &cids)
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		if participant, ok := participants[scid]; ok {
			for _, cid := range cids {
				if !slices.Contains(participant.Cids, cid) {
					participant.Cids = append(participant.Cids, cid)
				}
			}
		} else {
			participants[scid] = &types.SocketParticipant{
				Scid: scid,
				Cids: cids,
			}
		}
	}

	err = tx.SetDbVar("sock_topic", "")
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return participants, nil
}

func (db *Database) GetSocketParticipantDetails(tx IDatabaseTx, participants SocketParticipants) (SocketParticipants, error) {

	for userSub, details := range participants {
		// Get user anon info
		err := tx.QueryRow(`
			SELECT
				LEFT(u.first_name, 1) || LEFT(u.last_name, 1) as name,
				r.name as role
			FROM dbtable_schema.users u
			JOIN dbtable_schema.group_users gu ON gu.user_id = u.id
			JOIN dbtable_schema.group_roles gr ON gr.external_id = gu.external_id
			JOIN dbtable_schema.roles r ON r.id = gr.role_id
			WHERE u.sub = $1
		`, userSub).Scan(&details.Name, &details.Role)
		if err != nil {
			return nil, util.ErrCheck(err)
		}
	}

	return participants, nil
}

func (db *Database) StoreTopicMessage(tx IDatabaseTx, connId string, message *SocketMessage) error {

	message.Store = false
	message.Historical = true
	message.Timestamp = time.Now().Local().String()

	_, err := tx.Exec(`
		INSERT INTO dbtable_schema.topic_messages (created_sub, topic, message, connection_id)
		SELECT created_sub, $2, $3, $1
		FROM dbtable_schema.sock_connections
		WHERE connection_id = $1
	`, connId, message.Topic, GenerateMessage(util.DefaultPadding, message))

	if err != nil {
		return util.ErrCheck(err)
	}

	return nil
}

func (db *Database) GetTopicMessages(tx IDatabaseTx, topic string, page, pageSize int) ([][]byte, error) {

	messages := make([][]byte, pageSize)

	paginatedQuery := util.WithPagination(`
		SELECT message FROM dbtable_schema.topic_messages
		WHERE topic = $1
		ORDER BY created_on DESC
	`, page, pageSize)

	err := tx.SetDbVar("sock_topic", topic)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	rows, err := tx.Query(paginatedQuery, topic)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	defer rows.Close()

	i := 0
	for rows.Next() {
		err := rows.Scan(&messages[i])
		if err != nil {
			return nil, util.ErrCheck(err)
		}
		i++
	}

	err = tx.SetDbVar("sock_topic", "")
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if messages[pageSize-1] != nil {
		messages = append(messages, GenerateMessage(util.DefaultPadding, &SocketMessage{
			Topic:  topic,
			Action: types.SocketActions_HAS_MORE_MESSAGES,
		}))
	}

	return messages, nil
}
