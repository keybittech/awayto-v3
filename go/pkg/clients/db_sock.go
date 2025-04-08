package clients

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/interfaces"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (db *Database) InitDbSocketConnection(connId, userSub, groupId, roles string) error {
	err := db.TxExec(func(tx interfaces.IDatabaseTx) error {
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
	err := db.TxExec(func(itx interfaces.IDatabaseTx) error {
		_, txErr := itx.Exec(`
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

func (db *Database) GetSocketAllowances(tx interfaces.IDatabaseTx, userSub, description, handle string) (bool, error) {

	var subscribed bool
	err := tx.QueryRow(`
		SELECT EXISTS (
			SELECT 1
			FROM dbtable_schema.bookings b
			JOIN dbtable_schema.schedule_bracket_slots sbs ON sbs.id = b.schedule_bracket_slot_id
			JOIN dbtable_schema.quotes q ON q.id = b.quote_id
			WHERE b.id = $2 AND (sbs.created_sub = $1 OR q.created_sub = $1)
		)
	`, userSub, handle).Scan(&subscribed)
	if err != nil {
		return false, util.ErrCheck(err)
	}

	// defer rows.Close()
	//
	// subscribed := false
	// var r util.IdStruct
	//
	// switch description {
	// case exchangeTextNumCheck,
	// 	exchangeCallNumCheck,
	// 	exchangeWhiteboardNumCheck:
	//
	// 	for rows.Next() {
	// 		rows.Scan(&r.Id)
	// 		if r.Id == handle {
	// 			subscribed = true
	// 		}
	// 	}
	// default:
	// 	return false, nil
	// }

	return subscribed, nil
}

func (db *Database) GetTopicMessageParticipants(tx interfaces.IDatabaseTx, topic string) (map[string]*types.SocketParticipant, error) {

	err := tx.SetDbVar("sock_topic", topic)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	participants := make(map[string]*types.SocketParticipant)

	topicRows, err := tx.Query(`
		SELECT
			created_sub,
			JSONB_AGG(connection_id)
		FROM dbtable_schema.topic_messages
		WHERE topic = $1
		GROUP BY created_sub
	`, topic)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	defer topicRows.Close()
	var userSub string
	var cids []string
	var cidsBytes []byte

	for topicRows.Next() {
		err = topicRows.Scan(&userSub, &cidsBytes)
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		cids = strings.Split(string(cidsBytes), " ")

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

	err = tx.SetDbVar("sock_topic", "")
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return participants, nil
}

func (db *Database) GetSocketParticipantDetails(tx interfaces.IDatabaseTx, participants map[string]*types.SocketParticipant) (map[string]*types.SocketParticipant, error) {

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

func (db *Database) StoreTopicMessage(tx interfaces.IDatabaseTx, connId string, message *types.SocketMessage) error {

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

func (db *Database) GetTopicMessages(tx interfaces.IDatabaseTx, topic string, page, pageSize int) ([][]byte, error) {

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
		messages = append(messages, util.GenerateMessage(util.DefaultPadding, &types.SocketMessage{
			Topic:  topic,
			Action: types.SocketActions_HAS_MORE_MESSAGES,
		}))
	}

	return messages, nil
}
