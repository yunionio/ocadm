package config

import (
	"k8s.io/klog"
	kubeadmscheme "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/scheme"
	kubeadmapiv1beta1 "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta1"
	kubeadmconfig "k8s.io/kubernetes/cmd/kubeadm/app/util/config"

	ocadmscheme "yunion.io/x/ocadm/pkg/apis/scheme"
	apiv1 "yunion.io/x/ocadm/pkg/apis/v1"
)

// LoadOrDefaultJoinConfiguration takes a path to a config file and a versioned configuration that can serve as the default config
// If cfgPath is specified, defaultversionedcfg will always get overridden. Otherwise, the default config (often populated by flags) will be used.
// Then the external, versioned configuration is defaulted and converted to the internal type.
// Right thereafter, the configuration is defaulted again with dynamic values (like IP addresses of a machine, etc)
// Lastly, the internal config is validated and returned.
func LoadOrDefaultJoinConfiguration(cfgPath string, defaultversionedcfg *apiv1.JoinConfiguration) (*apiv1.JoinConfiguration, error) {
	if cfgPath != "" {
		// Loads configuration from config file, if provided
		// Nb. --config overrides command line flags, TODO: fix this
		return LoadJoinConfigurationFromFile(cfgPath)
	}

	return DefaultedJoinConfiguration(defaultversionedcfg)
}

// LoadJoinConfigurationFromFile loads versioned JoinConfiguration from file, converts it to internal, defaults and validates it
func LoadJoinConfigurationFromFile(cfgPath string) (*apiv1.JoinConfiguration, error) {
	klog.V(1).Infof("loading configuration from %q", cfgPath)
	kubeadmCfg, err := kubeadmconfig.LoadJoinConfigurationFromFile(cfgPath)
	if err != nil {
		return nil, err
	}
	// TODO: load for file
	joinCfg := new(apiv1.JoinConfiguration)
	joinCfg.JoinConfiguration = *kubeadmCfg
	return joinCfg, nil
}

func DefaultedJoinConfiguration(defaultcfg *apiv1.JoinConfiguration) (*apiv1.JoinConfiguration, error) {
	internalcfg := &apiv1.JoinConfiguration{}
	ocadmscheme.Scheme.Default(internalcfg)
	ocadmscheme.Scheme.Convert(defaultcfg, internalcfg, nil)
	// TODO: set dynamic config
	kubeadmVersionCfg := &kubeadmapiv1beta1.JoinConfiguration{}
	kubeadmscheme.Scheme.Default(kubeadmVersionCfg)
	kubeadmscheme.Scheme.Convert(&internalcfg.JoinConfiguration, kubeadmVersionCfg, nil)
	kubeadmInternalCfg, err := kubeadmconfig.DefaultedJoinConfiguration(kubeadmVersionCfg)
	if err != nil {
		return nil, err
	}
	internalcfg.JoinConfiguration = *kubeadmInternalCfg
	return internalcfg, nil
}
