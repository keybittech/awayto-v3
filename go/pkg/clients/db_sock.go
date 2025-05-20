package clients

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (ds DbSession) InitDbSocketConnection(ctx context.Context, connId string) error {
	finish := util.RunTimer()
	defer finish()
	_, err := ds.SessionBatchExec(ctx, `
		INSERT INTO dbtable_schema.sock_connections (created_sub, connection_id)
		VALUES ($1::uuid, $2)
	`, ds.ConcurrentUserSession.GetUserSub(), connId)
	if err != nil {
		return util.ErrCheck(err)
	}

	return nil
}

func (ds DbSession) RemoveDbSocketConnection(ctx context.Context, connId string) error {
	finish := util.RunTimer()
	defer finish()
	_, err := ds.SessionBatchExec(ctx, `
		DELETE FROM dbtable_schema.sock_connections
		USING dbtable_schema.sock_connections sc
		LEFT OUTER JOIN dbtable_schema.topic_messages tm ON tm.connection_id = sc.connection_id
		WHERE dbtable_schema.sock_connections.id = sc.id AND tm.id IS NULL AND sc.connection_id = $1 
	`, connId)
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

func (ds DbSession) GetSocketAllowances(ctx context.Context, bookingId string) (bool, error) {
	finish := util.RunTimer()
	defer finish()
	row, done, err := ds.SessionBatchQueryRow(ctx, socketAllowanceQuery,
		ds.ConcurrentUserSession.GetUserSub(),
		bookingId)
	if err != nil {
		return false, util.ErrCheck(err)
	}
	defer done()

	var allowed bool
	err = row.Scan(&allowed)
	if err != nil {
		return false, util.ErrCheck(err)
	}

	return allowed, nil
}

func (ds DbSession) GetTopicMessageParticipants(ctx context.Context, participants map[string]*types.SocketParticipant) error {
	finish := util.RunTimer()
	defer finish()
	topicRows, done, err := ds.SessionBatchQuery(ctx, `
		SELECT
			created_sub,
			ARRAY_AGG(connection_id)
		FROM dbtable_schema.topic_messages
		WHERE topic = $1
		GROUP BY created_sub
	`, ds.Topic)
	if err != nil {
		return util.ErrCheck(err)
	}
	defer done()

	defer topicRows.Close()
	for topicRows.Next() {
		var userSub string
		var cids []string

		err = topicRows.Scan(&userSub, &cids)
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

	return nil
}

func (ds DbSession) GetSocketParticipantDetails(ctx context.Context, participants map[string]*types.SocketParticipant) error {
	finish := util.RunTimer()
	defer finish()
	lenPart := len(participants)
	if lenPart == 0 {
		return nil
	}

	subs := make([]string, 0, lenPart)
	for sub := range participants {
		subs = append(subs, sub)
	}

	rows, done, err := ds.SessionBatchQuery(ctx, `
		SELECT
			u.sub,
			LEFT(u.first_name, 1) || LEFT(u.last_name, 1) as name,
			r.name as role
		FROM dbtable_schema.users u
		JOIN dbtable_schema.group_users gu ON gu.user_id = u.id
		JOIN dbtable_schema.group_roles gr ON gr.external_id = gu.external_id
		JOIN dbtable_schema.roles r ON r.id = gr.role_id
		WHERE u.sub = ANY($1)
	`, subs)
	if err != nil {
		return util.ErrCheck(err)
	}
	defer done()

	for rows.Next() {
		var sub, name, role string
		err := rows.Scan(&sub, &name, &role)
		if err != nil {
			return util.ErrCheck(err)
		}

		if participant, ok := participants[sub]; ok {
			participant.Name = name
			participant.Role = role
		}
	}

	return nil
}

func (ds DbSession) StoreTopicMessage(ctx context.Context, connId string, message *types.SocketMessage) error {
	finish := util.RunTimer()
	defer finish()
	message.Store = false
	message.Historical = true
	message.Timestamp = time.Now().Local().String()

	_, err := ds.SessionBatchExec(ctx, `
		INSERT INTO dbtable_schema.topic_messages (created_sub, connection_id, topic, message)
		VALUES ($1, $2, $3, $4)
	`, ds.ConcurrentUserSession.GetUserSub(), connId, message.Topic, util.GenerateMessage(util.DefaultPadding, message))
	if err != nil {
		return util.ErrCheck(err)
	}

	return nil
}

func (ds DbSession) GetTopicMessages(ctx context.Context, page, pageSize int) ([][]byte, error) {
	finish := util.RunTimer()
	defer finish()
	messages := make([][]byte, pageSize)

	paginatedQuery := util.WithPagination(`
		SELECT message FROM dbtable_schema.topic_messages
		WHERE topic = $1
		ORDER BY created_on DESC
	`, page, pageSize)

	rows, done, err := ds.SessionBatchQuery(ctx, paginatedQuery, ds.Topic)
	if err != nil {
		return nil, util.ErrCheck(err)
	}
	defer done()

	i := 0
	for rows.Next() {
		err := rows.Scan(&messages[i])
		if err != nil {
			return nil, util.ErrCheck(err)
		}
		i++
	}

	if messages[pageSize-1] != nil {
		messages = append(messages, util.GenerateMessage(util.DefaultPadding, &types.SocketMessage{
			Topic:  ds.Topic,
			Action: types.SocketActions_HAS_MORE_MESSAGES,
		}))
	}

	return messages, nil
}
