package clients

import (
	"fmt"
	"slices"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"github.com/lib/pq"
)

func (db *Database) InitDbSocketConnection(connId, userSub, groupId, roles string) error {
	err := db.TxExec(func(tx PoolTx) error {
		_, err := tx.Exec(`
			INSERT INTO dbtable_schema.sock_connections (created_sub, connection_id)
			VALUES ($1::uuid, $2)
		`, userSub, connId)
		if err != nil {
			return util.ErrCheck(err)
		}
		return nil
	}, userSub, groupId, roles)
	if err != nil {
		return util.ErrCheck(err)
	}

	return nil
}

func (db *Database) RemoveDbSocketConnection(connId string) error {
	err := db.TxExec(func(tx PoolTx) error {
		txErr := db.SetDbVar("sock_topic", "")
		if txErr != nil {
			return util.ErrCheck(txErr)
		}

		_, txErr = tx.Exec(`
			DELETE FROM dbtable_schema.sock_connections
			USING dbtable_schema.sock_connections sc
			LEFT OUTER JOIN dbtable_schema.topic_messages tm ON tm.connection_id = sc.connection_id
			WHERE dbtable_schema.sock_connections.id = sc.id AND tm.id IS NULL AND sc.connection_id = $1 
		`, connId)
		if txErr != nil {
			return util.ErrCheck(txErr)
		}
		return nil
	}, "worker", "", "")
	if err != nil {
		return util.ErrCheck(err)
	}

	return nil
}

var (
	exchangeTextNumCheck       = "exchange/" + fmt.Sprint(types.ExchangeActions_EXCHANGE_TEXT.Number())
	exchangeCallNumCheck       = "exchange/" + fmt.Sprint(types.ExchangeActions_EXCHANGE_CALL.Number())
	exchangeWhiteboardNumCheck = "exchange/" + fmt.Sprint(types.ExchangeActions_EXCHANGE_WHITEBOARD.Number())
)

var socketAllowanceExec = `
	SELECT allowed
	FROM dbfunc_schema.session_query_13($1, $2, $3, $4, $5, $6)
	AS (allowed bool)
`

var socketAllowanceQuery = `
	SELECT EXISTS (
		SELECT 1 as allowed
		FROM dbtable_schema.bookings b
		JOIN dbtable_schema.schedule_bracket_slots sbs ON sbs.id = b.schedule_bracket_slot_id
		JOIN dbtable_schema.quotes q ON q.id = b.quote_id
		WHERE b.id = $2::uuid AND (sbs.created_sub = $1::uuid OR q.created_sub = $1::uuid)
		LIMIT 1
	)
`

func (db *Database) GetSocketAllowances(session *types.UserSession, bookingId string) (bool, error) {
	var allowed bool
	err := db.Client().QueryRow(
		socketAllowanceExec,
		session.UserSub,
		session.GroupId,
		session.Roles,
		socketAllowanceQuery,
		session.UserSub,
		bookingId,
	).Scan(&allowed)
	if err != nil {
		return false, fmt.Errorf("database function call failed: %w", err)
	}

	return allowed, nil
}

func (db *Database) GetTopicMessageParticipants(tx PoolTx, topic string, participants map[string]*types.SocketParticipant) error {
	err := db.SetDbVar("sock_topic", topic)
	if err != nil {
		return util.ErrCheck(err)
	}

	topicRows, err := tx.Query(`
		SELECT
			created_sub,
			ARRAY_AGG(connection_id)
		FROM dbtable_schema.topic_messages
		WHERE topic = $1
		GROUP BY created_sub
	`, topic)
	if err != nil {
		return util.ErrCheck(err)
	}

	defer topicRows.Close()
	for topicRows.Next() {
		var userSub string
		var cids []string

		err = topicRows.Scan(&userSub, (*pq.StringArray)(&cids))
		if err != nil {
			return util.ErrCheck(err)
		}

		if participant, ok := participants[userSub]; ok {
			for _, cid := range cids {
				if !slices.Contains(participant.Cids, cid) {
					participant.Cids = append(participant.Cids, cid)
				}
			}
		} else {
			participants[userSub] = &types.SocketParticipant{Scid: userSub, Cids: cids}
		}
	}

	err = db.SetDbVar("sock_topic", "")
	if err != nil {
		return util.ErrCheck(err)
	}

	return nil
}

func (db *Database) GetSocketParticipantDetails(tx PoolTx, participants map[string]*types.SocketParticipant) error {
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
			return util.ErrCheck(err)
		}
	}

	return nil
}

func (db *Database) StoreTopicMessage(tx PoolTx, connId string, message *types.SocketMessage) error {
	message.Store = false
	message.Historical = true
	message.Timestamp = time.Now().Local().String()

	_, err := tx.Exec(`
		INSERT INTO dbtable_schema.topic_messages (created_sub, topic, message, connection_id)
		SELECT created_sub, $2, $3, $1
		FROM dbtable_schema.sock_connections
		WHERE connection_id = $1
	`, connId, message.Topic, util.GenerateMessage(util.DefaultPadding, message))

	if err != nil {
		return util.ErrCheck(err)
	}

	return nil
}

func (db *Database) GetTopicMessages(tx PoolTx, topic string, page, pageSize int) ([][]byte, error) {
	messages := make([][]byte, pageSize)

	paginatedQuery := util.WithPagination(`
		SELECT message FROM dbtable_schema.topic_messages
		WHERE topic = $1
		ORDER BY created_on DESC
	`, page, pageSize)

	err := db.SetDbVar("sock_topic", topic)
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

	err = db.SetDbVar("sock_topic", "")
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if messages[pageSize-1] != nil {
		messages = append(messages, util.GenerateMessage(util.DefaultPadding, &types.SocketMessage{
			Topic:  topic,
			Action: types.SocketActions_HAS_MORE_MESSAGES,
		}))
	}

	return messages, nil
}
