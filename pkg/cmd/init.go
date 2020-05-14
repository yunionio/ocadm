package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/lithammer/dedent"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	clientset "k8s.io/client-go/kubernetes"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	"k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/validation"
	kubeadminitphases "k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/init"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"
	cmdutil "k8s.io/kubernetes/cmd/kubeadm/app/cmd/util"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"k8s.io/kubernetes/cmd/kubeadm/app/features"
	certsphase "k8s.io/kubernetes/cmd/kubeadm/app/phases/certs"
	kubeconfigphase "k8s.io/kubernetes/cmd/kubeadm/app/phases/kubeconfig"
	kubeadmutil "k8s.io/kubernetes/cmd/kubeadm/app/util"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/apiclient"
	kubeconfigutil "k8s.io/kubernetes/cmd/kubeadm/app/util/kubeconfig"

	ocadmscheme "yunion.io/x/ocadm/pkg/apis/scheme"
	v1 "yunion.io/x/ocadm/pkg/apis/v1"
	occmdutil "yunion.io/x/ocadm/pkg/cmd/util"
	"yunion.io/x/ocadm/pkg/occonfig"
	"yunion.io/x/ocadm/pkg/options"
	initphases "yunion.io/x/ocadm/pkg/phases/init"
	configutil "yunion.io/x/ocadm/pkg/util/config"
	"yunion.io/x/ocadm/pkg/util/kubectl"
	"yunion.io/x/ocadm/pkg/util/mysql"
	"yunion.io/x/ocadm/pkg/util/onecloud"
	"yunion.io/x/onecloud/pkg/mcclient"
)

var (
	initDoneTempl = template.Must(template.New("init").Parse(dedent.Dedent(`
		Your Kubernetes and Onecloud control-plane has initialized successfully!

		To start using your cluster, you need to run the following as a regular user:

		  mkdir -p $HOME/.kube
		  sudo cp -i {{.KubeConfigPath}} $HOME/.kube/config
		  sudo chown $(id -u):$(id -g) $HOME/.kube/config

		{{if .ControlPlaneEndpoint -}}
		{{if .UploadCerts -}}
		You can now join any number of the control-plane node running the following command on each as root:

		  {{.joinControlPlaneCommand}}

		Please note that the certificate-key gives access to cluster sensitive data, keep it secret!
		As a safeguard, uploaded-certs will be deleted in two hours; If necessary, you can use
		"ocadm init phase upload-certs --experimental-upload-certs" to reload certs afterward.

		{{else -}}
		You can now join any number of control-plane nodes by copying certificate authorities
		and service account keys on each node and then running the following as root:

		  {{.joinControlPlaneCommand}}

		{{end}}{{end}}Then you can join any number of worker nodes by running the following on each as root:

		{{.joinWorkerCommand}}
		`)))

	ocEvictionHard = map[string]string{
		"memory.available":  "100Mi",
		"nodefs.available":  "10%",
		"nodefs.inodesFree": "5%",
		"imagefs.available": "5%", //default is 15%
	}
)

// initOptions defines all the init options exposed via flags by ocadm init.
// Please note that this structure includes the public kubeadm config API, but only a subset of the options
// supported by this api will exposed as a flag
type initOptions struct {
	cfgPath                 string
	skipTokenPrint          bool
	dryRun                  bool
	kubeconfigDir           string
	kubeconfigPath          string
	featureGatesString      string
	ignorePreflightErrors   []string
	bto                     *options.BootstrapTokenOptions
	externalCfg             *v1.InitConfiguration
	hostCfg                 *onecloud.HostCfg
	uploadCerts             bool
	certificateKey          string
	skipCertificateKeyPrint bool
	printAddonYaml          bool
	operatorVersion         string
	nodeIP                  string
	glanceNode              bool
	baremetalNode           bool
	esxiNode                bool
}

var _ initphases.InitData = &initData{}

