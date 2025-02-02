package collector

import (
	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/collector/fluentd"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/tls"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ReconcileCollectorConfig reconciles a collector config specifically for the collector defined by the factory
func (f *Factory) ReconcileCollectorConfig(er record.EventRecorder, k8sClient client.Client, namespace, name, collectorConfig string, owner metav1.OwnerReference) error {
	log.V(3).Info("Updating ConfigMap and Secrets")
	opensslConf := tls.OpenSSLConf(k8sClient)
	if f.CollectorType == logging.LogCollectionTypeFluentd {
		collectorConfigMap := runtime.NewConfigMap(
			namespace,
			name,
			map[string]string{
				"fluent.conf":         collectorConfig,
				"run.sh":              fluentd.RunScript,
				"cleanInValidJson.rb": fluentd.CleanInValidJson,
				"openssl.cnf":         opensslConf,
			},
		)
		utils.AddOwnerRefToObject(collectorConfigMap, owner)
		return reconcile.Configmap(k8sClient, k8sClient, collectorConfigMap)
	} else if f.CollectorType == logging.LogCollectionTypeVector {
		secret := runtime.NewSecret(
			namespace,
			constants.CollectorConfigSecretName,
			map[string][]byte{
				"vector.toml": []byte(collectorConfig),
				"openssl.cnf": []byte(opensslConf),
			})
		return reconcile.Secret(er, k8sClient, secret)
	}

	return nil
}
