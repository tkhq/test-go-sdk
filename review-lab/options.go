package main

import "encoding/json"

type Options struct {
	Replicas int `json:"replicas,omitempty"`
}

func NewOptions(replicas *int) Options {
	if replicas == nil {
		return Options{}
	}
	value := *replicas
	return Options{Replicas: value}
}

func WriteOptions(replicas *int) ([]byte, error) {
	return json.Marshal(NewOptions(replicas))
}
