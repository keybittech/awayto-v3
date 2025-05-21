package clients

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"

	"github.com/redis/go-redis/v9"
)

var defaultTrackDuration, _ = time.ParseDuration("86400s")

const socketServerConnectionsKey = "socket_server_connections"

type Redis struct {
	RedisClient *redis.Client
}

func (r *Redis) Client() *redis.Client {
	return r.RedisClient
}

func InitRedis() *Redis {

	redisPass, err := util.GetEnvFilePath("REDIS_PASS_FILE", 128)
	if err != nil {
		util.ErrorLog.Println(util.ErrCheck(err))
	}

	rc := redis.NewClient(&redis.Options{
		Addr:     util.E_REDIS_URL,
		Password: redisPass,
		DB:       0,
		Protocol: 3,
	})

	r := &Redis{
		RedisClient: rc,
	}

	r.InitKeys(context.Background())

	util.DebugLog.Println("Redis Init")
	return r
}

func ParticipantTopicsKey(topic string) (string, error) {
	if topic == "" {
		return "", util.ErrCheck(errors.New("malformed topic"))
	}
	return "participant_topics:" + topic, nil
}

func SocketIdTopicsKey(socketId string) (string, error) {
	if socketId == "" {
		return "", util.ErrCheck(errors.New("malformed topic"))
	}
	return "socket_id:" + socketId + ":topics", nil
}

func (r *Redis) InitKeys(ctx context.Context) {
	_, err := r.Client().Del(ctx, socketServerConnectionsKey).Result()
	if err != nil {
		panic(err)
	}
}

func (r *Redis) InitRedisSocketConnection(ctx context.Context, socketId string) error {
	finish := util.RunTimer()
	defer finish()
	_, err := r.Client().SAdd(ctx, socketServerConnectionsKey, socketId).Result()
	if err != nil {
		return util.ErrCheck(err)
	}

	return nil
}

func (r *Redis) HandleUnsub(ctx context.Context, socketId string) (map[string]string, error) {
	finish := util.RunTimer()
	defer finish()

	removedTopics := make(map[string]string)

	r.Client().SRem(ctx, socketServerConnectionsKey, socketId)

	socketIdTopicsKey, err := SocketIdTopicsKey(socketId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	participantTopics, err := r.Client().SMembers(ctx, socketIdTopicsKey).Result()
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	for _, participantTopic := range participantTopics {

		r.Client().SRem(ctx, participantTopic, socketId)

		socketIds, err := r.Client().SMembers(ctx, participantTopic).Result()
		if err != nil {
			continue
		}

		var targets string

		for _, socketId := range socketIds {
			if len(socketId) < 37 {
				continue
			}
			connId := socketId[37:]

			if strings.Index(targets, connId) == -1 {
				targets += connId
			}
		}

		if len(targets) == 0 {
			continue
		}

		topic := participantTopic[strings.Index(participantTopic, ":"):strings.LastIndex(participantTopic, ":")]

		removedTopics[topic] = targets[:len(targets)-1]
	}

	r.Client().Del(ctx, socketIdTopicsKey)

	return removedTopics, nil
}

func (r *Redis) RemoveTopicFromConnection(ctx context.Context, socketId, topic string) error {
	finish := util.RunTimer()
	defer finish()
	participantTopicsKey, err := ParticipantTopicsKey(topic)
	if err != nil {
		return util.ErrCheck(err)
	}

	_, err = r.Client().SRem(ctx, participantTopicsKey, socketId).Result()
	if err != nil {
		return util.ErrCheck(err)
	}

	socketIdTopicsKey, err := SocketIdTopicsKey(socketId)
	if err != nil {
		return util.ErrCheck(err)
	}

	_, err = r.Client().SRem(ctx, socketIdTopicsKey, participantTopicsKey).Result()
	if err != nil {
		return util.ErrCheck(err)
	}

	return nil
}

func (r *Redis) GetCachedParticipants(ctx context.Context, topic string, targetsOnly bool) (map[string]*types.SocketParticipant, string, error) {
	finish := util.RunTimer()
	defer finish()
	participantTopicsKey, err := ParticipantTopicsKey(topic)
	if err != nil {
		return nil, "", util.ErrCheck(err)
	}

	topicSocketIds, err := r.Client().SMembers(ctx, participantTopicsKey).Result()
	if err != nil {
		return nil, "", util.ErrCheck(err)
	}

	sps := make(map[string]*types.SocketParticipant, len(topicSocketIds))

	var participantTargets strings.Builder

	// socketId should be a userSub uuid of 36 characters, plus a colon, plus another connId uuid, so 37 is ok to check
	for _, socketId := range topicSocketIds {
		if len(socketId) < 37 {
			continue
		}
		connId := socketId[37:]
		participantTargets.Write([]byte(connId))
		if targetsOnly {
			continue
		}

		userSub := socketId[0:36]

		if participant, ok := sps[userSub]; ok {
			participant.Cids = append(participant.Cids, connId)
		} else {
			sps[userSub] = &types.SocketParticipant{
				Scid:   userSub,
				Cids:   []string{connId},
				Online: true,
			}
		}
	}

	return sps, participantTargets.String(), nil
}

func (r *Redis) TrackTopicParticipant(ctx context.Context, topic, socketId string) error {
	finish := util.RunTimer()
	defer finish()
	participantTopicsKey, err := ParticipantTopicsKey(topic)
	if err != nil {
		return util.ErrCheck(err)
	}

	err = r.Client().SAdd(ctx, participantTopicsKey, socketId).Err()
	if err != nil {
		return util.ErrCheck(err)
	}

	err = r.Client().Expire(ctx, participantTopicsKey, defaultTrackDuration).Err()
	if err != nil {
		return util.ErrCheck(err)
	}

	socketIdTopicsKey, err := SocketIdTopicsKey(socketId)
	if err != nil {
		return util.ErrCheck(err)
	}

	err = r.Client().SAdd(ctx, socketIdTopicsKey, participantTopicsKey).Err()
	if err != nil {
		return util.ErrCheck(err)
	}

	return nil
}

// New function to check if a user is already subscribed to a topic
func (r *Redis) HasTracking(ctx context.Context, topic, socketId string) (bool, error) {
	finish := util.RunTimer()
	defer finish()
	// Get the key for this socket's subscribed topics
	socketIdTopicsKey, err := SocketIdTopicsKey(socketId)
	if err != nil {
		return false, util.ErrCheck(err)
	}

	// Get the participant topics key for this topic
	participantTopicsKey, err := ParticipantTopicsKey(topic)
	if err != nil {
		return false, util.ErrCheck(err)
	}

	// Check if this socket is already subscribed to this topic
	isMember, err := r.Client().SIsMember(ctx, socketIdTopicsKey, participantTopicsKey).Result()
	if err != nil {
		// On error, return false (safer to do the DB check)
		return false, util.ErrCheck(err)
	}

	return isMember, nil
}
