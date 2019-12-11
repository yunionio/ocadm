package join

import (
	"fmt"

	"github.com/pkg/errors"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"

	"yunion.io/x/ocadm/pkg/apis/constants"
	"yunion.io/x/ocadm/pkg/phases/uploadconfig"
)

func NewControlPlaneJoinPhase() workflow.Phase {
	return workflow.Phase{
		Name:  "oc-control-plane-join",
		Short: "Join a machine as a oc control plane instance",
		Phases: []workflow.Phase{
			newUpdateStatusSubphase(),
		},
	}
}

func newUpdateStatusSubphase() workflow.Phase {
	return workflow.Phase{
		Name: "update-status",
		Short: fmt.Sprintf(
			"Register the new control-plane node into the %s maintained in the %s ConfigMap",
			constants.ClusterConfigurationConfigMapKey,
			constants.OnecloudAdminConfigConfigMap,
		),
		Run: runUpdateStatusPhase,
	}
}

func runUpdateStatusPhase(c workflow.RunData) error {
	data, ok := c.(JoinData)
	if !ok {
		return errors.New("oc-control-plane-join phase invoked with an invalid data struct")
	}

	if data.Cfg().ControlPlane == nil {
		return nil
	}

	client, err := data.ClientSet()
	if err != nil {
		return errors.Wrap(err, "couldn't create Kubernetes client")
	}

	cfg, err := data.OnecloudInitCfg()
	if err != nil {
		return err
	}

	if err := uploadconfig.UploadConfiguration(cfg, client); err != nil {
		return errors.Wrap(err, "error uploading configuration")
	}

	return nil
}
