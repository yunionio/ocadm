package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/lithammer/dedent"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/sets"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/klog"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	"k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/validation"
	kubeadmjoinphases "k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/join"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"
	cmdutil "k8s.io/kubernetes/cmd/kubeadm/app/cmd/util"
	"k8s.io/kubernetes/cmd/kubeadm/app/discovery"
	kubeadmutil "k8s.io/kubernetes/cmd/kubeadm/app/util"
	kubeconfigutil "k8s.io/kubernetes/cmd/kubeadm/app/util/kubeconfig"
	operatorconstants "yunion.io/x/onecloud-operator/pkg/apis/constants"

	"yunion.io/x/ocadm/pkg/apis/constants"
	ocadmscheme "yunion.io/x/ocadm/pkg/apis/scheme"
	apiv1 "yunion.io/x/ocadm/pkg/apis/v1"
	"yunion.io/x/ocadm/pkg/options"
	"yunion.io/x/ocadm/pkg/phases/addons/keepalived"
	joinphases "yunion.io/x/ocadm/pkg/phases/join"
	configutil "yunion.io/x/ocadm/pkg/util/config"
	"yunion.io/x/ocadm/pkg/util/onecloud"
)

var (
	joinWorkerNodeDoneMsg = dedent.Dedent(`
		This node has joined the cluster:
		* Certificate signing request was sent to apiserver and a response was received.
		* The Kubelet was informed of the new secure connection details.

		Run 'kubectl get nodes' on the control-plane to see this node join the cluster.

		`)

	joinControPlaneDoneTemp = template.Must(template.New("join").Parse(dedent.Dedent(`
		This node has joined the cluster and a new control plane instance was created:

		* Certificate signing request was sent to apiserver and approval was received.
		* The Kubelet was informed of the new secure connection details.
		* Control plane (master) label and taint were applied to the new node.
		* The Kubernetes control plane instances scaled up.
		{{.etcdMessage}}

		To start administering your cluster from this node, you need to run the following as a regular user:

			mkdir -p $HOME/.kube
			sudo cp -i {{.KubeConfigPath}} $HOME/.kube/config
			sudo chown $(id -u):$(id -g) $HOME/.kube/config

		Run 'kubectl get nodes' to see this node join the cluster.

		`)))
)

// joinOptions defines all the options exposed via flags by kubeadm join.
// Please note that this structure includes the public kubeadm config API, but only a subset of the options
// supported by this api will be exposed as a flag.
type joinOptions struct {
	cfgPath               string
	token                 string
	controlPlane          bool
	ignorePreflightErrors []string
	externalcfg           *apiv1.JoinConfiguration
	hostCfg               *onecloud.HostCfg
	certificateKey        string
	asOnecloudController  bool
	nodeIP                string
	highAvailabilityVIP   string
	keepalivedVersionTag  string
	glanceNode            bool
	baremetalNode         bool
	esxiNode              bool
	upgradeFromV2         bool
	hostInterface         string
}

// compile-time assert that the local data object satisfies the phases data interface.
var _ joinphases.JoinData = &joinData{}

// joinData defines all the runtime information used when running the kubeadm join worklow;
// this data is shared across all the phases that are included in the workflow.
type joinData struct {
	cfg                   *apiv1.JoinConfiguration
	skipTokenPrint        bool
	initCfg               *apiv1.InitConfiguration
	tlsBootstrapCfg       *clientcmdapi.Config
	clientSet             *clientset.Clientset
	ignorePreflightErrors sets.String
	outputWriter          io.Writer
	certificateKey        string
	enableHostAgent       bool
	asOnecloudController  bool
	nodeIP                string
	highAvailabilityVIP   string
	keepalivedVersionTag  string
	glanceNode            bool
	baremetalNode         bool
	esxiNode              bool
	hostInterface         string
}

