package options

import (
	"github.com/spf13/pflag"
	kubeadmoptions "k8s.io/kubernetes/cmd/kubeadm/app/cmd/options"

	"yunion.io/x/ocadm/pkg/apis/v1"
)

// AddConfigFlag adds the --config flag to the given flagset
func AddConfigFlag(fs *pflag.FlagSet, cfgPath *string) {
	fs.StringVar(cfgPath, CfgPath, *cfgPath, "Path to a ocadm configuration file.")
}

// AddKubeadmConfigFlag adds the --kubeadm-config flag to the given flagset
func AddKubeadmConfigFlag(fs *pflag.FlagSet, kubeadmCfgPath *string) {
	fs.StringVar(kubeadmCfgPath, KubeadmCfgPath, *kubeadmCfgPath, "Path to a kubeadm configuration file.")
}

// AddOnecloudVersion adds the --onecloud-version flag
func AddOnecloudVersion(fs *pflag.FlagSet, onecloudVersion *string) {
	fs.StringVar(
		onecloudVersion, "onecloud-version", *onecloudVersion,
		`Choose a specific Onecloud version for the control plane.`,
	)
}

// AddImageMetaFlags adds the --image-repository flag to the given flagset
func AddImageMetaFlags(fs *pflag.FlagSet, imageRepository *string) {
	fs.StringVar(imageRepository, ImageRepository, v1.DefaultImageRepository, "Choose a container registry to pull control plane images from")
}

var (
	// AddKubeConfigFlag adds the --kubeconfig flag to the given flagset
	AddKubeConfigFlag = kubeadmoptions.AddKubeConfigFlag
	// AddKubeConfigDirFlag adds the --kubeconfig-dir flag to the given flagset
	AddKubeConfigDirFlag = kubeadmoptions.AddKubeConfigDirFlag
	// AddIgnorePreflightErrorsFlag adds the --ignore-preflight-errors flag to the given flagset
	AddIgnorePreflightErrorsFlag = kubeadmoptions.AddIgnorePreflightErrorsFlag
	// AddControlPlanExtraArgsFlags adds the ExtraArgs flags for control plane components
	AddControlPlanExtraArgsFlags = kubeadmoptions.AddControlPlanExtraArgsFlags
	// AddKubernetesVersionFlag adds the --kubernetes-version flag to the given flagset
	AddKubernetesVersionFlag = kubeadmoptions.AddKubernetesVersionFlag
	// AddFeatureGatesStringFlag adds the --feature-gates flag to the given flagset
	AddFeatureGatesStringFlag = kubeadmoptions.AddFeatureGatesStringFlag
)
