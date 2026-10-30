package main

import "encoding/json"

type Options struct {
	Replicas *int `json:"replicas,omitempty"`
}

func copyReplicas(replicas *int) *int {
	if replicas == nil {
		return nil
	}
	value := *replicas
	return &value
}

func NewOptions(replicas *int) Options {
	return Options{Replicas: copyReplicas(replicas)}
}

func WriteOptions(replicas *int) ([]byte, error) {
	return json.Marshal(NewOptions(replicas))
}
