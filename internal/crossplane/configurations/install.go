package configurations

import (
	"context"
	"fmt"
	"github.com/platfornow/lash/internal/catalog"
	"github.com/platfornow/lash/internal/core"
	"github.com/platfornow/lash/internal/eventbus"
	"github.com/platfornow/lash/internal/events"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
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

	data, err := catalog.FetchManifest(opts.Info)
	if err != nil {
		return err
	}

	if opts.Verbose && opts.EventBus != nil {
		opts.EventBus.Publish(events.NewDebugEvent("Retrieved YAML ... \n%s", data))
	}

	// decode the YAML
	obj, gvk, err := core.DecodeYAML(data)
	if err != nil {
		return err
	}

	if !isCrossplaneProvider(gvk) {
		return fmt.Errorf("%s is not a provider", obj.GetName())
	}

	obj.SetLabels(map[string]string{
		core.InstalledByLabel: core.InstalledByValue,
	})

	// install the object
	opts.EventBus.Publish(events.NewStartWaitEvent("Installing %s %s", opts.Info.Name, opts.Info.Manifest))
	err = core.Apply(ctx, core.ApplyOpts{RESTConfig: opts.RESTConfig, Object: obj, GVK: *gvk})
	if err != nil {
		return err
	}

	opts.EventBus.Publish(events.NewDoneEvent("Installed %s %s", opts.Info.Name, opts.Info.Manifest))

	// wait for it
	return WaitUntilHealtyAndInstalled(ctx, opts.RESTConfig, obj.GetName())
}

func isCrossplaneProvider(gvk *schema.GroupVersionKind) bool {
	return gvk.Group == "pkg.crossplane.io"
}
