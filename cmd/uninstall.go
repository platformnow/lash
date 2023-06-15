package cmd

import (
	"context"
	"flag"
	"github.com/platfornow/lash/internal/argocd"
	"github.com/platfornow/lash/internal/claims"
	"github.com/platfornow/lash/internal/clusterrolebindings"
	"github.com/platfornow/lash/internal/clusterroles"
	"github.com/platfornow/lash/internal/core"
	"github.com/platfornow/lash/internal/crds"
	"github.com/platfornow/lash/internal/crossplane"
	"github.com/platfornow/lash/internal/crossplane/compositeresourcedefinitions"
	"github.com/platfornow/lash/internal/crossplane/compositions"
	"github.com/platfornow/lash/internal/crossplane/configurations"
	"github.com/platfornow/lash/internal/crossplane/controllerconfigs"
	"github.com/platfornow/lash/internal/crossplane/lock"
	"github.com/platfornow/lash/internal/crossplane/managed"
	"github.com/platfornow/lash/internal/crossplane/providerrevisions"
	"github.com/platfornow/lash/internal/crossplane/providers"
	"github.com/platfornow/lash/internal/eventbus"
	"github.com/platfornow/lash/internal/events"
	"github.com/platfornow/lash/internal/log"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"os"
)

type RemoveFinalizersOpts struct {
	RESTConfig *rest.Config
}

