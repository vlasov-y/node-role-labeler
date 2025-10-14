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
	"os"
	"path"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GetProjectDir", Ordered, func() {
	Context("when getting project directory", func() {
		It("should return a valid path", func() {
			dir, err := GetProjectDir()
			ExpectWithOffset(1, err).ToNot(HaveOccurred())
			info, err := os.Stat(dir)
			ExpectWithOffset(1, err).ToNot(HaveOccurred())
			ExpectWithOffset(2, info.IsDir()).To(BeTrue())
			info, err = os.Stat(path.Join(dir, "go.mod"))
			ExpectWithOffset(1, err).ToNot(HaveOccurred())
			ExpectWithOffset(2, info.Mode().IsRegular()).To(BeTrue())
		})

		It("should return an error if impossible to find", func() {
			cwd, err := os.Getwd()
			ExpectWithOffset(1, err).ToNot(HaveOccurred())
			Expect(os.Chdir("/")).To(Succeed())
			_, err = GetProjectDir()
			ExpectWithOffset(1, err).To(HaveOccurred())
			Expect(os.Chdir(cwd)).To(Succeed())
		})
	})
})
