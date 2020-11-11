// Copyright 2020 THL A29 Limited, a Tencent company.
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

package crd

import (
	"time"

	apiv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	extv1beta1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
)

// WaitForEstablishedCRD blocks until CRD comes to an Established state.
// Has a deadline of 60 seconds for this to occur.
func WaitForEstablishedCRD(crdGetter extv1beta1.CustomResourceDefinitionInterface, name string) error {
	return wait.PollImmediate(time.Second, 60*time.Second, func() (done bool, err error) {
		crd, err := crdGetter.Get(name, metav1.GetOptions{})
		if err != nil {
			return false, nil
		}

		for _, cond := range crd.Status.Conditions {
			if cond.Type != apiv1beta1.Established {
				continue
			}
			if cond.Status != apiv1beta1.ConditionTrue {
				continue
			}
			klog.Infof("custom resource definition:%v established", crd.Name)
			return true, err
		}

		return false, nil
	})
}
