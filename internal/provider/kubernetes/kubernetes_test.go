//go:build integration
// +build integration

package kubernetes

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
)

const (
	defaultWait = time.Second * 10
	defaultTick = time.Millisecond * 20
)

func TestProvider(t *testing.T) {
	// Setup the test environment.
	testEnv, cliCfg, err := startEnv()
	require.NoError(t, err)

	// Setup and start the kube provider.
	svr, err := config.NewDefaultServer()
	require.NoError(t, err)
	provider, err := New(cliCfg, svr, new(ResourceTable))
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(ctrl.SetupSignalHandler())
	go func() {
		require.NoError(t, provider.Start(ctx))
	}()

	// Stop the kube provider.
	defer func() {
		cancel()
		require.NoError(t, testEnv.Stop())
	}()

	testcases := map[string]func(context.Context, *testing.T, client.Client){
		"gatewayclass controller name": testGatewayClassController,
		"gatewayclass accepted status": testGatewayClassAcceptedStatus,
		"gateway of gatewayclass":      testGatewayOfClass,
	}
	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			tc(ctx, t, provider.manager.GetClient())
		})
	}
}

func startEnv() (*envtest.Environment, *rest.Config, error) {
	log.SetLogger(zap.New(zap.WriteTo(os.Stderr), zap.UseDevMode(true)))
	crd := filepath.Join("..", "testdata", "in")
	env := &envtest.Environment{
		CRDDirectoryPaths: []string{crd},
	}
	cfg, err := env.Start()
	if err != nil {
		return nil, nil, err
	}
	return env, cfg, nil
}

func testGatewayClassController(ctx context.Context, t *testing.T, cli client.Client) {
	gc := &gwapiv1b1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-gc-controllername",
		},
		Spec: gwapiv1b1.GatewayClassSpec{
			ControllerName: v1alpha1.GatewayControllerName,
		},
	}
	require.NoError(t, cli.Create(ctx, gc))

	defer func() {
		require.NoError(t, cli.Delete(ctx, gc))
	}()

	require.Eventually(t, func() bool {
		return cli.Get(ctx, types.NamespacedName{Name: gc.Name}, gc) == nil
	}, defaultWait, defaultTick)
	assert.Equal(t, gc.ObjectMeta.Generation, int64(1))
}

func testGatewayClassAcceptedStatus(ctx context.Context, t *testing.T, cli client.Client) {
	gc := &gwapiv1b1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-gc-accepted-status",
		},
		Spec: gwapiv1b1.GatewayClassSpec{
			ControllerName: v1alpha1.GatewayControllerName,
		},
	}
	require.NoError(t, cli.Create(ctx, gc))

	defer func() {
		require.NoError(t, cli.Delete(ctx, gc))
	}()

	require.Eventually(t, func() bool {
		if err := cli.Get(ctx, types.NamespacedName{Name: gc.Name}, gc); err != nil {
			return false
		}

		for _, cond := range gc.Status.Conditions {
			if cond.Type == string(gwapiv1b1.GatewayClassConditionStatusAccepted) && cond.Status == metav1.ConditionTrue {
				return true
			}
		}

		return false
	}, defaultWait, defaultTick)
}

func testGatewayOfClass(ctx context.Context, t *testing.T, cli client.Client) {
	gc := &gwapiv1b1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-gc-of-class",
		},
		Spec: gwapiv1b1.GatewayClassSpec{
			ControllerName: gwapiv1b1.GatewayController(v1alpha1.GatewayControllerName),
		},
	}
	require.NoError(t, cli.Create(ctx, gc))

	defer func() {
		require.NoError(t, cli.Delete(ctx, gc))
	}()

	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test-gw-of-class"}}
	require.NoError(t, cli.Create(ctx, ns))

	gw := &gwapiv1b1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-gw-of-class",
			Namespace: ns.Name,
		},
		Spec: gwapiv1b1.GatewaySpec{
			GatewayClassName: gwapiv1b1.ObjectName(gc.Name),
			Listeners: []gwapiv1b1.Listener{
				{
					Name:     "test",
					Port:     gwapiv1b1.PortNumber(int32(8080)),
					Protocol: gwapiv1b1.HTTPProtocolType,
				},
			},
		},
	}
	require.NoError(t, cli.Create(ctx, gw))
}
