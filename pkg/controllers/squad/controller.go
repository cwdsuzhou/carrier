// Copyright 2019 The Kubernetes Authors.
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

package squad

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	carrierv1alpha1 "github.com/ocgi/carrier/pkg/apis/carrier/v1alpha1"
	"github.com/ocgi/carrier/pkg/client/clientset/versioned"
	getterv1alpha1 "github.com/ocgi/carrier/pkg/client/clientset/versioned/typed/carrier/v1alpha1"
	"github.com/ocgi/carrier/pkg/client/informers/externalversions"
	listerv1alpha1 "github.com/ocgi/carrier/pkg/client/listers/carrier/v1alpha1"
	"github.com/ocgi/carrier/pkg/util"
)

// Controller is a the GameServerSet controller
type Controller struct {
	crdGetter           v1beta1.CustomResourceDefinitionInterface
	gameServerLister    listerv1alpha1.GameServerLister
	gameServerSetGetter getterv1alpha1.GameServerSetsGetter
	gameServerSetLister listerv1alpha1.GameServerSetLister
	gameServerSetSynced cache.InformerSynced
	squadGetter         getterv1alpha1.SquadsGetter
	squadLister         listerv1alpha1.SquadLister
	squadSynced         cache.InformerSynced
	workerQueue         workqueue.RateLimitingInterface
	recorder            record.EventRecorder
}

// NewController returns a new squads crd controller
func NewController(
	kubeClient kubernetes.Interface,
	carrierClient versioned.Interface,
	carrierInformerFactory externalversions.SharedInformerFactory) *Controller {

	gameServers := carrierInformerFactory.Carrier().V1alpha1().GameServers()
	gameServerSets := carrierInformerFactory.Carrier().V1alpha1().GameServerSets()
	gsSetInformer := gameServerSets.Informer()

	squads := carrierInformerFactory.Carrier().V1alpha1().Squads()
	squadsInformer := squads.Informer()

	c := &Controller{
		gameServerLister:    gameServers.Lister(),
		gameServerSetGetter: carrierClient.CarrierV1alpha1(),
		gameServerSetLister: gameServerSets.Lister(),
		gameServerSetSynced: gsSetInformer.HasSynced,
		squadGetter:         carrierClient.CarrierV1alpha1(),
		squadLister:         squads.Lister(),
		squadSynced:         squadsInformer.HasSynced,
	}
	c.workerQueue = workqueue.NewRateLimitingQueue(
		workqueue.NewItemFastSlowRateLimiter(20*time.Millisecond, 500*time.Millisecond, 5))
	s := scheme.Scheme
	// Register operator types with the runtime scheme.
	s.AddKnownTypes(carrierv1alpha1.SchemeGroupVersion, &carrierv1alpha1.Squad{})
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(
		&typedcorev1.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})
	c.recorder = eventBroadcaster.NewRecorder(s, corev1.EventSource{Component: "squad-controller"})

	squadsInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.enqueueGameSquad,
		UpdateFunc: c.updateGameSquad,
		DeleteFunc: c.deleteGamSquad,
	})

	gsSetInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: c.gameServerSetEventHandler,
		UpdateFunc: func(_, newObj interface{}) {
			c.gameServerSetEventHandler(newObj)
		},
		DeleteFunc: c.deleteGameServerSet,
	})

	return c
}

// Run the Squad controller. Will block until stop is closed.
// Runs threadiness number workers to process the rate limited queue
func (c *Controller) Run(workers int, stop <-chan struct{}) error {
	klog.V(4).Info("Wait for cache sync")
	if !cache.WaitForCacheSync(stop, c.squadSynced, c.gameServerSetSynced) {
		return errors.New("failed to wait for caches to sync")
	}
	for i := 0; i < workers; i++ {
		go wait.Until(c.worker, time.Second, stop)
	}
	<-stop
	return nil
}

// obj could be a Squad, or a DeletionFinalStateUnknown marker item.
func (c *Controller) updateGameSquad(old, cur interface{}) {
	c.enqueueGameSquad(cur)
}

// obj could be a GamServerSet, or a DeletionFinalStateUnknown marker item.
func (c *Controller) enqueueGameSquad(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("couldn't get key for object %+v: %v", obj, err))
		return
	}

	// Requests are always added to queue with resyncPeriod delay.  If there's already
	// request for the Squad in the queue then c new request is always dropped. Requests spend resync
	// interval in queue so Squads are processed every resync interval.
	c.workerQueue.AddRateLimited(key)
}

