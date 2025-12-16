package resource

import (
	"user-service/pkg/config"
	"user-service/pkg/kafka"
	"user-service/pkg/manager"
)

// KafkaResource wraps the global Kafka client into a Resource.
type KafkaResource struct{}

// KafkaResourcePlugin registers Kafka as a resource plugin.
type KafkaResourcePlugin struct{}

func (p *KafkaResourcePlugin) Name() string { return "kafka" }

func (p *KafkaResourcePlugin) MustCreateResource() manager.Resource { return &KafkaResource{} }

func (r *KafkaResource) MustOpen() {
	cfg := config.GetGlobalConfig()
	if cfg == nil || !cfg.Kafka.Enabled {
		return
	}
	kafka.DefaultClient().MustOpen()
	// Ensure follow events topic exists; ignore error in dev environments.
	_ = kafka.DefaultClient().EnsureTopic(kafka.FollowEventsTopic, 3, 1)
}

func (r *KafkaResource) Close() {
	kafka.DefaultClient().Close()
}
