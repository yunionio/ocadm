package component

import (
	"io"
	"os"

	"github.com/spf13/cobra"

	"yunion.io/x/onecloud-operator/pkg/apis/constants"
	onecloud "yunion.io/x/onecloud-operator/pkg/apis/onecloud/v1alpha1"
)

const (
	WebCEImageName        = "web"
	APIGatewayCEImageName = "apigateway"
	WebEEImageName        = "web-ee"
	APIGatewayEEImageName = "apigateway-ee"
)

func NewCmdWeb() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "web",
		Short: "Manage web UI component",
	}
	NewWebCmd(cmd, os.Stdout).Bind()
	return cmd
}

type WebCmd struct {
	*baseCmd
}

func NewWebCmd(cmd *cobra.Command, out io.Writer) *WebCmd {
	return &WebCmd{
		baseCmd: newBaseCmd(cmd, out),
	}
}

func (w *WebCmd) Bind() {
	w.baseCmd.AddCmd(w.newSubCmd("use-ce", w.useCE))
	w.baseCmd.AddCmd(w.newSubCmd("use-ee", w.useEE))
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

func (w *WebCmd) useCE(data *componentsData, _ io.Writer) error {
	// change onecloud web component image
	// disable yunion-agent
	return w.updateOnecloudCluster(func(oc *onecloud.OnecloudCluster) *onecloud.OnecloudCluster {
		web := &oc.Spec.Web
		SetCEdition(web)
		web.ImageName = WebCEImageName

		apiGateway := &oc.Spec.APIGateway
		apiGateway.ImageName = APIGatewayCEImageName
		SetCEdition(apiGateway)

		yunionagent := &oc.Spec.Yunionagent
		yunionagent.Disable = true
		return oc
	})
}

func (w *WebCmd) useEE(data *componentsData, _ io.Writer) error {
	// change onecloud web component image to ee
	// enable yunion-agent
	return w.updateOnecloudCluster(func(oc *onecloud.OnecloudCluster) *onecloud.OnecloudCluster {
		web := &oc.Spec.Web
		web.ImageName = WebEEImageName
		SetEEdition(web)

		apiGateway := &oc.Spec.APIGateway
		apiGateway.ImageName = APIGatewayEEImageName
		SetEEdition(apiGateway)

		yunionagent := &oc.Spec.Yunionagent
		yunionagent.Disable = false
		return oc
	})
}
