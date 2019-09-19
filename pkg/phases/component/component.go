package component

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"
	kubeadmutil "k8s.io/kubernetes/cmd/kubeadm/app/util"
	kubeconfigutil "k8s.io/kubernetes/cmd/kubeadm/app/util/kubeconfig"

	onecloud "yunion.io/x/onecloud-operator/pkg/apis/onecloud/v1alpha1"
	"yunion.io/x/onecloud-operator/pkg/client/clientset/versioned"
	"yunion.io/x/onecloud/pkg/mcclient"

	"yunion.io/x/ocadm/pkg/apis/constants"
	"yunion.io/x/ocadm/pkg/occonfig"
	"yunion.io/x/ocadm/pkg/phases/cluster"
)

type IComponent interface {
	GetName() string
	GetComponentType() onecloud.ComponentType

	NewDeployment(*onecloud.OnecloudCluster) (*apps.Deployment, error)
	NewConfigMap(*onecloud.OnecloudCluster, *OnecloudComponentsConfig) (*corev1.ConfigMap, error)
	NewService(*onecloud.OnecloudCluster) *corev1.Service
	NewCloudUser(*OnecloudComponentsConfig) *onecloud.CloudUser
	NewDBConfig(*OnecloudComponentsConfig) *onecloud.DBConfig
	NewDBConfig2(*OnecloudComponentsConfig) *onecloud.DBConfig
	NewCloudEndpoint() *CloudEndpoint

	ToEnableCmd() *cobra.Command
	ToEnablePhase() workflow.Phase
	ToDisableCmd() *cobra.Command
	ToDisablePhase() workflow.Phase
}

type BaseComponent struct {
	componentType onecloud.ComponentType
	componentObj  IComponent
}

func NewBaseComponent(cType onecloud.ComponentType, comp IComponent) *BaseComponent {
	return &BaseComponent{
		componentType: cType,
		componentObj:  comp,
	}
}

func (c BaseComponent) GetName() string {
	return c.componentType.String()
}

func (c BaseComponent) GetComponentType() onecloud.ComponentType {
	return c.componentType
}

func (c BaseComponent) NewDeployment(_ *onecloud.OnecloudCluster) (*apps.Deployment, error) {
	return nil, errors.Errorf("component %s not NewDeployment implemented", c.GetName())
}

func (c BaseComponent) NewConfigMap(_ *onecloud.OnecloudCluster, _ *OnecloudComponentsConfig) *corev1.ConfigMap {
	return nil
}

func (c BaseComponent) NewService(_ *onecloud.OnecloudCluster) *corev1.Service {
	return nil
}

func (c BaseComponent) NewCloudEndpoint() *CloudEndpoint {
	return nil
}

func (c BaseComponent) NewCloudUser(_ *OnecloudComponentsConfig) *onecloud.CloudUser {
	return nil
}

func (c BaseComponent) NewDBConfig(_ *OnecloudComponentsConfig) *onecloud.DBConfig {
	return nil
}

func (c BaseComponent) NewDBConfig2(_ *OnecloudComponentsConfig) *onecloud.DBConfig {
	return nil
}

type componentsOptions struct{}

func newComponentsOptions() *componentsOptions {
	return &componentsOptions{}
}

type componentsData struct {
	client        clientset.Interface
	clusterClient versioned.Interface
	oc            *onecloud.OnecloudCluster
	cfg           *OnecloudComponentsConfig
}

func newComponentsData(cmd *cobra.Command, args []string, opt *componentsOptions, out io.Writer) (*componentsData, error) {
	kubeConfigFile := constants.GetAdminKubeConfigPath()
	if _, err := os.Stat(kubeConfigFile); err != nil {
		return nil, err
	}
	tlsBootstrapCfg, err := clientcmd.LoadFromFile(kubeConfigFile)
	if err != nil {
		return nil, errors.Wrapf(err, "Error loading %s", kubeConfigFile)
	}

	kubeCli, err := kubeconfigutil.ToClientSet(tlsBootstrapCfg)
	if err != nil {
		return nil, errors.Wrap(err, "New kubernetes client")
	}
	clusterCli, err := cluster.NewClusterClient(tlsBootstrapCfg)
	if err != nil {
		return nil, errors.Wrap(err, "New onecloud cluster client")
	}
	oc, err := clusterCli.OnecloudV1alpha1().OnecloudClusters(constants.OnecloudNamespace).Get(cluster.DefaultClusterName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "Get %s onecloud cluster", cluster.DefaultClusterName)
	}
	data := &componentsData{
		clusterClient: clusterCli,
		client:        kubeCli,
		oc:            oc,
	}
	cfg, err := data.NewOnecloudComponentsConfig()
	if err != nil {
		return nil, errors.Wrap(err, "new components data")
	}
	data.cfg = cfg
	return data, nil
}

func (d *componentsData) NewOnecloudComponentsConfig() (*OnecloudComponentsConfig, error) {
	var (
		cfg *OnecloudComponentsConfig
		err error
	)
	oc := d.OnecloudCluster()
	ns := oc.GetNamespace()
	cfgMapName := ComponentsConfigMapName(oc)
	cli := d.KubernetesClient()
	cfgObj, err := cli.CoreV1().ConfigMaps(ns).Get(cfgMapName, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, errors.Wrapf(err, "get %s configmap from namespace %s", cfgMapName, ns)
		}
		// configmap not exists, new components config directly
		cfg, err = NewOnecloudComponentsConfig(nil)
	} else {
		// load config object from exists configmap
		cfg, err = NewOnecloudComponentsConfigFromConfigMap(cfgObj)
		if err != nil {
			return nil, errors.Wrapf(err, "new components config from configmap %s", cfgObj.GetName())
		}
	}
	obj, err := cfg.ToConfigMap(d.oc)
	if err != nil {
		return nil, errors.Wrap(err, "convert to configmap")
	}
	if err := SyncConfigMap(d.client, d.oc, obj); err != nil {
		return nil, errors.Wrapf(err, "sync config map %s", obj.GetName())
	}
	return cfg, err
}

