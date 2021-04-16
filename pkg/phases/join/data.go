package join

import (
	phases "k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/join"

	apiv1 "yunion.io/x/ocadm/pkg/apis/v1"
)

// JoinData is the interface to use for join phases.
// The "joinData" type from "cmd/join.go" must satisfy this interface.
type JoinData interface {
	phases.JoinData
	OnecloudInitCfg() (*apiv1.InitConfiguration, error)
	OnecloudJoinCfg() *apiv1.JoinConfiguration
	GetHighAvailabilityVIP() string
	GetKeepalivedVersionTag() string
	GetNodeIP() string
	GetHostInterface() string
}
