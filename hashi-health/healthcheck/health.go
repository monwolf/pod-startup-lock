/*
 * Copyright 2018, Oath Inc.
 * Licensed under the terms of the MIT license. See LICENSE file in the project root for terms.
 */

package healthcheck

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/nomad/api"
	"github.com/monwolf/pod-startup-lock/common/util"
	"github.com/monwolf/pod-startup-lock/hashi-health/config"
	"github.com/monwolf/pod-startup-lock/hashi-health/hashi"
)

type HealthChecker struct {
	hashi     *hashi.Client
	conf      config.Config
	node      *api.Node
	isHealthy bool
	nodeId    string
}

func NewHealthChecker(appConfig config.Config, hashi *hashi.Client) *HealthChecker {
	nodeId := hashi.GetNodeId(appConfig.NodeName)
	nodeInfo := hashi.GetNodeInfo(nodeId)
	for _, sj := range nodeInfo.Meta {
		log.Printf("Label discovered: %s", sj)
	}

	return &HealthChecker{hashi, appConfig, nodeInfo, false, nodeId}
}

func (h *HealthChecker) HealthFunction() func() bool {
	return func() bool {
		return h.isHealthy
	}
}

func (h *HealthChecker) Run() {
	for {
		if h.check() {
			log.Print("HealthCheck passed")
			h.isHealthy = true
			time.Sleep(h.conf.HealthPassTimeout)
		} else {
			log.Print("HealthCheck failed")
			h.isHealthy = false
			time.Sleep(h.conf.HealthFailTimeout)
		}
	}
}

func (h *HealthChecker) check() bool {
	log.Print("---")
	log.Print("HealthCheck:")
	systemJobs := h.hashi.GetSystemJobs(h.conf.Namespace)
	for _, sj := range systemJobs {

		log.Printf("System Job discovered: %s", sj.Name)

	}
	// This only check if nomad can schedule the jobs
	if !h.checkAllSystemJobsReady(systemJobs) {
		return false
	}

	nodeAllocations := h.hashi.GetNodeAllocations(h.nodeId)
	if len(nodeAllocations) <= 0 {
		log.Printf("No allocations found running in : '%v'", h.node.Name)
		return true
	}
	// for _, na := range nodeAllocations {
	// 	log.Print(na.JobID)
	// }

	return h.checkAllSystemJobsAllocationsAvailableOnNode(systemJobs, nodeAllocations)
}

func (h *HealthChecker) checkAllSystemJobsReady(systemJobs []*api.JobListStub) bool {
	for _, sysjob := range systemJobs {
		if required, reason := h.checkRequired(sysjob); !required {
			log.Print(reason)
			continue
		}

		status := sysjob.Status
		if status != "running" {
			log.Printf("'%v' systemJob not running: '%v'", sysjob.Name, status)
			return false
		}
		log.Printf("Job: '%v': running", sysjob.Name)
	}
	log.Print("All SystemJobs are in state running")
	return true
}

func (h *HealthChecker) checkAllSystemJobsAllocationsAvailableOnNode(systemJobs []*api.JobListStub, allocations []*api.AllocationListStub) bool {
	for _, systemJob := range systemJobs {
		if required, reason := h.checkRequired(systemJob); !required {
			log.Print(reason)
			continue
		}
		log.Printf("'%v' systemJob: Looking for allocations on node", systemJob.Name)
		allocation, found := findSystemJobAllocations(systemJob, allocations)
		if !found {
			log.Printf("'%v' systemJob: No Allocation found", systemJob.Name)
			return false
		}
		log.Printf("'%v' systemJob: Found Allocation: '%v'", systemJob.Name, allocation.Name)
		if !isAllocationReady(allocation) {
			return false
		}
	}
	log.Print("All SystemJobs Allocations available on node")
	return true
}

func (h *HealthChecker) checkRequired(sysjob *api.JobListStub) (bool, string) {
	reason := fmt.Sprintf("'%v' systemJob Excluded from healthcheck: ", sysjob.Name)
	if len(h.conf.ExcludeSystemJobs) > 0 && util.ArrayContains(h.conf.ExcludeSystemJobs, sysjob.Name) {
		return false, reason + "matches exclude job name"
	}
	if len(h.conf.IncludeSystemJobs) > 0 && !util.ArrayContains(h.conf.ExcludeSystemJobs, sysjob.Name) {
		return false, reason + "not matches job name"
	}
	nodeSelector := sysjob.Datacenters
	if !util.ArrayContains(nodeSelector, h.node.Datacenter) {
		return false, reason + "not eligible for scheduling on node"
	}
	return true, fmt.Sprintf("'%v' systemJob healthcheck required", sysjob.Name)
}

func findSystemJobAllocations(systemJob *api.JobListStub, allocations []*api.AllocationListStub) (*api.AllocationListStub, bool) {
	for _, allocation := range allocations {
		if isAllocationOwnedByJob(allocation, systemJob) {
			return allocation, true
		}
	}
	return nil, false
}

func isAllocationReady(allocation *api.AllocationListStub) bool {
	// if allocation.DeploymentStatus.Healthy == nil || *allocation.DeploymentStatus.Healthy == false {
	// 	log.Printf(" %s('%v') Allocation Not Healthy", allocation.JobID, allocation.Name)
	// 	return false
	// }

	for _, taskState := range allocation.TaskStates {
		if taskState.Failed {
			log.Printf(" %s('%v') Allocation Not Healthy", allocation.JobID, allocation.Name)
			return false
		}
	}
	return true

	// for _, cond := range allocation.Status.Conditions {
	// 	if cond.Type == "Ready" && cond.Status == "True" {
	// 		log.Printf("'%v' Pod: Ready", allocation.Name)
	// 		return true
	// 	}
	// }
}

func isAllocationOwnedByJob(allocation *api.AllocationListStub, job *api.JobListStub) bool {
	return allocation.JobID == job.ID
}