type initData struct {
	cfg                     *v1.InitConfiguration
	skipTokenPrint          bool
	dryRun                  bool
	kubeconfigDir           string
	kubeconfigPath          string
	ignorePreflightErrors   sets.String
	certificatesDir         string
	dryRunDir               string
	externalCA              bool
	client                  clientset.Interface
	kubectlClient           *kubectl.Client
	ocClient                *mcclient.ClientSession
	waiter                  apiclient.Waiter
	outputWriter            io.Writer
	uploadCerts             bool
	certificateKey          string
	skipCertificateKeyPrint bool
	enableHostAgent         bool
	printAddonYaml          bool
	operatorVersion         string
	nodeIP                  string
}

// NewCmdInit returns "deployer init" command
func NewCmdInit(out io.Writer, initOptions *initOptions) *cobra.Command {
	if initOptions == nil {
		initOptions = newInitOptions()
	}
	initRunner := workflow.NewRunner()

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Run this command in order to set up Kubernetes and OneCloud control plane",
		Run: func(cmd *cobra.Command, args []string) {
			c, err := initRunner.InitData(args)
			kubeadmutil.CheckErr(err)

			data := c.(*initData)
			data.enableHostAgent = initOptions.hostCfg.EnableHost
			fmt.Printf("[init] Using Kubernetes and Onecloud version: %s & %s\n", data.Cfg().KubernetesVersion, data.OnecloudCfg().OnecloudVersion)

			err = initRunner.Run(args)
			kubeadmutil.CheckErr(err)
			err = onecloud.GenerateDefaultHostConfig(initOptions.hostCfg)
			if err != nil {
				fmt.Printf("Generate host config error: %s", err)
			}

			err = showJoinCommand(data, out)
			kubeadmutil.CheckErr(err)
		},
		Args: cobra.NoArgs,
	}

	// add flags to the init command
	// init command local flags could be eventually inherited by the sub-commands automatically generated for phases
	externalKubeadmCfg := &initOptions.externalCfg.InitConfiguration
	externalCfg := initOptions.externalCfg
	AddInitConfigFlags(cmd.Flags(), initOptions.externalCfg)
	AddHostConfigFlags(cmd.Flags(), initOptions.hostCfg)
	AddKubeadmInitConfigFlags(cmd.Flags(), externalKubeadmCfg, &initOptions.featureGatesString)
	AddInitOtherFlags(cmd.Flags(), initOptions)
	initOptions.bto.AddTokenFlag(cmd.Flags())
	initOptions.bto.AddTTLFlag(cmd.Flags())
	options.AddImageMetaFlags(cmd.Flags(), &externalCfg.ImageRepository)

	// defines additional flag that are not used by the init command but that could be eventually used
	// by the sub-commands automatically generated for phases
	initRunner.SetAdditionalFlags(func(flags *flag.FlagSet) {
		options.AddKubeConfigFlag(flags, &initOptions.kubeconfigPath)
		options.AddKubeConfigDirFlag(flags, &initOptions.kubeconfigDir)
		options.AddControlPlanExtraArgsFlags(
			flags,
			&externalKubeadmCfg.APIServer.ExtraArgs,
			&externalKubeadmCfg.ControllerManager.ExtraArgs,
			&externalKubeadmCfg.Scheduler.ExtraArgs)
	})

	initRunner.AppendPhase(initphases.NewPreflightPhase())
	initRunner.AppendPhase(kubeadminitphases.NewKubeletStartPhase())
	initRunner.AppendPhase(kubeadminitphases.NewCertsPhase())
	initRunner.AppendPhase(kubeadminitphases.NewKubeConfigPhase())
	initRunner.AppendPhase(kubeadminitphases.NewControlPlanePhase())
	initRunner.AppendPhase(kubeadminitphases.NewEtcdPhase())
	initRunner.AppendPhase(kubeadminitphases.NewWaitControlPlanePhase())
	initRunner.AppendPhase(kubeadminitphases.NewUploadConfigPhase())
	initRunner.AppendPhase(kubeadminitphases.NewUploadCertsPhase())
	initRunner.AppendPhase(kubeadminitphases.NewMarkControlPlanePhase())
	initRunner.AppendPhase(kubeadminitphases.NewBootstrapTokenPhase())
	initRunner.AppendPhase(initphases.NewUploadConfigPhase())
	initRunner.AppendPhase(kubeadminitphases.NewAddonPhase())
	initRunner.AppendPhase(initphases.NewOCAddonPhase())
	initRunner.AppendPhase(initphases.NodeEnableHostAgent())

	// sets the data builder function, that will be used by the runner
	// both when running the entire workflow or single phases
	initRunner.SetDataInitializer(func(cmd *cobra.Command, args []string) (workflow.RunData, error) {
		return newInitData(cmd, args, initOptions, out)
	})

	// binds the Runner to deployer init command by altering
	// command help, adding --skip-phases flag and by adding phases subcommands
	initRunner.BindToCommand(cmd)

	return cmd
}

