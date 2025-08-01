package clustercollectorsgroup

import (
	"context"

	"github.com/odigos-io/odigos/api/k8sconsts"
	odigosv1 "github.com/odigos-io/odigos/api/odigos/v1alpha1"
	"github.com/odigos-io/odigos/k8sutils/pkg/env"
	"github.com/odigos-io/odigos/k8sutils/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func newClusterCollectorGroup(namespace string, resourcesSettings *odigosv1.CollectorsGroupResourcesSettings, serviceGraphDisabled *bool, clusterMetricsEnabled *bool) *odigosv1.CollectorsGroup {
	return &odigosv1.CollectorsGroup{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CollectorsGroup",
			APIVersion: "odigos.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      k8sconsts.OdigosClusterCollectorCollectorGroupName,
			Namespace: namespace,
		},
		Spec: odigosv1.CollectorsGroupSpec{
			Role:                    odigosv1.CollectorsGroupRoleClusterGateway,
			CollectorOwnMetricsPort: k8sconsts.OdigosClusterCollectorOwnTelemetryPortDefault,
			ResourcesSettings:       *resourcesSettings,
			ServiceGraphDisabled:    serviceGraphDisabled,
			ClusterMetricsEnabled:   clusterMetricsEnabled,
		},
	}
}

func sync(ctx context.Context, c client.Client) error {

	namespace := env.GetCurrentNamespace()

	odigosConfiguration, err := utils.GetCurrentOdigosConfiguration(ctx, c)
	if err != nil {
		return err
	}
	resourceSettings := getGatewayResourceSettings(&odigosConfiguration)

	// default servicegraph is enabled (serviceGraphDisabled to false)
	serviceGraphDisabled := odigosConfiguration.CollectorGateway.ServiceGraphDisabled
	if serviceGraphDisabled == nil {
		result := false
		serviceGraphDisabled = &result
	}

	// default cluster metrics is disabled (clusterMetricsEnabled to false)
	clusterMetricsEnabled := odigosConfiguration.CollectorGateway.ClusterMetricsEnabled
	if clusterMetricsEnabled == nil {
		result := false
		clusterMetricsEnabled = &result
	}

	// cluster collector is always set and never deleted at the moment.
	// this is to accelerate spinup time and avoid errors while things are gradually being reconciled
	// and started.
	// in the future we might want to support a deployment of instrumentations only and allow user
	// to setup their own collectors, then we would avoid adding the cluster collector by default.
	err = utils.ApplyCollectorGroup(ctx, c, newClusterCollectorGroup(namespace, resourceSettings, serviceGraphDisabled, clusterMetricsEnabled))
	if err != nil {
		return err
	}

	return nil
}
