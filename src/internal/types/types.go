// Copyright 2025 The Node Role Labeler Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package types defines common data structures and constants used throughout the application.
package types

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	annotationDomain = "node-role-labeler.io"
	// JSON state storage
	AnnotationState = annotationDomain + "/state"
	// Can be used to exclude nodes
	AnnotationEnable = annotationDomain + "/enable"
	// Default custom role label prefix
	CustomRoleLabelPrefix = "node-role.cluster.local/"
)

type Reconciler struct {
	client.Client
	Config                  *rest.Config
	Scheme                  *runtime.Scheme
	Recorder                record.EventRecorder
	MaxConcurrentReconciles int
}

func NewReconciler(mgr manager.Manager) (r Reconciler) {
	return Reconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Config:   mgr.GetConfig(),
		Recorder: mgr.GetEventRecorderFor("NodeRoleLabeler"),
	}
}
