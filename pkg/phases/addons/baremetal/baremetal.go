package baremetal

import (
	clientset "k8s.io/client-go/kubernetes"

	"yunion.io/x/onecloud/pkg/mcclient"

	apiv1 "yunion.io/x/ocadm/pkg/apis/v1"
)

func EnsureBaremetalAddon(cfg *apiv1.InitConfiguration, kubeCli clientset.Interface, session *mcclient.ClientSession) error {
	return nil
}
