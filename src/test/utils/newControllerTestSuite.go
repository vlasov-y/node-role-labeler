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

package utils

import (
	"context"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

type ControllerTestSuite struct {
	Client   client.Client
	Config   *rest.Config
	Ctx      context.Context
	Cancel   context.CancelFunc
	Recorder *record.FakeRecorder
	TestEnv  *envtest.Environment
}

func NewControllerTestSuite() (suite *ControllerTestSuite) {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	scheme := runtime.NewScheme()
	appsv1.AddToScheme(scheme)
	batchv1.AddToScheme(scheme)
	corev1.AddToScheme(scheme)

	recorder := record.NewFakeRecorder(5)
	ctx, cancel := context.WithCancel(context.Background())

	testEnv := &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crds")},
		ErrorIfCRDPathMissing: false,
	}

	cmd := exec.Command("task", "tools:setup-envtest")
	err := Run(cmd)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	cwd, err := GetProjectDir()
	ExpectWithOffset(1, err).ToNot(HaveOccurred())

	data, err := os.ReadFile(path.Join(cwd, ".task/envtest-path"))
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	testEnv.BinaryAssetsDirectory = strings.TrimSpace(string(data))

	rc, err := testEnv.Start()
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	ExpectWithOffset(2, rc).NotTo(BeNil())

	c, err := client.New(rc, client.Options{Scheme: scheme})
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	ExpectWithOffset(2, c).NotTo(BeNil())

	return &ControllerTestSuite{
		Client:   c,
		Config:   rc,
		Ctx:      ctx,
		Cancel:   cancel,
		Recorder: recorder,
		TestEnv:  testEnv,
	}
}

func (ts *ControllerTestSuite) Teardown() {
	ts.Cancel()
	err := ts.TestEnv.Stop()
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
}
