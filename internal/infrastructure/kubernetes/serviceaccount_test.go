package kubernetes

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/ir"
)

func TestCreateServiceAccountIfNeeded(t *testing.T) {
	testCases := []struct {
		name    string
		in      *ir.Infra
		current *corev1.ServiceAccount
		out     *Resources
		expect  bool
	}{
		{
			name: "create-sa",
			in: &ir.Infra{
				Proxy: &ir.ProxyInfra{
					Name:      "test",
					Namespace: "test",
				},
			},
			out: &Resources{
				ServiceAccount: &corev1.ServiceAccount{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ServiceAccount",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace:       "test",
						Name:            "test",
						ResourceVersion: "1",
					},
				},
			},
			expect: true,
		},
		{
			name: "sa-exists",
			in: &ir.Infra{
				Proxy: &ir.ProxyInfra{
					Name:      "test",
					Namespace: "test",
				},
			},
			current: &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:       "test",
					Name:            "test",
					ResourceVersion: "34",
				},
			},
			out: &Resources{
				ServiceAccount: &corev1.ServiceAccount{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ServiceAccount",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace:       "test",
						Name:            "test",
						ResourceVersion: "34",
					},
				},
			},
			expect: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			kube := &Infra{
				mu: sync.Mutex{},
			}
			if tc.current != nil {
				kube.Client = fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects(tc.current).Build()
			} else {
				kube.Client = fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).Build()
			}
			err := kube.createServiceAccountIfNeeded(context.Background(), tc.in)
			if !tc.expect {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, *tc.out.ServiceAccount, *kube.Resources.ServiceAccount)
			}
		})
	}
}