func newUninstallCmd() *cobra.Command {
	o := uninstallOpts{
		bus:     eventbus.New(),
		verbose: false,
	}

	cmd := &cobra.Command{
		Use:                   "uninstall",
		DisableSuggestions:    true,
		DisableFlagsInUseLine: true,
		Args:                  cobra.NoArgs,
		Short:                 "Uninstall Landscape",
		SilenceErrors:         true,
		Example:               "  lash uninstall",
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

	return cmd
}

type uninstallOpts struct {
	kubeconfig        string
	kubeconfigContext string
	bus               eventbus.Bus
	restConfig        *rest.Config
	namespace         string
	verbose           bool
	dryRun            bool
}

func (o *uninstallOpts) complete() (err error) {
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

func (o *uninstallOpts) run() error {
	ctx := context.TODO()

	err := o.deleteArgoApplications(ctx)

	err = o.deleteArgoProjects(ctx)

	if err := o.deletePackages(ctx); err != nil {
		return err
	}

	if err := o.deleteProviders(ctx); err != nil {
		return err
	}

	if err := o.deleteReleases(ctx); err != nil {
		return err
	}

	if err := o.deleteControllerConfigs(ctx); err != nil {
		return err
	}

	if err := o.deleteXRDs(ctx); err != nil {
		return err
	}

	if err := o.deleteCrossplane(ctx); err != nil {
		return err
	}

	o.bus.Publish(events.NewStartWaitEvent("Finishing cleaning..."))
	o.deleteCompositions(ctx)
	o.deleteCRDsQuietly(ctx)
	o.deleteClusterRoleBindingsQuietly(ctx)
	o.deleteClusterRolesQuietly(ctx)
	o.deleteManagedResources(ctx)
	o.deleteCrossplaneLock(ctx)
	o.deleteProviderRevisions(ctx)
	err = o.deleteNamespace(ctx)
	o.bus.Publish(events.NewStartWaitEvent("Cleaning done"))

	return err
}

func (o *uninstallOpts) deleteArgoApplications(ctx context.Context) error {
	all, err := argocd.ListApplications(ctx, o.restConfig)
	if err != nil {
		return err
	}

	if len(all) == 0 {
		return nil
	}

	if o.dryRun {
		o.bus.Publish(events.NewDebugEvent("found [%d] applications", len(all)))
	}

	for _, el := range all {
		if o.dryRun {
			o.bus.Publish(events.NewDebugEvent(" > %s", el.GetName()))
			continue
		}

		o.bus.Publish(events.NewStartWaitEvent("removing application %s...", el.GetName()))

		err = RemoveFinalizers(ctx, &el, o.restConfig)
		if err != nil {
			return err
		}

		err := core.Delete(ctx, core.DeleteOpts{
			RESTConfig: o.restConfig,
			Object:     &el,
		})
		if err != nil {
			return err
		}

		o.bus.Publish(events.NewDoneEvent("application %s uninstalled", el.GetName()))
	}

	return nil

}

func (o *uninstallOpts) deleteArgoProjects(ctx context.Context) error {
	all, err := argocd.ListProjects(ctx, o.restConfig)
	if err != nil {
		return err
	}

	if len(all) == 0 {
		return nil
	}

	if o.dryRun {
		o.bus.Publish(events.NewDebugEvent("found [%d] projects", len(all)))
	}

	for _, el := range all {
		if o.dryRun {
			o.bus.Publish(events.NewDebugEvent(" > %s", el.GetName()))
			continue
		}

		o.bus.Publish(events.NewStartWaitEvent("removing project %s...", el.GetName()))

		err = RemoveFinalizers(ctx, &el, o.restConfig)
		if err != nil {
			return err
		}

		err := core.Delete(ctx, core.DeleteOpts{
			RESTConfig: o.restConfig,
			Object:     &el,
		})
		if err != nil {
			return err
		}

		o.bus.Publish(events.NewDoneEvent("project %s uninstalled", el.GetName()))
	}

	return nil

}

func (o *uninstallOpts) deleteXRDs(ctx context.Context) error {
	all, err := compositeresourcedefinitions.List(ctx, o.restConfig)
	if err != nil {
		return err
	}

	if len(all) == 0 {
		return nil
	}

	if o.dryRun {
		o.bus.Publish(events.NewDebugEvent("found [%d] composite resource definitions", len(all)))
	}

	for _, el := range all {
		if o.dryRun {
			o.bus.Publish(events.NewDebugEvent(" > %s", el.GetName()))
			continue
		}

		o.bus.Publish(events.NewStartWaitEvent("removing xrd %s...", el.GetName()))

		err = RemoveFinalizers(ctx, &el, o.restConfig)
		if err != nil {
			return err
		}

		err := core.Delete(ctx, core.DeleteOpts{
			RESTConfig: o.restConfig,
			Object:     &el,
		})
		if err != nil {
			return err
		}

		o.bus.Publish(events.NewDoneEvent("provider config %s uninstalled", el.GetName()))
	}

	return nil

}

func (o *uninstallOpts) deleteManagedResources(ctx context.Context) error {
	// Remove all managed resources modules
	coreModule := claims.NewCore("core")
	gitopsModule := claims.NewGitops("core-argo-cd")

	if err := o.deleteCompositeResources(ctx, coreModule); err != nil {
		return err
	}
	if err := o.deleteCompositeResources(ctx, gitopsModule); err != nil {
		return err
	}

	// Remove Releases
	err := o.deleteReleases(ctx)
	if err != nil {
		return err
	}

	// Remove all managed resources (kubectl get managed)
	// todo Only removing kubernetes objects for now, we may need to add other types of managed resources
	all, err := managed.ListK8Objects(ctx, o.restConfig)
	if err != nil {
		return err
	}

	if len(all) == 0 {
		return nil
	}

	if o.dryRun {
		o.bus.Publish(events.NewDebugEvent("found [%d] composite resource definitions", len(all)))
	}

	for _, el := range all {
		if o.dryRun {
			o.bus.Publish(events.NewDebugEvent(" > %s", el.GetName()))
			continue
		}

		o.bus.Publish(events.NewStartWaitEvent("removing xrd %s...", el.GetName()))

		err = RemoveFinalizers(ctx, &el, o.restConfig)
		if err != nil {
			return err
		}

		err := core.Delete(ctx, core.DeleteOpts{
			RESTConfig: o.restConfig,
			Object:     &el,
		})
		if err != nil {
			return err
		}

		o.bus.Publish(events.NewDoneEvent("provider config %s uninstalled", el.GetName()))
	}

	return nil

}

func (o *uninstallOpts) deleteCompositeResources(ctx context.Context, resource claims.ManagedResource) error {

	all, err := claims.List(ctx, o.restConfig, resource)
	if err != nil {
		return err
	}

	if len(all) == 0 {
		return nil
	}

	if o.dryRun {
		o.bus.Publish(events.NewDebugEvent("found [%d] composite resource definitions", len(all)))
	}

	for _, el := range all {
		if o.dryRun {
			o.bus.Publish(events.NewDebugEvent(" > %s", el.GetName()))
			continue
		}

		o.bus.Publish(events.NewStartWaitEvent("removing xrd %s...", el.GetName()))

		err = RemoveFinalizers(ctx, &el, o.restConfig)
		if err != nil {
			return err
		}

		err := core.Delete(ctx, core.DeleteOpts{
			RESTConfig: o.restConfig,
			Object:     &el,
		})
		if err != nil {
			return err
		}

		o.bus.Publish(events.NewDoneEvent("provider config %s uninstalled", el.GetName()))
	}

	return nil

}

func (o *uninstallOpts) deleteReleases(ctx context.Context) error {
	all, err := providers.ListHelmReleases(ctx, o.restConfig)
	if err != nil {
		return err
	}

	if len(all) == 0 {
		return nil
	}

	if o.dryRun {
		o.bus.Publish(events.NewDebugEvent("found [%d] releases", len(all)))
	}

	for _, el := range all {
		if o.dryRun {
			o.bus.Publish(events.NewDebugEvent(" > %s", el.GetName()))
			continue
		}

		o.bus.Publish(events.NewStartWaitEvent("removing release %s...", el.GetName()))

		err = RemoveFinalizers(ctx, &el, o.restConfig)
		if err != nil {
			return err
		}

		err := core.Delete(ctx, core.DeleteOpts{
			RESTConfig: o.restConfig,
			Object:     &el,
		})
		if err != nil {
			return err
		}

		o.bus.Publish(events.NewDoneEvent("release %s uninstalled", el.GetName()))
	}

	return nil

}

func (o *uninstallOpts) deleteCrossplane(ctx context.Context) error {
	pod, err := crossplane.InstalledPOD(ctx, o.restConfig)
	if err != nil {
		return err
	}

	if pod == nil {
		if o.verbose {
			o.bus.Publish(events.NewDebugEvent("crossplane not found"))
		}
		return nil
	}

	ver, err := crossplane.PODImageVersion(pod)
	if err != nil {
		return err
	}

	if o.dryRun {
		o.bus.Publish(events.NewDebugEvent(
			"found crossplane (%s) pod: %s in namespace: %s",
			ver, pod.GetName(), pod.GetNamespace()))
		return nil
	}

	o.bus.Publish(events.NewStartWaitEvent("uninstalling crossplane %s...", ver))

	err = crossplane.Uninstall(crossplane.UninstallOpts{
		RESTConfig: o.restConfig,
		EventBus:   o.bus,
		Namespace:  pod.GetNamespace(),
		Verbose:    o.verbose,
	})
	if err != nil {
		return err
	}

	o.bus.Publish(events.NewDoneEvent("crossplane %s uninstalled", ver))

	return nil
}

func (o *uninstallOpts) deleteHelmProviderConfigs(ctx context.Context) error {
	all, err := providers.ListHelmProviderConfigs(ctx, o.restConfig)
	if err != nil {
		return err
	}

	if len(all) == 0 {
		return nil
	}

	if o.dryRun {
		o.bus.Publish(events.NewDebugEvent("found [%d] provider configs for Helm", len(all)))
	}

	for _, el := range all {
		if o.dryRun {
			o.bus.Publish(events.NewDebugEvent(" > %s", el.GetName()))
			continue
		}

		o.bus.Publish(events.NewStartWaitEvent("removing provider config %s...", el.GetName()))

		err = RemoveFinalizers(ctx, &el, o.restConfig)
		if err != nil {
			return err
		}

		err := core.Delete(ctx, core.DeleteOpts{
			RESTConfig: o.restConfig,
			Object:     &el,
		})
		if err != nil {
			return err
		}

		o.bus.Publish(events.NewDoneEvent("provider config %s uninstalled", el.GetName()))
	}

	return nil

}

func (o *uninstallOpts) deleteK8ProviderConfigs(ctx context.Context) error {
	all, err := providers.ListK8ProviderConfigs(ctx, o.restConfig)
	if err != nil {
		return err
	}

	if len(all) == 0 {
		return nil
	}

	if o.dryRun {
		o.bus.Publish(events.NewDebugEvent("found [%d] provider configs for Helm", len(all)))
	}

	for _, el := range all {
		if o.dryRun {
			o.bus.Publish(events.NewDebugEvent(" > %s", el.GetName()))
			continue
		}

		o.bus.Publish(events.NewStartWaitEvent("removing provider config %s...", el.GetName()))

		err = RemoveFinalizers(ctx, &el, o.restConfig)
		if err != nil {
			return err
		}

		err := core.Delete(ctx, core.DeleteOpts{
			RESTConfig: o.restConfig,
			Object:     &el,
		})
		if err != nil {
			return err
		}

		o.bus.Publish(events.NewDoneEvent("provider config %s uninstalled", el.GetName()))
	}

	return nil

}

func (o *uninstallOpts) deleteProviderRevisions(ctx context.Context) error {
	all, err := providerrevisions.List(ctx, o.restConfig)
	if err != nil {
		return err
	}

	if len(all) == 0 {
		return nil
	}

	if o.dryRun {
		o.bus.Publish(events.NewDebugEvent("found [%d] provider revisions", len(all)))
	}

	for _, el := range all {
		if o.dryRun {
			o.bus.Publish(events.NewDebugEvent(" > %s", el.GetName()))
			continue
		}

		o.bus.Publish(events.NewStartWaitEvent("removing provider revisions %s...", el.GetName()))

		err = RemoveFinalizers(ctx, &el, o.restConfig)
		if err != nil {
			return err
		}

		err := core.Delete(ctx, core.DeleteOpts{
			RESTConfig: o.restConfig,
			Object:     &el,
		})
		if err != nil {
			return err
		}

		o.bus.Publish(events.NewDoneEvent("provider revision %s uninstalled", el.GetName()))
	}

	return nil

}
func (o *uninstallOpts) deleteProviders(ctx context.Context) error {
	all, err := providers.List(ctx, o.restConfig)
	if err != nil {
		return err
	}

	if len(all) == 0 {
		return nil
	}

	if o.dryRun {
		o.bus.Publish(events.NewDebugEvent("found [%d] packages", len(all)))
	}

	for _, el := range all {
		if o.dryRun {
			o.bus.Publish(events.NewDebugEvent(" > %s", el.GetName()))
			continue
		}

		o.bus.Publish(events.NewStartWaitEvent("uninstalling package %s...", el.GetName()))

		err := core.Delete(ctx, core.DeleteOpts{
			RESTConfig: o.restConfig,
			Object:     &el,
		})
		if err != nil {
			return err
		}

		o.bus.Publish(events.NewDoneEvent("package %s uninstalled", el.GetName()))
	}

	return nil
}

func (o *uninstallOpts) deleteControllerConfigs(ctx context.Context) error {
	all, err := controllerconfigs.ListAll(ctx, o.restConfig)
	if err != nil {
		return err
	}

	if len(all) == 0 {
		return nil
	}

	if o.dryRun {
		o.bus.Publish(events.NewDebugEvent("found [%d] controller configs", len(all)))
	}

	for _, el := range all {
		if o.dryRun {
			o.bus.Publish(events.NewDebugEvent(" > %s", el.GetName()))
			continue
		}

		err := controllerconfigs.Delete(ctx, controllerconfigs.DeleteOpts{
			RESTConfig: o.restConfig,
			Name:       el.GetName(),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (o *uninstallOpts) deletePackages(ctx context.Context) error {
	all, err := configurations.List(ctx, o.restConfig)
	if err != nil {
		return err
	}
	if len(all) == 0 {
		return nil
	}

	if o.dryRun {
		o.bus.Publish(events.NewDebugEvent("found [%d] packages", len(all)))
	}

	for _, el := range all {
		if o.dryRun {
			o.bus.Publish(events.NewDebugEvent(" > %s", el.GetName()))
			continue
		}

		o.bus.Publish(events.NewStartWaitEvent("uninstalling package %s...", el.GetName()))
		err := core.Delete(ctx, core.DeleteOpts{
			RESTConfig: o.restConfig,
			Object:     &el,
		})
		if err != nil {
			return err
		}
		o.bus.Publish(events.NewDoneEvent("package %s uninstalled", el.GetName()))
	}

	return nil
}

func (o *uninstallOpts) deleteCompositions(ctx context.Context) {
	all, err := compositions.List(ctx, o.restConfig)
	if err != nil {
		return
	}

	if len(all) == 0 {
		return
	}

	if o.dryRun {
		o.bus.Publish(events.NewDebugEvent("found [%d] compositions", len(all)))
	}

	for _, el := range all {
		if o.dryRun {
			o.bus.Publish(events.NewDebugEvent(" > %s", el.GetName()))
			continue
		}
		_ = core.Delete(ctx, core.DeleteOpts{
			RESTConfig: o.restConfig,
			Object:     &el,
		})
	}
}

func (o *uninstallOpts) deleteCrossplaneLock(ctx context.Context) error {
	all, err := lock.List(ctx, o.restConfig)
	if err != nil {
		return err
	}

	if len(all) == 0 {
		return nil
	}

	if o.dryRun {
		o.bus.Publish(events.NewDebugEvent("found [%d] lock", len(all)))
	}

	for _, el := range all {
		if o.dryRun {
			o.bus.Publish(events.NewDebugEvent(" > %s", el.GetName()))
			continue
		}

		o.bus.Publish(events.NewStartWaitEvent("removing lock %s...", el.GetName()))

		err = RemoveFinalizers(ctx, &el, o.restConfig)
		if err != nil {
			return err
		}

		err := core.Delete(ctx, core.DeleteOpts{
			RESTConfig: o.restConfig,
			Object:     &el,
		})
		if err != nil {
			return err
		}

		o.bus.Publish(events.NewDoneEvent("lock %s uninstalled", el.GetName()))
	}

	return nil

}
func (o *uninstallOpts) deleteCRDsQuietly(ctx context.Context) {
	all, err := crds.List(ctx, o.restConfig)
	if err != nil || len(all) == 0 {
		return
	}

	items := []unstructured.Unstructured{}
	for _, el := range all {
		res := crds.CRDInstances(ctx, o.restConfig, el.GetName())
		if res != nil {
			items = append(items, res...)
		}
	}

	if o.dryRun {
		o.bus.Publish(events.NewDebugEvent("found [%d] custom resources", len(items)))
		for _, el := range items {
			o.bus.Publish(events.NewDebugEvent(" > %s (%s)", el.GetName(), el.GetAPIVersion()))
		}
	} else {
		for _, el := range items {
			o.bus.Publish(events.NewDebugEvent(" > %s", el.GetName()))
			err = crds.PatchAndDelete(ctx, o.restConfig, &el)
		}
	}

	if o.dryRun {
		o.bus.Publish(events.NewDebugEvent("found [%d] custom resource definitions", len(all)))
		for _, el := range all {
			o.bus.Publish(events.NewDebugEvent(" > %s (%s)", el.GetName(), el.GetAPIVersion()))
		}
		return
	}

	for _, el := range all {
		err = crds.PatchAndDelete(ctx, o.restConfig, &el)
		err = crds.PatchAndDeleteMergeType(ctx, o.restConfig, &el)
	}
}

func (o *uninstallOpts) deleteClusterRoleBindingsQuietly(ctx context.Context) {
	all, err := clusterrolebindings.List(ctx, o.restConfig)
	if err != nil {
		return
	}

	res, err := core.Filter(all, func(obj unstructured.Unstructured) bool {
		accept := (obj.GetName() == "provider-helm-admin-binding")
		accept = accept || (obj.GetName() == "provider-kubernetes-admin-binding")
		accept = accept || (obj.GetName() == "argocd-server-repo-server")
		accept = accept || (obj.GetName() == "argocd-server-server")
		accept = accept || (obj.GetName() == "argocd-server-application-controller")

		return accept
	})

	if len(res) == 0 || err != nil {
		return
	}

	if o.dryRun {
		o.bus.Publish(events.NewDebugEvent("found [%d] cluster role bindings", len(res)))
	}

	for _, el := range res {
		if o.dryRun {
			o.bus.Publish(events.NewDebugEvent("> %s", el.GetName()))
			continue
		}
		_ = clusterrolebindings.Delete(ctx, clusterrolebindings.DeleteOpts{
			RESTConfig: o.restConfig,
			Name:       el.GetName(),
		})
	}
}
func (o *uninstallOpts) deleteClusterRolesQuietly(ctx context.Context) {
	all, err := clusterroles.List(ctx, o.restConfig)
	if err != nil {
		return
	}

	res, err := core.Filter(all, func(obj unstructured.Unstructured) bool {
		accept := (obj.GetName() == "argocd-server-aggregate-to-admin")
		accept = accept || (obj.GetName() == "argocd-server-aggregate-to-edit")
		accept = accept || (obj.GetName() == "argocd-server-aggregate-to-view")
		accept = accept || (obj.GetName() == "argocd-server-application-controller")
		accept = accept || (obj.GetName() == "argocd-server-repo-server")
		accept = accept || (obj.GetName() == "argocd-server-server")

		return accept
	})

	if len(res) == 0 || err != nil {
		return
	}

	if o.dryRun {
		o.bus.Publish(events.NewDebugEvent("found [%d] cluster roles", len(res)))
	}

	for _, el := range res {
		if o.dryRun {
			o.bus.Publish(events.NewDebugEvent("> %s", el.GetName()))
			continue
		}
		_ = clusterroles.Delete(ctx, clusterroles.DeleteOpts{
			RESTConfig: o.restConfig,
			Name:       el.GetName(),
		})
	}
}

func (o *uninstallOpts) deleteNamespace(ctx context.Context) error {
	obj, err := core.Get(ctx, core.GetOpts{
		RESTConfig: o.restConfig,
		GVK: schema.GroupVersionKind{
			Version: "v1", Kind: "Namespace",
		},
		Name: o.namespace,
	})
	if err != nil || obj == nil {
		return err
	}

	err = RemoveFinalizers(ctx, obj, o.restConfig)
	if err != nil {
		return err
	}

	err = core.Delete(ctx, core.DeleteOpts{
		RESTConfig: o.restConfig,
		Object:     obj,
	})

	return nil
}

func RemoveFinalizers(ctx context.Context, obj *unstructured.Unstructured, restConfig *rest.Config) error {

	// Remove finalizers from metadata
	err := core.Patch(ctx, core.PatchOpts{
		RESTConfig: restConfig,
		GVK:        obj.GroupVersionKind(),
		PatchData:  []byte(`{"metadata":{"finalizers":[]}}`),
		PatchType:  types.MergePatchType,
		Name:       obj.GetName(),
		Namespace:  obj.GetNamespace(),
	})

	if err != nil {
		return err
	}

	return err
}
