package queue

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/go-redis/redis"
)

type MsgHandler struct {
}

func (m *MsgHandler) HandleMessage(msg *Message) error {
	fmt.Println(msg)
	return nil
}

func TestRedisQueue(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		DB:       0,
		Password: "",
	})

	// msg produce
	p := NewProducer(client)
	err := p.Publish("task", []byte("this is a test task body"))
	t.Log(err)

	// msg consume
	h := &MsgHandler{}
	ctx := context.Background()
	c := NewConsumer(ctx, client, "task")
	c.SetHandler(h)

	wg := sync.WaitGroup{}
	wg.Add(1)

	if ctx.Err() != nil {
		wg.Done()
	}
	wg.Wait()

	t.Log("done")
}

func TestRedisDelayQueue(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		DB:       0,
		Password: "",
	})

	// msg produce
	p := NewProducer(client)
	err := p.PublishDelayMsg("task_delay", []byte("this is a test delay task body"), time.Second*5)
	t.Log(err)

	err = p.PublishDelayMsg("task_delay", []byte("this is a test delay task body2"), time.Second*10)
	t.Log(err)

	// msg consume
	ctx := context.Background()
	hd := &MsgHandler{}
	cd := NewConsumer(ctx, client, "task_delay")
	cd.SetHandler(hd)

	pushMsg := func() {
		for {
			r := time.Duration(rand.Intn(5) + 5)
			ticker := time.NewTicker(time.Second * r)
			select {
			case <-ticker.C:
				delay := time.Second * r
				body := fmt.Sprintf("this is a test delay task body:%v %v", delay, time.Now().Unix())
				err = p.PublishDelayMsg("task_delay",
					[]byte(body), time.Duration(delay))
				fmt.Printf("%v push test body:%v\n", time.Now().Format("2006/01/02 15:04:05"), body)
			}
		}
	}
	go pushMsg()

	wg := sync.WaitGroup{}
	wg.Add(1)

	if ctx.Err() != nil {
		wg.Done()
	}
	wg.Wait()

	t.Log("done")
}