func (c *Controller) worker() {
	for c.processNextWorkItem() {
	}
	klog.Infof("GameServerSet controller worker shutting down")
}

func (c *Controller) deleteGamSquad(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("couldn't get key for object %+v: %v", obj, err))
		return
	}

	c.workerQueue.Forget(key)
}

func (c *Controller) processNextWorkItem() bool {
	key, quit := c.workerQueue.Get()
	if quit {
		return false
	}
	defer c.workerQueue.Done(key)

	err := c.syncSquad(key.(string))
	if err != nil {
		c.workerQueue.AddRateLimited(key)
		utilruntime.HandleError(err)
		return true
	}
	c.workerQueue.Forget(key)
	return true
}

// syncSquad synchronised the squad CRDs and configures/updates
// backing GameServerSets
func (c *Controller) syncSquad(key string) error {
	startTime := time.Now()
	klog.V(4).InfoS("Started syncing Squaq", "name", key, "start", startTime)
	defer func() {
		klog.V(4).InfoS("Finished syncing Squad", "name", key, "cost", time.Since(startTime))
	}()

	// // Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		// don't return an error, as we don't want this retried
		runtime.HandleError(errors.Wrapf(err, "invalid resource key: %s", key))
		return nil
	}

	squad, err := c.squadLister.Squads(namespace).Get(name)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			klog.Info("Squad is no longer available for syncing")
			return nil
		}
		return errors.Wrapf(err, "error retrieving squad %s from namespace %s", name, namespace)
	}

	// TODO
	// ensureDefaults setting default value for squad.
	// Remove this when webhook are supported.
	c.ensureDefaults(squad)

	gsSetList, err := c.listGameServerSetsByOwner(squad)
	if err != nil {
		return err
	}
	// List all GameServers owned by this Squad
	gsMap, err := c.getGameServerMapForSquad(squad, gsSetList)
	if err != nil {
		return err
	}

	if squad.DeletionTimestamp != nil {
		return c.syncStatusOnly(squad, gsSetList)
	}

	if err = c.checkPausedConditions(squad); err != nil {
		return err
	}

	if squad.Spec.Paused {
		return c.sync(squad, gsSetList)
	}

	if getRollbackTo(squad) != nil {
		return c.rollback(squad, gsSetList)
	}

	scalingEvent, err := c.isScalingEvent(squad, gsSetList)
	if err != nil {
		return err
	}
	if scalingEvent {
		klog.V(4).InfoS("Scaling squad", "name", klog.KObj(squad))
		return c.sync(squad, gsSetList)
	}

	klog.V(4).InfoS("Reconciling squad", "name", klog.KObj(squad))

	switch squad.Spec.Strategy.Type {
	case carrierv1alpha1.RecreateSquadStrategyType:
		return c.rolloutRecreate(squad, gsSetList, gsMap)
	case carrierv1alpha1.RollingUpdateSquadStrategyType:
		return c.rolloutRolling(squad, gsSetList)
	case carrierv1alpha1.CanaryUpdateSquadStrategyType:
		return c.rolloutCanary(squad, gsSetList)
	case carrierv1alpha1.InplaceUpdateSquadStrategyType:
		return c.rolloutInplace(squad, gsSetList)
	}
	return errors.Errorf("unexpected squad strategy type: %s", squad.Spec.Strategy.Type)
}

// getGameServerMapForSquad returns the GameServers managed by a Squad.
func (c *Controller) getGameServerMapForSquad(
	squad *carrierv1alpha1.Squad,
	gsSetList []*carrierv1alpha1.GameServerSet) (map[types.UID][]*carrierv1alpha1.GameServer, error) {
	// Get all GameServer that potentially belong to this Squad.
	gsLst, err := c.listGameServersBySquadOwner(squad)
	if err != nil {
		return nil, err
	}
	// Group GameServers by their controller (if it's in gsSetList).
	gsMap := make(map[types.UID][]*carrierv1alpha1.GameServer, len(gsSetList))
	for _, gsSet := range gsSetList {
		gsMap[gsSet.UID] = []*carrierv1alpha1.GameServer{}
	}
	for _, gs := range gsLst {
		controllerRef := metav1.GetControllerOf(gs)
		if controllerRef == nil {
			continue
		}
		// Only append if we care about this UID.
		if _, ok := gsMap[controllerRef.UID]; ok {
			gsMap[controllerRef.UID] = append(gsMap[controllerRef.UID], gs)
		}
	}
	return gsMap, nil
}

