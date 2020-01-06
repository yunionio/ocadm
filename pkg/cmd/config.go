package cmd

import (
	"fmt"
	"io"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	cmdutil "k8s.io/kubernetes/cmd/kubeadm/app/cmd/util"
	"k8s.io/kubernetes/cmd/kubeadm/app/features"
	kubeadmutil "k8s.io/kubernetes/cmd/kubeadm/app/util"
	//configutil "k8s.io/kubernetes/cmd/kubeadm/app/util/config"
	utilruntime "k8s.io/kubernetes/cmd/kubeadm/app/util/runtime"
	utilsexec "k8s.io/utils/exec"

	"yunion.io/x/ocadm/pkg/apis/scheme"
	apiv1 "yunion.io/x/ocadm/pkg/apis/v1"
	"yunion.io/x/ocadm/pkg/images"
	"yunion.io/x/ocadm/pkg/options"
	configutil "yunion.io/x/ocadm/pkg/util/config"
)

func NewCmdConfig(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
		RunE:  cmdutil.SubCmdRunE("config"),
	}

	cmd.AddCommand(NewCmdConfigImages(out))
	return cmd
}

// NewCmdConfigImages returns the "ocadm config images" command
func NewCmdConfigImages(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "images",
		Short: "Interact with container images used by ocadm.",
		RunE:  cmdutil.SubCmdRunE("images"),
	}
	cmd.AddCommand(NewCmdConfigImagesList(out, nil))
	cmd.AddCommand(NewCmdConfigImagesPull())
	return cmd
}

// NewCmdConfigImagesPull returns the `kubeadm config images pull` command
func NewCmdConfigImagesPull() *cobra.Command {
	externalcfg := &apiv1.InitConfiguration{}
	scheme.Scheme.Default(externalcfg)
	var cfgPath, featureGatesString, operatorVersion string
	var err error

	cmd := &cobra.Command{
		Use:   "pull",
		Short: "Pull images used by ocadm.",
		Run: func(_ *cobra.Command, _ []string) {
			externalcfg.InitConfiguration.ClusterConfiguration.FeatureGates, err = features.NewFeatureGate(&features.InitFeatureGates, featureGatesString)
			kubeadmutil.CheckErr(err)
			internalcfg, err := configutil.LoadOrDefaultInitConfiguration(cfgPath, externalcfg)
			kubeadmutil.CheckErr(err)
			containerRuntime, err := utilruntime.NewContainerRuntime(utilsexec.New(), internalcfg.NodeRegistration.CRISocket)
			kubeadmutil.CheckErr(err)
			imagesPull := NewImagesPull(containerRuntime, images.GetAllImages(&internalcfg.ClusterConfiguration, &internalcfg.InitConfiguration.ClusterConfiguration, operatorVersion))
			kubeadmutil.CheckErr(imagesPull.PullAll())
		},
	}
	AddImagesCommonConfigFlags(cmd.PersistentFlags(), externalcfg, &cfgPath, &featureGatesString, &operatorVersion)
	cmdutil.AddCRISocketFlag(cmd.PersistentFlags(), &externalcfg.NodeRegistration.CRISocket)

	return cmd
}

// ImagesPull is the struct used to hold information relating to image pulling
type ImagesPull struct {
	runtime utilruntime.ContainerRuntime
	images  []string
}

// NewImagesPull initializes and returns the `kubeadm config images pull` command
func NewImagesPull(runtime utilruntime.ContainerRuntime, images []string) *ImagesPull {
	return &ImagesPull{
		runtime: runtime,
		images:  images,
	}
}

// PullAll pulls all images that the ImagesPull knows about
func (ip *ImagesPull) PullAll() error {
	for _, image := range ip.images {
		fmt.Printf("[config/images] Pulling %s\n", image)
		if err := ip.runtime.PullImage(image); err != nil {
			return errors.Wrapf(err, "failed to pull image %q", image)
		}
		fmt.Printf("[config/images] Pulled %s\n", image)
	}
	return nil
}

// NewCmdConfigImagesList returns the "ocadm config images list" command
func NewCmdConfigImagesList(out io.Writer, mockK8sVersion *string) *cobra.Command {
	externalCfg := &apiv1.InitConfiguration{}
	scheme.Scheme.Default(externalCfg)
	var cfgPath, featureGatesString, operatorVersion string

	if mockK8sVersion != nil {
		externalCfg.KubernetesVersion = *mockK8sVersion
	}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Print a list of images ocadm will use. The configuration file is used in case any images or image repositories are customized.",
		Run: func(_ *cobra.Command, _ []string) {
			var err error
			externalCfg.InitConfiguration.ClusterConfiguration.FeatureGates, err = features.NewFeatureGate(&features.InitFeatureGates, featureGatesString)
			imagesList, err := NewImagesList(cfgPath, externalCfg, operatorVersion)
			kubeadmutil.CheckErr(err)
			kubeadmutil.CheckErr(imagesList.Run(out))
		},
	}
	AddImagesCommonConfigFlags(cmd.PersistentFlags(), externalCfg, &cfgPath, &featureGatesString, &operatorVersion)
	return cmd
}

func NewImagesList(cfgPath string, cfg *apiv1.InitConfiguration, operatorVersion string) (*ImagesList, error) {
	// TODO: load configuration
	initcfg, err := configutil.LoadOrDefaultInitConfiguration(cfgPath, cfg)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert cfg to an internal cfg")
	}
	return &ImagesList{
		cfg:             initcfg,
		operatorVersion: operatorVersion,
	}, nil
}

type ImagesList struct {
	cfg             *apiv1.InitConfiguration
	operatorVersion string
}

func (i *ImagesList) Run(out io.Writer) error {
	imgs := images.GetAllImages(&i.cfg.ClusterConfiguration, &i.cfg.InitConfiguration.ClusterConfiguration, i.operatorVersion)
	for _, img := range imgs {
		fmt.Fprintln(out, img)
	}

	return nil
}

// AddImagesCommonConfigFlags adds the flags that configure kubeadm (and affect the images kubeadm will use)
func AddImagesCommonConfigFlags(flagSet *flag.FlagSet, cfg *apiv1.InitConfiguration, cfgPath *string, featureGatesString *string, operatorVersion *string) {
	options.AddKubernetesVersionFlag(flagSet, &cfg.KubernetesVersion)
	options.AddOnecloudVersion(flagSet, &cfg.OnecloudVersion)
	options.AddFeatureGatesStringFlag(flagSet, featureGatesString)
	options.AddImageMetaFlags(flagSet, &cfg.ImageRepository)
	options.AddKubeadmConfigFlag(flagSet, cfgPath)
	options.AddOperatorVersionFlags(flagSet, operatorVersion)
}
