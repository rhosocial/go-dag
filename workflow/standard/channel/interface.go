// Copyright (c) 2023 - 2024. vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

// Package channel helps the workflow manage the transit channel.
package channel

// Transit represents a transit entity with name, listening channels, and sending channels.
type Transit struct {
	Name              string
	ListeningChannels []string
	SendingChannels   []string
}

// NewTransit creates a new Transit with the given name, listening channels, and sending channels.
func NewTransit(name string, listeningChannels, sendingChannels []string) *Transit {
	return &Transit{
		Name:              name,
		ListeningChannels: listeningChannels,
		SendingChannels:   sendingChannels,
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

	// Create edges based on listening and sending channels
	for _, transit := range transits {
		for _, listeningChannel := range transit.ListeningChannels {
			for _, t := range transits {
				if t.Name != transit.Name && contains(t.SendingChannels, listeningChannel) {
					nodeMap[transit.Name].Incoming = append(nodeMap[transit.Name].Incoming, t.Name)
					nodeMap[t.Name].Outgoing = append(nodeMap[t.Name].Outgoing, transit.Name)
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
