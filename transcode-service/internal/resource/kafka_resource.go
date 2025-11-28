package resource

import (
	"transcode-service/pkg/kafka"
	"transcode-service/pkg/manager"
)

type KafkaResource struct{}

type KafkaResourcePlugin struct{}

func (p *KafkaResourcePlugin) Name() string { return "kafka" }

func (p *KafkaResourcePlugin) MustCreateResource() manager.Resource { return &KafkaResource{} }

func (r *KafkaResource) MustOpen() { kafka.DefaultClient().MustOpen() }

func (r *KafkaResource) Close() { kafka.DefaultClient().Close() }
