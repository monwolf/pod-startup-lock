/*
 * Copyright 2018, Oath Inc.
 * Licensed under the terms of the MIT license. See LICENSE file in the project root for terms.
 */

package hashi

import (
	"log"

	"github.com/hashicorp/nomad/api"
	. "github.com/monwolf/pod-startup-lock/common/util"
	. "github.com/monwolf/pod-startup-lock/hashi-health/config"
)

type Client struct {
	nomadClient *api.Client
}

func NewClient(appConfig Config) *Client {
	config := api.DefaultConfig()
	nomadcli, err := api.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}
	return &Client{nomadClient: nomadcli}

}

// Get Node ID from nodeName
func (c *Client) GetNodeId(nodeName string) string {
	nodes, _, err := c.nomadClient.Nodes().List(&api.QueryOptions{})
	if err != nil {
		log.Fatalf("Problem retrieving node information: %s", err.Error())
	}
	for _, node := range nodes {
		if node.Name == nodeName {
			return node.ID
		}
	}
	log.Fatal("Problem retrieving node information: node not found")
	return ""
}

func (c *Client) GetNodeInfo(nodeId string) *api.Node {
	node := (*RetryOrPanicDefault(func() (interface{}, error) {
		node, _, err := c.nomadClient.Nodes().Info(nodeId, nil)
		return node, err
	})).(*api.Node)
	return node
}

func (c *Client) GetSystemJobs(namespace string) []*api.JobListStub {
	systemJobList := (*RetryOrPanicDefault(func() (interface{}, error) {
		joblist, _, err := c.nomadClient.Jobs().List(&api.QueryOptions{Namespace: namespace, Filter: `Type == "system"`})
		if err != nil {
			return nil, err
		}
		// Workarround for old nomads.
		var jl []*api.JobListStub
		for _, job := range joblist {
			//log.Printf("Nomad Job Name: %s Nomad type: %s ", job.Name, job.Type)
			if job.Type == "system" {
				jl = append(jl, job)
			}
		}
		return jl, nil
	})).([]*api.JobListStub)
	return systemJobList
}

func (c *Client) GetNodeAllocations(nodeId string) []*api.AllocationListStub {

	allocationList := (*RetryOrPanicDefault(func() (interface{}, error) {
		allocations, _, err := c.nomadClient.Allocations().List(&api.QueryOptions{Filter: `NodeID == "` + nodeId + `"`})
		if err != nil {
			return nil, err
		}

		// Workarround for old nomads.
		var jl []*api.AllocationListStub
		for _, job := range allocations {
			if job.NodeID == nodeId {
				jl = append(jl, job)
			}
		}
		return jl, nil
	})).([]*api.AllocationListStub)
	return allocationList
}
