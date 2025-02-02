package reconcile

import (
	"context"
	"fmt"
	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Secret reconciles a Secret to the desired spec returning an error
// if there is an issue creating or updating to the desired state
func Secret(er record.EventRecorder, k8Client client.Client, desired *corev1.Secret) error {
	reason := constants.EventReasonGetObject
	updateReason := ""
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		current := &corev1.Secret{}
		key := client.ObjectKeyFromObject(desired)
		if err := k8Client.Get(context.TODO(), key, current); err != nil {
			if errors.IsNotFound(err) {
				reason = constants.EventReasonCreateObject
				return k8Client.Create(context.TODO(), desired)
			}
			return fmt.Errorf("failed to get %v Secret: %w", key, err)
		}
		if reflect.DeepEqual(current.Data, desired.Data) {
			// identical; no need to update.
			log.V(3).Info("Secret are the same skipping update")
			return nil
		}
		current.Data = desired.Data
		reason = constants.EventReasonUpdateObject
		return k8Client.Update(context.TODO(), current)
	})

	eventType := corev1.EventTypeNormal
	msg := fmt.Sprintf("%s Secret %s/%s", reason, desired.Namespace, desired.Name)
	if updateReason != "" {
		msg = fmt.Sprintf("%s because of change in %s.", msg, updateReason)
	}
	if retryErr != nil {
		eventType = corev1.EventTypeWarning
		msg = fmt.Sprintf("Unable to %s: %v", msg, retryErr)
	}
	er.Event(desired, eventType, reason, msg)
	return retryErr
}
