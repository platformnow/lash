//go:build integration
// +build integration

package compositeresourcedefinitions

import (
	"context"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/platfornow/lash/internal/core"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/tools/clientcmd"
)

func TestGet(t *testing.T) {
	kubeconfig, err := ioutil.ReadFile(clientcmd.RecommendedHomeFile)
	assert.Nil(t, err, "expecting nil error loading kubeconfig")

	restConfig, err := core.RESTConfigFromBytes(kubeconfig, "")
	assert.Nil(t, err, "expecting nil error creating rest.Config")

	obj, err := Get(context.TODO(), restConfig, "core.pkg.platformnow.io")
	assert.Nil(t, err, "expecting nil error getting composite resource definition")
	assert.NotNil(t, obj, "expecting not nil getting composite resource definition")

	fmt.Printf("%+v\n", obj)
}
