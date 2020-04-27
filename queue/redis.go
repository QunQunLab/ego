package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/satori/go.uuid"

	"github.com/QunQunLab/ego/log"
)

const (
	listSuffix = ":list"
	zsetSuffix = ":zset"
)

type Message struct {
	ID        string `json:"id"`
	Body      []byte `json:"body"`
	Timestamp int64  `json:"timestamp"`
	DelayTime int64  `json:"delayTime"`
}

func (m Message) String() string {
	return fmt.Sprintf("ID:%v body:[%v] t:%v dt:%v", m.ID, string(m.Body), m.Timestamp, m.DelayTime)
}

func NewMessage(id string, body []byte) *Message {
	if id == "" {
		guid, _ := uuid.NewV4()
		id = guid.String()
	}
	return &Message{
		ID:        id,
		Body:      body,
		Timestamp: time.Now().Unix(),
		DelayTime: time.Now().Unix(),
	}
}

type Handler interface {
	HandleMessage(msg *Message) error
}

type consumer struct {
	once            sync.Once
	redisCmd        redis.Cmdable
	ctx             context.Context
	topicName       string
	handler         Handler
	rateLimitPeriod time.Duration
	options         ConsumerOptions
}

type ConsumerOptions struct {
	RateLimitPeriod time.Duration
}

type Consumer = *consumer

func (s *consumer) SetHandler(handler Handler) {
	s.once.Do(func() {
		s.processQueueMsg()
		s.processDelayQueueMsg()
	})
	s.handler = handler
}

func (s *consumer) processQueueMsg() {
	go func() {
		ticker := time.NewTicker(s.options.RateLimitPeriod)
		defer func() {
			log.Info("TOPIC:%v stop process queue msg.", s.topicName)
			ticker.Stop()
		}()
		topicName := s.topicName + listSuffix
		for {
			select {
			case <-s.ctx.Done():
				log.Info("TOPIC:%v context Done msg: %#v \n", s.topicName, s.ctx.Err())
				return
			case <-ticker.C:
				// first check handler
				if s.handler == nil {
					continue
				}

				revBody, err := s.redisCmd.LPop(topicName).Bytes()
				// first check redis nil
				if err == redis.Nil {
					continue
				}
				if err != nil {
					log.Error("TOPIC:%v LPop error: %#v", s.topicName, err.Error())
					continue
				}
				if len(revBody) == 0 {
					continue
				}

				msg := &Message{}
				err = json.Unmarshal(revBody, msg)
				if err != nil {
					log.Error("TOPIC:%v unmarshal msg:[%v] err:%v", s.topicName, string(revBody), err.Error())
					continue
				}
				handleMsg := func() {
					log.Info("TOPIC:%v process msg:[%v]", s.topicName, msg)
					err = s.handler.HandleMessage(msg)
					if err != nil {
						log.Error("TOPIC:%v process msgID:%v done with err:%v", s.topicName, msg.ID, err.Error())
					}
				}
				go handleMsg()
			}
		}
	}()
}

func (s *consumer) processDelayQueueMsg() {
	go func() {
		ticker := time.NewTicker(s.options.RateLimitPeriod)
		defer func() {
			log.Info("TOPIC:%v stop process msg.", s.topicName)
			ticker.Stop()
		}()
		topicName := s.topicName + zsetSuffix
		for {
			currentTime := time.Now().UnixNano() / 1000 / 1000
			select {
			case <-s.ctx.Done():
				log.Error("TOPIC:%v context Done msg: %#v", s.topicName, s.ctx.Err())
				return
			case <-ticker.C:
				// first check handler
				if s.handler == nil {
					continue
				}

				var valuesCmd *redis.ZSliceCmd
				_, err := s.redisCmd.TxPipelined(func(pip redis.Pipeliner) error {
					valuesCmd = pip.ZRangeByScoreWithScores(topicName,
						&redis.ZRangeBy{
							Min: "0",
							Max: strconv.FormatInt(currentTime, 10),
						})
					pip.ZRemRangeByScore(topicName, "0", strconv.FormatInt(currentTime, 10))
					return nil
				})
				if err != nil {
					log.Error("TOPIC:%v zset pip error: %#v", s.topicName, err.Error())
					continue
				}

				rev := valuesCmd.Val()
				for _, revBody := range rev {
					msg := &Message{}
					err := json.Unmarshal([]byte(revBody.Member.(string)), msg)
					if err != nil {
						log.Error("TOPIC:%v unmarshal msg:[%v] err:%v", s.topicName, revBody.Member.(string), err.Error())
						continue
					}

					handleMsg := func() {
						log.Info("TOPIC:%v process msg:%v", s.topicName, msg)
						err = s.handler.HandleMessage(msg)
						if err != nil {
							log.Error("TOPIC:%v process msgID:%v done with err:%v", s.topicName, msg.ID, err.Error())
						}
					}
					go handleMsg()
				}
			}
		}
	}()
}

func NewConsumer(ctx context.Context, redisCmd redis.Cmdable, topicName string, op ...ConsumerOptions) Consumer {
	consumer := &consumer{
		redisCmd:  redisCmd,
		ctx:       ctx,
		topicName: topicName,
	}
	if consumer.options.RateLimitPeriod == 0 {
		consumer.options.RateLimitPeriod = time.Millisecond * 200
	}
	if len(op) > 0 && op[0].RateLimitPeriod > 0 {
		consumer.options.RateLimitPeriod = op[0].RateLimitPeriod
	}
	return consumer
}

type Producer struct {
	redisCmd redis.Cmdable
}

func (p *Producer) Publish(topicName string, body []byte) error {
	msg := NewMessage("", body)
	sendData, _ := json.Marshal(msg)
	return p.redisCmd.RPush(topicName+listSuffix, string(sendData)).Err()
}

func (p *Producer) PublishDelayMsg(topicName string, body []byte, delay time.Duration) error {
	if delay <= 0 {
		return errors.New("delay need great than zero")
	}
	tm := time.Now().Add(delay)
	msg := NewMessage("", body)
	msg.DelayTime = tm.Unix()

	sendData, _ := json.Marshal(msg)
	score := float64(tm.UnixNano() / 1000 / 1000)
	return p.redisCmd.ZAdd(
		topicName+zsetSuffix,
		&redis.Z{
			Score:  score,
			Member: string(sendData)},
	).Err()
}

func NewProducer(cmd redis.Cmdable) *Producer {
	return &Producer{redisCmd: cmd}
}
