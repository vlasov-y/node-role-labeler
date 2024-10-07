/*
Copyright 2024.

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

package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/vlasov-y/node-role-labeler/internal/controller/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// NodeReconciler reconciles a Node object
type NodeReconciler struct {
	client.Client
	Config   *rest.Config
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups="",resources=nodes/status,verbs=get;update;patch

func (r *NodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	log := log.FromContext(ctx)
	node := corev1.Node{}
	if err = r.Get(ctx, req.NamespacedName, &node); err != nil {
		// Object does not exist, ignore the event and return
		if !errors.IsNotFound(err) {
			msg := "cannot get the node"
			log.V(1).Error(err, msg)
			r.Recorder.Eventf(&node, corev1.EventTypeWarning, "GetNodeFailed", "%s: %s", msg, err.Error())
		}
		return result, client.IgnoreNotFound(err)
	}
	log = log.WithValues("node", node.Name)

	// Define role prefixes
	officialRolePrefix := "node-role.kubernetes.io/"
	customRolePrefix := "node-role.cluster.local/"
	if v := os.Getenv("NODE_ROLE_PREFIX"); v != "" {
		customRolePrefix = v
	}
	if customRolePrefix == officialRolePrefix {
		// Error: Custom role prefix matches the official one
		msg := "custom node role prefix cannot match official node-role.kubernetes.io"
		log.V(1).Error(fmt.Errorf(msg), msg)
		r.Recorder.Eventf(&node, corev1.EventTypeWarning, "OperatorMisconfigured", "check operator logs")
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

	// Initialize state roles map from the annotation
	stateAnnotation := "node-role-labeler.io/state"
	stateRoles := map[string]string{}
	if a, exists := annotations[stateAnnotation]; exists && a != "" {
		if err = json.Unmarshal([]byte(a), &stateRoles); err != nil {
			// Error: Cannot unmarshal the state annotation
			msg := "cannot unmarshal node's labels state annotation"
			log.V(1).Error(err, msg)
			r.Recorder.Eventf(&node, corev1.EventTypeWarning, "StateUnmarshalFailed", "%s: %s", msg, err.Error())
		}
	}

	// If the state roles map is empty, initialize it with custom and official role labels from node's labels
	if len(stateRoles) == 0 {
		for k, v := range labels {
			role := strings.TrimPrefix(strings.TrimPrefix(k, customRolePrefix), officialRolePrefix)
			customRole := fmt.Sprintf("%s%s", customRolePrefix, role)
			officialRole := fmt.Sprintf("%s%s", officialRolePrefix, role)
			if strings.HasPrefix(k, customRolePrefix) || strings.HasPrefix(k, officialRolePrefix) {
				stateRoles[role] = v
				labels[customRole] = v
				labels[officialRole] = v
			}
		}
		// Log and record event for initialization
		msg := "initialized the state"
		log.V(1).Info(msg)
		r.Recorder.Eventf(&node, corev1.EventTypeNormal, "Initialization", "created %s annotation", stateAnnotation)
	}

	// Iterate over labels to manage custom and official role labels
	for k, v := range labels {
		role := strings.TrimPrefix(strings.TrimPrefix(k, customRolePrefix), officialRolePrefix)
		customRole := fmt.Sprintf("%s%s", customRolePrefix, role)
		officialRole := fmt.Sprintf("%s%s", officialRolePrefix, role)
		if strings.HasPrefix(k, customRolePrefix) {
			// Custom role label found
			if _, exists := labels[officialRole]; !exists {
				// Official role does not exist
				if _, exists := stateRoles[role]; exists {
					// Official role was deleted
					delete(labels, customRole)
					delete(stateRoles, role)
					msg := fmt.Sprintf("deleted label %s=%s", customRole, v)
					log.V(1).Info(msg)
					r.Recorder.Eventf(&node, corev1.EventTypeNormal, "LabelDeleted", msg)
				} else {
					// Official role does not exist, add it
					labels[officialRole] = v
					stateRoles[role] = v
					msg := fmt.Sprintf("added label %s=%s", officialRole, v)
					log.V(1).Info(msg)
					r.Recorder.Eventf(&node, corev1.EventTypeNormal, "LabelAdded", msg)
				}
			} else {
				// There is a matching official label
				// Checking the state
				if _, exists := stateRoles[role]; !exists {
					stateRoles[role] = v
				}
			}
		} else if strings.HasPrefix(k, officialRolePrefix) {
			// We have found an official role label
			if _, exists := labels[customRole]; !exists {
				// There is no matching custom label
				if _, exists := stateRoles[role]; exists {
					// Custom role has been deleted
					delete(labels, officialRole)
					delete(stateRoles, role)
					msg := fmt.Sprintf("deleted label %s=%s", officialRole, v)
					log.V(1).Info(msg)
					r.Recorder.Eventf(&node, corev1.EventTypeNormal, "LabelDeleted", msg)
				} else {
					// Official role has been created
					labels[customRole] = v
					stateRoles[role] = v
					msg := fmt.Sprintf("added label %s=%s", customRole, v)
					log.V(1).Info(msg)
					r.Recorder.Eventf(&node, corev1.EventTypeNormal, "LabelAdded", msg)
				}
			} else {
				// There is a matching official label
				// Checking the state
				if _, exists := stateRoles[role]; !exists {
					stateRoles[role] = v
				}
			}
		}
	}

	var stateMarshaled []byte
	if stateMarshaled, err = json.Marshal(stateRoles); err != nil {
		msg := "failed to marshal the state"
		log.Error(err, msg)
		r.Recorder.Eventf(&node, corev1.EventTypeWarning, "StateMarshalFailed", "%s", err.Error())
		return
	}
	annotations[stateAnnotation] = string(stateMarshaled)
	node.SetLabels(labels)
	node.SetAnnotations(annotations)

	if err = r.Client.Update(ctx, &node); err != nil {
		msg := "failed to update the node"
		log.Error(err, msg)
		r.Recorder.Eventf(&node, corev1.EventTypeWarning, "NodeUpdateFailed", "%s", err.Error())
		return
	}

	// Update node labels with the managed state
	if err = r.Client.Status().Update(ctx, &node); err != nil {
		msg := "failed to update the node status"
		log.Error(err, msg)
		r.Recorder.Eventf(&node, corev1.EventTypeWarning, "NodeStatusUpdateFailed", "%s", err.Error())
		return
	}

	return
}

// SetupWithManager sets up the controller with the Manager.
func (r *NodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Node{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: 10}).
		WithEventFilter(utils.IgnoreOutOfOrder()).
		WithEventFilter(utils.IgnoreDeletionPredicate()).
		Complete(r)
}
