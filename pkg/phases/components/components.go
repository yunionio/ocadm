package components

import (
	"crypto/rsa"
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/klog"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"
	kubeadmutil "k8s.io/kubernetes/cmd/kubeadm/app/util"
	kubeconfigutil "k8s.io/kubernetes/cmd/kubeadm/app/util/kubeconfig"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/pkiutil"

	"yunion.io/x/onecloud/pkg/mcclient"

	"yunion.io/x/ocadm/pkg/apis/constants"
	apiv1 "yunion.io/x/ocadm/pkg/apis/v1"
	"yunion.io/x/ocadm/pkg/occonfig"
	"yunion.io/x/ocadm/pkg/phases/certs"
	configutil "yunion.io/x/ocadm/pkg/util/config"
	"yunion.io/x/ocadm/pkg/util/mysql"
	"yunion.io/x/ocadm/pkg/util/onecloud"
)

type GetDBInfo func(interface{}) *apiv1.DBInfo

type GetServiceAccount func(configOpt interface{}) *ServiceAccount

type SetupFunc func(s *mcclient.ClientSession, svcCfg interface{}, clusterCfg *apiv1.ClusterConfiguration, localCfg *apiv1.HostLocalInfo) error

type WaitRunningFunc func(waiter onecloud.Waiter) error

type SysInitFunc func(s *mcclient.ClientSession, clusterCfg *apiv1.ClusterConfiguration, hostLocalCfg *apiv1.HostLocalInfo) error

type PreUninstallFunc func(s *mcclient.ClientSession, clusterCfg *apiv1.ClusterConfiguration, hostLocalCfg *apiv1.HostLocalInfo) error

type ConfigurationFactory func(authConfig *occonfig.RCAdminConfig, clusterCfg *apiv1.ClusterConfiguration, localCfg *apiv1.HostLocalInfo, certDir string) (interface{}, interface{}, error)

type Component struct {
	// Name is onecloud component name
	Name string

	// ServiceName is component service name
	ServiceName string

	// ServiceType is component service type
	ServiceType string

	// CertConfig is TLS cert config
	CertConfig *CertConfig

	// ConfigDir define config file directory
	ConfigDir string

	// ConfigFileName define config file base name
	ConfigFileName string

	// ConfigurationFactory define how to create component configuration
	ConfigurationFactory ConfigurationFactory

	// GetDBInfo define service need db initialize
	GetDBInfo GetDBInfo

	// GetServiceAccount define service service account
	GetServiceAccount GetServiceAccount

	// UseSession define service use mcclient.ClientSession
	UseSession bool

	// SetupFunc invoked before create pod
	SetupFunc SetupFunc

	// WaitRunningFunc wait service pod healthy and running
	WaitRunningFunc WaitRunningFunc

	// SysInitFunc do post service system setup actions after service pod running
	SysInitFunc SysInitFunc

	// PreUninstallFunc do pre uninstall component
	PreUninstallFunc PreUninstallFunc

	alreadyInstall bool
}

func (c *Component) createTLSCerts(data ComponentData) error {
	if c.CertConfig == nil {
		return nil
	}
	cert := certs.NewOcServiceCert(constants.CACertAndKeyBaseName, c.Name, c.CertConfig.CertName)

	if _, err := pkiutil.TryLoadCertFromDisk(data.OnecloudCertificateDir(), cert.BaseName); err == nil {
		if _, err := pkiutil.TryLoadKeyFromDisk(data.OnecloudCertificateDir(), cert.BaseName); err == nil {
			fmt.Printf("[certs] Using existing %s certificate authority\n", cert.BaseName)
			return nil
		}
		fmt.Printf("[certs] Using existing %s keyless certificate authority\n", cert.BaseName)
		return nil
	}

	var caKey *rsa.PrivateKey
	ic := data.OnecloudCfg()
	caCert, err := pkiutil.TryLoadCertFromDisk(ic.OnecloudCertificatesDir, cert.CAName)
	if err != nil {
		return errors.Wrapf(err, "failed to load ca %s %s", ic.OnecloudCertificatesDir, cert.CAName)
	}
	if !caCert.IsCA {
		return errors.Errorf("certificate %q is not a CA", cert.CAName)
	}
	caKey, err = pkiutil.TryLoadKeyFromDisk(ic.OnecloudCertificatesDir, cert.CAName)
	if err != nil {
		return errors.Wrapf(err, "failed to load ca key %s %s", ic.OnecloudCertificatesDir, cert.CAName)
	}

	// if dryrunning, write certificates authority to a temporary folder (and defer restore to the path originally specified by the user)
	cfg := data.OnecloudCfg()
	cfg.CertificatesDir = data.OnecloudCertificateWriteDir()
	defer func() { cfg.CertificatesDir = data.OnecloudCertificateDir() }()

	// create the new certificate authority (or use existing)
	return cert.CreateFromCA(cfg, caCert, caKey)
}

