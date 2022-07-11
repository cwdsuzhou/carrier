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
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"

	carrierv1alpha1 "github.com/ocgi/carrier/pkg/apis/carrier/v1alpha1"
	"github.com/ocgi/carrier/pkg/util"
	"github.com/ocgi/carrier/pkg/util/kube"
)

// inplace update GameServer inplace
func (c *Controller) rolloutInplace(squad *carrierv1alpha1.Squad, gsSetList []*carrierv1alpha1.GameServerSet) error {
	newGSSet, isFirstCreate, err := c.findOrCreateGameServerSet(squad, gsSetList)
	if err != nil {
		return err
	}

	var allGSSet []*carrierv1alpha1.GameServerSet
	allGSSet = append(allGSSet, gsSetList...)

	// scaling squad
	scaled, _, err := c.scaleGameServerSetAndRecordEvent(newGSSet, squad.Spec.Replicas, squad)
	if err != nil {
		return err
	}
	if scaled {
		return c.syncRolloutStatus(allGSSet, newGSSet, squad)
	}

	// updating squad
	threshold := InplaceThreshold(*squad)
	if threshold == 0 || isFirstCreate {
		// Do nothing if the threshold is zero
		// or the first creation
		if isFirstCreate {
			if err := c.clearInplaceUpdateStrategy(squad); err != nil {
				klog.ErrorS(err, "clear threshold failed", "name", klog.KObj(squad))
				return err
			}
			allGSSet = append(allGSSet, newGSSet)
		}
		return c.syncRolloutStatus(allGSSet, newGSSet, squad)
	}
	// update GameServerSet
	SetGameServerSetInplaceUpdateAnnotations(newGSSet, squad)
	newGSSet.Spec.Template.Spec.Template.Spec = *squad.Spec.Template.Spec.Template.Spec.DeepCopy()
	SetGameServerTemplateHashLabels(newGSSet)
	_, err = c.gameServerSetGetter.GameServerSets(newGSSet.Namespace).Update(context.TODO(), newGSSet, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	if threshold < squad.Spec.Replicas {
		return c.syncRolloutStatus(allGSSet, newGSSet, squad)
	}
	if SquadComplete(squad, &squad.Status) {
		// We need to update squad first.
		// Fix the wrong status:
		// when the GameServerSet update is successful but squad update fails,
		// the annotations of GameServerSet will be reset in the next synchronization.
		if err := c.clearInplaceUpdateStrategy(squad); err != nil {
			return err
		}
		// Make sure update GameServerSet annotations success or failed after retry.
		if err = wait.PollImmediate(50*time.Millisecond, 1*time.Second, func() (done bool, err error) {
			updateErr := c.cleanupGameServerSet(newGSSet)
			if updateErr == nil {
				return true, nil
			}
			return false, nil
		}); err != nil {
			return err
		}
	}
	// Sync Squad status
	return c.syncRolloutStatus(allGSSet, newGSSet, squad)
}

func (c *Controller) cleanupGameServerSet(gsSet *carrierv1alpha1.GameServerSet) error {
	klog.V(4).InfoS("Cleans up inplace update annotations", "GameServerSet", klog.KObj(gsSet))
	delete(gsSet.Annotations, util.GameServerInPlaceUpdateAnnotation)
	delete(gsSet.Annotations, util.GameServerInPlaceUpdatedReplicasAnnotation)
	_, err := c.gameServerSetGetter.GameServerSets(gsSet.Namespace).Update(context.TODO(), gsSet, metav1.UpdateOptions{})
	return err
}

// clearInplaceUpdateThreshold sets .spec.strategy.inplaceUpdate.threshold to zero and update the input Squad
func (c *Controller) clearInplaceUpdateStrategy(squad *carrierv1alpha1.Squad) error {
	klog.V(4).InfoS("Cleans up threshold", "threshold", squad.Spec.Strategy.InplaceUpdate, "name", klog.KObj(squad))
	squadCopy := squad.DeepCopy()
	threshold := intstr.FromInt(0)
	squadCopy.Spec.Strategy.InplaceUpdate.Threshold = &threshold
	patch, err := kube.CreateMergePatch(squad, squadCopy)
	if err != nil {
		return err
	}
	_, err = c.squadGetter.Squads(squad.Namespace).Patch(context.TODO(), squad.Name, types.MergePatchType, patch, metav1.PatchOptions{})
	return err
}
