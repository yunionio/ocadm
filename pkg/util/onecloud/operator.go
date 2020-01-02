package onecloud

import (
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"yunion.io/x/log"
	"yunion.io/x/onecloud-operator/pkg/apis/constants"
	onecloud "yunion.io/x/onecloud-operator/pkg/apis/onecloud/v1alpha1"
	"yunion.io/x/onecloud-operator/pkg/client/clientset/versioned"
)

const (
	WebCEImageName        = "web"
	APIGatewayCEImageName = "apigateway"
	WebEEImageName        = "web-ee"
	APIGatewayEEImageName = "apigateway-ee"
)

func SetOCUseCE(oc *onecloud.OnecloudCluster) *onecloud.OnecloudCluster {
	web := &oc.Spec.Web
	SetCEdition(web)
	web.ImageName = WebCEImageName

	apiGateway := &oc.Spec.APIGateway
	apiGateway.ImageName = APIGatewayCEImageName
	SetCEdition(apiGateway)

	yunionagent := &oc.Spec.Yunionagent
	yunionagent.Disable = true
	return oc
}

func SetOCUseEE(oc *onecloud.OnecloudCluster) *onecloud.OnecloudCluster {
	web := &oc.Spec.Web
	web.ImageName = WebEEImageName
	SetEEdition(web)

	apiGateway := &oc.Spec.APIGateway
	apiGateway.ImageName = APIGatewayEEImageName
	SetEEdition(apiGateway)

	yunionagent := &oc.Spec.Yunionagent
	yunionagent.Disable = false
	return oc
}

func setEdition(spec *onecloud.DeploymentSpec, edition string) {
	if spec.Annotations == nil {
		spec.Annotations = map[string]string{}
	}
	spec.Annotations[constants.OnecloudEditionAnnotationKey] = edition
}

func SetCEdition(spec *onecloud.DeploymentSpec) {
	setEdition(spec, constants.OnecloudCommunityEdition)
}

func SetEEdition(spec *onecloud.DeploymentSpec) {
	setEdition(spec, constants.OnecloudEnterpriseEdition)
}

func isDeploymentImageUpdated(
	globalRepo string,
	globalVersion string,
	spec *onecloud.DeploymentSpec,
	curStatus *onecloud.DeploymentStatus) (bool, string) {
	repo := globalRepo
	version := globalVersion
	if spec.Repository != "" {
		repo = spec.Repository
	}
	if spec.Tag != "" {
		version = spec.Tag
	}
	curRepo := curStatus.ImageStatus.Repository
	curVersion := curStatus.ImageStatus.Tag
	if repo != curRepo {
		return false, fmt.Sprintf("current repo %s => expected repo %s", curRepo, repo)
	}
	if version != curVersion {
		return false, fmt.Sprintf("current version %s => expected version %s", curVersion, version)
	}
	return true, ""
}

func IsDeploymentUpdated(
	globalRepo string,
	globalVersion string,
	spec *onecloud.DeploymentSpec,
	curStatus *onecloud.DeploymentStatus) (bool, string) {
	if updated, reason := isDeploymentImageUpdated(globalRepo, globalVersion, spec, curStatus); !updated {
		return false, reason
	}
	if curStatus.Phase == onecloud.NormalPhase {
		return true, ""
	}
	return false, fmt.Sprintf("%s is upgrading", curStatus.ImageStatus.ImageName)
}

type SpecStatusPair struct {
	Name   string
	Getter func(*onecloud.OnecloudCluster) (onecloud.DeploymentSpec, onecloud.DeploymentStatus)
}

var SpecsStatus []SpecStatusPair = []SpecStatusPair{
	{
		Name: "keystone",
		Getter: func(oc *onecloud.OnecloudCluster) (onecloud.DeploymentSpec, onecloud.DeploymentStatus) {
			return oc.Spec.Keystone.DeploymentSpec, oc.Status.Keystone.DeploymentStatus
		},
	},
	{
		Name: "glance",
		Getter: func(oc *onecloud.OnecloudCluster) (onecloud.DeploymentSpec, onecloud.DeploymentStatus) {
			return oc.Spec.Glance.DeploymentSpec, oc.Status.Glance.DeploymentStatus
		},
	},
	{
		Name: "region",
		Getter: func(oc *onecloud.OnecloudCluster) (onecloud.DeploymentSpec, onecloud.DeploymentStatus) {
			return oc.Spec.RegionServer.DeploymentSpec, oc.Status.RegionServer.DeploymentStatus
		},
	},
	{
		Name: "scheduler",
		Getter: func(oc *onecloud.OnecloudCluster) (onecloud.DeploymentSpec, onecloud.DeploymentStatus) {
			return oc.Spec.Scheduler, oc.Status.Scheduler
		},
	},
	{
		Name: "apigateway",
		Getter: func(oc *onecloud.OnecloudCluster) (onecloud.DeploymentSpec, onecloud.DeploymentStatus) {
			return oc.Spec.APIGateway, oc.Status.APIGateway
		},
	},
	{
		Name: "web",
		Getter: func(oc *onecloud.OnecloudCluster) (onecloud.DeploymentSpec, onecloud.DeploymentStatus) {
			return oc.Spec.Web, oc.Status.Web
		},
	},
	{
		Name: "cloudnet",
		Getter: func(oc *onecloud.OnecloudCluster) (onecloud.DeploymentSpec, onecloud.DeploymentStatus) {
			return oc.Spec.Cloudnet, oc.Status.Cloudnet
		},
	},
}

func IsClusterUpdated(oc *onecloud.OnecloudCluster) (bool, string) {
	for _, ss := range SpecsStatus {
		curSpec, curStatus := ss.Getter(oc)
		if updated, reason := IsDeploymentUpdated(oc.Spec.ImageRepository, oc.Spec.Version, &curSpec, &curStatus); !updated {
			return false, fmt.Sprintf("%s: %s", ss.Name, reason)
		}
	}
	return true, ""
}

func WaitOnecloudDeploymentUpdated(
	cli versioned.Interface,
	name string,
	namespace string,
	timeout time.Duration,
) error {
	return wait.PollImmediate(10*time.Second, timeout, func() (bool, error) {
		oc, err := cli.OnecloudV1alpha1().OnecloudClusters(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		ok, reason := IsClusterUpdated(oc)
		if ok {
			return true, nil
		}
		log.Infof("Wait: %s", reason)
		return false, nil
	})
}