func DeleteFile(filePath string) error {
	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return nil
}

func DeleteconfigFile(c *Component) error {
	filePath := occonfig.YAMLConfigFilePath(c.ConfigDir, c.ConfigFileName)
	return DeleteFile(filePath)
}

func GetStaticPodManifestPath(c *Component) string {
	manifestDir := constants.GetStaticPodDirectory()
	return constants.GetStaticPodFilepath(c.Name, manifestDir)
}

func IsStaticPodExists(c *Component) bool {
	path := GetStaticPodManifestPath(c)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func StopStaticPod(c *Component) error {
	return DeleteFile(GetStaticPodManifestPath(c))
}

func WriteConfigFile(c *Component, configOpt interface{}) error {
	return occonfig.WriteOnecloudConfigFile(c.ConfigDir, c.ConfigFileName, configOpt)
}

func SetupDB(dbConn *mysql.Connection, createDBInfo *apiv1.DBInfo) error {
	if err := configutil.InitDBUser(dbConn, *createDBInfo); err != nil {
		return errors.Wrap(err, "init database and user")
	}
	return nil
}

func SetupServiceAccount(s *mcclient.ClientSession, sa *ServiceAccount) error {
	if err := occonfig.InitServiceAccount(s, sa.AdminUser, sa.AdminPassword); err != nil {
		return errors.Wrap(err, "create service account")
	}
	return nil
}

func (c *Component) PreInstall(rd workflow.RunData) error {
	_, ok := rd.(ComponentData)
	if !ok {
		return errors.Errorf("%s install phase invoked with an invalid data", c.Name)
	}
	if IsStaticPodExists(c) {
		c.alreadyInstall = true
		klog.Infof("static pod %s exists, already installed.", GetStaticPodManifestPath(c))
		return nil
	}
	return nil
}

func (c *Component) RunSetup(rd workflow.RunData) error {
	if c.alreadyInstall {
		return nil
	}

	data, ok := rd.(ComponentData)
	if !ok {
		return errors.Errorf("%s install phase invoked with an invalid data", c.Name)
	}

	var session *mcclient.ClientSession
	var err error
	var adminConfig *occonfig.RCAdminConfig

	initCfg := data.OnecloudCfg()
	clusterCfg := &initCfg.ClusterConfiguration
	localCfg := &initCfg.HostLocalInfo
	certDir := data.OnecloudCertificateDir()

	if c.NeedSession() {
		adminConfig, err = occonfig.NewRCAdminConfigByFile(data.OnecloudAdminConfigPath())
		if err != nil {
			return errors.Wrap(err, "create admin config")
		}
		session, err = data.OnecloudClientSession()
		if err != nil {
			return errors.Wrap(err, "create client session")
		}
	}

	if err := c.createTLSCerts(data); err != nil {
		return errors.Wrap(err, "create TLS certs")
	}

	var (
		configObj interface{}
		configOpt interface{}
	)

	if c.ConfigurationFactory != nil {
		configObj, configOpt, err = c.ConfigurationFactory(adminConfig, clusterCfg, localCfg, certDir)
		if err != nil {
			return errors.Wrap(err, "generate config")
		}
		if err := WriteConfigFile(c, configOpt); err != nil {
			return errors.Wrap(err, "write config file")
		}
	}

	if c.GetDBInfo != nil {
		dbInfo := c.GetDBInfo(configObj)
		dbConn, err := data.RootDBConnection()
		if err != nil {
			return errors.Wrap(err, "init mysql connection")
		}
		if err := SetupDB(dbConn, dbInfo); err != nil {
			return err
		}
	}

	if c.GetServiceAccount != nil {
		sa := c.GetServiceAccount(configObj)
		if err := SetupServiceAccount(session, sa); err != nil {
			return err
		}
	}

	if c.SetupFunc == nil {
		return nil
	}
	return c.SetupFunc(session, configObj, clusterCfg, localCfg)
}

func (c *Component) RunStart(rd workflow.RunData) error {
	data, ok := rd.(ComponentData)
	if !ok {
		return errors.Errorf("invalid component data")
	}
	clusterCfg := data.OnecloudCfg().ClusterConfiguration
	if err := CreateStaticPodFiles(data.ManifestDir(), &clusterCfg, c.Name); err != nil {
		return errors.Wrapf(err, "Create %s static pod mainifest", c.Name)
	}
	timeout := data.Cfg().ClusterConfiguration.APIServer.TimeoutForControlPlane.Duration
	fmt.Printf("[wait-%s-start] Waiting for %s static pod from direcotry %q. This can take up to %v\n", c.Name, c.Name, data.ManifestDir(), timeout)
	kubeCli, err := data.Client()
	if err != nil {
		return errors.Wrap(err, "get kubernetes client")
	}
	waiter := onecloud.NewOCWaiter(
		kubeCli,
		data.OnecloudClientSession,
		timeout,
		data.OutputWriter(),
	)
	if err := waiter.WaitForServicePods(c.Name); err != nil {
		return errors.Wrapf(err, "wait service pod %s to running", c.Name)
	}
	if c.WaitRunningFunc == nil {
		return nil
	}
	return c.WaitRunningFunc(waiter)
}

func (c *Component) RunSysInit(rd workflow.RunData) error {
	if c.alreadyInstall {
		return nil
	}
	data, ok := rd.(ComponentData)
	if !ok {
		return errors.Errorf("%s component phase invoked with an invalid data", c.Name)
	}
	session, err := data.OnecloudClientSession()
	if err != nil {
		return err
	}
	clusterCfg := data.OnecloudCfg().ClusterConfiguration
	hostLocalCfg := data.OnecloudCfg().HostLocalInfo
	if c.SysInitFunc == nil {
		return nil
	}
	return c.SysInitFunc(session, &clusterCfg, &hostLocalCfg)
}

func (c *Component) RunInstall(rd workflow.RunData) error {
	_, ok := rd.(ComponentData)
	if !ok {
		return errors.Errorf("%s component phase invoked with an invalid data", c.Name)
	}
	// append cert init to DefaultCertsList

	return nil
}

type componentsOptions struct {
	cfgPath string
}

type componentsData struct {
	cfg             *apiv1.InitConfiguration
	client          clientset.Interface
	tlsBootstrapCfg *clientcmdapi.Config
	dryRun          bool
	dryRunDir       string
	outputWriter    io.Writer
}

func newComponentsOptions() *componentsOptions {
	return &componentsOptions{}
}

func newComponentsData(cmd *cobra.Command, args []string, opt *componentsOptions, out io.Writer) (*componentsData, error) {
	// if the admin.conf file already exists, use it for skipping the discovery process.
	var tlsBootstrapCfg *clientcmdapi.Config
	var err error

	for _, kubeConfigFile := range []string{
		constants.GetAdminKubeConfigPath(),
		constants.GetKubeletKubeConfigPath(),
	} {
		if _, err := os.Stat(kubeConfigFile); err == nil {
			// use the admin.conf as tlsBootstrapCfg, that is the kubeconfig file used for reading the ocadm-config during dicovery
			klog.V(1).Infof("[preflight] found %s. Use it for skipping discovery", kubeConfigFile)
			tlsBootstrapCfg, err = clientcmd.LoadFromFile(kubeConfigFile)
			if err != nil {
				return nil, errors.Wrapf(err, "Error loading %s", kubeConfigFile)
			}
		}
	}

	if tlsBootstrapCfg == nil {
		return nil, errors.New("Not found valid kubeconfig, run `ocadm join` node firstly.")
	}

	kubeCli, err := kubeconfigutil.ToClientSet(tlsBootstrapCfg)
	if err != nil {
		return nil, err
	}

	cfg, err := FetchInitConfiguration(tlsBootstrapCfg)
	if err != nil {
		return nil, err
	}

	data := &componentsData{
		cfg:          cfg,
		client:       kubeCli,
		outputWriter: out,
	}
	return data, nil
}

func (d *componentsData) Cfg() *kubeadmapi.InitConfiguration {
	return &d.cfg.InitConfiguration
}

// TLSBootstrapCfg returns the cluster-info (kubeconfig).
func (d *componentsData) TLSBootstrapCfg() (*clientcmdapi.Config, error) {
	return d.tlsBootstrapCfg, nil
}

// OutputWriter returns the io.Writer used to write output to by this command.
func (d *componentsData) OutputWriter() io.Writer {
	return d.outputWriter
}

// Client returns a Kubernetes client to be used by kubeadm.
// This function is implemented as a singleton, thus avoiding to recreate the client when it is used by different phases.
// Important. This function must be called after the admin.conf kubeconfig file is created.
func (d *componentsData) Client() (clientset.Interface, error) {
	return d.client, nil
}

func (d *componentsData) OnecloudCfg() *apiv1.InitConfiguration {
	return d.cfg
}

func (d *componentsData) OnecloudCertificateWriteDir() string {
	return d.OnecloudCfg().OnecloudCertificatesDir
}

func (d *componentsData) OnecloudCertificateDir() string {
	return d.OnecloudCfg().OnecloudCertificatesDir
}

func (d *componentsData) ManifestDir() string {
	return constants.GetStaticPodDirectory()
}

func (d *componentsData) RootDBConnection() (*mysql.Connection, error) {
	info := &d.cfg.MysqlConnection
	return mysql.NewConnection(info)
}

func (d *componentsData) LocalAddress() string {
	return d.OnecloudCfg().HostLocalInfo.ManagementNetInterface.IPAddress()
}

func (d *componentsData) OnecloudAdminConfigPath() string {
	return occonfig.AdminConfigFilePath()
}

func (d *componentsData) OnecloudClientSession() (*mcclient.ClientSession, error) {
	return occonfig.ClientSessionFromFile(d.OnecloudAdminConfigPath())
}

func runComponentFunc(runner *workflow.Runner) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		_, err := runner.InitData(args)
		kubeadmutil.CheckErr(err)

		err = runner.Run(args)
		kubeadmutil.CheckErr(err)
	}
}