// NewCmdJoin returns "ocadm join" command
// NB. joinOptions is exposed as parameter for allowing unit testing of
//     the newJoinData method, that implements all the command options validation logic
func NewCmdJoin(out io.Writer, joinOptions *joinOptions) *cobra.Command {
	if joinOptions == nil {
		joinOptions = newJoinOptions()
	}
	joinRunner := workflow.NewRunner()

	cmd := &cobra.Command{
		Use:   "join [api-server-endpoint]",
		Short: "Run this on any machine you wish to join an existing cluster",
		Run: func(cmd *cobra.Command, args []string) {

			c, err := joinRunner.InitData(args)
			kubeadmutil.CheckErr(err)

			data := c.(*joinData)
			data.enableHostAgent = joinOptions.hostCfg.EnableHost
			data.glanceNode = joinOptions.glanceNode
			data.esxiNode = joinOptions.esxiNode
			data.baremetalNode = joinOptions.baremetalNode
			// by default, control plane node as onecloud controller
			if joinOptions.asOnecloudController {
				data.asOnecloudController = true
			}
			err = joinRunner.Run(args)
			kubeadmutil.CheckErr(err)

			if !joinOptions.upgradeFromV2 {
				err = onecloud.GenerateDefaultHostConfig(joinOptions.hostCfg)
				kubeadmutil.CheckErr(err)
			}

			// if the node is hosting a new control plane instance
			if data.cfg.ControlPlane != nil {
				// outputs the join control plane done message and exit
				etcdMessage := ""
				if data.initCfg.Etcd.External == nil {
					etcdMessage = "* A new etcd member was added to the local/stacked etcd cluster."
				}

				ctx := map[string]string{
					"KubeConfigPath": constants.GetAdminKubeConfigPath(),
					"etcdMessage":    etcdMessage,
				}
				_ = joinControPlaneDoneTemp.Execute(data.outputWriter, ctx)

			} else {
				// otherwise, if the node joined as a worker node;
				// outputs the join done message and exit
				_, _ = fmt.Fprintf(data.outputWriter, joinWorkerNodeDoneMsg)
			}
		},
		// We accept the control-plane location as an optional positional argument
		Args: cobra.MaximumNArgs(1),
	}

	addJoinConfigFlags(cmd.Flags(), joinOptions.externalcfg)
	addJoinOtherFlags(cmd.Flags(), joinOptions)
	AddHostConfigFlags(cmd.Flags(), joinOptions.hostCfg)
	joinRunner.AppendPhase(kubeadmjoinphases.NewPreflightPhase())
	joinRunner.AppendPhase(keepalived.NewKeepalivedPhase())
	joinRunner.AppendPhase(kubeadmjoinphases.NewControlPlanePreparePhase())
	// joinRunner.AppendPhase(joinphases.NewNodePreparePhase())
	joinRunner.AppendPhase(kubeadmjoinphases.NewCheckEtcdPhase())
	joinRunner.AppendPhase(kubeadmjoinphases.NewKubeletStartPhase())
	joinRunner.AppendPhase(kubeadmjoinphases.NewControlPlaneJoinPhase())
	joinRunner.AppendPhase(joinphases.NewControlPlaneJoinPhase())

	// sets the data builder function, that will be used by the runner
	// both when running the entire workflow or single phases
	joinRunner.SetDataInitializer(func(cmd *cobra.Command, args []string) (workflow.RunData, error) {
		return newJoinData(cmd, args, joinOptions, out)
	})

	// binds the Runner to kubeadm join command by altering
	// command help, adding --skip-phases flag and by adding phases subcommands
	joinRunner.BindToCommand(cmd)

	return cmd
}

// addJoinConfigFlags adds join flags bound to the config to the specified flagset
func addJoinConfigFlags(flagSet *flag.FlagSet, cfg *apiv1.JoinConfiguration) {

	flagSet.StringVar(
		&cfg.NodeRegistration.Name, options.NodeName, cfg.NodeRegistration.Name,
		`Specify the node name.`,
	)
	// add control plane endpoint flags to the specified flagset
	flagSet.StringVar(
		&cfg.ControlPlane.LocalAPIEndpoint.AdvertiseAddress, options.APIServerAdvertiseAddress, cfg.ControlPlane.LocalAPIEndpoint.AdvertiseAddress,
		"If the node should host a new control plane instance, the IP address the API Server will advertise it's listening on. If not set the default network interface will be used.",
	)
	flagSet.Int32Var(
		&cfg.ControlPlane.LocalAPIEndpoint.BindPort, options.APIServerBindPort, cfg.ControlPlane.LocalAPIEndpoint.BindPort,
		"If the node should host a new control plane instance, the port for the API Server to bind to.",
	)
	// adds bootstrap token specific discovery flags to the specified flagset
	flagSet.StringVar(
		&cfg.Discovery.BootstrapToken.Token, options.TokenDiscovery, "",
		"For token-based discovery, the token used to validate cluster information fetched from the API server.",
	)
	flagSet.StringSliceVar(
		&cfg.Discovery.BootstrapToken.CACertHashes, options.TokenDiscoveryCAHash, []string{},
		"For token-based discovery, validate that the root CA public key matches this hash (format: \"<type>:<value>\").",
	)
	flagSet.BoolVar(
		&cfg.Discovery.BootstrapToken.UnsafeSkipCAVerification, options.TokenDiscoverySkipCAHash, false,
		"For token-based discovery, allow joining without --discovery-token-ca-cert-hash pinning.",
	)
	//	discovery via kube config file flag
	flagSet.StringVar(
		&cfg.Discovery.File.KubeConfigPath, options.FileDiscovery, "",
		"For file-based discovery, a file or URL from which to load cluster information.",
	)
	flagSet.StringVar(
		&cfg.Discovery.TLSBootstrapToken, options.TLSBootstrapToken, cfg.Discovery.TLSBootstrapToken,
		`Specify the token used to temporarily authenticate with the Kubernetes Control Plane while joining the node.`,
	)
	cmdutil.AddCRISocketFlag(flagSet, &cfg.NodeRegistration.CRISocket)
}

