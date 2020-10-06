/*
Copyright: 2020 Jed Record

This program is free software; you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation; Version 2 (GPLv2)

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License along
with this program; if not, write to the Free Software Foundation, Inc.,
51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.

Full license text at: https://gnu.org/licenses/gpl-2.0.txt
*/

// Package resources Data structures and functions to manipulate Kubernetes utilization resources
package resources

import (
	"fmt"
	"strings"

	"github.com/jedrecord/kutil/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Restat A resource statistic to measure
type Restat struct {
	Req   int64
	Limit int64
	Avail int64
	Cap   int64
	Util  int64
}

// Nodemetrics Node resource metrics
type Nodemetrics struct {
	Sched  bool
	Label  string
	Status string
	Taints string
	Cpu    Restat
	Mem    Restat
	Pods   Imetric
}

// Nsmetrics Namespace resource metrics
type Nsmetrics struct {
	Cpu  Restat
	Mem  Restat
	Pods Imetric
}

// Clustermetrics Cluster resource metrics
type Clustermetrics struct {
	Namespaces map[string]*Nsmetrics
	Nodes      map[string]*Nodemetrics
	Cpu        Restat
	Mem        Restat
	Pods       Imetric
}

// Imetric Holder for simple metrics
type Imetric struct {
	Inuse int64
	Avail int64
	Cap   int64
	Util  int64
}

// NewCluster constructor
func NewCluster() *Clustermetrics {
	var c Clustermetrics
	c.Namespaces = make(map[string]*Nsmetrics)
	c.Nodes = make(map[string]*Nodemetrics)
	return &c
}

// NewNodemetrics constructor
func NewNodemetrics() *Nodemetrics {
	var n Nodemetrics
	return &n
}

// NewNsmetrics constructor
func NewNsmetrics() *Nsmetrics {
	var n Nsmetrics
	return &n
}

// Load Retrieve kubernetes resource data into the Clustermetrics object
func (c *Clustermetrics) Load(cs *kubernetes.Clientset) {
	// Retrieve a list of nodes from the cluster as type nodelist
	mynodes, err := cs.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		utils.LogError("There was a problem connecting with the API")
	}
	// Loop through the nodes to collect utilization data
	if len(mynodes.Items) > 0 {
		// Node loop - Begin collecting node data
		for _, mynode := range mynodes.Items {
			n := mynode.Name
			// Loop over labels and assign node role
			var role string
			for label := range mynode.Labels {
				pair := strings.Split(label, "/")
				// Get the node-role (ie: master, worker, infra) if available
				if pair[0] == "node-role.kubernetes.io" {
					if len(role) > 0 {
						role = role + ","
					}
					role = role + pair[1]
				}
			}
			// Loop over spec.taints["effect"] = "NoSchedule"
			var nodetaint string
			var nodesched bool = true
			for _, t := range mynode.Spec.Taints {
				tkey := strings.Split(t.Key, "/")
				if tkey[0] == "node-role.kubernetes.io" || tkey[0] == "node.kubernetes.io" {
					if len(nodetaint) > 0 {
						nodetaint = nodetaint + ","
					}
					// Ommitting redundant taintlabel "master" on master nodes
					if tkey[1] == "master" {
						nodetaint = nodetaint + string(t.Effect)
					} else {
						nodetaint = nodetaint + tkey[1] + ":" + string(t.Effect)
					}
				}
				if t.Effect == "NoSchedule" || t.Effect == "NoExecute" {
					nodesched = false
				}
			}
			// Loop over status.conditions
			var nstatus string
			for _, cond := range mynode.Status.Conditions {
				if cond.Status == "True" {
					if len(nstatus) > 0 {
						nstatus = nstatus + ","
					}
					nstatus = nstatus + string(cond.Type)
				}
			}
			ndata := NewNodemetrics()
			ndata.Label = role
			ndata.Taints = nodetaint
			ndata.Sched = nodesched
			ndata.Status = nstatus
			cpuAvail := mynode.Status.Allocatable["cpu"]
			memAvail := mynode.Status.Allocatable["memory"]
			podsAvail := mynode.Status.Allocatable["pods"]
			cpuCap := mynode.Status.Capacity["cpu"]
			memCap := mynode.Status.Capacity["memory"]
			podsCap := mynode.Status.Capacity["pods"]
			ndata.Cpu.Avail = cpuAvail.MilliValue()
			ndata.Mem.Avail = memAvail.Value()
			ndata.Pods.Avail = podsAvail.Value()
			ndata.Cpu.Cap = cpuCap.MilliValue()
			ndata.Mem.Cap = memCap.Value()
			ndata.Pods.Cap = podsCap.Value()
			c.UpdateNode(n, ndata)
			c.Cpu.Cap += cpuCap.MilliValue()
			c.Mem.Cap += memCap.Value()
			c.Pods.Cap += podsCap.Value()
			// Don't increase aggregate cluster resources if this node is unschedualable
			if nodesched {
				c.Cpu.Avail += cpuAvail.MilliValue()
				c.Mem.Avail += memAvail.Value()
				c.Pods.Avail += podsAvail.Value()
			}
		}
	} else {
		utils.LogError("No nodes discovered")
	}

	// Retrieve a list of pods from the cluster as type podlist
	mypods, err := cs.CoreV1().Pods("").List(metav1.ListOptions{})
	if err != nil {
		utils.LogError("There was a problem connecting with the API")
	}
	if len(mypods.Items) > 0 {
		// Loop through the pods to collect utilization data
		for _, mypod := range mypods.Items {
			ns := mypod.Namespace
			no := mypod.Spec.NodeName

			// Initialize namespace and node data structs to hold the data
			nsdata := NewNsmetrics()
			ndata := NewNodemetrics()

			// slice to hold the names of active containers in each pod
			var activeContainers []string

			// Loop through the status of each container in the pod
			// We want to collect metrics from live containers only
			for _, cons := range mypod.Status.ContainerStatuses {
				if cons.Ready == true {
					activeContainers = append(activeContainers, cons.Name)
				}
			}
			// Loop through the container specs to collect cpu/memory requests and limits
			for _, con := range mypod.Spec.Containers {
				ok := false
				// Make sure this is an active container
				for _, a := range activeContainers {
					if a == con.Name {
						ok = true
					}
				}
				if ok {
					cpuReq := con.Resources.Requests["cpu"]
					cpuLim := con.Resources.Limits["cpu"]
					memReq := con.Resources.Requests["memory"]
					memLim := con.Resources.Limits["memory"]
					nsdata.Cpu.Req += cpuReq.MilliValue()
					nsdata.Cpu.Limit += cpuLim.MilliValue()
					nsdata.Mem.Req += memReq.Value()
					nsdata.Mem.Limit += memLim.Value()
					ndata.Cpu.Req += cpuReq.MilliValue()
					ndata.Cpu.Limit += cpuLim.MilliValue()
					ndata.Mem.Req += memReq.Value()
					ndata.Mem.Req += memLim.Value()
					c.Cpu.Req += cpuReq.MilliValue()
					c.Cpu.Limit += cpuLim.MilliValue()
					c.Mem.Req += memReq.Value()
					c.Mem.Limit += memLim.Value()
					if c.Nodes[no].Sched == false {
						c.Cpu.Avail += cpuReq.MilliValue()
						c.Mem.Avail += memReq.Value()
					}
				}
			}
			// If we've got at least 1 active container, add this pod to our pod stats
			if len(activeContainers) > 0 {
				nsdata.Pods.Inuse++
				ndata.Pods.Inuse++
				c.Pods.Inuse++
				if c.Nodes[no].Sched == false {
					c.Pods.Avail++
				}
			}
			// These update functions take the structs we just collected and update
			// the clustermetrics object (which is also as struct)
			c.UpdateNamespace(ns, nsdata)
			c.UpdateNode(no, ndata)
		}
	} else {
		utils.LogError("No pods discovered")
	}

	// Calculate totals for namespaces
	for n, m := range c.Namespaces {
		nsdata := NewNsmetrics()
		cu := utils.CalcPct(c.Cpu.Avail, m.Cpu.Req)
		mu := utils.CalcPct(c.Mem.Avail, m.Mem.Req)
		pu := utils.CalcPct(c.Pods.Avail, m.Pods.Inuse)
		nsdata.Cpu.Util = cu
		nsdata.Mem.Util = mu
		nsdata.Pods.Util = pu
		c.UpdateNamespace(n, nsdata)
	}

	// Calculate totals for nodes
	for n, m := range c.Nodes {
		ndata := NewNodemetrics()
		cu := utils.CalcPct(m.Cpu.Avail, m.Cpu.Req)
		mu := utils.CalcPct(m.Mem.Avail, m.Mem.Req)
		pu := utils.CalcPct(m.Pods.Avail, m.Pods.Inuse)
		ndata.Cpu.Util = cu
		ndata.Mem.Util = mu
		ndata.Pods.Util = pu
		c.UpdateNode(n, ndata)
	}

	// Calculate totals for the cluster
	cu := utils.CalcPct(c.Cpu.Avail, c.Cpu.Req)
	mu := utils.CalcPct(c.Mem.Avail, c.Mem.Req)
	pu := utils.CalcPct(c.Pods.Avail, c.Pods.Inuse)
	c.Cpu.Util = cu
	c.Mem.Util = mu
	c.Pods.Util = pu
}

