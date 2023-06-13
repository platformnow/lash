package claims

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
)

type ModuleOpts struct {
	RESTConfig *rest.Config
	Data       map[string]interface{}
}

type ManagedResource interface {
	Apply(ctx context.Context, opts ModuleOpts) error
	GetGroupVersionKind() schema.GroupVersionKind
	GetGroupVersionResource() schema.GroupVersionResource
	getName() string
}

type Core struct {
	name string
}

func (c Core) getName() string {
	return c.name
}

func (c Core) SetName(s string) {
	c.name = s
}

type GitOps struct {
	name string
}

func (c GitOps) getName() string {
	return c.name
}

func (c GitOps) SetName(s string) {
	c.name = s
}

func ApplyModule(ctx context.Context, opts ModuleOpts, resource ManagedResource) error {
	return resource.Apply(ctx, opts)
}