// addJoinOtherFlags adds join flags that are not bound to a configuration file to the given flagset
func addJoinOtherFlags(flagSet *flag.FlagSet, joinOptions *joinOptions) {
	flagSet.StringVar(
		&joinOptions.cfgPath, options.CfgPath, joinOptions.cfgPath,
		"Path to ocadm config file.",
	)
	flagSet.StringSliceVar(
		&joinOptions.ignorePreflightErrors, options.IgnorePreflightErrors, joinOptions.ignorePreflightErrors,
		"A list of checks whose errors will be shown as warnings. Example: 'IsPrivilegedUser,Swap'. Value 'all' ignores errors from all checks.",
	)
	flagSet.StringVar(
		&joinOptions.token, options.TokenStr, "",
		"Use this token for both discovery-token and tls-bootstrap-token when those values are not provided.",
	)
	flagSet.BoolVar(
		&joinOptions.controlPlane, options.ControlPlane, joinOptions.controlPlane,
		"Create a new control plane instance on this node",
	)
	flagSet.StringVar(
		&joinOptions.certificateKey, options.CertificateKey, "",
		"Use this key to decrypt the certificate secrets uploaded by init.",
	)
	flagSet.BoolVar(
		&joinOptions.asOnecloudController, options.AsOnecloudController, joinOptions.asOnecloudController,
		"Join node and set node as onecloud controller",
	)
	flagSet.StringVar(
		&joinOptions.nodeIP, options.NodeIP, joinOptions.nodeIP,
		"Join node IP",
	)
	flagSet.StringVar(
		&joinOptions.highAvailabilityVIP, options.HighAvailabilityVIP, joinOptions.highAvailabilityVIP,
		"high Availability VIP",
	)
	flagSet.StringVar(
		&joinOptions.keepalivedVersionTag, options.KeepalivedVersionTag, joinOptions.keepalivedVersionTag,
		fmt.Sprintf(`keepalived docker image tag within yunion aliyun registry. (default: "%s")`, constants.DefaultKeepalivedVersionTag),
	)
	options.AddGlanceNodeLabelFlag(flagSet, &joinOptions.glanceNode, &joinOptions.baremetalNode, &joinOptions.esxiNode)
	options.AddUpgradeFromV2Flags(flagSet, &joinOptions.upgradeFromV2)
}

// newJoinOptions returns a struct ready for being used for creating cmd join flags.
func newJoinOptions() *joinOptions {
	// initialize the public ocadm config API by applying defaults
	externalcfg := &apiv1.JoinConfiguration{}

	// Add optional config objects to host flags.
	// un-set objects will be cleaned up afterwards (into newJoinData func)
	externalcfg.Discovery.File = &kubeadmapi.FileDiscovery{}
	externalcfg.Discovery.BootstrapToken = &kubeadmapi.BootstrapTokenDiscovery{}
	externalcfg.ControlPlane = &kubeadmapi.JoinControlPlane{}

	// Apply defaults
	ocadmscheme.Scheme.Default(externalcfg)

	return &joinOptions{
		externalcfg: externalcfg,
		hostCfg:     new(onecloud.HostCfg),
	}
}

