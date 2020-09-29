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

	"github.com/jedrecord/kutil/pkg/resources"
	"github.com/jedrecord/kutil/pkg/utils"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func showVersion() {
	fmt.Println("kutil version 0.9.1")
	fmt.Println("Copyright (C) 2020 Jed Record")
	fmt.Println("License GNU GPL version 2 <https://gnu.org/licenses/gpl-2.0.html>")
	fmt.Println("This is free software; you are free to change and redistribute it.")
	fmt.Println("There is NO WARRANTY, to the extent permitted by law.")
	fmt.Println("")
	fmt.Println("Bugs: Jed Record <jed@jedrecord.com>")
	fmt.Println("WWW:  https://github.com/jedrecord/kutil")
}

func main() {
	// Set boolean flags for display output
	nsflag := flag.Bool("namespaces", false, "Display utilization by namespaces")
	nodesflag := flag.Bool("nodes", false, "Display utilization by nodes")
	clusterflag := flag.Bool("cluster", false, "Display utilization for the cluster")
	versionflag := flag.Bool("version", false, "Display program version")

	// User can provide a kubeconfig file or use default (~/.kube/config)
	kubeconfig := flag.String("kubeconfig", filepath.Join(os.Getenv("HOME"), "/.kube/config"), "kubeconfig file")

	flag.Parse()
	if *versionflag {
		showVersion()
		os.Exit(0)
	}

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		utils.LogError("Could not access kubeconfig file")
	}
	config.AcceptContentTypes = "application/vnd.kubernetes.protobuf, application/json"
	config.ContentType = "application/vnd.kubernetes.protobuf"

	// Build a valid set of credentials for a kubernetes cluster, returns pointer or err
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		utils.LogError("There was a problem parsing kubeconfig")
	}

	// Create a Clustermetrics object (struct) to hold the current k8s resources data state
	// (Clustermetrics{} defined in pkg/resources/resources.go)
	mycluster := resources.NewCluster()

	// Connect with the cluster and collect current state
	// Requires a pointer to a valid clientset
	// See Clustermetrics{} functions in pkg/resources/resources.go
	mycluster.Load(clientset)

	// Determine output based on flag options (-namespaces, -nodes, -cluster)
	if *nsflag {
		mycluster.PrintNamespaceSummary()
	}
	if *nodesflag {
		mycluster.PrintNodeSummary()
	}
	if *clusterflag {
		mycluster.PrintClusterSummary()
	}

	// If no options selected default output is node and cluster summary
	if !*nsflag && !*nodesflag && !*clusterflag {
		mycluster.PrintNodeSummary()
		fmt.Println()
		mycluster.PrintClusterSummary()
	}
}

