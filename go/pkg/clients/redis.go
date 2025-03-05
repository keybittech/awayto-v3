package clients

import (
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"context"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	RedisClient IRedisClient
}

type SocketParticipants map[string]*types.SocketParticipant

func InitRedis() IRedis {

	redisPass, err := util.EnvFile(os.Getenv("REDIS_PASS_FILE"))
	if err != nil {
		util.ErrorLog.Println(util.ErrCheck(err))
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_URL"),
		Password: redisPass,
		DB:       0,
		Protocol: 3,
	})

	connLen := 0

	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				ctx := context.Background()
				defer ctx.Done()
				socketConnections, err := redisClient.SMembers(ctx, "socket_server_connections").Result()
				sockLen := len(socketConnections)
				if err != nil {
					println("reading socket connection err", err.Error())
				}
				if connLen != sockLen {
					connLen = sockLen
					fmt.Printf("got socket connection list new count :%d %+v\n", len(socketConnections), socketConnections)
				}
			}
		}
	}()

	r := &Redis{
		RedisClient: redisClient,
	}
	r.InitKeys()

	return r
}

var socketServerConnectionsKey = "socket_server_connections"

func ParticipantTopicsKey(topic string) string {
	return fmt.Sprintf("participant_topics:%s", topic)
}

func SocketIdTopicsKey(socketId string) string {
	return fmt.Sprintf("socket_id:%s:topics", socketId)
}

func (r *Redis) Client() IRedisClient {
	return r.RedisClient
}

func (r *Redis) InitKeys() {
	ctx := context.Background()
	defer ctx.Done()
	_, err := r.Client().Del(ctx, socketServerConnectionsKey).Result()
	if err != nil {
		panic(err)
	}
}

func (r *Redis) InitRedisSocketConnection(socketId string) error {
	ctx := context.Background()
	defer ctx.Done()

	_, err := r.Client().SAdd(ctx, socketServerConnectionsKey, socketId).Result()
	if err != nil {
		return util.ErrCheck(err)
	}

	return nil
}

func (r *Redis) HandleUnsub(socketId string) (map[string][]string, error) {
	ctx := context.Background()
	defer ctx.Done()
	removedTopics := make(map[string][]string)

	r.Client().SRem(ctx, socketServerConnectionsKey, socketId)

	socketIdTopicsKey := SocketIdTopicsKey(socketId)

	participantTopics, err := r.Client().SMembers(ctx, socketIdTopicsKey).Result()
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	for _, participantTopic := range participantTopics {

		r.Client().SRem(ctx, participantTopic, socketId)

		socketIds, err := r.Client().SMembers(ctx, participantTopic).Result()
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		topic := strings.SplitN(participantTopic, ":", 2)[1]

		targets := []string{}

		for _, socketId := range socketIds {
			_, connId, err := util.SplitSocketId(socketId)
			if err != nil {
				util.ErrCheck(err)
				continue
			}

			if !slices.Contains(targets, connId) {
				targets = append(targets, connId)
			}
		}

		removedTopics[topic] = targets
	}

	r.Client().Del(ctx, socketIdTopicsKey)

	return removedTopics, nil
}

func (r *Redis) RemoveTopicFromConnection(socketId, topic string) error {
	ctx := context.Background()
	defer ctx.Done()

	participantTopicsKey := ParticipantTopicsKey(topic)

	_, err := r.Client().SRem(ctx, participantTopicsKey, socketId).Result()
	if err != nil {
		return util.ErrCheck(err)
	}

	socketIdTopicsKey := SocketIdTopicsKey(socketId)

	_, err = r.Client().SRem(ctx, socketIdTopicsKey, participantTopicsKey).Result()
	if err != nil {
		return util.ErrCheck(err)
	}

	return nil
}

func (r *Redis) GetCachedParticipants(ctx context.Context, topic string) SocketParticipants {
	participantTopicsKey := ParticipantTopicsKey(topic)
	topicSocketIds, err := r.Client().SMembers(ctx, participantTopicsKey).Result()
	if err != nil {
		util.ErrCheck(err)
		return nil
	}

	sps := make(SocketParticipants)

	for _, socketId := range topicSocketIds {
		userSub, connId, err := util.SplitSocketId(socketId)
		if err != nil {
			util.ErrCheck(err)
			continue
		}

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

	return sps
}

func (r *Redis) GetParticipantTargets(participants SocketParticipants) []string {
	var topicCids []string
	for _, participant := range participants {
		topicCids = append(topicCids, participant.Cids...)
	}
	return topicCids
}

func (r *Redis) TrackTopicParticipant(ctx context.Context, topic, socketId string) {
	participantTopicsKey := ParticipantTopicsKey(topic)

	err := r.Client().SAdd(ctx, participantTopicsKey, socketId).Err()
	if err != nil {
		util.ErrCheck(err)
	}

	defaultDuration, _ := time.ParseDuration("86400s")
	err = r.Client().Expire(ctx, participantTopicsKey, defaultDuration).Err()
	if err != nil {
		util.ErrCheck(err)
	}

	socketIdTopicsKey := SocketIdTopicsKey(socketId)
	err = r.Client().SAdd(ctx, socketIdTopicsKey, participantTopicsKey).Err()
	if err != nil {
		util.ErrCheck(err)
	}
}
