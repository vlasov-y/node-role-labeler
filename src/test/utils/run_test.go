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
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Run", Ordered, func() {
	Context("when running a command", func() {
		It("should succeed on a valid command", func() {
			Expect(Run(exec.Command("date"))).To(Succeed())
		})

		It("should return an error for the failed command", func() {
			Expect(Run(exec.Command("exit", "1"))).ToNot(Succeed())
		})
	})
})
