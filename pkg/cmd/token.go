package cmd

import (
	"io"
	"os"

	"github.com/lithammer/dedent"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	clientset "k8s.io/client-go/kubernetes"
	bootstrapapi "k8s.io/cluster-bootstrap/token/api"
	"k8s.io/klog"
	kubeadmscheme "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/scheme"
	kubeadmapiv1beta1 "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta1"
	"k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/validation"
	kubeadmcmd "k8s.io/kubernetes/cmd/kubeadm/app/cmd"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/options"
	cmdutil "k8s.io/kubernetes/cmd/kubeadm/app/cmd/util"
	kubeadmutil "k8s.io/kubernetes/cmd/kubeadm/app/util"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/apiclient"
	kubeconfigutil "k8s.io/kubernetes/cmd/kubeadm/app/util/kubeconfig"

	configutil "yunion.io/x/ocadm/pkg/util/config"
)

// NewCmdToken returns cobra.Command for token management
func NewCmdToken(out io.Writer, errW io.Writer) *cobra.Command {
	var kubeConfigFile string
	var dryRun bool
	tokenCmd := &cobra.Command{
		Use:   "token",
		Short: "Manage bootstrap tokens.",
		Long: dedent.Dedent(`
			This command manages bootstrap tokens. It is optional and needed only for advanced use cases.

			In short, bootstrap tokens are used for establishing bidirectional trust between a client and a server.
			A bootstrap token can be used when a client (for example a node that is about to join the cluster) needs
			to trust the server it is talking to. Then a bootstrap token with the "signing" usage can be used.
			bootstrap tokens can also function as a way to allow short-lived authentication to the API Server
			(the token serves as a way for the API Server to trust the client), for example for doing the TLS Bootstrap.

			What is a bootstrap token more exactly?
			 - It is a Secret in the kube-system namespace of type "bootstrap.kubernetes.io/token".
			 - A bootstrap token must be of the form "[a-z0-9]{6}.[a-z0-9]{16}". The former part is the public token ID,
			   while the latter is the Token Secret and it must be kept private at all circumstances!
			 - The name of the Secret must be named "bootstrap-token-(token-id)".

			You can read more about bootstrap tokens here:
			  https://kubernetes.io/docs/admin/bootstrap-tokens/
		`),

		// Without this callback, if a user runs just the "token"
		// command without a subcommand, or with an invalid subcommand,
		// cobra will print usage information, but still exit cleanly.
		// We want to return an error code in these cases so that the
		// user knows that their command was invalid.
		RunE: cmdutil.SubCmdRunE("token"),
	}

	options.AddKubeConfigFlag(tokenCmd.PersistentFlags(), &kubeConfigFile)
	tokenCmd.PersistentFlags().BoolVar(&dryRun,
		"dry-run", dryRun, "Whether to enable dry-run mode or not")

	var cfgPath string
	var printJoinCommand bool
	bto := options.NewBootstrapTokenOptions()

	createCmd := &cobra.Command{
		Use:                   "create [token]",
		DisableFlagsInUseLine: true,
		Short:                 "Create bootstrap tokens on the server.",
		Long: dedent.Dedent(`
			This command will create a bootstrap token for you.
			You can specify the usages for this token, the "time to live" and an optional human friendly description.

			The [token] is the actual token to write.
			This should be a securely generated random token of the form "[a-z0-9]{6}.[a-z0-9]{16}".
			If no [token] is given, kubeadm will generate a random token instead.
		`),
		Run: func(tokenCmd *cobra.Command, args []string) {
			if len(args) > 0 {
				bto.TokenStr = args[0]
			}
			klog.V(1).Infoln("[token] validating mixed arguments")
			err := validation.ValidateMixedArguments(tokenCmd.Flags())
			kubeadmutil.CheckErr(err)

			klog.V(1).Infoln("[token] getting Clientsets from kubeconfig file")
			kubeConfigFile = cmdutil.GetKubeConfigPath(kubeConfigFile)
			client, err := getClientset(kubeConfigFile, dryRun)
			kubeadmutil.CheckErr(err)

			cfg, err := configutil.FetchInitConfigurationFromCluster(client, out, "token", true)
			kubeadmutil.CheckErr(err)

			kubeadmcfg := &kubeadmapiv1beta1.InitConfiguration{}
			err = kubeadmscheme.Scheme.Convert(&cfg.InitConfiguration, kubeadmcfg, nil)
			kubeadmutil.CheckErr(err)

			err = bto.ApplyTo(kubeadmcfg)
			kubeadmutil.CheckErr(err)

			err = kubeadmcmd.RunCreateToken(out, client, cfgPath, kubeadmcfg, printJoinCommand, kubeConfigFile)
			kubeadmutil.CheckErr(err)
		},
	}

	options.AddConfigFlag(createCmd.Flags(), &cfgPath)
	createCmd.Flags().BoolVar(&printJoinCommand,
		"print-join-command", false, "Instead of printing only the token, print the full 'kubeadm join' flag needed to join the cluster using the token.")
	bto.AddTTLFlagWithName(createCmd.Flags(), "ttl")
	bto.AddUsagesFlag(createCmd.Flags())
	bto.AddGroupsFlag(createCmd.Flags())
	bto.AddDescriptionFlag(createCmd.Flags())

	tokenCmd.AddCommand(createCmd)
	tokenCmd.AddCommand(NewCmdTokenGenerate(out))

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List bootstrap tokens on the server.",
		Long: dedent.Dedent(`
			This command will list all bootstrap tokens for you.
		`),
		Run: func(tokenCmd *cobra.Command, args []string) {
			kubeConfigFile = cmdutil.GetKubeConfigPath(kubeConfigFile)
			client, err := getClientset(kubeConfigFile, dryRun)
			kubeadmutil.CheckErr(err)

			err = kubeadmcmd.RunListTokens(out, errW, client)
			kubeadmutil.CheckErr(err)
		},
	}
	tokenCmd.AddCommand(listCmd)

	deleteCmd := &cobra.Command{
		Use:                   "delete [token-value]",
		DisableFlagsInUseLine: true,
		Short:                 "Delete bootstrap tokens on the server.",
		Long: dedent.Dedent(`
			This command will delete a given bootstrap token for you.

			The [token-value] is the full Token of the form "[a-z0-9]{6}.[a-z0-9]{16}" or the
			Token ID of the form "[a-z0-9]{6}" to delete.
		`),
		Run: func(tokenCmd *cobra.Command, args []string) {
			if len(args) < 1 {
				kubeadmutil.CheckErr(errors.Errorf("missing subcommand; 'token delete' is missing token of form %q", bootstrapapi.BootstrapTokenIDPattern))
			}
			kubeConfigFile = cmdutil.GetKubeConfigPath(kubeConfigFile)
			client, err := getClientset(kubeConfigFile, dryRun)
			kubeadmutil.CheckErr(err)

			err = kubeadmcmd.RunDeleteToken(out, client, args[0])
			kubeadmutil.CheckErr(err)
		},
	}
	tokenCmd.AddCommand(deleteCmd)

	return tokenCmd
}

// NewCmdTokenGenerate returns cobra.Command to generate new token
func NewCmdTokenGenerate(out io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "generate",
		Short: "Generate and print a bootstrap token, but do not create it on the server.",
		Long: dedent.Dedent(`
			This command will print out a randomly-generated bootstrap token that can be used with
			the "init" and "join" commands.

			You don't have to use this command in order to generate a token. You can do so
			yourself as long as it is in the format "[a-z0-9]{6}.[a-z0-9]{16}". This
			command is provided for convenience to generate tokens in the given format.

			You can also use "kubeadm init" without specifying a token and it will
			generate and print one for you.
		`),
		Run: func(cmd *cobra.Command, args []string) {
			err := kubeadmcmd.RunGenerateToken(out)
			kubeadmutil.CheckErr(err)
		},
	}
}

func getClientset(file string, dryRun bool) (clientset.Interface, error) {
	if dryRun {
		dryRunGetter, err := apiclient.NewClientBackedDryRunGetterFromKubeconfig(file)
		if err != nil {
			return nil, err
		}
		return apiclient.NewDryRunClient(dryRunGetter, os.Stdout), nil
	}
	return kubeconfigutil.ClientSetFromFile(file)
}
