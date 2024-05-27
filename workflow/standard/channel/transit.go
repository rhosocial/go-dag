// Copyright (c) 2023 - 2024. vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

// Package channel helps the workflow manage the transit channel.
package channel

// Transit represents a transit entity with name, incoming channels, and outgoing channels.
//
// Unlike Node, the relationship between a Transit is defined by a listening and sending channel.
// Therefore, a BuildGraphFromTransits build graph is required.
type Transit interface {
	// Node is responsible for saving and handling relationships with Transit.
	Node

	// AppendListeningChannels appends a listening channel name(s) to the current node.
	//
	// key refers to the name of the listening node, and value refers to the channel name of the node.
	AppendListeningChannels(key string, value ...string)

	// AppendSendingChannels appends a sending channel name(s) to the current node.
	//
	// key refers to the name of the sending node, and value refers to the channel name of the node.
	AppendSendingChannels(key string, value ...string)
}

// SimpleTransit is a simple implementation of the Transit interface.
type SimpleTransit struct {
	ListeningChannels map[string][]string
	SendingChannels   map[string][]string
	Node
}

// AppendListeningChannels appends a listening channel name(s) to the current node.
//
// key refers to the name of the listening node, and value refers to the channel name of the node.
func (t *SimpleTransit) AppendListeningChannels(key string, value ...string) {
	t.ListeningChannels[key] = append(t.ListeningChannels[key], value...)
}

// AppendSendingChannels appends a sending channel name(s) to the current node.
//
// key refers to the name of the sending node, and value refers to the channel name of the node.
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
	nodeMap := make(map[string]Node)

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