// gameServerSetEventHandler enqueues the owning Squad for this GameServerSet,
// assuming that it has one
func (c *Controller) gameServerSetEventHandler(obj interface{}) {
	gsSet := obj.(*carrierv1alpha1.GameServerSet)
	if gsSet.DeletionTimestamp != nil {
		// On a restart of the controller manager, it's possible for an object to
		// show up in a state that is already pending deletion.
		c.deleteGameServerSet(gsSet)
		return
	}
	ref := metav1.GetControllerOf(gsSet)
	if ref == nil {
		return
	}
	squad := c.resolveControllerRef(gsSet.Namespace, ref)
	if squad == nil {
		return
	}
	c.enqueueGameSquad(squad)
}

// deleteGameServerSet enqueues the squad that manages a GameServerSet when
// the GameServerSet is deleted.
func (c *Controller) deleteGameServerSet(obj interface{}) {
	gsSet, ok := obj.(*carrierv1alpha1.GameServerSet)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("Couldn't get object from tombstone %#v", obj))
			return
		}
		gsSet, ok = tombstone.Obj.(*carrierv1alpha1.GameServerSet)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("Tombstone contained object that is not a GameServerSet %#v", obj))
			return
		}
	}

	controllerRef := metav1.GetControllerOf(gsSet)
	if controllerRef == nil {
		// No controller should care about orphans being deleted.
		return
	}
	squad := c.resolveControllerRef(gsSet.Namespace, controllerRef)
	if squad == nil {
		return
	}
	klog.V(4).InfoS("GameServerSet deleted.", "name", klog.KObj(gsSet))
	c.enqueueGameSquad(squad)
}

// resolveControllerRef returns the controller referenced by a ControllerRef,
// or nil if the ControllerRef could not be resolved to a matching controller
// of the correct Kind.
func (c *Controller) resolveControllerRef(namespace string, controllerRef *metav1.OwnerReference) *carrierv1alpha1.Squad {
	// We can't look up by UID, so look up by Name and then verify UID.
	// Don't even try to look up by Name if it's the wrong Kind.
	if controllerRef.Kind != controllerKind.Kind {
		return nil
	}
	squad, err := c.squadLister.Squads(namespace).Get(controllerRef.Name)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			klog.Warning("Owner Squad no longer available for syncing")
		} else {
			runtime.HandleError(errors.Wrapf(err, "error retrieving GameServerSet owner, ref: %v", controllerRef))
		}
		return nil
	}
	if squad.UID != controllerRef.UID {
		// The controller we found with this Name is not the same one that the
		// ControllerRef points to.
		return nil
	}
	return squad
}

func (c *Controller) ensureDefaults(squad *carrierv1alpha1.Squad) {
	// setting revision history limit
	if squad.Spec.RevisionHistoryLimit == nil {
		squad.Spec.RevisionHistoryLimit = new(int32)
		*squad.Spec.RevisionHistoryLimit = 10
	}
	// setting selector
	if squad.Spec.Selector == nil {
		squad.Spec.Selector = &metav1.LabelSelector{}
	}
	if squad.Spec.Selector.MatchLabels == nil {
		squad.Spec.Selector.MatchLabels = map[string]string{
			util.SquadNameLabelKey: squad.Name,
		}
	}
	// setting update strategy
	strategy := &squad.Spec.Strategy
	// Set default carrierv1alpha1.RollingUpdateSquadStrategyType as RollingUpdate.
	if strategy.Type == "" {
		strategy.Type = carrierv1alpha1.RollingUpdateSquadStrategyType
	}
	if strategy.Type == carrierv1alpha1.RollingUpdateSquadStrategyType {
		if strategy.RollingUpdate == nil {
			rollingUpdate := carrierv1alpha1.RollingUpdateSquad{}
			strategy.RollingUpdate = &rollingUpdate
		}
		if strategy.RollingUpdate.MaxUnavailable == nil {
			// Set default MaxUnavailable as 25% by default.
			maxUnavailable := intstr.FromString("25%")
			strategy.RollingUpdate.MaxUnavailable = &maxUnavailable
		}
		if strategy.RollingUpdate.MaxSurge == nil {
			// Set default MaxSurge as 25% by default.
			maxSurge := intstr.FromString("25%")
			strategy.RollingUpdate.MaxSurge = &maxSurge
		}
	}
}
