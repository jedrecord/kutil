[![License: GPL v2](https://img.shields.io/badge/License-GPL%20v2-blue.svg)](https://www.gnu.org/licenses/old-licenses/gpl-2.0.en.html)
[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/jedrecord/kutil)](https://github.com/jedrecord/kutil)
[![kubernetes version](https://img.shields.io/badge/kubernetes-v1.13+-blue)](https://github.com/jedrecord/kutil)
[![OpenShift version](https://img.shields.io/badge/OpenShift-v4.1+-EE0000?logo=Red-Hat-Open-Shift)](https://github.com/jedrecord/kutil)
[![Twitter Follow](https://img.shields.io/twitter/follow/jedrecord?label=follow&style=social)](https://twitter.com/jedrecord)

# kutil
Display a summary of Kubernetes node, namespace, and cluster resource utilization by memory, cpu, and pods

## Screenshots
<image src="https://github.com/jedrecord/kutil/blob/master/assets/screenshot1.jpg" alt="screenshot 1" width="600">

<image src="https://github.com/jedrecord/kutil/blob/master/assets/screenshot2.jpg" alt="screenshot 2" width="600">

## Installation
```
# To install the latest pre-built binary for linux:
sudo curl -L https://github.com/jedrecord/kutil/releases/download/v0.9.3/kutil-linux-amd64 \
    -o /usr/local/bin/kutil && sudo chmod 755 /usr/local/bin/kutil

To install from source (requires Go)
make build
sudo make install

or with go get:
GO111MODULE=on go get github.com/jedrecord/kutil/cmd/kutil
```

## Source
The source code is well commented with the main command package located in the project cmd/kutil directory. You will find the meat of this program is in the resources package located in the pkg/resources directory. To build a binary from source, navigate to the cmd/kutil directory and run "go build".

## Platforms
Currently, in the first release, only linux is supported. Look for Windows support in an upcoming release.
