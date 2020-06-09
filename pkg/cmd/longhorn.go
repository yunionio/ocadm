package cmd

import (
	"io"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	kubeadmutil "k8s.io/kubernetes/cmd/kubeadm/app/util"
	kubeconfigutil "k8s.io/kubernetes/cmd/kubeadm/app/util/kubeconfig"

	"yunion.io/x/ocadm/pkg/apis/constants"
	v1 "yunion.io/x/ocadm/pkg/apis/v1"
	"yunion.io/x/ocadm/pkg/options"
	"yunion.io/x/ocadm/pkg/phases/longhorn"
)

func NewCmdLonghorn(out io.Writer) *cobra.Command {
	cmds := &cobra.Command{
		Use:   "longhorn",
		Short: "Deployment longhorn management",
	}

	cmds.AddCommand(cmdLonghornEnable())
	cmds.AddCommand(cmdMigratePvToLonghorn())
	return cmds
}

func cmdLonghornEnable() *cobra.Command {
	opt := &longHornConfig{}
	runner := workflow.NewRunner()
	cmd := &cobra.Command{
		Use:   "enable",
		Short: "Run this command to enable longhorn",
		PreRun: func(cmd *cobra.Command, args []string) {
		},
		Run: func(cmd *cobra.Command, args []string) {
			_, err := runner.InitData(args)
			kubeadmutil.CheckErr(err)

			err = runner.Run(args)
			kubeadmutil.CheckErr(err)
		},
		Args: cobra.NoArgs,
	}
	AddLonghornFlags(cmd.Flags(), opt)
	runner.AppendPhase(longhorn.InstallLonghornPhase())

	runner.SetDataInitializer(func(cmd *cobra.Command, args []string) (workflow.RunData, error) {
		_, err := opt.VersionedClient()
		if err != nil {
			return nil, err
		}
		err = opt.VerifyNode()
		if err != nil {
			return nil, err
		}
		opt.kubeconfigPath = kubeadmconstants.GetAdminKubeConfigPath()
		_, err = opt.KubectlClient()
		if err != nil {
			return nil, err
		}
		return opt, nil
	})
	runner.BindToCommand(cmd)
	return cmd
}

type longHornConfig struct {
	nodesBaseData
	kubectlCli
	DataPath                   string
	OverProvisioningPercentage int
	ReplicaCount               int
	ImageRepository            string
}

func AddLonghornFlags(flagSet *flag.FlagSet, o *longHornConfig) {
	// longhorn configuration
	flagSet.StringVar(
		&o.DataPath, options.LonghornDataPath,
		constants.LonghornDefaultDataPath, "Longhorn data path, default /var/lib/longhorn",
	)
	flagSet.IntVar(
		&o.OverProvisioningPercentage,
		options.LonghornOverProvisioningPercentage,
		constants.LonghornDefaultOverProvisioningPercentage,
		"Longhorn disk over provisioning percentage, default 100",
	)
	//flagSet.IntVar(
	//	&o.ReplicaCount, options.LonghornReplicaCount,
	//	constants.LonghornDefaultReplicaCount, "Longhorn replica count, default 3",
	//)
	flagSet.StringVar(
		&o.ImageRepository, options.ImageRepository,
		v1.DefaultImageRepository, "Image repository",
	)
	o.AddNodesFlags(flagSet)
}

func (d *longHornConfig) GetImageRepository() string {
	return d.ImageRepository
}

func (d *longHornConfig) LonghornConfig() *longhorn.LonghornConfig {
	return &longhorn.LonghornConfig{
		DataPath:                    d.DataPath,
		OverProviosioningPercentage: d.OverProvisioningPercentage,
		ReplicaCount:                d.ReplicaCount,
	}
}

func (d *longHornConfig) VerifyNode() error {
	cli, err := d.ClientSet()
	if err != nil {
		return err
	}
	if len(d.nodes) == 0 {
		return nil
	} else {
		for i := 0; i < len(d.nodes); i++ {
			_, err = cli.CoreV1().Nodes().Get(d.nodes[i], metav1.GetOptions{})
			if err != nil {
				return errors.Wrapf(err, "get node %s", d.nodes[i])
			}
		}
		return nil
	}
}

type migratePvConfig struct {
	kubectlCli
	sourcePVC string
	component string

	ImageRepository          string
	clientSet                *clientset.Clientset
	deleteMigartePodInTheEnd bool
}

func (d *migratePvConfig) SourcePVC() string {
	return d.sourcePVC
}

func (d *migratePvConfig) GetImageRepository() string {
	return d.ImageRepository
}

func (d *migratePvConfig) DeleteMigartePodInTheEnd() bool {
	return d.deleteMigartePodInTheEnd
}

// ClientSet returns the ClientSet for accessing the cluster with the identity defined in admin.conf.
func (d *migratePvConfig) ClientSet() (*clientset.Clientset, error) {
	if d.clientSet != nil {
		return d.clientSet, nil
	}
	path := constants.GetAdminKubeConfigPath()
	client, err := kubeconfigutil.ClientSetFromFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "[preflight] couldn't create Kubernetes client")
	}
	d.clientSet = client
	return client, nil
}

// migrate data from a pv to new pv which created by longhorn
func cmdMigratePvToLonghorn() *cobra.Command {
	opt := &migratePvConfig{}
	runner := workflow.NewRunner()
	cmd := &cobra.Command{
		Use:   "migrate-from",
		Short: "Run this command to migrate pv to longhorn",
		PreRun: func(cmd *cobra.Command, args []string) {
		},
		Run: func(cmd *cobra.Command, args []string) {
			_, err := runner.InitData(args)
			kubeadmutil.CheckErr(err)

			err = runner.Run(args)
			kubeadmutil.CheckErr(err)
		},
		Args: cobra.NoArgs,
	}
	AddMigrateToLonghornFlags(cmd.Flags(), opt)
	runner.AppendPhase(longhorn.MigrateToLonghornPhase())
	runner.SetDataInitializer(func(cmd *cobra.Command, args []string) (workflow.RunData, error) {
		if len(opt.sourcePVC) == 0 {
			return nil, errors.New("missing source pvc")
		}
		return opt, nil
	})
	runner.BindToCommand(cmd)
	return cmd
}

func AddMigrateToLonghornFlags(flagSet *flag.FlagSet, o *migratePvConfig) {
	flagSet.StringVar(
		&o.sourcePVC, options.PVCMigrateToLonghorn,
		o.sourcePVC, "PVC migrate to longhorn",
	)
	flagSet.StringVar(
		&o.ImageRepository, options.ImageRepository,
		v1.DefaultImageRepository, "Image repository",
	)
	flagSet.BoolVar(&o.deleteMigartePodInTheEnd, "delete-migrate-pod",
		true, "Delete migrate pod in the end")
}
