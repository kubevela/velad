package utils

import (
	"context"
	v1 "k8s.io/api/core/v1"
	"strings"

	core "github.com/oam-dev/kubevela/apis/core.oam.dev"
	"github.com/oam-dev/kubevela/apis/core.oam.dev/v1alpha2"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// GetClient returns a client.Client
func GetClient() (client.Client, error) {
	restConf, err := config.GetConfig()
	if err != nil {
		return nil, err
	}
	scheme := runtime.NewScheme()
	_ = core.AddToScheme(scheme)
	_ = v1.AddToScheme(scheme)
	return client.New(restConf, client.Options{
		Scheme: scheme,
	})

}

// EditGatewayDefinition edits the gateway trait definition. In VelaD, we use Traefik instead of Nginx.
func EditGatewayDefinition() error {
	cli, err := GetClient()
	if err != nil {
		return err
	}
	gateway := &v1alpha2.TraitDefinition{}
	ctx := context.Background()
	err = cli.Get(ctx, client.ObjectKey{
		Name:      "gateway",
		Namespace: "vela-system",
	}, gateway)
	if err != nil {
		return err
	}
	gateway.Spec.Schematic.CUE.Template = strings.ReplaceAll(gateway.Spec.Schematic.CUE.Template, "*\"nginx\"", "*\"traefik\"")
	return cli.Update(ctx, gateway)
}