// UpdateNamespace Adder for the Namespaces
func (c *Clustermetrics) UpdateNamespace(name string, metrics *Nsmetrics) {
	if met, ok := c.Namespaces[name]; ok {
		met.Cpu.Req += metrics.Cpu.Req
		met.Cpu.Limit += metrics.Cpu.Limit
		met.Mem.Req += metrics.Mem.Req
		met.Mem.Limit += metrics.Mem.Limit
		met.Pods.Inuse += metrics.Pods.Inuse
		if metrics.Cpu.Util > 0 {
			met.Cpu.Util = metrics.Cpu.Util
		}
		if metrics.Mem.Util > 0 {
			met.Mem.Util = metrics.Mem.Util
		}
		if metrics.Pods.Util > 0 {
			met.Pods.Util = metrics.Pods.Util
		}
	} else {
		c.Namespaces[name] = metrics
	}
}

// UpdateNode Adder for the Nodes
func (c *Clustermetrics) UpdateNode(name string, metrics *Nodemetrics) {
	if met, ok := c.Nodes[name]; ok {

		// Node metrics incrementally updated by looping through pods
		met.Cpu.Req += metrics.Cpu.Req
		met.Cpu.Limit += metrics.Cpu.Limit
		met.Mem.Req += metrics.Mem.Req
		met.Mem.Limit += metrics.Mem.Limit
		met.Pods.Inuse += metrics.Pods.Inuse

		// Node metrics set when looping through nodes
		met.Cpu.Avail += metrics.Cpu.Avail
		met.Cpu.Cap += metrics.Cpu.Cap
		met.Mem.Avail += metrics.Mem.Avail
		met.Mem.Cap += metrics.Mem.Cap
		if len(metrics.Label) > 0 {
			met.Label = metrics.Label
		}
		if len(metrics.Status) > 0 {
			met.Status = metrics.Label
			met.Sched = metrics.Sched
		}
		if metrics.Cpu.Util > 0 {
			met.Cpu.Util = metrics.Cpu.Util
		}
		if metrics.Mem.Util > 0 {
			met.Mem.Util = metrics.Mem.Util
		}
		if metrics.Pods.Avail > 0 {
			met.Pods.Avail = metrics.Pods.Avail
		}
		if metrics.Pods.Cap > 0 {
			met.Pods.Cap = metrics.Pods.Cap
		}
		if metrics.Pods.Util > 0 {
			met.Pods.Util = metrics.Pods.Util
		}
	} else {
		c.Nodes[name] = metrics
	}
}

