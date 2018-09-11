package main

import (
	"log"
	"strings"
	"time"

	"github.com/gospotcheck/aiven"
	"github.com/hashicorp/terraform/helper/resource"
)

// KafkaTopicChangeWaiter is used to refresh the Aiven Kafka Topic endpoints when
// provisioning.
type KafkaTopicChangeWaiter struct {
	Client      *aiven.Client
	Project     string
	ServiceName string
	Topic       string
}

// RefreshFunc will call the Aiven client and refresh it's state.
func (w *KafkaTopicChangeWaiter) RefreshFunc() resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		topic, err := w.Client.KafkaTopics.Get(
			w.Project,
			w.ServiceName,
			w.Topic,
		)

		if err != nil {
			// Handle this special case as it takes a while for topics to be created.
			log.Printf("[DEBUG] Got %#v error while waiting for topic to be up.", err)
			if strings.Compare(err.Error(), "Topic '"+w.Topic+"' does not exist") == 0 {
				return nil, "CONFIGURING", nil
			}
			return nil, "", err
		}

		log.Printf("[DEBUG] Got %s state while waiting for topic to be up.", topic.State)

		return topic, topic.State, nil
	}
}

// Conf sets up the configuration to refresh.
func (w *KafkaTopicChangeWaiter) Conf() *resource.StateChangeConf {
	state := &resource.StateChangeConf{
		Pending: []string{"CONFIGURING"},
		Target:  []string{"ACTIVE"},
		Refresh: w.RefreshFunc(),
	}
	state.Delay = 10 * time.Second
	state.Timeout = 10 * time.Minute
	state.MinTimeout = 2 * time.Second
	return state
}
