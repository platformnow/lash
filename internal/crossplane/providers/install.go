package providers

import (
	"context"
	"fmt"
	"github.com/platfornow/lash/internal/catalog"
	"github.com/platfornow/lash/internal/core"
	"github.com/platfornow/lash/internal/eventbus"
	"github.com/platfornow/lash/internal/events"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"
	"path"
	"strings"
)

type InstallOpts struct {
	RESTConfig *rest.Config
	Info       *catalog.PackageInfo
	Namespace  string
	EventBus   eventbus.Bus
	Verbose    bool
}

type Yaml struct {
	name string
	url  string
}

func InstallFromRepo(ctx context.Context, opts InstallOpts) error {

	pp := &catalog.PackageInfo{}
	*pp = *opts.Info

	var yamls [4]Yaml
	yamls[0] = Yaml{name: "provider", url: pp.Manifest}
	yamls[1] = Yaml{name: "controller-config", url: strings.Replace(pp.Manifest, path.Base(pp.Manifest), "controller-config.yaml", -1)}
	yamls[2] = Yaml{name: "service-account", url: strings.Replace(pp.Manifest, path.Base(pp.Manifest), "service-account.yaml", -1)}
	yamls[3] = Yaml{name: "cluster-role-binding", url: strings.Replace(pp.Manifest, path.Base(pp.Manifest), "cluster-role-binding.yaml", -1)}

	for _, yaml := range yamls {
		yamlData, err := catalog.FetchManifestFromUrl(yaml.url)
		if err != nil {
			return err
		}

		if opts.Verbose && opts.EventBus != nil {
			opts.EventBus.Publish(events.NewDebugEvent("Retrieved YAML ... \n%s", yamlData))
		}

		// decode the YAML
		obj, gvk, err := core.DecodeYAML(yamlData)
		if err != nil {
			return err
		}

		if yaml.name == "provider" && !isCrossplaneProvider(gvk) {
			return fmt.Errorf("%s is not a provider", obj.GetName())
		}

		// update the controller config with labels, so we can watch based on them
		if yaml.name == "controller-config" {

			metadata := map[string]interface{}{
				"labels": map[string]interface{}{
					core.InstalledByLabel: core.InstalledByValue,
					core.PackageNameLabel: opts.Info.Name,
				},
			}

			err = unstructured.SetNestedField(obj.Object, metadata, "spec", "metadata")
			if err != nil {
				return err
			}

			obj.SetLabels(map[string]string{
				core.InstalledByLabel: core.InstalledByValue,
			})

		}

		opts.EventBus.Publish(events.NewStartWaitEvent("Installing %s %s", pp.Name, yaml.name))
		err = core.Apply(ctx, core.ApplyOpts{RESTConfig: opts.RESTConfig, Object: obj, GVK: *gvk})
		if err != nil {
			return err
		}

		opts.EventBus.Publish(events.NewDoneEvent("Installed %s %s", pp.Name, yaml.name))
	}

	// wait for it
	return waitUntilProviderIsReady(ctx, opts.RESTConfig, opts.Info.Name, opts.Namespace)
}

func waitUntilProviderIsReady(ctx context.Context, restConfig *rest.Config, name, namespace string) error {
	req, err := labels.NewRequirement(core.PackageNameLabel, selection.Equals, []string{name})
	if err != nil {
		return err
	}

	sel := labels.NewSelector()
	sel = sel.Add(*req)

	stopFn := func(et watch.EventType, obj *unstructured.Unstructured) (bool, error) {
		pod := &corev1.Pod{}
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), &pod)
		if err != nil {
			return false, err
		}

		for _, cond := range pod.Status.Conditions {
			if (cond.Type == corev1.PodReady) && (cond.Status == corev1.ConditionTrue) {
				return true, nil
			}
		}

		return false, nil
	}

	return core.Watch(ctx, core.WatchOpts{
		RESTConfig: restConfig,
		GVR:        schema.GroupVersionResource{Version: "v1", Resource: "pods"},
		Namespace:  namespace,
		Selector:   sel,
		StopFn:     stopFn,
	})
}

func isCrossplaneProvider(gvk *schema.GroupVersionKind) bool {
	return gvk.Group == "pkg.crossplane.io"
}
