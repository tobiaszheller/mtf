package pubsub

import (
	"context"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	log "github.com/sirupsen/logrus"
)

const (
	queueSize = 100
	projectID = "boring-cloud"
)

func timeNow() string {
	return time.Now().Format("15:04:05:0000")
}

func NewPubsub() *Pubsub {
	ps := &Pubsub{
		messages: make(chan proto.Message, queueSize),
	}
	ctx := context.Background()
	conn, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		panic(err)
	}

	ps.topic = conn.Topic("first-topic")
	exists, err := ps.topic.Exists(ctx)
	if err != nil {
		panic(err)
	}

	if !exists {
		ps.topic, err = conn.CreateTopic(ctx, "first-topic")
		if err != nil {
			panic(err)
		}
	}

	sub := conn.Subscription("sub1")
	ok, err := sub.Exists(ctx)
	if err != nil {
		panic(err)
	}

	if !ok {
		sub, err = conn.CreateSubscription(ctx, "sub1", pubsub.SubscriptionConfig{
			Topic:       ps.topic,
			AckDeadline: time.Second * 10,
		})
		if err != nil {
			panic(err)
		}
	}
	//	sub.ReceiveSettings.Synchronous = true
	go func() {
		err := sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
			m := any.Any{}
			err := proto.Unmarshal(msg.Data, &m)
			if err != nil {
				log.Infof("recived messge ID: %v Data:%v\n", msg.ID, string(msg.Data))
				panic(err)
			}

			dynAny := ptypes.DynamicAny{}
			if err := ptypes.UnmarshalAny(&m, &dynAny); err != nil {
				panic(err)
			}
			msg.Ack()
			ps.messages <- dynAny.Message

		})
		if err != nil {
			panic(err)
		}
	}()
	return ps
}

type Pubsub struct {
	messages chan proto.Message
	queue    []proto.Message
	topic    *pubsub.Topic
}

func (p *Pubsub) Receive(i interface{}) {
	expected, ok := i.(proto.Message)
	if !ok {
		panic("message is not a proto.Message")
	}
	for _, m := range p.queue {
		if proto.Equal(expected, m) {
			log.Info("Got expected message in queue")
			return
		}
	}

	deadlineC := time.Tick(time.Second * 30)
	for {
		select {
		case msg := <-p.messages:
			if proto.MessageName(msg) != proto.MessageName(expected) {
				log.Debug("messge don't eq queueing message")
				p.queue = append(p.queue, msg)
				continue
			}

			if proto.Equal(expected, msg) {
				log.Info("Message eq")
				return
			}

		case <-deadlineC:
			panic("timout during pubsub.receive")
		}
	}
}

func (p *Pubsub) Send(i interface{}) {
	msg, ok := i.(proto.Message)
	if !ok {
		panic("message is not a proto.Message")
	}

	ctx := context.Background()
	anyMsg, err := ptypes.MarshalAny(msg)
	if err != nil {
		panic(err)
	}

	buf, err := proto.Marshal(anyMsg)
	if err != nil {
		panic(err)
	}

	m := &pubsub.Message{
		Data: buf,
	}

	if _, err := p.topic.Publish(ctx, m).Get(ctx); err != nil {
		panic(err)
	}
}