func runComponentSetDataInitializer(runner *workflow.Runner, copt *componentsOptions) {
	runner.SetDataInitializer(func(cmd *cobra.Command, args []string) (workflow.RunData, error) {
		return newComponentsData(cmd, args, copt, os.Stdout)
	})
}

func (c *Component) ToInstallCmd() *cobra.Command {
	runner := workflow.NewRunner()
	cOpt := newComponentsOptions()

	cmd := &cobra.Command{
		Use:   c.Name,
		Short: fmt.Sprintf("Install the %s component to cluster", c.Name),
		Run:   runComponentFunc(runner),
		Args:  cobra.NoArgs,
	}

	runner.AppendPhase(c.ToInstallPhase())
	runComponentSetDataInitializer(runner, cOpt)

	runner.BindToCommand(cmd)

	return cmd
}

func (c *Component) ToInstallPhase() workflow.Phase {
	phase := workflow.Phase{
		Name:  c.Name,
		Short: fmt.Sprintf("Init and setup onecloud %s %s service", c.ServiceName, c.ServiceType),
		Run:   c.RunInstall,
		Phases: []workflow.Phase{
			{
				Name:  "check",
				Short: fmt.Sprintf("Pre install check"),
				Run:   c.PreInstall,
			},
			{
				Name:  "setup",
				Short: fmt.Sprintf("Create %s database and setup config", c.Name),
				Run:   c.RunSetup,
			},
			{
				Name:  "start",
				Short: fmt.Sprintf("Create %s static pod manifest and start service", c.Name),
				Run:   c.RunStart,
			},
			{
				Name:  "sysinit",
				Short: "post start and do system init",
				Run:   c.RunSysInit,
			},
		},
		//InheritFlags: c.InheritFlags,
	}
	return phase
}

