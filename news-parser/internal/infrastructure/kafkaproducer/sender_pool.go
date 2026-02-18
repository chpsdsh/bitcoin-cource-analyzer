package kafkaproducer

import (
	"context"
	"news-parser/internal/domain"
	"sync"
)

const NumKafkaSenders = 5

type Sender interface {
	sendData(dto domain.ResultDto)
}

type KafkaSender struct {
	records  chan domain.ResultDto
	producer Sender
}

func (s KafkaSender) sendDataToKafka(ctx context.Context) {
	for {
		select {
		case dto, ok := <-s.records:
			if !ok {
				return
			}
			s.producer.sendData(dto)
		case <-ctx.Done():
			return
		}
	}
}

func NewSenderPool(wg *sync.WaitGroup, ctx context.Context, records chan domain.ResultDto, producer Sender) {
	for i := 0; i < NumKafkaSenders; i++ {
		worker := KafkaSender{records: records, producer: producer}
		wg.Go(func() { worker.sendDataToKafka(ctx) })
	}
}
