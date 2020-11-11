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

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	time "time"

	carrierv1alpha1 "github.com/ocgi/carrier/pkg/apis/carrier/v1alpha1"
	versioned "github.com/ocgi/carrier/pkg/client/clientset/versioned"
	internalinterfaces "github.com/ocgi/carrier/pkg/client/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/ocgi/carrier/pkg/client/listers/carrier/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// WebhookConfigurationInformer provides access to a shared informer and lister for
// WebhookConfigurations.
type WebhookConfigurationInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.WebhookConfigurationLister
}

type webhookConfigurationInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewWebhookConfigurationInformer constructs a new informer for WebhookConfiguration type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewWebhookConfigurationInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredWebhookConfigurationInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredWebhookConfigurationInformer constructs a new informer for WebhookConfiguration type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredWebhookConfigurationInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.CarrierV1alpha1().WebhookConfigurations(namespace).List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.CarrierV1alpha1().WebhookConfigurations(namespace).Watch(options)
			},
		},
		&carrierv1alpha1.WebhookConfiguration{},
		resyncPeriod,
		indexers,
	)
}

func (f *webhookConfigurationInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredWebhookConfigurationInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *webhookConfigurationInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&carrierv1alpha1.WebhookConfiguration{}, f.defaultInformer)
}

func (f *webhookConfigurationInformer) Lister() v1alpha1.WebhookConfigurationLister {
	return v1alpha1.NewWebhookConfigurationLister(f.Informer().GetIndexer())
}
