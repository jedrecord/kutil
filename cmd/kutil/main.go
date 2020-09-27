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

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jedrecord/kutil/pkg/resources"
	"github.com/jedrecord/kutil/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func showVersion() {
	fmt.Println("kutil version 0.9.0")
	fmt.Println("Copyright (C) 2020 Jed Record")
	fmt.Println("License GNU GPL version 2 <https://gnu.org/licenses/gpl-2.0.html>")
	fmt.Println("This is free software; you are free to change and redistribute it.")
	fmt.Println("There is NO WARRANTY, to the extent permitted by law.")
	fmt.Println("")
	fmt.Println("Bugs: Jed Record <jed@jedrecord.com>")
	fmt.Println("WWW:  https://github.com/jedrecord/kutil")
}

func main() {
	nsflag := flag.Bool("namespaces", false, "Display utilization by namespaces")
	nodesflag := flag.Bool("nodes", false, "Display utilization by nodes")
	clusterflag := flag.Bool("cluster", false, "Display utilization for the cluster")
	versionflag := flag.Bool("version", false, "Display program version")
	kubeconfig := flag.String("kubeconfig", filepath.Join(os.Getenv("HOME"), "/.kube/config"), "kubeconfig file")
	flag.Parse()
	if *versionflag {
		showVersion()
		os.Exit(0)
	}
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		logError("Could not access kubeconfig file")
	}
	config.AcceptContentTypes = "application/vnd.kubernetes.protobuf, application/json"
	config.ContentType = "application/vnd.kubernetes.protobuf"
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logError("There was a problem parsing kubeconfig")
	}

	mycluster := resources.NewCluster()

	mynodes, err := clientset.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		logError("There was a problem connecting with the API")
	}
	// Loop through the nodes to collect utilization data
	if len(mynodes.Items) > 0 {
		//fmt.Printf("NODE\t\tLABEL\t\t\tCPU\t\tRAM\t\tPODS\n")
		for _, mynode := range mynodes.Items {
			n := mynode.Name
			var role string
			for label := range mynode.Labels {
				pair := strings.Split(label, "/")
				multirole := false
				if pair[0] == "node-role.kubernetes.io" {
					if multirole {
						role = role + ","
					}
					role = role + pair[1]
					multirole = true
				}
			}
			ndata := resources.NewNodemetrics()
			ndata.Label = role
			ndata.Sched = true
			ndata.Status = "Ready"
			cpuAvail := mynode.Status.Allocatable["cpu"]
			memAvail := mynode.Status.Allocatable["memory"]
			podsAvail := mynode.Status.Allocatable["pods"]
			ndata.Cpu.Avail = cpuAvail.MilliValue()
			ndata.Mem.Avail = memAvail.Value()
			ndata.Pods.Avail = podsAvail.Value()
			mycluster.UpdateNode(n, ndata)
			mycluster.Cpu.Avail += cpuAvail.MilliValue()
			mycluster.Mem.Avail += memAvail.Value()
			mycluster.Pods.Avail += podsAvail.Value()
		}
	} else {
		logError("No nodes discovered")
	}

	mypods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
	if err != nil {
		logError("There was a problem connecting with the API")
	}
	// Loop through the pods to collect utilization data
	if len(mypods.Items) > 0 {
		// do something
		for _, mypod := range mypods.Items {
			ns := mypod.Namespace
			no := mypod.Spec.NodeName
			nsdata := resources.NewNsmetrics()
			ndata := resources.NewNodemetrics()
			var good []string
			for _, cons := range mypod.Status.ContainerStatuses {
				if cons.Ready == true {
					good = append(good, cons.Name)
				}
			}
			for _, con := range mypod.Spec.Containers {
				ok := false
				for _, g := range good {
					if g == con.Name {
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
					mycluster.Cpu.Req += cpuReq.MilliValue()
					mycluster.Cpu.Limit += cpuLim.MilliValue()
					mycluster.Mem.Req += memReq.Value()
					mycluster.Mem.Limit += memLim.Value()
				}
			}
			if len(good) > 0 {
				nsdata.Pods.Inuse++
				ndata.Pods.Inuse++
				mycluster.Pods.Inuse++
			}
			mycluster.UpdateNamespace(ns, nsdata)
			mycluster.UpdateNode(no, ndata)
		}
	} else {
		logError("No pods discovered")
	}

	for n, m := range mycluster.Namespaces {
		nsdata := resources.NewNsmetrics()
		cu := utils.CalcPct(mycluster.Cpu.Avail, m.Cpu.Req)
		mu := utils.CalcPct(mycluster.Mem.Avail, m.Mem.Req)
		pu := utils.CalcPct(mycluster.Pods.Avail, m.Pods.Inuse)
		nsdata.Cpu.Util = cu
		nsdata.Mem.Util = mu
		nsdata.Pods.Util = pu
		mycluster.UpdateNamespace(n, nsdata)
	}

	for n, m := range mycluster.Nodes {
		ndata := resources.NewNodemetrics()
		cu := utils.CalcPct(m.Cpu.Avail, m.Cpu.Req)
		mu := utils.CalcPct(m.Mem.Avail, m.Mem.Req)
		pu := utils.CalcPct(m.Pods.Avail, m.Pods.Inuse)
		ndata.Cpu.Util = cu
		ndata.Mem.Util = mu
		ndata.Pods.Util = pu
		mycluster.UpdateNode(n, ndata)
	}

	cu := utils.CalcPct(mycluster.Cpu.Avail, mycluster.Cpu.Req)
	mu := utils.CalcPct(mycluster.Mem.Avail, mycluster.Mem.Req)
	pu := utils.CalcPct(mycluster.Pods.Avail, mycluster.Pods.Inuse)
	mycluster.Cpu.Util = cu
	mycluster.Mem.Util = mu
	mycluster.Pods.Util = pu

	if *nsflag {
		mycluster.PrintNamespaceSummary()
	}
	if *nodesflag {
		mycluster.PrintNodeSummary()
	}
	if *clusterflag {
		mycluster.PrintClusterSummary()
	}
	if !*nsflag && !*nodesflag && !*clusterflag {
		mycluster.PrintNodeSummary()
		fmt.Println()
		mycluster.PrintClusterSummary()
	}
}

func logError(msg string) {
	fmt.Printf("Error: %s\n", msg)
	os.Exit(1)
}
