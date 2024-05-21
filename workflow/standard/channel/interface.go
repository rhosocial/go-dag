// Copyright (c) 2023 - 2024. vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

// Package channel helps the workflow manage the transit channel.
package channel

// Transit represents a transit entity with name, incoming channels, and outgoing channels.
type Transit struct {
	Node
	ListeningChannels map[string][]string
	SendingChannels   map[string][]string
}

// NewTransit creates a new Transit with the given name, incoming channels, and outgoing channels.
func NewTransit(name string, incoming, outgoing []string) *Transit {
	return &Transit{
		Node:              *NewNode(name, incoming, outgoing),
		ListeningChannels: make(map[string][]string),
		SendingChannels:   make(map[string][]string),
	}
}

// BuildGraphFromTransits constructs a Graph from a list of Transits.
func BuildGraphFromTransits(transits []*Transit, sourceName, sinkName string) (*Graph, error) {
	transitMap := make(map[string]*Transit)
	nodeMap := make(map[string]*Node)

	// Create a map of Transit for easy lookup.
	for _, transit := range transits {
		transitMap[transit.Name] = transit
	}

	// Initialize nodes from transits
	for _, transit := range transits {
		nodeMap[transit.Name] = NewNode(transit.Name, []string{}, []string{})
	}

	// Create edges based on incoming and outgoing channels
	for _, transit := range transits {
		for _, incomingChannel := range transit.Incoming {
			for _, t := range transits {
				if t.Name != transit.Name && contains(t.Outgoing, incomingChannel) {
					nodeMap[transit.Name].Incoming = append(nodeMap[transit.Name].Incoming, t.Name)
					nodeMap[t.Name].Outgoing = append(nodeMap[t.Name].Outgoing, transit.Name)
					transit.ListeningChannels[t.Name] = append(transit.ListeningChannels[t.Name], incomingChannel)
					t.SendingChannels[transit.Name] = append(t.SendingChannels[transit.Name], incomingChannel)
				}
			}
		}
	}

	// Convert nodeMap to slice
	var nodes []*Node
	for _, node := range nodeMap {
		nodes = append(nodes, node)
	}

	// Create the graph
	return NewGraph(sourceName, sinkName, nodes)
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