// AddInitConfigFlags adds init flags bound to the config to specified flagset
func AddInitConfigFlags(flagSet *flag.FlagSet, cfg *v1.InitConfiguration) {
	flagSet.StringVar(
		&cfg.OnecloudVersion, options.OnecloudVersion, cfg.OnecloudVersion,
		`Choose a specific Onecloud version for the control plane.`,
	)
	flagSet.StringVar(
		&cfg.Region, options.Region, cfg.Region,
		"Onecloud init region",
	)
	flagSet.StringVar(
		&cfg.Zone, options.Zone, cfg.Zone,
		"Onecloud init zone",
	)
	flagSet.StringVar(
		&cfg.MysqlConnection.Server, options.MysqlAddress, cfg.MysqlConnection.Server,
		"The IP address of mysql to connect.",
	)
	flagSet.StringVar(
		&cfg.MysqlConnection.Username, options.MysqlUser, cfg.MysqlConnection.Username,
		"The username of mysql to connect",
	)
	flagSet.StringVar(
		&cfg.MysqlConnection.Password, options.MysqlPassword, cfg.MysqlConnection.Password,
		"The password of mysql to connect",
	)
	flagSet.IntVar(
		&cfg.MysqlConnection.Port, options.MysqlPort, cfg.MysqlConnection.Port,
		"The port of mysql server",
	)
}

func AddHostConfigFlags(flagSet *flag.FlagSet, o *onecloud.HostCfg) {
	flagSet.StringArrayVar(
		&o.LocalImagePath, options.HostLocalImagePath, o.LocalImagePath,
		"Host configure: local image path",
	)
	flagSet.StringVar(
		&o.Hostname, options.Hostname, o.Hostname,
		"Host configure: host name",
	)
	flagSet.StringArrayVar(
		&o.Networks, options.HostNetworks, o.Networks,
		"Host configure: networks",
	)
	flagSet.BoolVar(
		&o.EnableHost, "enable-host-agent", o.EnableHost,
		"Enable host agent",
	)
}

// AddKubeadmInitConfigFlags adds init flags bound to the config to the specified flagset
func AddKubeadmInitConfigFlags(flagSet *flag.FlagSet, cfg *kubeadmapi.InitConfiguration, featureGatesString *string) {
	flagSet.StringVar(
		&cfg.LocalAPIEndpoint.AdvertiseAddress, options.APIServerAdvertiseAddress, cfg.LocalAPIEndpoint.AdvertiseAddress,
		"The IP address the API Server will advertise it's listening on. If not set the default network interface will be used.",
	)
	flagSet.Int32Var(
		&cfg.LocalAPIEndpoint.BindPort, options.APIServerBindPort, cfg.LocalAPIEndpoint.BindPort,
		"Port for the API Server to bind to.",
	)
	flagSet.StringVar(
		&cfg.Networking.ServiceSubnet, options.NetworkingServiceSubnet, cfg.Networking.ServiceSubnet,
		"Use alternative range of IP address for service VIPs.",
	)
	flagSet.StringVar(
		&cfg.Networking.PodSubnet, options.NetworkingPodSubnet, cfg.Networking.PodSubnet,
		"Specify range of IP addresses for the pod network. If set, the control plane will automatically allocate CIDRs for every node.",
	)
	flagSet.StringVar(
		&cfg.Networking.DNSDomain, options.NetworkingDNSDomain, cfg.Networking.DNSDomain,
		`Use alternative domain for services, e.g. "myorg.internal".`,
	)
	flagSet.StringVar(
		&cfg.KubernetesVersion, options.KubernetesVersion, cfg.KubernetesVersion,
		`Choose a specific Kubernetes version for the control plane.`,
	)
	flagSet.StringSliceVar(
		&cfg.APIServer.CertSANs, options.APIServerCertSANs, cfg.APIServer.CertSANs,
		`Optional extra Subject Alternative Names (SANs) to use for the API Server serving certificate. Can be both IP addresses and DNS names.`,
	)
	cmdutil.AddCRISocketFlag(flagSet, &cfg.NodeRegistration.CRISocket)
	flagSet.StringVar(featureGatesString, options.FeatureGatesString, *featureGatesString, "A set of key=value pairs that describe feature gates for various features. "+
		"Options are:\n"+strings.Join(features.KnownFeatures(&features.InitFeatureGates), "\n"))
	flagSet.StringVar(
		&cfg.ClusterConfiguration.ControlPlaneEndpoint, options.ControlPlaneEndpoint, "",
		"The load balancer vip for control plane master nodes.",
	)
}

