package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	kubeadmapiv1beta1 "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta1"
	"k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/validation"

	//"k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/validation"
	kubeadmutil "k8s.io/kubernetes/cmd/kubeadm/app/util"
	kubeadmconfig "k8s.io/kubernetes/cmd/kubeadm/app/util/config"

	"yunion.io/x/ocadm/pkg/apis/constants"
	ocadmscheme "yunion.io/x/ocadm/pkg/apis/scheme"
	apiv1 "yunion.io/x/ocadm/pkg/apis/v1"
	ocadmutil "yunion.io/x/ocadm/pkg/util"
	netutil "yunion.io/x/ocadm/pkg/util/net"
)

// MarshalInitConfigurationToBytes marshals the internal InitConfiguration object to bytes. It writes the embedded
// ClusterConfiguration object with ComponentConfigs out as separate YAML documents
func MarshalInitConfigurationToBytes(cfg *apiv1.InitConfiguration, gv schema.GroupVersion) ([]byte, error) {
	initbytes, err := kubeadmutil.MarshalToYamlForCodecs(cfg, gv, ocadmscheme.Codecs)
	if err != nil {
		return []byte{}, err
	}
	allFiles := [][]byte{initbytes}

	// Exception: If the specified groupversion is targeting the internal type, don't print embedded ClusterConfiguration contents
	// This is mostly used for unit testing. In a real scenario the internal version of the API is never marshalled as-is.
	if gv.Version != runtime.APIVersionInternal {
		kubeadmBytes, err := kubeadmconfig.MarshalInitConfigurationToBytes(&cfg.InitConfiguration, kubeadmapiv1beta1.SchemeGroupVersion)
		if err != nil {
			return []byte{}, err
		}
		clusterbytes, err := MarshalClusterConfigurationToBytes(&cfg.ClusterConfiguration, gv)
		if err != nil {
			return []byte{}, err
		}
		allFiles = append(allFiles, kubeadmBytes, clusterbytes)
	}
	return bytes.Join(allFiles, []byte(constants.YAMLDocumentSeparator)), nil
}

// MarshalClusterConfigurationToBytes marshals the internal ClusterConfiguration object to bytes. It writes the embedded
// ComponentConfiguration objects out as separate YAML documents
func MarshalClusterConfigurationToBytes(clustercfg *apiv1.ClusterConfiguration, gv schema.GroupVersion) ([]byte, error) {
	clusterbytes, err := kubeadmutil.MarshalToYamlForCodecs(clustercfg, gv, ocadmscheme.Codecs)
	if err != nil {
		return []byte{}, err
	}
	allFiles := [][]byte{clusterbytes}
	return bytes.Join(allFiles, []byte(constants.YAMLDocumentSeparator)), nil
}

func LoadInitConfigurationFromFile(cfgPath string) (*apiv1.InitConfiguration, error) {
	klog.V(1).Infof("loading configuration from %q", cfgPath)

	b, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to read config from %q ", cfgPath)
	}
	return BytesToInitConfiguration(b)
}

func LoadOrDefaultInitConfiguration(cfgPath string, defaultcfg *apiv1.InitConfiguration) (*apiv1.InitConfiguration, error) {
	if cfgPath != "" {
		return LoadInitConfigurationFromFile(cfgPath)
	}

	return DefaultedInitConfiguration(defaultcfg)
}

func DefaultedInitConfiguration(defaultcfg *apiv1.InitConfiguration) (*apiv1.InitConfiguration, error) {
	internalcfg := &apiv1.InitConfiguration{}
	ocadmscheme.Scheme.Default(internalcfg)
	ocadmscheme.Scheme.Convert(defaultcfg, internalcfg, nil)

	// Applies dynamic defaults to settings not provided with flags
	if err := kubeadmconfig.SetInitDynamicDefaults(&internalcfg.InitConfiguration); err != nil {
		return nil, err
	}
	// Validates cfg (flags/configs + defaults + dynamic defaults)
	if err := validation.ValidateInitConfiguration(&internalcfg.InitConfiguration).ToAggregate(); err != nil {
		return nil, err
	}
	if err := SetInitDynamicDefaults(internalcfg); err != nil {
		return nil, err
	}
	return internalcfg, nil
}

// BytesToInitConfiguration converts a byte slice to an internal, defaulted and validated InitConfiguration object.
// The map may contain many different YAML documents. These YAML documents are parsed one-by-one
// and well-known ComponentConfig GroupVersionKinds are stored inside of the internal InitConfiguration struct.
// The resulting InitConfiguration is then dynamically defaulted and validated prior to return.
func BytesToInitConfiguration(b []byte) (*apiv1.InitConfiguration, error) {
	kubeadmInitCfg, err := kubeadmconfig.BytesToInitConfiguration(b)
	if err != nil {
		return nil, err
	}

	gvkmap, err := ocadmutil.SplitYAMLDocuments(b)
	if err != nil {
		return nil, err
	}

	return documentMapToInitConfiguration(gvkmap, kubeadmInitCfg, false)
}

