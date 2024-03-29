package options

import (
	kubeadmoptions "k8s.io/kubernetes/cmd/kubeadm/app/cmd/options"
)

const (
	// CfgPath flag sets the path to onecloud config file.
	CfgPath = "config"

	// KubeadmCfgPath flag sets the path to kubeadm config file.
	KubeadmCfgPath = "kubeadm-config"
)

const (
	MysqlAddress                       = "mysql-host"
	MysqlUser                          = "mysql-user"
	MysqlPassword                      = "mysql-password"
	MysqlPort                          = "mysql-port"
	Region                             = "region"
	Zone                               = "zone"
	OnecloudVersion                    = "onecloud-version"
	ControlPlaneEndpoint               = "control-plane-endpoint"
	HostLocalImagePath                 = "host-local-image-path"
	Hostname                           = "host-name"
	HostNetworks                       = "host-networks"
	AsOnecloudController               = "as-onecloud-controller"
	PrintAddonYaml                     = "print-addon-yaml"
	OperatorVersion                    = "operator-version"
	NodeIP                             = "node-ip"
	AddonCalicoIpAutodetectionMethod   = "addon-calico-ip-autodetection-method"
	AddonCalicoiFelixChaininsertmode   = "addon-calico-felix-chaininsertmode"
	NodeCIDRMaskSize                   = "node-cidr-mask-size"
	AddonCalicoIPV4BlockSize           = "addon-calico-ipv4-block-size"
	HighAvailabilityVIP                = "high-availability-vip"
	KeepalivedVersionTag               = "keepalived-version-tag"
	LonghornDataPath                   = "longhorn-data-path"
	LonghornOverProvisioningPercentage = "longhorn-over-provisioning-percentage"
	LonghornReplicaCount               = "longhorn-replica-count"
	PVCMigrateToLonghorn               = "source-pvc"
)

