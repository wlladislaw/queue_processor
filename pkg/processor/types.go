package processor
import "time"

type Task interface {
	Ack(multiple bool) error
	Nack(multiple bool, requeue bool) error
	Body() []byte
	GetCreatedAt() time.Time
}

type Tuber interface {
	GetTubeName() string
	GetTasks() <-chan Task
	GetMesgsCount() int
}

type Processor interface {
	Publish(qName string, data []byte) error
	Consume(qName string) (Tuber, error)
	Close() error
}
