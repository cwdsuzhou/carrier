// Copyright 2021 THL A29 Limited, a Tencent company.
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

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/ocgi/carrier/pkg/apis/carrier/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// WebhookConfigurationLister helps list WebhookConfigurations.
type WebhookConfigurationLister interface {
	// List lists all WebhookConfigurations in the indexer.
	List(selector labels.Selector) (ret []*v1alpha1.WebhookConfiguration, err error)
	// WebhookConfigurations returns an object that can list and get WebhookConfigurations.
	WebhookConfigurations(namespace string) WebhookConfigurationNamespaceLister
	WebhookConfigurationListerExpansion
}

// webhookConfigurationLister implements the WebhookConfigurationLister interface.
type webhookConfigurationLister struct {
	indexer cache.Indexer
}

// NewWebhookConfigurationLister returns a new WebhookConfigurationLister.
func NewWebhookConfigurationLister(indexer cache.Indexer) WebhookConfigurationLister {
	return &webhookConfigurationLister{indexer: indexer}
}

// List lists all WebhookConfigurations in the indexer.
func (s *webhookConfigurationLister) List(selector labels.Selector) (ret []*v1alpha1.WebhookConfiguration, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.WebhookConfiguration))
	})
	return ret, err
}

// WebhookConfigurations returns an object that can list and get WebhookConfigurations.
func (s *webhookConfigurationLister) WebhookConfigurations(namespace string) WebhookConfigurationNamespaceLister {
	return webhookConfigurationNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// WebhookConfigurationNamespaceLister helps list and get WebhookConfigurations.
type WebhookConfigurationNamespaceLister interface {
	// List lists all WebhookConfigurations in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1alpha1.WebhookConfiguration, err error)
	// Get retrieves the WebhookConfiguration from the indexer for a given namespace and name.
	Get(name string) (*v1alpha1.WebhookConfiguration, error)
	WebhookConfigurationNamespaceListerExpansion
}

// webhookConfigurationNamespaceLister implements the WebhookConfigurationNamespaceLister
// interface.
type webhookConfigurationNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all WebhookConfigurations in the indexer for a given namespace.
func (s webhookConfigurationNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.WebhookConfiguration, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.WebhookConfiguration))
	})
	return ret, err
}

// Get retrieves the WebhookConfiguration from the indexer for a given namespace and name.
func (s webhookConfigurationNamespaceLister) Get(name string) (*v1alpha1.WebhookConfiguration, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("webhookconfiguration"), name)
	}
	return obj.(*v1alpha1.WebhookConfiguration), nil
}
