package init

import (
	"fmt"

	"github.com/pkg/errors"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"

	"yunion.io/x/onecloud/pkg/mcclient"

	apis "yunion.io/x/ocadm/pkg/apis/v1"
	"yunion.io/x/ocadm/pkg/util/mysql"
	"yunion.io/x/ocadm/pkg/util/onecloud"
)

type SetupFunc func(sqlConn *mysql.Connection, s *mcclient.ClientSession, clusterCfg *apis.ClusterConfiguration, hostLocalCfg *apis.HostLocalInfo, certDir string) error

type WaitRunningFunc func(waiter onecloud.Waiter) error

type SysInitFunc func(s *mcclient.ClientSession, clusterCfg *apis.ClusterConfiguration, hostLocalCfg *apis.HostLocalInfo) error

type ServiceBasePhase struct {
	// Name is service name
	Name string
	// Type is service type
	Type string

	// InheritFlags is workflow.InheritFlags
	InheritFlags []string

	// SetupFunc do service setup process
	SetupFunc SetupFunc

	SetupUseSession bool

	// WaitRunningFunc wait service pod healthy and running
	WaitRunningFunc WaitRunningFunc

	// SysInitFunc do post service system setup actions
	SysInitFunc SysInitFunc
}

func (p *ServiceBasePhase) ToPhase() workflow.Phase {
	phase := workflow.Phase{
		Name:  p.Name,
		Short: fmt.Sprintf("Init and setup onecloud %s %s service", p.Name, p.Type),
		Run:   p.runInit,
		Phases: []workflow.Phase{
			{
				Name:  "setup",
				Short: fmt.Sprintf("Create %s database and setup config", p.Name),
				Run:   p.runSetup,
			},
			{
				Name:  "start",
				Short: fmt.Sprintf("Create %s static pod manifest and start service", p.Name),
				Run:   p.runStart,
			},
			{
				Name:  "sysinit",
				Short: "post start and do system init",
				Run:   p.runSysInit,
			},
		},
		InheritFlags: p.InheritFlags,
	}
	return phase
}

func (p *ServiceBasePhase) runInit(c workflow.RunData) error {
	data, ok := c.(InitData)
	if !ok {
		return errors.Errorf("%s init phase invoked with an invalid data", p.Name)
	}
	cfg := data.OnecloudCfg()
	region := cfg.ClusterConfiguration.Region
	mysqlInfo := cfg.MysqlConnection

	fmt.Printf("[keystone] Using mysql connection %#v, region: %s", mysqlInfo, region)

	// append cert init to DefaultCertsList

	return nil
}

func (p *ServiceBasePhase) runSetup(c workflow.RunData) error {
	data, ok := c.(InitData)
	if !ok {
		return errors.Errorf("%s init phase invoked with an invalid data", p.Name)
	}
	dbConn, err := data.RootDBConnection()
	if err != nil {
		return errors.Wrap(err, "init mysql connection")
	}
	cfg := data.OnecloudCfg()
	certDir := cfg.OnecloudCertificatesDir
	var s *mcclient.ClientSession
	if p.SetupUseSession {
		s, err = data.OnecloudClientSession()
		if err != nil {
			return errors.Wrapf(err, "%s get client session", p.Name)
		}
	}
	return p.SetupFunc(dbConn, s, &cfg.ClusterConfiguration, &cfg.HostLocalInfo, certDir)
}

func (p *ServiceBasePhase) runStart(c workflow.RunData) error {
	if err := runControlPlaneSubphase(p.Name)(c); err != nil {
		return errors.Wrapf(err, "Create %s static pod mainifest", p.Name)
	}
	data := c.(InitData)
	timeout := data.Cfg().ClusterConfiguration.APIServer.TimeoutForControlPlane.Duration
	fmt.Printf("[wait-%s-start] Waiting for %s static pod from direcotry %q. This can take up to %v\n", p.Name, p.Name, data.ManifestDir(), timeout)
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
	if err := waiter.WaitForServicePods(p.Name); err != nil {
		return errors.Wrapf(err, "wait service pod %s to running", p.Name)
	}
	return p.WaitRunningFunc(waiter)
}

func (p *ServiceBasePhase) runSysInit(c workflow.RunData) error {
	data, ok := c.(InitData)
	if !ok {
		return errors.Errorf("%s init phase invoked with an invalid data", p.Name)
	}
	session, err := data.OnecloudClientSession()
	if err != nil {
		return err
	}
	clusterCfg := data.OnecloudCfg().ClusterConfiguration
	hostLocalCfg := data.OnecloudCfg().HostLocalInfo
	return p.SysInitFunc(session, &clusterCfg, &hostLocalCfg)
}
