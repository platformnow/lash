package providers

import (
	"context"

	"github.com/platfornow/lash/internal/core"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
)

func List(ctx context.Context, restConfig *rest.Config) ([]unstructured.Unstructured, error) {
	return core.List(ctx, core.ListOpts{
		RESTConfig: restConfig,
		GVK: schema.GroupVersionKind{
			Group:   "pkg.crossplane.io",
			Version: "v1",
			Kind:    "Provider",
		},
	})
}

func ListHelmReleases(ctx context.Context, restConfig *rest.Config) ([]unstructured.Unstructured, error) {
	return core.List(ctx, core.ListOpts{
		RESTConfig: restConfig,
		GVK: schema.GroupVersionKind{
			Group:   "helm.crossplane.io",
			Version: "v1beta1",
			Kind:    "Release",
		},
	})
}

func ListHelmProviderConfigs(ctx context.Context, restConfig *rest.Config) ([]unstructured.Unstructured, error) {
	return core.List(ctx, core.ListOpts{
		RESTConfig: restConfig,
		GVK: schema.GroupVersionKind{
			Group:   "helm.crossplane.io",
			Version: "v1alpha1",
			Kind:    "ProviderConfig",
		},
	})
}

func ListK8ProviderConfigs(ctx context.Context, restConfig *rest.Config) ([]unstructured.Unstructured, error) {
	return core.List(ctx, core.ListOpts{
		RESTConfig: restConfig,
		GVK: schema.GroupVersionKind{
			Group:   "kubernetes.crossplane.io",
			Version: "v1alpha1",
			Kind:    "ProviderConfig",
		},
	})
}
