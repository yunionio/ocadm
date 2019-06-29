package init

import (
	"fmt"
	"io"
	"time"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/apiclient"

	"yunion.io/x/ocadm/pkg/apis/constants"
	"yunion.io/x/ocadm/pkg/options"
	"yunion.io/x/ocadm/pkg/phases/keystone"
	"yunion.io/x/onecloud/pkg/mcclient"
	"yunion.io/x/onecloud/pkg/mcclient/modules"
)

var (
	controlPlanePhaseProperties = map[string]struct {
		name  string
		short string
	}{
		constants.OnecloudKeystone: {
			name:  "keystone",
			short: getPhaseDescription(constants.OnecloudKeystone),
		},
	}
)

func getPhaseDescription(component string) string {
	return fmt.Sprintf("Generates the %s static Pod manifest", component)
}

// NewKeystonePhase creates a ocadm workflow phase that implements handing of etcd
func NewKeystonePhase() workflow.Phase {
	phase := workflow.Phase{
		Name:  "keystone",
		Short: "Init and setup onecloud keystone identity service",
		Run:   runKeystoneInit,
		Phases: []workflow.Phase{
			{
				Name:         "setup",
				Short:        "Create keystone database and setup config",
				InheritFlags: []string{},
				Example:      "",
				Run:          runKeystoneSetup,
			},
			{
				Name:  "start",
				Short: "Create keystone static pod manifest and start service",
				Run:   runKeystoneStart,
			},
			{
				Name:  "sysinit",
				Short: "Inject init data like: policy, public endpoint",
				Run:   runKeystoneSysInit,
			},
		},
		InheritFlags: []string{
			options.CfgPath,
			options.MysqlAddress,
			options.MysqlPort,
			options.MysqlUser,
			options.MysqlPassword,
			options.Region,
		},
	}
	return phase
}

func runKeystoneInit(c workflow.RunData) error {
	data, ok := c.(InitData)
	if !ok {
		return errors.New("keystone init phase invoked with an invalid data")
	}
	cfg := data.OnecloudCfg()
	region := cfg.ClusterConfiguration.Region
	mysqlInfo := cfg.MysqlConnection
	fmt.Printf("[keystone] Using mysql connection %#v, region: %s", mysqlInfo, region)
	return nil
}

func runKeystoneSetup(c workflow.RunData) error {
	data, ok := c.(InitData)
	if !ok {
		return errors.New("keystone init phase invoked with an invalid data")
	}
	dbConn, err := data.RootDBConnection()
	if err != nil {
		return errors.Wrap(err, "init mysql connection")
	}
	cfg := data.OnecloudCfg()
	localAddress := data.LocalAddress()
	certDir := cfg.OnecloudCertificatesDir
	return keystone.SetupKeystone(dbConn, cfg.Keystone, cfg.Region, localAddress, certDir)
}

func runKeystoneStart(c workflow.RunData) error {
	if err := runControlPlaneSubphase(constants.OnecloudKeystone)(c); err != nil {
		return errors.Wrap(err, "Create keystone static pod mainifest")
	}
	return waitKeystoneRunning(c.(InitData))
}

func runKeystoneSysInit(c workflow.RunData) error {
	data, ok := c.(InitData)
	if !ok {
		return errors.New("keystone init phase invoked with an invalid data")
	}
	session, err := data.OnecloudClientSession()
	if err != nil {
		return err
	}
	clusterCfg := data.OnecloudCfg().ClusterConfiguration
	address := data.LocalAddress()
	return keystone.DoSysInit(session, &clusterCfg, address)
}

type Waiter interface {
	apiclient.Waiter

	WaitForKeystone() error
	WaitForRegion() error
	WaitForGlance() error
}

type OCWaiter struct {
	apiclient.Waiter

	sessionFactory func() (*mcclient.ClientSession, error)
	timeout        time.Duration
	writer         io.Writer
}

// NewOCWaiter returns a new Onecloud waiter object that check service healthy
func NewOCWaiter(
	kubeClient clientset.Interface,
	sessionFactory func() (*mcclient.ClientSession, error),
	timeout time.Duration,
	writer io.Writer,
) Waiter {
	return &OCWaiter{
		Waiter:         apiclient.NewKubeWaiter(kubeClient, timeout, writer),
		sessionFactory: sessionFactory,
		timeout:        timeout,
		writer:         writer,
	}
}

func (w *OCWaiter) WaitForKeystone() error {
	start := time.Now()
	return wait.PollImmediate(constants.APICallRetryInterval, w.timeout, func() (bool, error) {
		session, err := w.sessionFactory()
		w.timeout.Seconds()
		if err != nil {
			duration := time.Since(start).Seconds()
			if (duration + float64(10*time.Second)) > w.timeout.Seconds() {
				fmt.Fprintf(w.writer, "[keystone] Error get auth session: %v", err)
			}
			return false, nil
		}
		if _, err := modules.Policies.List(session, nil); err != nil {
			return false, errors.Wrap(err, "Failed to get policy")
		}
		fmt.Printf("[keystone] healthy after %f seconds\n", time.Since(start).Seconds())
		return true, nil
	})
}

func (w *OCWaiter) WaitForRegion() error {
	return fmt.Errorf("not impl")
}

func (w *OCWaiter) WaitForGlance() error {
	return fmt.Errorf("not impl")
}

func waitKeystoneRunning(data InitData) error {
	timeout := data.Cfg().ClusterConfiguration.APIServer.TimeoutForControlPlane.Duration
	fmt.Printf("[wait-keystone-start] Waiting for keystone static pod from direcotry %q. This can take up to %v\n", data.ManifestDir(), timeout)
	kubeCli, err := data.Client()
	if err != nil {
		return err
	}
	waiter := NewOCWaiter(
		kubeCli,
		data.OnecloudClientSession,
		timeout,
		data.OutputWriter(),
	)
	if err := waiter.WaitForPodsWithLabel("component=" + constants.OnecloudKeystone); err != nil {
		return errors.Wrap(err, "wait keystone pod running")
	}
	return waiter.WaitForKeystone()
}
