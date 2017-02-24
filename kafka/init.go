package kafka

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Shopify/sarama"
)

var (
	kAsyncProducer sarama.AsyncProducer
	kClient        sarama.Client
)

const WALType = "WAL"

var (
	traceTopic  string
	topicSuffix string
	instanceId  string
	host        string
	walTopic    string
)

// related stream processor:
// https://github.com/xtaci/sp
type WAL struct {
	Type       string      `json:"type"`
	InstanceId string      `json:"instanceId"`
	Table      string      `json:"table"`
	Host       string      `json:"host"`
	Key        string      `json:"key"`
	CreatedAt  time.Time   `json:"created_at"`
	Data       interface{} `json:"data"`
}

func initKafka(brokers []string, waltopic, tracetopic, id string) {
	addrs := brokers        //c.StringSlice("kafka-brokers")
	walTopic = waltopic     // c.String("wal-topic")
	traceTopic = tracetopic //c.String("trace-topic")
	instanceId = id         // c.String("id")
	host, _ = os.Hostname()
	topicSuffix = "_" + id // c.String("id")

	config := sarama.NewConfig()
	config.Producer.Return.Successes = false
	config.Producer.Return.Errors = false
	producer, err := sarama.NewAsyncProducer(addrs, config)
	if err != nil {
		log.Fatalln(err)
	}

	kAsyncProducer = producer
	cli, err := sarama.NewClient(addrs, nil)
	if err != nil {
		log.Fatalln(err)
	}
	kClient = cli
}

func Init(brokers []string, waltopic, tracetopic, id string) {
	initKafka(brokers, waltopic, tracetopic, id)
}

// Trace user events
func Trace(content map[string]*json.RawMessage) {
	if bts, err := json.Marshal(&content); err == nil {
		msg := &sarama.ProducerMessage{Topic: traceTopic, Value: sarama.ByteEncoder(bts)}
		kAsyncProducer.Input() <- msg
	} else {
		log.Println(err)
	}
}

// WAL
func CommitUpdate(key, data interface{}, tblname string) {
	wal := &WAL{}
	wal.Type = WALType
	wal.InstanceId = instanceId
	wal.Table = tblname
	wal.Host = host
	wal.Data = data
	wal.Key = fmt.Sprint(key)
	wal.CreatedAt = time.Now()

	// for log compaction
	kafkaKey := fmt.Sprintf("%v-%v-%v-%v-%v", wal.Type, wal.InstanceId, wal.Table, wal.Host, wal.Key)

	if bts, err := json.Marshal(wal); err == nil {
		msg := &sarama.ProducerMessage{Topic: walTopic, Key: sarama.StringEncoder(kafkaKey), Value: sarama.ByteEncoder(bts)}
		kAsyncProducer.Input() <- msg
	} else {
		log.Println(err)
	}
}

func NewConsumer() (sarama.Consumer, error) {
	return sarama.NewConsumerFromClient(kClient)
}