// Return the length of the longest entry in a list
func (c *Clustermetrics) maxW(field string, min int) int {
	var w int
	switch field {
	case "name":
		for n := range c.Nodes {
			w = utils.MaxInt(w, len(n))
		}
	case "status":
		for _, m := range c.Nodes {
			w = utils.MaxInt(w, len(m.Status))
		}
	case "label":
		for _, m := range c.Nodes {
			w = utils.MaxInt(w, len(m.Label))
		}
	case "taints":
		for _, m := range c.Nodes {
			w = utils.MaxInt(w, len(m.Taints))
		}
	case "namespace":
		for n := range c.Namespaces {
			w = utils.MaxInt(w, len(n))
		}
	}
	return utils.MaxInt(w, min)
}

// PrintNodeSummary Print utilization summary of each node in the cluster
func (c *Clustermetrics) PrintNodeSummary() {
	nw := c.maxW("name", 4)
	sw := c.maxW("status", 6)
	tw := c.maxW("taints", 6)
	lw := c.maxW("label", 5)
	fmt.Printf("%-*s  %-*s  %-*s  %-*s  %s  %s  %s\n", nw, "NODE", sw, "STATUS", lw, "LABEL", tw, "TAINTS", "CPU REQ", "MEM REQ", "PODS")
	for name, n := range c.Nodes {
		if name != "" {
			fmt.Printf("%-*v  %-*v  %-*s  %-*v  %-7v  %-7v  %v\n", nw, name, sw, n.Status, lw, n.Label, tw, n.Taints, utils.FmtPct(n.Cpu.Util), utils.FmtPct(n.Mem.Util), utils.FmtPct(n.Pods.Util))
		}
	}
}

