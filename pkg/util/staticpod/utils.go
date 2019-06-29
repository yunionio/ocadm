package staticpod

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/staticpod"
)

var (
	VolumeMountMapToSlice = staticpod.VolumeMountMapToSlice
	ComponentResources    = staticpod.ComponentResources
	WriteStaticPodToDisk  = staticpod.WriteStaticPodToDisk
)

func ComponentPodWithInit(initContainer, container *v1.Container, volumes map[string]v1.Volume) v1.Pod {
	pod := staticpod.ComponentPod(*container, volumes)
	if initContainer != nil {
		pod.Spec.InitContainers = []v1.Container{*initContainer}
	}
	return pod
}
