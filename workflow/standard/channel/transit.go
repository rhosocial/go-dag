// Copyright (c) 2023 - 2024. vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

// Package channel helps the workflow manage the transit channel.
package channel

// Transit represents a transit entity with name, incoming channels, and outgoing channels.
type Transit interface {
	Node
	AppendListeningChannels(key string, value ...string)
	AppendSendingChannels(key string, value ...string)
}

type SimpleTransit struct {
	ListeningChannels map[string][]string
	SendingChannels   map[string][]string
	Node
}

func (t *SimpleTransit) AppendListeningChannels(key string, value ...string) {
	t.ListeningChannels[key] = append(t.ListeningChannels[key], value...)
}

func (t *SimpleTransit) AppendSendingChannels(key string, value ...string) {
	t.SendingChannels[key] = append(t.SendingChannels[key], value...)
}

// NewTransit creates a new Transit with the given name, incoming channels, and outgoing channels.
func NewTransit(name string, incoming, outgoing []string) Transit {
	return &SimpleTransit{
		Node:              NewNode(name, incoming, outgoing),
		ListeningChannels: make(map[string][]string),
		SendingChannels:   make(map[string][]string),
	}
}

// BuildGraphFromTransits constructs a Graph from a list of Transits.
func BuildGraphFromTransits(sourceName, sinkName string, transits ...Transit) (*Graph, error) {
	transitMap := make(map[string]Transit)
	nodeMap := make(map[string]Node)

	// Create a map of Transit for easy lookup.
	for _, transit := range transits {
		transitMap[transit.GetName()] = transit
	}

	// Initialize nodes from transits
	for _, transit := range transits {
		nodeMap[transit.GetName()] = NewNode(transit.GetName(), []string{}, []string{})
	}

	// Create edges based on incoming and outgoing channels
	for _, transit := range transits {
		for _, incomingChannel := range transit.GetIncoming() {
			for _, t := range transits {
				if t.GetName() != transit.GetName() && contains(t.GetOutgoing(), incomingChannel) {
					nodeMap[transit.GetName()].AppendIncoming(t.GetName())
					nodeMap[t.GetName()].AppendOutgoing(transit.GetName())
					transit.AppendListeningChannels(t.GetName(), incomingChannel)
					t.AppendSendingChannels(transit.GetName(), incomingChannel)
				}
			}
		}
	}

	// Convert nodeMap to slice
	var nodes []Node
	for _, node := range nodeMap {
		nodes = append(nodes, node)
	}

	// Create the graph
	return NewGraph(sourceName, sinkName, nodes...)
}

// Utility function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}