// PrintNamespaceSummary Print utilization summary of each namespace in the cluster
func (c *Clustermetrics) PrintNamespaceSummary() {
	nsw := c.maxW("namespace", 9)
	fmt.Printf("%-*s  %-7s  %-4s  %-9s  %-4s  %-4s  %s\n", nsw, "NAMESPACE", "CPU REQ", "UTIL", "MEM REQ", "UTIL", "PODS", "UTIL")
	for name, n := range c.Namespaces {
		if name != "" {
			fmt.Printf("%-*v  %-7v  %-4v  %-9v  %-4v  %-4v  %v\n", nsw, name, utils.FmtMilli(n.Cpu.Req), utils.FmtPct(n.Cpu.Util), utils.FmtMiB(n.Mem.Req), utils.FmtPct(n.Mem.Util), (n.Pods.Inuse), utils.FmtPct(n.Pods.Util))
		}
	}
}

// PrintClusterSummary Print utilization summary for the cluster
func (c *Clustermetrics) PrintClusterSummary() {
	memreq := utils.FmtGiB(c.Mem.Req)
	memavail := utils.FmtGiB(c.Mem.Avail)
	memcap := utils.FmtGiB(c.Mem.Cap)
	cpureq := utils.FmtCPU(c.Cpu.Req)
	cpuavail := utils.FmtCPU(c.Cpu.Avail)
	cpucap := utils.FmtCPU(c.Cpu.Cap)
	fmt.Printf("%-17s  %-10s %-10s %-10s %s\n", "CLUSTER RESOURCES", "REQUESTED", "AVAILABLE", "CAPACITY", "UTIL")
	fmt.Printf("%-17s  %-10v %-10v %-10v %s\n", "CPU", cpureq, cpuavail, cpucap, utils.FmtPct(c.Cpu.Util))
	fmt.Printf("%-17s  %-10v %-10v %-10v %s\n", "MEMORY", memreq, memavail, memcap, utils.FmtPct(c.Mem.Util))
	fmt.Printf("%-17v  %-10v %-10v %-10v %v\n", "PODS", c.Pods.Inuse, c.Pods.Avail, c.Pods.Cap, utils.FmtPct(c.Pods.Util))
}