func (d *componentsData) ComponentsConfig() *OnecloudComponentsConfig {
	return d.cfg
}

func (d *componentsData) KubernetesClient() kubernetes.Interface {
	return d.client
}

func (d *componentsData) ClusterClient() versioned.Interface {
	return d.clusterClient
}

func (d *componentsData) OnecloudCluster() *onecloud.OnecloudCluster {
	return d.oc
}

func (d *componentsData) OnecloudRCAdminConfig() *occonfig.RCAdminConfig {
	return &occonfig.RCAdminConfig{
		AuthUrl:     fmt.Sprintf("https://%s:%d/v3", d.oc.Spec.LoadBalancerEndpoint, constants.KeystoneAdminPort),
		Region:      d.oc.Spec.Region,
		Username:    "sysadmin",
		Password:    d.oc.Spec.Keystone.BootstrapPassword,
		ProjectName: "system",
		Insecure:    true,
		Timeout:     600,
		// EndpointType: "publicURL",
	}
}

func (d *componentsData) OnecloudClientSession() (*mcclient.ClientSession, error) {
	cfg := d.OnecloudRCAdminConfig()
	cli := mcclient.NewClient(cfg.AuthUrl, cfg.Timeout, cfg.Debug, cfg.Insecure, cfg.CertFile, cfg.KeyFile)
	token, err := cli.AuthenticateWithSource(cfg.Username, cfg.Password, cfg.DomainName, cfg.ProjectName, cfg.ProjectDomain, mcclient.AuthSourceCli)
	if err != nil {
		return nil, err
	}
	session := cli.NewSession(context.Background(), cfg.Region, "", constants.EndpointTypePublic, token, "")
	return session, nil
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

func (c BaseComponent) ToEnableCmd() *cobra.Command {
	runner := workflow.NewRunner()
	cOpt := newComponentsOptions()

	cmd := &cobra.Command{
		Use:   c.GetName(),
		Short: fmt.Sprintf("Install and enable the %s component to cluster", c.GetName()),
		Run:   runComponentFunc(runner),
		Args:  cobra.NoArgs,
	}

	runner.AppendPhase(c.ToEnablePhase())
	runComponentSetDataInitializer(runner, cOpt)

	runner.BindToCommand(cmd)

	return cmd
}

func (c BaseComponent) ToEnablePhase() workflow.Phase {
	phase := workflow.Phase{
		Name:  c.GetName(),
		Short: fmt.Sprintf("Enable and setup onecloud %s component", c.GetName()),
		Run:   c.RunEnable,
	}
	return phase
}

func (c BaseComponent) getOperator(rd workflow.RunData) (*Operator, error) {
	data, ok := rd.(*componentsData)
	if !ok {
		return nil, errors.Errorf("%s component phase invoked with an invalid data", c.GetName())
	}
	s, err := data.OnecloudClientSession()
	if err != nil {
		return nil, errors.Wrap(err, "get onecloud session")
	}
	componentsCfg := data.ComponentsConfig()
	return NewOperator(c.componentObj, data.KubernetesClient(), s, data.OnecloudCluster(), componentsCfg), nil
}

func (c BaseComponent) runAction(rd workflow.RunData, action string, rf func(*Operator) error) error {
	klog.Infof("Start %s %s", action, c.GetName())
	operator, err := c.getOperator(rd)
	if err != nil {
		return err
	}
	if err := rf(operator); err != nil {
		return err
	}
	klog.Infof("End %s %s", action, c.GetName())
	return nil
}

func (c BaseComponent) RunEnable(rd workflow.RunData) error {
	return c.runAction(rd, "enable", func(opt *Operator) error { return opt.Enable() })
}

func (c BaseComponent) ToDisableCmd() *cobra.Command {
	runner := workflow.NewRunner()
	cOpt := newComponentsOptions()

	cmd := &cobra.Command{
		Use:   c.GetName(),
		Short: fmt.Sprintf("Disable the %s component to cluster", c.GetName()),
		Run:   runComponentFunc(runner),
		Args:  cobra.NoArgs,
	}

	runner.AppendPhase(c.ToDisablePhase())
	runComponentSetDataInitializer(runner, cOpt)

	runner.BindToCommand(cmd)

	return cmd
}

func (c BaseComponent) ToDisablePhase() workflow.Phase {
	phase := workflow.Phase{
		Name:  c.GetName(),
		Short: fmt.Sprintf("Disable onecloud %s component", c.GetName()),
		Run:   c.RunDisable,
	}
	return phase
}

func (c BaseComponent) RunDisable(rd workflow.RunData) error {
	return c.runAction(rd, "disable", func(opt *Operator) error { return opt.Disable() })
}

type Operator struct {
	component IComponent
	manager   *ComponentManager
	oc        *onecloud.OnecloudCluster
}

func NewOperator(
	comp IComponent,
	kubeCli kubernetes.Interface,
	s *mcclient.ClientSession,
	oc *onecloud.OnecloudCluster,
	cfg *OnecloudComponentsConfig) *Operator {
	manager := NewComponentManager(kubeCli, s, cfg)
	return &Operator{
		component: comp,
		manager:   manager,
		oc:        oc,
	}
}

func (o *Operator) Enable() error {
	return o.manager.SyncComponent(o.oc, o.component)
}

func (o *Operator) Disable() error {
	return o.manager.DisableComponent(o.oc, o.component)
}
