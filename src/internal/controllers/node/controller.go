/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package node

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	. "github.com/vlasov-y/node-role-labeler/internal/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// NodeReconciler reconciles a Node object
type NodeReconciler struct {
	Reconciler
}

// +kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups="",resources=nodes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;get;list;patch;update;watch

func (r *NodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	log := logf.FromContext(ctx)

	node := corev1.Node{}
	if err = r.Get(ctx, req.NamespacedName, &node); err != nil {
		// Object does not exist, ignore the event and return
		if !errors.IsNotFound(err) {
			msg := "cannot get the node"
			log.V(1).Error(err, msg)
		}
		return result, client.IgnoreNotFound(err)
	}
	log = log.WithValues("node", node.Name)

	// Define role label prefixes
	officialRolePrefix := "node-role.kubernetes.io/"
	customRolePrefix := CustomRoleLabelPrefix
	if v := os.Getenv("NODE_ROLE_PREFIX"); v != "" {
		customRolePrefix = v
	}
	if customRolePrefix == officialRolePrefix {
		// Custom role prefix cannot match the official one
		err = fmt.Errorf("%s must not be used as a custom role prefix", officialRolePrefix)
		log.Error(err, "reserved role prefix used")
		return
	}

	// Initialize labels and annotations maps if they are nil
	labels := node.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}
	annotations := node.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	// Handle enable annotation
	if a, exists := annotations[AnnotationEnable]; exists {
		var enabled bool
		if enabled, err = strconv.ParseBool(a); err != nil {
			msg := fmt.Sprintf("failed to parse %s annotation", AnnotationEnable)
			r.Recorder.Eventf(&node, corev1.EventTypeWarning, "EnableAnnotationBroken", msg)
			return
		}

		if !enabled {
			log.V(1).Info("skipping the node because enable=false")
			return
		}
	}

	// Initialize state roles map from the annotation
	stateRoles := map[string]string{}
	if a, exists := annotations[AnnotationState]; exists && a != "" {
		if err = json.Unmarshal([]byte(a), &stateRoles); err != nil {
			msg := "cannot unmarshal node's state annotation"
			r.Recorder.Eventf(&node, corev1.EventTypeWarning, "BadState", "%s: %s", msg, err.Error())
			stateRoles = map[string]string{}
		}
	}

	// Iterate over labels to manage custom and official role labels
	for k := range node.DeepCopy().Labels {
		// Skipping non-role labels
		if !strings.HasPrefix(k, customRolePrefix) && !strings.HasPrefix(k, officialRolePrefix) {
			continue
		}

		// Stripping any role prefix to get role name...
		role := strings.TrimPrefix(strings.TrimPrefix(k, customRolePrefix), officialRolePrefix)
		// ...and building new role labels
		customRole := fmt.Sprintf("%s%s", customRolePrefix, role)
		officialRole := fmt.Sprintf("%s%s", officialRolePrefix, role)

		// Gathering information
		stateRoleValue, stateRoleExists := stateRoles[role]
		customRoleValue, customRoleExists := labels[customRole]
		officialRoleValue, officialRoleExists := labels[officialRole]

		// Processing
		if stateRoleExists {
			// Role was present before...
			if !customRoleExists || !officialRoleExists {
				// ...but it is absent now in at least one of the role labels...
				// ...then this role has to be deleted, since someone...
				// ...has removed one or both role labels, but not the state
				delete(labels, customRole)
				delete(labels, officialRole)
				delete(stateRoles, role)
				msg := fmt.Sprintf("deleted role %s", role)
				log.V(1).Info(msg)
				r.Recorder.Eventf(&node, corev1.EventTypeNormal, "RoleDeleted", msg)
			} else {
				// ...and it is still present in both official and custom labels...
				// ...so we have to compare role label values, maybe someone has changed it
				if stateRoleValue != customRoleValue || stateRoleValue != officialRoleValue {
					// So state role value differs from one of the roles, so either state has been broken...
					stateIsBroken := customRoleValue == officialRoleValue
					if stateIsBroken {
						// ...and we have to repair it...
						stateRoles[role] = customRoleValue
						msg := fmt.Sprintf("repairing state role %s value", role)
						log.V(1).Info(msg)
						r.Recorder.Eventf(&node, corev1.EventTypeWarning, "StateRepair", msg)
					} else {
						// ...or someone changed one of the role's value...
						var newRoleValue, oldRoleValue string
						if stateRoleValue == customRoleValue {
							// ...and official role now holds a new value
							newRoleValue = officialRoleValue
							oldRoleValue = customRoleValue
						} else {
							// ...and custom role now holds a new value
							newRoleValue = customRoleValue
							oldRoleValue = officialRoleValue
						}
						labels[customRole] = newRoleValue
						labels[officialRole] = newRoleValue
						stateRoles[role] = newRoleValue
						// Just cosmetic change for logging
						if newRoleValue == "" {
							newRoleValue = "<empty"
						}
						if oldRoleValue == "" {
							oldRoleValue = "<empty"
						}
						msg := fmt.Sprintf("changed role %s value from %s to %s", role, oldRoleValue, newRoleValue)
						log.V(1).Info(msg)
						r.Recorder.Eventf(&node, corev1.EventTypeNormal, "RoleChanged", msg)
					}
				}
			}
		} else {
			// We did not have the role in the state, so we have to add it...
			var newRoleValue string
			if officialRoleExists {
				newRoleValue = officialRoleValue
			} else {
				newRoleValue = customRoleValue
			}
			labels[customRole] = newRoleValue
			labels[officialRole] = newRoleValue
			stateRoles[role] = newRoleValue
			msg := fmt.Sprintf("role %s=%s is added", role, newRoleValue)
			if newRoleValue == "" {
				msg = fmt.Sprintf("role %s is added", role)
			}
			log.V(1).Info(msg)
			r.Recorder.Eventf(&node, corev1.EventTypeNormal, "RoleAdded", msg)
		}
	}

	// Removing roles that are listed in the state only
	for role := range stateRoles {
		_, customRoleExists := labels[fmt.Sprintf("%s%s", customRolePrefix, role)]
		_, officialRoleExists := labels[fmt.Sprintf("%s%s", officialRolePrefix, role)]
		if !customRoleExists && !officialRoleExists {
			delete(stateRoles, role)
		}
	}

	// Saving new state
	var stateMarshaled []byte
	if stateMarshaled, err = json.Marshal(stateRoles); err != nil {
		log.Error(err, "failed to marshal the state")
		return
	}
	annotations[AnnotationState] = string(stateMarshaled)
	node.SetLabels(labels)
	node.SetAnnotations(annotations)

	if err = r.Update(ctx, &node); err != nil {
		if strings.Contains(err.Error(), "please apply your changes to the latest version and try again") {
			err = nil
			log.V(1).Info("requeue because of the update conflict")
			return ctrl.Result{RequeueAfter: time.Second * 5}, err
		}
		log.Error(err, "failed to update the node object")
		return
	}

	return
}

// SetupWithManager sets up the controller with the Manager.
func (r *NodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Node{}).
		Named("node").
		Complete(r)
}