const (
	// APIServerAdvertiseAddress flag sets the IP address the API Server will advertise it's listening on. Specify '0.0.0.0' to use the address of the default network interface.
	APIServerAdvertiseAddress = kubeadmoptions.APIServerAdvertiseAddress

	// APIServerBindPort flag sets the port for the API Server to bind to.
	APIServerBindPort = kubeadmoptions.APIServerBindPort

	// APIServerCertSANs flag sets extra Subject Alternative Names (SANs) to use for the API Server serving certificate. Can be both IP addresses and DNS names.
	APIServerCertSANs = kubeadmoptions.APIServerCertSANs

	// APIServerExtraArgs flag sets a extra flags to pass to the API Server or override default ones in form of <flagname>=<value>.
	APIServerExtraArgs = kubeadmoptions.APIServerExtraArgs

	// CertificatesDir flag sets the path where to save and read the certificates.
	CertificatesDir = kubeadmoptions.CertificatesDir

	// ControllerManagerExtraArgs flag sets extra flags to pass to the Controller Manager or override default ones in form of <flagname>=<value>.
	ControllerManagerExtraArgs = kubeadmoptions.ControllerManagerExtraArgs

	// DryRun flag instruct kubeadm to don't apply any changes; just output what would be done.
	DryRun = kubeadmoptions.DryRun

	// FeatureGatesString flag sets key=value pairs that describe feature gates for various features.
	FeatureGatesString = kubeadmoptions.FeatureGatesString

	// IgnorePreflightErrors sets the path a list of checks whose errors will be shown as warnings. Example: 'IsPrivilegedUser,Swap'. Value 'all' ignores errors from all checks.
	IgnorePreflightErrors = kubeadmoptions.IgnorePreflightErrors

	// ImageRepository sets the container registry to pull control plane images from.
	ImageRepository = kubeadmoptions.ImageRepository

	// KubeconfigDir flag sets the path where to save the kubeconfig file.
	KubeconfigDir = kubeadmoptions.KubeconfigDir

	// KubeconfigPath flag sets the kubeconfig file to use when talking to the cluster. If the flag is not set, a set of standard locations are searched for an existing KubeConfig file.
	KubeconfigPath = kubeadmoptions.KubeconfigPath

	// KubernetesVersion flag sets the Kubernetes version for the control plane.
	KubernetesVersion = kubeadmoptions.KubernetesVersion

	// NetworkingDNSDomain flag sets the domain for services, e.g. "myorg.internal".
	NetworkingDNSDomain = kubeadmoptions.NetworkingDNSDomain

	// NetworkingServiceSubnet flag sets the range of IP address for service VIPs.
	NetworkingServiceSubnet = kubeadmoptions.NetworkingServiceSubnet

	// NetworkingPodSubnet flag sets the range of IP addresses for the pod network. If set, the control plane will automatically allocate CIDRs for every node.
	NetworkingPodSubnet = kubeadmoptions.NetworkingPodSubnet

	// NodeCRISocket flag sets the CRI socket to connect to.
	NodeCRISocket = kubeadmoptions.NodeCRISocket

	// NodeName flag sets the node name.
	NodeName = kubeadmoptions.NodeName

	// SchedulerExtraArgs flag sets extra flags to pass to the Scheduler or override default ones in form of <flagname>=<value>".
	SchedulerExtraArgs = kubeadmoptions.SchedulerExtraArgs

	// SkipTokenPrint flag instruct kubeadm to skip printing of the default bootstrap token generated by 'kubeadm init'.
	SkipTokenPrint = kubeadmoptions.SkipTokenPrint

	// CSROnly flag instructs kubeadm to create CSRs instead of automatically creating or renewing certs
	CSROnly = kubeadmoptions.CSROnly

	// CSRDir flag sets the location for CSRs and flags to be output
	CSRDir = kubeadmoptions.CSRDir

	// TokenStr flags sets both the discovery-token and the tls-bootstrap-token when those values are not provided
	TokenStr = kubeadmoptions.TokenStr

	// TokenTTL flag sets the time to live for token
	TokenTTL = kubeadmoptions.TokenTTL

	// TokenUsages flag sets the usages of the token
	TokenUsages = kubeadmoptions.TokenUsages

	// TokenGroups flag sets the authentication groups of the token
	TokenGroups = kubeadmoptions.TokenGroups

	// TokenDescription flag sets the description of the token
	TokenDescription = kubeadmoptions.TokenDescription

	// TLSBootstrapToken flag sets the token used to temporarily authenticate with the Kubernetes Control Plane to submit a certificate signing request (CSR) for a locally created key pair
	TLSBootstrapToken = kubeadmoptions.TLSBootstrapToken

	// TokenDiscovery flag sets the token used to validate cluster information fetched from the API server (for token-based discovery)
	TokenDiscovery = kubeadmoptions.TokenDiscovery

	// TokenDiscoveryCAHash flag instruct kubeadm to validate that the root CA public key matches this hash (for token-based discovery)
	TokenDiscoveryCAHash = kubeadmoptions.TokenDiscoveryCAHash

	// TokenDiscoverySkipCAHash flag instruct kubeadm to skip CA hash verification (for token-based discovery)
	TokenDiscoverySkipCAHash = kubeadmoptions.TokenDiscoverySkipCAHash

	// FileDiscovery flag sets the file or URL from which to load cluster information (for file-based discovery)
	FileDiscovery = kubeadmoptions.FileDiscovery

	// ControlPlane flag instruct kubeadm to create a new control plane instance on this node
	ControlPlane = kubeadmoptions.ControlPlane

	// UploadCerts flag instruct kubeadm to upload certificates
	UploadCerts = kubeadmoptions.UploadCerts

	// CertificateKey flag sets the key used to encrypt and decrypt certificate secrets
	CertificateKey = kubeadmoptions.CertificateKey

	// SkipCertificateKeyPrint flag instruct kubeadm to skip printing certificate key used to encrypt certs by 'kubeadm init'.
	SkipCertificateKeyPrint = kubeadmoptions.SkipCertificateKeyPrint

	ForceReset = kubeadmoptions.ForceReset
)