func (c *Component) ToUninstallCmd() *cobra.Command {
	runner := workflow.NewRunner()
	cOpt := newComponentsOptions()

	cmd := &cobra.Command{
		Use:   c.Name,
		Short: fmt.Sprintf("Uninstall the %s component in cluster", c.Name),
		Run: func(cmd *cobra.Command, args []string) {
			_, err := runner.InitData(args)
			kubeadmutil.CheckErr(err)

			err = runner.Run(args)
			kubeadmutil.CheckErr(err)
		},
		Args: cobra.NoArgs,
	}
	runner.AppendPhase(c.ToUninstallPhase())
	runner.SetDataInitializer(func(cmd *cobra.Command, args []string) (workflow.RunData, error) {
		return newComponentsData(cmd, args, cOpt, os.Stdout)
	})

	runner.BindToCommand(cmd)

	return cmd
}

func (c *Component) ToUninstallPhase() workflow.Phase {
	phase := workflow.Phase{
		Name:  c.Name,
		Short: fmt.Sprintf("Uninstall onecloud %s component", c.Name),
		Run:   c.RunUninstall,
	}
	return phase
}

func (c *Component) NeedSession() bool {
	return c.GetServiceAccount != nil || c.UseSession
}

func (c *Component) RunUninstall(rd workflow.RunData) error {
	klog.Infof("Start uninstall %s", c.Name)
	data, ok := rd.(ComponentData)
	if !ok {
		return errors.Errorf("%s component phase invoked with an invalid data", c.Name)
	}
	cfg := data.OnecloudCfg()
	clusterCfg := &cfg.ClusterConfiguration
	localCfg := &cfg.HostLocalInfo

	var session *mcclient.ClientSession
	var err error

	if c.NeedSession() {
		session, err = data.OnecloudClientSession()
		if err != nil {
			return err
		}
	}

	if c.PreUninstallFunc != nil {
		if err := c.PreUninstallFunc(session, clusterCfg, localCfg); err != nil {
			return err
		}
	}

	if err := StopStaticPod(c); err != nil {
		return err
	}

	if err := DeleteconfigFile(c); err != nil {
		return err
	}
	return nil
}
