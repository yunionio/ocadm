package kube

import (
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// GetConfig returns a Kubernetes client config
func GetConfig(kubeconfig, context, namespace string) *genericclioptions.ConfigFlags {
	cf := genericclioptions.NewConfigFlags(true)
	cf.Namespace = &namespace
	cf.Context = &context
	cf.KubeConfig = &kubeconfig
	return cf
}
