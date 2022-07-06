// Copyright 2021 The OCGI Authors.
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

package util

import (
	"strconv"

	carrierv1alpha1 "github.com/ocgi/carrier/pkg/apis/carrier/v1alpha1"
)

// Merge helps merge labels or annotations
func Merge(one, two map[string]string) map[string]string {
	three := make(map[string]string)
	for k, v := range one {
		three[k] = v
	}
	for k, v := range two {
		three[k] = v
	}
	return three
}

// GetDesiredReplicasAnnotation returns the number of desired replicas
func GetDesiredReplicasAnnotation(gsSet *carrierv1alpha1.GameServerSet) (int32, bool) {
	return GetIntFromAnnotation(gsSet, DesiredReplicasAnnotation)
}

func GetIntFromAnnotation(gsSet *carrierv1alpha1.GameServerSet, annotationKey string) (int32, bool) {
	annotationValue, ok := gsSet.Annotations[annotationKey]
	if !ok {
		return int32(0), false
	}
	intValue, err := strconv.Atoi(annotationValue)
	if err != nil {
		return int32(0), false
	}
	return int32(intValue), true
}
