package argocd

import (
	"context"

	"github.com/platfornow/lash/internal/core"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
)

func ListApplications(ctx context.Context, restConfig *rest.Config) ([]unstructured.Unstructured, error) {
	return core.List(ctx, core.ListOpts{
		RESTConfig: restConfig,
		GVK: schema.GroupVersionKind{
			Group:   "argoproj.io",
			Version: "v1alpha1",
			Kind:    "Application",
		},
	})
}

func ListProjects(ctx context.Context, restConfig *rest.Config) ([]unstructured.Unstructured, error) {
	return core.List(ctx, core.ListOpts{
		RESTConfig: restConfig,
		GVK: schema.GroupVersionKind{
			Group:   "argoproj.io",
			Version: "v1alpha1",
			Kind:    "AppProject",
		},
	})
}