// AddInitOtherFlags adds init flags that are not bound to a configuration file to the given flagset
// Note: All flags that are not bound to the cfg object should be allowed in cmd/kubeadm/app/apis/kubeadm/validation/validation.go
func AddInitOtherFlags(flagSet *flag.FlagSet, initOptions *initOptions) {
	options.AddConfigFlag(flagSet, &initOptions.cfgPath)
	flagSet.StringSliceVar(
		&initOptions.ignorePreflightErrors, options.IgnorePreflightErrors, initOptions.ignorePreflightErrors,
		"A list of checks whose errors will be shown as warnings. Example: 'IsPrivilegedUser,Swap'. Value 'all' ignores errors from all checks.",
	)
	flagSet.StringVar(
		&initOptions.nodeIP, options.NodeIP, initOptions.nodeIP,
		"Init Node IP",
	)
	flagSet.BoolVar(
		&initOptions.dryRun, options.DryRun, initOptions.dryRun,
		"Don't apply any changes; just output what would be done.",
	)
	flagSet.BoolVar(
		&initOptions.printAddonYaml, options.PrintAddonYaml, initOptions.printAddonYaml,
		"Print addon yaml manifest",
	)
	options.AddOperatorVersionFlags(flagSet, &initOptions.operatorVersion)
	options.AddGlanceNodeLabelFlag(flagSet, &initOptions.glanceNode, &initOptions.baremetalNode, &initOptions.esxiNode)
}

// newInitOptions returns a struct ready for being used for creating cmd init flags.
func newInitOptions() *initOptions {
	// initialize the public ocadm config API by applying defaults
	externalCfg := &v1.InitConfiguration{}
	ocadmscheme.Scheme.Default(externalCfg)

	// Create the options object for the bootstrap token-related flags, and override the default value for .Description
	bto := options.NewBootstrapTokenOptions()
	bto.Description = "The default bootstrap token generated by 'ocadm init'."

	return &initOptions{
		externalCfg:    externalCfg,
		bto:            bto,
		kubeconfigDir:  kubeadmconstants.KubernetesDir,
		kubeconfigPath: kubeadmconstants.GetAdminKubeConfigPath(),
		uploadCerts:    true, // always upload certs
		hostCfg:        new(onecloud.HostCfg),
	}
}