// newJoinData returns a new joinData struct to be used for the execution of the kubeadm join workflow.
// This func takes care of validating joinOptions passed to the command, and then it converts
// options into the internal JoinConfiguration type that is used as input all the phases in the kubeadm join workflow
func newJoinData(cmd *cobra.Command, args []string, opt *joinOptions, out io.Writer) (*joinData, error) {
	// Re-apply defaults to the public kubeadm API (this will set only values not exposed/not set as a flags)
	ocadmscheme.Scheme.Default(opt.externalcfg)

	// Validate standalone flags values and/or combination of flags and then assigns
	// validated values to the public kubeadm config API when applicable

	// if a token is provided, use this value for both discovery-token and tls-bootstrap-token when those values are not provided
	if len(opt.token) > 0 {
		if len(opt.externalcfg.Discovery.TLSBootstrapToken) == 0 {
			opt.externalcfg.Discovery.TLSBootstrapToken = opt.token
		}
		if len(opt.externalcfg.Discovery.BootstrapToken.Token) == 0 {
			opt.externalcfg.Discovery.BootstrapToken.Token = opt.token
		}
	}

	// if a file or URL from which to load cluster information was not provided, unset the Discovery.File object
	if len(opt.externalcfg.Discovery.File.KubeConfigPath) == 0 {
		opt.externalcfg.Discovery.File = nil
	}

	// if an APIServerEndpoint from which to retrieve cluster information was not provided, unset the Discovery.BootstrapToken object
	if len(args) == 0 {
		opt.externalcfg.Discovery.BootstrapToken = nil
	} else {
		if len(opt.cfgPath) == 0 && len(args) > 1 {
			klog.Warningf("[preflight] WARNING: More than one API server endpoint supplied on command line %v. Using the first one.", args)
		}
		opt.externalcfg.Discovery.BootstrapToken.APIServerEndpoint = args[0]
	}

	// if not joining a control plane, unset the ControlPlane object
	if !opt.controlPlane {
		opt.externalcfg.ControlPlane = nil
	}

	// if the admin.conf file already exists, use it for skipping the discovery process.
	// NB. this case can happen when we are joining a control-plane node only (and phases are invoked atomically)
	var adminKubeConfigPath = constants.GetAdminKubeConfigPath()
	var tlsBootstrapCfg *clientcmdapi.Config
	if _, err := os.Stat(adminKubeConfigPath); err == nil && opt.controlPlane {
		// use the admin.conf as tlsBootstrapCfg, that is the kubeconfig file used for reading the kubeadm-config during discovery
		klog.V(1).Infof("[preflight] found %s. Use it for skipping discovery", adminKubeConfigPath)
		tlsBootstrapCfg, err = clientcmd.LoadFromFile(adminKubeConfigPath)
		if err != nil {
			return nil, errors.Wrapf(err, "Error loading %s", adminKubeConfigPath)
		}
	}

	ignorePreflightErrorsSet, err := validation.ValidateIgnorePreflightErrors(opt.ignorePreflightErrors, nil)
	if err != nil {
		return nil, err
	}

	if err = validation.ValidateMixedArguments(cmd.Flags()); err != nil {
		return nil, err
	}

	// Either use the config file if specified, or convert public kubeadm API to the internal JoinConfiguration
	// and validates JoinConfiguration
	if opt.externalcfg.NodeRegistration.Name == "" {
		klog.V(1).Infoln("[preflight] found NodeName empty; using OS hostname as NodeName")
	}

	if opt.externalcfg.ControlPlane != nil && opt.externalcfg.ControlPlane.LocalAPIEndpoint.AdvertiseAddress == "" {
		klog.V(1).Infoln("[preflight] found advertiseAddress empty; using default interface's IP address as advertiseAddress")
	}

	// in case the command doesn't have flags for discovery, makes the join cfg validation pass checks on discovery
	if cmd.Flags().Lookup(options.FileDiscovery) == nil {
		if _, err := os.Stat(adminKubeConfigPath); os.IsNotExist(err) {
			return nil, errors.Errorf("File %s does not exists. Please use 'kubeadm join phase control-plane-prepare' subcommands to generate it.", adminKubeConfigPath)
		}
		klog.V(1).Infof("[preflight] found discovery flags missing for this command. using FileDiscovery: %s", adminKubeConfigPath)
		opt.externalcfg.Discovery.File = &kubeadmapi.FileDiscovery{KubeConfigPath: adminKubeConfigPath}
		opt.externalcfg.Discovery.BootstrapToken = nil //NB. this could be removed when we get better control on args (e.g. phases without discovery should have NoArgs )
	}

	cfg, err := configutil.LoadOrDefaultJoinConfiguration(opt.cfgPath, opt.externalcfg)
	if err != nil {
		return nil, err
	}

	// override node name and CRI socket from the command line opt
	if opt.externalcfg.NodeRegistration.Name != "" {
		cfg.NodeRegistration.Name = opt.externalcfg.NodeRegistration.Name
	}
	if opt.externalcfg.NodeRegistration.CRISocket != "" {
		cfg.NodeRegistration.CRISocket = opt.externalcfg.NodeRegistration.CRISocket
	}

	if cfg.ControlPlane != nil {
		if err := configutil.VerifyAPIServerBindAddress(cfg.ControlPlane.LocalAPIEndpoint.AdvertiseAddress); err != nil {
			return nil, err
		}
	}

	return &joinData{
		cfg:                   cfg,
		tlsBootstrapCfg:       tlsBootstrapCfg,
		ignorePreflightErrors: ignorePreflightErrorsSet,
		outputWriter:          out,
		certificateKey:        opt.certificateKey,
		nodeIP:                opt.nodeIP,
		highAvailabilityVIP:   opt.highAvailabilityVIP,
		keepalivedVersionTag:  opt.keepalivedVersionTag,
		hostInterface:         strings.Split(opt.hostCfg.Networks[0], "/")[0],
	}, nil
}

