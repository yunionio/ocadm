package onecloud

import (
	"yunion.io/x/onecloud-operator/pkg/apis/constants"
	onecloud "yunion.io/x/onecloud-operator/pkg/apis/onecloud/v1alpha1"
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
