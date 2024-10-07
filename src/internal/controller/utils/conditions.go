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

package utils

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConditionNotFound is a custom error type indicating that a condition was not found
type ConditionNotFound struct {
	ConditionType string
}

// Error returns the error message for ConditionNotFound
func (e *ConditionNotFound) Error() string {
	return fmt.Sprintf("Condition '%s' not found", e.ConditionType)
}

// Get condition status string from bool
func ToConditionStatus(b *bool) metav1.ConditionStatus {
	if b == nil {
		return metav1.ConditionUnknown
	}
	if *b {
		return metav1.ConditionTrue
	} else {
		return metav1.ConditionFalse
	}
}

// Adds condition if it did not exist
func AddConditions(cs []metav1.Condition, c ...metav1.Condition) *[]metav1.Condition {
	var exists bool
	for _, new := range c {
		exists = false
		for _, old := range cs {
			if old.Type == new.Type {
				exists = true
				break
			}
		}
		if !exists {
			cs = append(cs, new)
		}
	}
	return &cs
}

// Adds condition or replace existing one and returns new list of conditions
func AddOrUpdateConditions(cs []metav1.Condition, c ...metav1.Condition) *[]metav1.Condition {
	var updated bool
	for _, new := range c {
		updated = false
		for i, old := range cs {
			if old.Type == new.Type {
				cs[i] = new
				updated = true
				break
			}
		}
		if !updated {
			cs = append(cs, new)
		}
	}
	return &cs
}

// Remove condition from the list of conditions and returns new list
func RemoveCondition(cs []metav1.Condition, c metav1.Condition) *[]metav1.Condition {
	var updatedConditions []metav1.Condition
	for _, condition := range cs {
		if condition.Type != c.Type {
			updatedConditions = append(updatedConditions, condition)
		}
	}
	cs = updatedConditions
	return &cs
}

// Find condition in the list by its type and returns it
// If not found, returns ConditionNotFound error
func GetConditionByType(cs *[]metav1.Condition, t string) (*metav1.Condition, error) {
	for _, condition := range *cs {
		if condition.Type == t {
			return &condition, nil
		}
	}
	return nil, &ConditionNotFound{
		ConditionType: t,
	}
}
