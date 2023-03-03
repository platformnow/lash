package cmd

import (
	"context"
	"flag"
	"github.com/platfornow/lash/internal/core"
	"github.com/platfornow/lash/internal/eventbus"
	"github.com/platfornow/lash/internal/events"
	"github.com/platfornow/lash/internal/log"
	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"os"
)

func newCreateCmd() *cobra.Command {
	o := createOpts{
		bus:     eventbus.New(),
		verbose: false,
	}

	cmd := &cobra.Command{
		Use:                   "create",
		DisableSuggestions:    true,
		DisableFlagsInUseLine: true,
		Args:                  cobra.NoArgs,
		Short:                 "create a catalog repository",
		SilenceErrors:         true,
		Example:               "  lash create",
		RunE: func(cmd *cobra.Command, args []string) error {
			l := log.GetInstance()
			if o.verbose {
				l.SetLevel(log.DebugLevel)
			}

			handler := events.LogHandler(l)
			o.bus = eventbus.New()
			eids := []eventbus.Subscription{
				o.bus.Subscribe(events.StartWaitEventID, handler),
				o.bus.Subscribe(events.StopWaitEventID, handler),
				o.bus.Subscribe(events.DoneEventID, handler),
				o.bus.Subscribe(events.DebugEventID, handler),
			}
			defer func() {
				for _, e := range eids {
					o.bus.Unsubscribe(e)
				}
			}()

			if err := o.complete(); err != nil {
				return err
			}

			return o.run()
		},
	}

	defaultKubeconfig := os.Getenv(clientcmd.RecommendedConfigPathEnvVar)
	if len(defaultKubeconfig) == 0 {
		defaultKubeconfig = clientcmd.RecommendedHomeFile
	}

	cmd.Flags().BoolVarP(&o.verbose, "verbose", "v", false, "dump verbose output")
	cmd.Flags().BoolVar(&o.dryRun, "dry-run", false, "preview the object that would be deleted, without really deleting it")
	cmd.Flags().StringVar(&o.kubeconfig, clientcmd.RecommendedConfigPathFlag, defaultKubeconfig, "absolute path to the kubeconfig file")
	cmd.Flags().StringVar(&o.kubeconfigContext, "context", "", "kubeconfig context to use")
	cmd.Flags().StringVarP(&o.namespace, "namespace", "n", "landscape-system", "namespace where to install landscape idp")
	cmd.Flags().StringVar(&o.namespace, "source", "", "repository source")
	cmd.Flags().StringVar(&o.namespace, "destination", "", "repository destination")

	return cmd
}

type createOpts struct {
	kubeconfig        string
	kubeconfigContext string
	bus               eventbus.Bus
	restConfig        *rest.Config
	namespace         string
	verbose           bool
	dryRun            bool
	source            string
	destination       string
}

func (o *createOpts) complete() (err error) {
	flag.Set("logtostderr", "false")
	flag.Parse()
	klog.InitFlags(nil)

	yml, err := os.ReadFile(o.kubeconfig)
	if err != nil {
		return err
	}

	o.restConfig, err = core.RESTConfigFromBytes(yml, o.kubeconfigContext)
	if err != nil {
		return err
	}

	return nil
}

func (o *createOpts) run() error {
	ctx := context.TODO()

	o.bus.Publish(events.NewStartWaitEvent("finishing cleaning..."))

	o.createRepository(ctx)

	o.bus.Publish(events.NewDoneEvent("Created"))

	return nil
}

func (o *createOpts) createRepository(ctx context.Context) error {

	return nil

}

