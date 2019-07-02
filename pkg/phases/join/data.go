package join

import (
	phases "k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/join"
)

// JoinData is the interface to use for join phases.
// The "joinData" type from "cmd/join.go" must satisfy this interface.
type JoinData interface {
	phases.JoinData
}
