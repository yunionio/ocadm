package options

import (
	"github.com/spf13/pflag"
	kubeadmoptions "k8s.io/kubernetes/cmd/kubeadm/app/cmd/options"
)

// AddConfigFlag adds the --config flag to the given flagset
func AddConfigFlag(fs *pflag.FlagSet, cfgPath *string) {
	fs.StringVar(cfgPath, CfgPath, *cfgPath, "Path to a kubeadm configuration file.")
}

// AddKubeadmConfigFlag adds the --kubeadm-config flag to the given flagset
func AddKubeadmConfigFlag(fs *pflag.FlagSet, kubeadmCfgPath *string) {
	fs.StringVar(kubeadmCfgPath, KubeadmCfgPath, *kubeadmCfgPath, "Path to a kubeadm configuration file.")
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
	// AddImageMetaFlags adds the --image-repository flag to the given flagset
	AddImageMetaFlags = kubeadmoptions.AddImageMetaFlags
)
