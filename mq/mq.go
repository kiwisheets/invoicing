package mq

import (
	"fmt"

	c_mq "github.com/cheshir/go-mq"
	"github.com/kiwisheets/util"
	"github.com/sirupsen/logrus"
)

type MQ struct {
	mq             c_mq.MQ
	CreateProducer c_mq.SyncProducer
	RenderProducer c_mq.SyncProducer
}

func (m *MQ) Close() {
	m.mq.Close()
}

func setupProducers(m *MQ) {
	var err error
	m.CreateProducer, err = m.mq.SyncProducer("invoice_create")
	if err != nil {
		panic(fmt.Errorf("failed to create producer: invoice_create: %s", err))
	}

	m.RenderProducer, err = m.mq.SyncProducer("invoice_render")
	if err != nil {
		panic(fmt.Errorf("failed to create producer: invoice_render: %s", err))
	}
}

func setupConsumers(m *MQ) {

}

func Init() *MQ {
	m := MQ{
		mq: util.NewMQ(),
	}
	setupConsumers(&m)
	setupProducers(&m)

	go handleMQErrors(m.mq.Error())

	return &m
}

func handleMQErrors(errors <-chan error) {
	for err := range errors {
		logrus.Errorf("mq error: %s", err)
	}
}
