package ir

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
)

func TestValidateInfra(t *testing.T) {
	testCases := []struct {
		name   string
		infra  *Infra
		expect bool
	}{
		{
			name:   "default",
			infra:  NewInfra(),
			expect: true,
		},
		{
			name: "no-name",
			infra: &Infra{
				Proxy: &ProxyInfra{
					Name:      "",
					Namespace: "test",
					Image:     "image",
					Listeners: NewProxyListeners(),
				},
			},
			expect: false,
		},
		{
			name: "no-namespace",
			infra: &Infra{
				Proxy: &ProxyInfra{
					Name:      "test",
					Namespace: "",
					Image:     "image",
					Listeners: NewProxyListeners(),
				},
			},
			expect: false,
		},
		{
			name: "no-listeners",
			infra: &Infra{
				Proxy: &ProxyInfra{
					Name:      "test",
					Namespace: "test",
					Image:     "image",
				},
			},
			expect: true,
		},
		{
			name: "no-listener-ports",
			infra: &Infra{
				Proxy: &ProxyInfra{
					Name:      "test",
					Namespace: "test",
					Image:     "image",
					Listeners: []ProxyListener{
						{
							Ports: []ListenerPort{},
						},
					},
				},
			},
			expect: false,
		},
		{
			name: "no-listener-port-name",
			infra: &Infra{
				Proxy: &ProxyInfra{
					Name:      "test",
					Namespace: "test",
					Image:     "image",
					Listeners: []ProxyListener{
						{
							Ports: []ListenerPort{
								{
									Port: DefaultHTTPListenerPort,
								},
							},
						},
					},
				},
			},
			expect: false,
		},
		{
			name: "no-listener-port-number",
			infra: &Infra{
				Proxy: &ProxyInfra{
					Name:      "test",
					Namespace: "test",
					Image:     "image",
					Listeners: []ProxyListener{
						{
							Ports: []ListenerPort{
								{
									Name: "http",
								},
							},
						},
					},
				},
			},
			expect: false,
		},
		{
			name: "no-image",
			infra: &Infra{
				Proxy: &ProxyInfra{
					Name:      "test",
					Namespace: "test",
					Image:     "",
				},
			},
			expect: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateInfra(tc.infra)
			if !tc.expect {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNewInfra(t *testing.T) {
	testCases := []struct {
		name     string
		expected *Infra
	}{
		{
			name: "default infra",
			expected: &Infra{
				// Kube is the only supported provider type.
				Provider: v1alpha1.ProviderTypePtr(v1alpha1.ProviderTypeKubernetes),
				Proxy:    NewProxyInfra(),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := NewInfra()
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestNewProxyInfra(t *testing.T) {
	testCases := []struct {
		name     string
		expected *ProxyInfra
	}{
		{
			name: "default infra",
			expected: &ProxyInfra{
				Name:      DefaultProxyName,
				Namespace: DefaultProxyNamespace,
				Image:     DefaultProxyImage,
				Listeners: NewProxyListeners(),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := NewProxyInfra()
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestObjectName(t *testing.T) {
	defaultInfra := NewInfra()

	testCases := []struct {
		name     string
		infra    *Infra
		expected string
	}{
		{
			name:     "default infra",
			infra:    defaultInfra,
			expected: "envoy-default",
		},
		{
			name: "defined infra",
			infra: &Infra{
				Proxy: &ProxyInfra{
					Name: "foo",
				},
			},
			expected: "envoy-foo",
		},
		{
			name: "unspecified infra name",
			infra: &Infra{
				Proxy: &ProxyInfra{},
			},
			expected: "envoy-default",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.infra.Proxy.ObjectName()
			require.Equal(t, tc.expected, actual)
		})
	}
}