// GetHighAvailabilityVIP return the highAvailabilityVIP
func (j *joinData) GetHighAvailabilityVIP() string {
	return j.highAvailabilityVIP
}

// GetKeepalivedVersionTag return the keepalivedVersionTag
func (j *joinData) GetKeepalivedVersionTag() string {
	return j.keepalivedVersionTag
}

// GetHostInterface return the hostInterface
func (j *joinData) GetHostInterface() string {
	return j.hostInterface
}

// EnableHostAgent return is enable host agent
func (j *joinData) EnabledHostAgent() bool {
	return j.enableHostAgent
}

// CertificateKey returns the key used to encrypt the certs.
func (j *joinData) CertificateKey() string {
	return j.certificateKey
}

// Cfg returns the JoinConfiguration.
func (j *joinData) Cfg() *kubeadmapi.JoinConfiguration {
	return &j.cfg.JoinConfiguration
}

func (j *joinData) OnecloudJoinCfg() *apiv1.JoinConfiguration {
	return j.cfg
}

// TLSBootstrapCfg returns the cluster-info (kubeconfig).
func (j *joinData) TLSBootstrapCfg() (*clientcmdapi.Config, error) {
	if j.tlsBootstrapCfg != nil {
		return j.tlsBootstrapCfg, nil
	}
	klog.V(1).Infoln("[preflight] Discovering cluster-info")
	tlsBootstrapCfg, err := discovery.For(&j.cfg.JoinConfiguration)
	j.tlsBootstrapCfg = tlsBootstrapCfg
	return tlsBootstrapCfg, err
}

// OnecloudInitCfg returns the InitConfiguration.
func (j *joinData) OnecloudInitCfg() (*apiv1.InitConfiguration, error) {
	if j.initCfg != nil {
		return j.initCfg, nil
	}
	if _, err := j.TLSBootstrapCfg(); err != nil {
		return nil, err
	}
	klog.V(1).Infoln("[preflight] Fetching init configuration")
	initCfg, err := fetchInitConfigurationFromJoinConfiguration(j.cfg, j.tlsBootstrapCfg)
	if err != nil {
		return nil, err
	}
	initCfg.NodeRegistration.KubeletExtraArgs = customizeKubeletExtarArgs(
		j.enableHostAgent, j.glanceNode, j.baremetalNode, j.esxiNode, j.asOnecloudController, j.nodeIP)
	j.initCfg = initCfg
	return j.initCfg, nil
}