// documentMapToInitConfiguration converts a map of GVKs and YAML documents to defaulted and validated configuration object.
func documentMapToInitConfiguration(gvkmap map[schema.GroupVersionKind][]byte, kubeadmInitCfg *kubeadmapi.InitConfiguration, allowDeprecated bool) (*apiv1.InitConfiguration, error) {
	var initCfg *apiv1.InitConfiguration
	var clusterCfg *apiv1.ClusterConfiguration

	for gvk, fileContent := range gvkmap {
		if ocadmutil.GroupVersionKindsHasInitConfiguration(gvk) {
			initCfg = &apiv1.InitConfiguration{}
			// Decode the bytes into the internal struct. Under the hood, the bytes will be unmarshalled into the
			// right external version, defaulted, and converted into the internal version.
			if err := runtime.DecodeInto(ocadmscheme.Codecs.UniversalDecoder(), fileContent, initCfg); err != nil {
				return nil, err
			}
			continue
		}
		if ocadmutil.GroupVersionKindsHasClusterConfiguration(gvk) {
			clusterCfg = &apiv1.ClusterConfiguration{}
			// Decode the bytes into the internal struct. Under the hood, the bytes will be unmarshalled into the
			// right external version, defaulted, and converted into the internal version.
			if err := runtime.DecodeInto(ocadmscheme.Codecs.UniversalDecoder(), fileContent, clusterCfg); err != nil {
				return nil, err
			}
			continue
		}

		fmt.Printf("[oc-config] WARNING: Ignored YAML document with GroupVersionKind %v\n", gvk)
	}

	// Enforce that InitConfiguration and/or ClusterConfiguration has to exist among the YAML documents
	if initCfg == nil && clusterCfg == nil {
		return nil, errors.New("no InitConfiguration or ClusterConfiguration kind was found in the YAML file")
	}

	// If InitConfiguration wasn't given, default it by creating an external struct instance, default it and convert into the internal type
	if initCfg == nil {
		extinitcfg := &apiv1.InitConfiguration{}
		ocadmscheme.Scheme.Default(extinitcfg)
		// Set initcfg to an empty struct value the deserializer will populate
		initCfg = &apiv1.InitConfiguration{}
		ocadmscheme.Scheme.Convert(extinitcfg, initCfg, nil)
	}
	// If ClusterConfiguration was given, populate it in the InitConfiguration struct
	if clusterCfg != nil {
		initCfg.ClusterConfiguration = *clusterCfg
	}

	if kubeadmInitCfg != nil {
		initCfg.InitConfiguration = *kubeadmInitCfg
	}

	// Applies dynamic defaults to settings not provided with flags
	if err := SetInitDynamicDefaults(initCfg); err != nil {
		return nil, err
	}

	// TODO: Validates cfg (flags/configs + defaults + dynamic defaults)
	return initCfg, nil
}

func SetInitDynamicDefaults(cfg *apiv1.InitConfiguration) error {
	if err := SetHostLocalDynamicDefaults(&cfg.HostLocalInfo, cfg.LocalAPIEndpoint.AdvertiseAddress); err != nil {
		return err
	}
	if err := SetServicesDynamicDefaults(&cfg.ClusterConfiguration, cfg.LocalAPIEndpoint.AdvertiseAddress); err != nil {
		return err
	}
	return nil
}

func SetHostLocalDynamicDefaults(info *apiv1.HostLocalInfo, kubeadmAPILocalAddress string) error {
	intf, ip, err := netutil.ChooseBindAddress(net.ParseIP(kubeadmAPILocalAddress))
	if err != nil {
		return errors.Wrapf(err, "Failed to choose address %s", kubeadmAPILocalAddress)
	}
	routes, err := netutil.GetAllDefaultRoutes()
	if err != nil {
		return errors.Wrap(err, "get default routes")
	}
	var gateway net.IP
	for _, route := range routes {
		if route.Interface == intf.Name {
			gateway = route.Gateway
			break
		}
	}
	info.ManagementNetInterface.Interface = intf.Name
	info.ManagementNetInterface.Address = ip
	info.ManagementNetInterface.Gateway = gateway
	addrs, err := intf.Addrs()
	if err != nil {
		return errors.Wrap(err, "get addrs")
	}
	var wantAddr net.Addr = nil
	for _, addr := range addrs {
		if strings.HasPrefix(addr.String(), ip.String()) {
			wantAddr = addr
			break
		}
	}
	if wantAddr == nil {
		return errors.Wrapf(err, "not found %s at %s", ip.String(), intf.Name)
	}
	parts := strings.Split(wantAddr.String(), "/")
	if len(parts) != 2 {
		return errors.Errorf("invalid addr %s", wantAddr)
	}
	maskLen, err := strconv.Atoi(parts[1])
	if err != nil {
		return errors.Wrapf(err, "invalid mask len %s", parts[1])
	}
	info.ManagementNetInterface.MaskLen = maskLen
	return nil
}

func SetServicesDynamicDefaults(cfg *apiv1.ClusterConfiguration, authAddress string) error {
	return nil
}
