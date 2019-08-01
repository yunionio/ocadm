package options

import (
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	kubeadmscheme "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/scheme"
	kubeadmapiv1beta2 "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta2"
	kubeadmoptions "k8s.io/kubernetes/cmd/kubeadm/app/cmd/options"
)

// NewBootstrapTokenOptions creates a new BootstrapTokenOptions object with the default values
func NewBootstrapTokenOptions() *BootstrapTokenOptions {
	bto := &BootstrapTokenOptions{
		BootstrapTokenOptions: kubeadmoptions.NewBootstrapTokenOptions(),
	}
	return bto
}

type BootstrapTokenOptions struct {
	*kubeadmoptions.BootstrapTokenOptions
}

func (bio *BootstrapTokenOptions) ApplyTo(cfg *kubeadmapi.InitConfiguration) error {
	externalCfg := &kubeadmapiv1beta2.InitConfiguration{}
	kubeadmscheme.Scheme.Default(externalCfg)
	if err := bio.BootstrapTokenOptions.ApplyTo(externalCfg); err != nil {
		return err
	}
	internalExternalCfg := &kubeadmapi.InitConfiguration{}
	if err := kubeadmscheme.Scheme.Convert(externalCfg, internalExternalCfg, nil); err != nil {
		return err
	}
	cfg.BootstrapTokens = internalExternalCfg.BootstrapTokens
	return nil
}
