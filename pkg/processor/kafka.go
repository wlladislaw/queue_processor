package processor

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync/atomic"
	"time"

	"github.com/IBM/sarama"
)

type Kafka struct {
	producer      sarama.SyncProducer
	groupConsumer sarama.ConsumerGroup
}

func NewKafka(brokers []string, retryMax int) (Processor, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Retry.Max = retryMax
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Version = sarama.V2_8_0_0
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRange()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	group := "queue_processor"

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		log.Fatalln("err init kafka producer:", err)
	}

	consumerGroup, err := sarama.NewConsumerGroup(brokers, group, config)
	if err != nil {
		log.Fatalf("err init consumer group: %v", err)
	}

	return &Kafka{
		producer:      producer,
		groupConsumer: consumerGroup,
	}, nil
}

func (k *Kafka) Publish(topic string, data []byte) error {
	msg := sarama.ByteEncoder(data)
	message := &sarama.ProducerMessage{
		Topic: topic,
		Value: msg,
	}

	partition, offset, err := k.producer.SendMessage(message)
	if err != nil {
		log.Println("err kafka producer send :", err)
		return err
	}

	fmt.Printf("messg producer: '%s' [ partition=%d, offset=%d ]\n", msg, partition, offset)
	return nil
}

func (k *Kafka) Consume(topic string) (Tuber, error) {
	topics := []string{topic}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		cancel()
	}()

	tasksCh := make(chan Task)
	tube := &KafkaTube{
		TubeName: topic,
		Tasks:    tasksCh,
	}
	handler := ConsumerGroupHandler{
		out:  tasksCh,
		tube: tube,
	}
	go func() {
		defer close(tasksCh)
		for {
			if err := k.groupConsumer.Consume(ctx, topics, handler); err != nil {
				log.Printf("err group consume: %v", err)
				time.Sleep(2 * time.Second)
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()
	return tube, nil
}

func (k *Kafka) Close() error {
	errProducer := k.producer.Close()
	if errProducer != nil {
		return errProducer
	}
	errConsumer := k.groupConsumer.Close()
	if errConsumer != nil {
		return errConsumer
	}
	return nil
}

type KafkaTube struct {
	TubeName string
	Tasks    <-chan Task
	Claim    sarama.ConsumerGroupClaim
	offset   int64
}

func (rt *KafkaTube) GetTasks() <-chan Task {
	return rt.Tasks
}
func (rt *KafkaTube) GetTubeName() string {
	return rt.TubeName
}
func (rt *KafkaTube) GetMesgsCount() int {
	if rt.Claim != nil {
		offset := atomic.LoadInt64(&rt.offset)
		return int(rt.Claim.HighWaterMarkOffset() - offset) - 1
	}
	return 0
}

type KafkaTask struct {
	sess      sarama.ConsumerGroupSession
	msg       *sarama.ConsumerMessage
	createdAt time.Time
}

func (kt *KafkaTask) Ack(bool) error {
	kt.sess.MarkMessage(kt.msg, "")
	return nil
}
func (kt *KafkaTask) Nack(bool, bool) error {
	kt.sess.MarkMessage(kt.msg, "")
	return nil
}
func (kt *KafkaTask) Body() []byte {
	return kt.msg.Value
}
func (kt *KafkaTask) GetCreatedAt() time.Time {
	return kt.createdAt
}

type ConsumerGroupHandler struct {
	out  chan Task
	tube *KafkaTube
}

func (ConsumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (ConsumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (h ConsumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		h.tube.Claim = claim
		atomic.StoreInt64(&h.tube.offset, msg.Offset)

		fmt.Printf("Messg: topic=%s partition=%d offset=%d value=%s\n",
			msg.Topic, msg.Partition, msg.Offset, string(msg.Value))
		h.out <- &KafkaTask{
			createdAt: msg.Timestamp,
			sess:      sess,
			msg:       msg,
		}
	}

	return nil
}
