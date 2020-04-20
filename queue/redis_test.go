package queue

import (
	"testing"
	"time"

	"github.com/go-redis/redis"
)

func TestRedisMq(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		DB:       0,
		Password: "",
	})
	p := NewProducer(client)
	err := p.Publish("task", []byte("this is a test task body"))
	t.Log(err)

	err = p.PublishDelayMsg("task_delay", []byte("this is a test delay task body"), time.Minute*2)
	t.Log(err)
}