// newInitData returns a new initData struct to be used for the execution of the ocadm init workflow.
// This func takes care of validating initOptions passed to the command, and then it converts
// options into the internal InitConfiguration type that is used as input all the phases in the ocadm init workflow
func newInitData(cmd *cobra.Command, args []string, options *initOptions, out io.Writer) (*initData, error) {
	// Re-apply defaults to the public kubeadm API (this will set only values not exposed/not set as a flags)
	ocadmscheme.Scheme.Default(options.externalCfg)

	// Validate standalone flags values and/or combination of flags and then assigns
	// validated values to the public kubeadm config API when applicable
	var err error
	if options.externalCfg.FeatureGates, err = features.NewFeatureGate(&features.InitFeatureGates, options.featureGatesString); err != nil {
		return nil, err
	}

	ignorePreflightErrorsSet, err := validation.ValidateIgnorePreflightErrors(options.ignorePreflightErrors, nil)
	if err != nil {
		return nil, err
	}

	if err = validation.ValidateMixedArguments(cmd.Flags()); err != nil {
		return nil, err
	}

	if err = options.bto.ApplyTo(&options.externalCfg.InitConfiguration); err != nil {
		return nil, err
	}

	// Either use the config file if specified, or convert public kubeadm API to the internal InitConfiguration
	// and validates InitConfiguration
	cfg, err := configutil.LoadOrDefaultInitConfiguration(options.cfgPath, options.externalCfg)
	if err != nil {
		return nil, err
	}
	cfg.ComponentConfigs.Kubelet.EvictionHard = ocEvictionHard

	// override node name and CRI socket from the command line options
	if options.externalCfg.NodeRegistration.Name != "" {
		cfg.NodeRegistration.Name = options.externalCfg.NodeRegistration.Name
	}
	if options.externalCfg.NodeRegistration.CRISocket != "" {
		cfg.NodeRegistration.CRISocket = options.externalCfg.NodeRegistration.CRISocket
	}
	cfg.ComponentConfigs.Kubelet.NodeStatusUpdateFrequency = metav1.Duration{
		Duration: time.Second * 4,
	}

	// init node always as onecloud controller
	cfg.NodeRegistration.KubeletExtraArgs = customizeKubeletExtarArgs(
		options.hostCfg.EnableHost, options.glanceNode, options.baremetalNode, options.esxiNode, true, options.nodeIP)

	if err := configutil.VerifyAPIServerBindAddress(cfg.LocalAPIEndpoint.AdvertiseAddress); err != nil {
		return nil, err
	}
	if err := features.ValidateVersion(features.InitFeatureGates, cfg.FeatureGates, cfg.KubernetesVersion); err != nil {
		return nil, err
	}

	// if dry running creates a temporary folder for saving kubeadm generated files
	dryRunDir := ""
	if options.dryRun {
		if dryRunDir, err = ioutil.TempDir("", "kubeadm-init-dryrun"); err != nil {
			return nil, errors.Wrap(err, "couldn't create a temporary directory")
		}
	}

	// Checks if an external CA is provided by the user (when the CA Cert is present but the CA Key is not)
	externalCA, err := certsphase.UsingExternalCA(&cfg.InitConfiguration.ClusterConfiguration)
	if externalCA {
		// In case the certificates signed by CA (that should be provided by the user) are missing or invalid,
		// returns, because kubeadm can't regenerate them without the CA Key
		if err != nil {
			return nil, errors.Wrapf(err, "invalid or incomplete external CA")
		}

		// Validate that also the required kubeconfig files exists and are invalid, because
		// kubeadm can't regenerate them without the CA Key
		kubeconfigDir := options.kubeconfigDir
		if options.dryRun {
			kubeconfigDir = dryRunDir
		}
		if err := kubeconfigphase.ValidateKubeconfigsForExternalCA(kubeconfigDir, &cfg.InitConfiguration); err != nil {
			return nil, err
		}
	}

	// Checks if an external Front-Proxy CA is provided by the user (when the Front-Proxy CA Cert is present but the Front-Proxy CA Key is not)
	externalFrontProxyCA, err := certsphase.UsingExternalFrontProxyCA(&cfg.InitConfiguration.ClusterConfiguration)
	if externalFrontProxyCA {
		// In case the certificates signed by Front-Proxy CA (that should be provided by the user) are missing or invalid,
		// returns, because kubeadm can't regenerate them without the Front-Proxy CA Key
		if err != nil {
			return nil, errors.Wrapf(err, "invalid or incomplete external front-proxy CA")
		}
	}

	if options.uploadCerts && (externalCA || externalFrontProxyCA) {
		return nil, errors.New("can't use upload-certs with an external CA or an external front-proxy CA")
	}

	data := &initData{
		cfg:                     cfg,
		certificatesDir:         cfg.CertificatesDir,
		skipTokenPrint:          options.skipTokenPrint,
		dryRun:                  options.dryRun,
		dryRunDir:               dryRunDir,
		kubeconfigDir:           options.kubeconfigDir,
		kubeconfigPath:          options.kubeconfigPath,
		ignorePreflightErrors:   ignorePreflightErrorsSet,
		externalCA:              externalCA,
		outputWriter:            out,
		uploadCerts:             options.uploadCerts,
		certificateKey:          options.certificateKey,
		skipCertificateKeyPrint: options.skipCertificateKeyPrint,
		printAddonYaml:          options.printAddonYaml,
		operatorVersion:         options.operatorVersion,
	}

	return data, nil
}