func customizeKubeletExtarArgs(
	enableHostAgent, glanceNode, baremetalNode, esxiNode, asOnecloudController bool, nodeIP string,
) map[string]string {
	if !enableHostAgent && !asOnecloudController {
		return nil
	}
	ret := make(map[string]string)
	if enableHostAgent {
		klog.V(1).Infoln("[preflight] Enable host agent")
		lableStr := fmt.Sprintf("%s=enable", operatorconstants.OnecloudEnableHostLabelKey)
		ret["node-labels"] = lableStr
	}
	if glanceNode {
		klog.V(1).Infoln("[preflight] As glance node")
		lableStr := ret["node-labels"]
		if len(lableStr) > 0 {
			lableStr += ","
		}
		lableStr += "onecloud.yunion.io/glance=enable"
		ret["node-labels"] = lableStr
	}
	if baremetalNode {
		klog.V(1).Infoln("[preflight] As baremetal node")
		lableStr := ret["node-labels"]
		if len(lableStr) > 0 {
			lableStr += ","
		}
		lableStr += "onecloud.yunion.io/baremetal=enable"
		ret["node-labels"] = lableStr
	}
	if esxiNode {
		klog.V(1).Infoln("[preflight] As esxi node")
		lableStr := ret["node-labels"]
		if len(lableStr) > 0 {
			lableStr += ","
		}
		lableStr += "onecloud.yunion.io/esxi=enable"
		ret["node-labels"] = lableStr
	}
	if asOnecloudController {
		klog.V(1).Infoln("[preflight] As onecloud controller")
		var labelStr string
		if _labelStr, ok := ret["node-labels"]; ok {
			labelStr = fmt.Sprintf("%s,%s=enable", _labelStr, operatorconstants.OnecloudControllerLabelKey)
		} else {
			labelStr = fmt.Sprintf("%s=enable", operatorconstants.OnecloudControllerLabelKey)
		}
		ret["node-labels"] = labelStr
	}

	if len(nodeIP) > 0 {
		ret["node-ip"] = nodeIP
	}

	return ret
}

// InitCfg returns the kubeadm InitConfiguration.
func (j *joinData) InitCfg() (*kubeadmapi.InitConfiguration, error) {
	initCfg, err := j.OnecloudInitCfg()
	if err != nil {
		return nil, err
	}
	return &initCfg.InitConfiguration, err
}

// ClientSet returns the ClientSet for accessing the cluster with the identity defined in admin.conf.
func (j *joinData) ClientSet() (*clientset.Clientset, error) {
	if j.clientSet != nil {
		return j.clientSet, nil
	}
	path := constants.GetAdminKubeConfigPath()
	client, err := kubeconfigutil.ClientSetFromFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "[preflight] couldn't create Kubernetes client")
	}
	j.clientSet = client
	return client, nil
}

// IgnorePreflightErrors returns the list of preflight errors to ignore.
func (j *joinData) IgnorePreflightErrors() sets.String {
	return j.ignorePreflightErrors
}

// OutputWriter returns the io.Writer used to write messages such as the "join done" message.
func (j *joinData) OutputWriter() io.Writer {
	return j.outputWriter
}

// GetNodeIP returns current node ip for join mode
func (j *joinData) GetNodeIP() string {
	return j.nodeIP
}

// fetchInitConfigurationFromJoinConfiguration retrieves the init configuration from a join configuration, performing the discovery
func fetchInitConfigurationFromJoinConfiguration(cfg *apiv1.JoinConfiguration, tlsBootstrapCfg *clientcmdapi.Config) (*apiv1.InitConfiguration, error) {
	// Retrieves the kubeadm configuration
	klog.V(1).Infoln("[preflight] Retrieving KubeConfig objects")
	initConfiguration, err := fetchInitConfiguration(tlsBootstrapCfg)
	if err != nil {
		return nil, err
	}

	// Create the final KubeConfig file with the cluster name discovered after fetching the cluster configuration
	clusterinfo := kubeconfigutil.GetClusterFromKubeConfig(tlsBootstrapCfg)
	tlsBootstrapCfg.Clusters = map[string]*clientcmdapi.Cluster{
		initConfiguration.ClusterName: clusterinfo,
	}
	tlsBootstrapCfg.Contexts[tlsBootstrapCfg.CurrentContext].Cluster = initConfiguration.ClusterName

	// injects into the kubeadm configuration the information about the joining node
	initConfiguration.NodeRegistration = cfg.NodeRegistration
	if cfg.ControlPlane != nil {
		initConfiguration.LocalAPIEndpoint = cfg.ControlPlane.LocalAPIEndpoint
	}
	initConfiguration.ComponentConfigs.Kubelet.EvictionHard = ocEvictionHard

	return initConfiguration, nil
}

// fetchInitConfiguration reads the cluster configuration from the kubeadm-admin configMap
func fetchInitConfiguration(tlsBootstrapCfg *clientcmdapi.Config) (*apiv1.InitConfiguration, error) {
	// creates a client to access the cluster using the bootstrap token identity
	tlsClient, err := kubeconfigutil.ToClientSet(tlsBootstrapCfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to access the cluster")
	}

	// Fetches the init configuration
	initConfiguration, err := configutil.FetchInitConfigurationFromCluster(tlsClient, os.Stdout, "preflight", true)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch the ocadm-config ConfigMap")
	}

	return initConfiguration, nil
}