// EnableHostAgent return is enable host agent
func (d *initData) EnabledHostAgent() bool {
	return d.enableHostAgent
}

// PrintAddonYaml only print onecloud addon yaml manifest
func (d *initData) PrintAddonYaml() bool {
	return d.printAddonYaml
}

// UploadCerts returns Uploadcerts flag.
func (d *initData) UploadCerts() bool {
	return d.uploadCerts
}

// CertificateKey returns the key used to encrypt the certs.
func (d *initData) CertificateKey() string {
	return d.certificateKey
}

// SetCertificateKey set the key used to encrypt the certs.
func (d *initData) SetCertificateKey(key string) {
	d.certificateKey = key
}

// SkipCertificateKeyPrint returns the skipCertificateKeyPrint flag.
func (d *initData) SkipCertificateKeyPrint() bool {
	return d.skipCertificateKeyPrint
}

// Cfg returns apis.InitConfiguration
func (d *initData) OnecloudCfg() *v1.InitConfiguration {
	return d.cfg
}

// kubeadmCfg returns kubeadmapi.InitConfiguration.
func (d *initData) Cfg() *kubeadmapi.InitConfiguration {
	return &d.cfg.InitConfiguration
}

// DryRun returns the DryRun flag.
func (d *initData) DryRun() bool {
	return d.dryRun
}

// SkipTokenPrint returns the SkipTokenPrint flag.
func (d *initData) SkipTokenPrint() bool {
	return d.skipTokenPrint
}

// IgnorePreflightErrors returns the IgnorePreflightErrors flag.
func (d *initData) IgnorePreflightErrors() sets.String {
	return d.ignorePreflightErrors
}

// CertificateWriteDir returns the path to the certificate folder or the temporary folder path in case of DryRun.
func (d *initData) CertificateWriteDir() string {
	if d.dryRun {
		return d.dryRunDir
	}
	return d.certificatesDir
}

// CertificateDir returns the CertificateDir as originally specified by the user.
func (d *initData) CertificateDir() string {
	return d.certificatesDir
}

// KubeConfigDir returns the path of the Kubernetes configuration folder or the temporary folder path in case of DryRun.
func (d *initData) KubeConfigDir() string {
	if d.dryRun {
		return d.dryRunDir
	}
	return d.kubeconfigDir
}

// KubeConfigPath returns the path to the kubeconfig file to use for connecting to Kubernetes
func (d *initData) KubeConfigPath() string {
	if d.dryRun {
		d.kubeconfigPath = filepath.Join(d.dryRunDir, kubeadmconstants.AdminKubeConfigFileName)
	}
	return d.kubeconfigPath
}

func (d *initData) KubectlClient() (*kubectl.Client, error) {
	if d.kubectlClient != nil {
		return d.kubectlClient, nil
	}
	cli, err := kubectl.NewClientFormKubeconfigFile(d.KubeConfigPath())
	if err != nil {
		return nil, err
	}
	d.kubectlClient = cli
	return d.kubectlClient, nil
}

// ManifestDir returns the path where manifest should be stored or the temporary folder path in case of DryRun.
func (d *initData) ManifestDir() string {
	if d.dryRun {
		return d.dryRunDir
	}
	return kubeadmconstants.GetStaticPodDirectory()
}

// KubeletDir returns path of the kubelet configuration folder or the temporary folder in case of DryRun.
func (d *initData) KubeletDir() string {
	if d.dryRun {
		return d.dryRunDir
	}
	return kubeadmconstants.KubeletRunDirectory
}

// ExternalCA returns true if an external CA is provided by the user.
func (d *initData) ExternalCA() bool {
	return d.externalCA
}

// OutputWriter returns the io.Writer used to write output to by this command.
func (d *initData) OutputWriter() io.Writer {
	return d.outputWriter
}

// Client returns a Kubernetes client to be used by kubeadm.
// This function is implemented as a singleton, thus avoiding to recreate the client when it is used by different phases.
// Important. This function must be called after the admin.conf kubeconfig file is created.
func (d *initData) Client() (clientset.Interface, error) {
	if d.client == nil {
		if d.dryRun {
			// If we're dry-running, we should create a faked client that answers some GETs in order to be able to do the full init flow and just logs the rest of requests
			dryRunGetter := apiclient.NewInitDryRunGetter(d.Cfg().NodeRegistration.Name, d.Cfg().Networking.ServiceSubnet)
			d.client = apiclient.NewDryRunClient(dryRunGetter, os.Stdout)
		} else {
			// If we're acting for real, we should create a connection to the API server and wait for it to come up
			var err error
			d.client, err = kubeconfigutil.ClientSetFromFile(d.KubeConfigPath())
			if err != nil {
				return nil, err
			}
		}
	}
	return d.client, nil
}

// Tokens returns an array of token strings.
func (d *initData) Tokens() []string {
	tokens := []string{}
	for _, bt := range d.Cfg().BootstrapTokens {
		tokens = append(tokens, bt.Token.String())
	}
	return tokens
}

func (d *initData) RootDBConnection() (*mysql.Connection, error) {
	info := &d.cfg.MysqlConnection
	return mysql.NewConnection(info)
}

func (d *initData) LocalAddress() string {
	return d.OnecloudCfg().HostLocalInfo.ManagementNetInterface.IPAddress()
}

func (d *initData) OnecloudAdminConfigPath() string {
	return occonfig.AdminConfigFilePath()
}

func (d *initData) OnecloudClientSession() (*mcclient.ClientSession, error) {
	var err error
	d.ocClient, err = occonfig.ClientSessionFromFile(d.OnecloudAdminConfigPath())
	if err != nil {
		return nil, err
	}
	return d.ocClient, nil
}

func (d *initData) OperatorVersion() string {
	return d.operatorVersion
}

func printJoinCommand(out io.Writer, adminKubeConfigPath, token string, i *initData) error {
	joinControlPlaneCommand, err := occmdutil.GetJoinControlPlaneCommand(adminKubeConfigPath, token, i.certificateKey, i.skipTokenPrint, i.skipCertificateKeyPrint)
	if err != nil {
		return err
	}

	joinWorkerCommand, err := occmdutil.GetJoinWorkerCommand(adminKubeConfigPath, token, i.skipTokenPrint)
	if err != nil {
		return err
	}

	ctx := map[string]interface{}{
		"KubeConfigPath":          adminKubeConfigPath,
		"ControlPlaneEndpoint":    i.Cfg().ControlPlaneEndpoint,
		"UploadCerts":             i.uploadCerts,
		"joinControlPlaneCommand": joinControlPlaneCommand,
		"joinWorkerCommand":       joinWorkerCommand,
	}

	return initDoneTempl.Execute(out, ctx)
}

// showJoinCommand prints the join command after all the phases in init have finished
func showJoinCommand(i *initData, out io.Writer) error {
	adminKubeConfigPath := i.KubeConfigPath()

	// Prints the join command, multiple times in case the user has multiple tokens
	for _, token := range i.Tokens() {
		if err := printJoinCommand(out, adminKubeConfigPath, token, i); err != nil {
			return errors.Wrap(err, "failed to print join command")
		}
	}

	return nil
}
