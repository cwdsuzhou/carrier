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

package gameservers

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	"github.com/ocgi/carrier/pkg/apis/carrier"
	carrierv1alpha1 "github.com/ocgi/carrier/pkg/apis/carrier/v1alpha1"
	"github.com/ocgi/carrier/pkg/util"
)

const (
	// ToBeDeletedTaint is a taint used to make the node unschedulable.
	ToBeDeletedTaint = "ToBeDeletedByClusterAutoscaler"
)

// ApplyDefaults applies default values to the GameServer if they are not already populated
func ApplyDefaults(gs *carrierv1alpha1.GameServer) {
	if gs.Annotations == nil {
		gs.Annotations = map[string]string{}
	}
	gs.Annotations[carrier.GroupName] = carrierv1alpha1.SchemeGroupVersion.String()
	gs.Finalizers = append(gs.Finalizers, carrier.GroupName)

	applySpecDefaults(gs)
}

// ApplyDefaults applies default values to the GameServerSpec if they are not already populated
func applySpecDefaults(gs *carrierv1alpha1.GameServer) {
	gss := &gs.Spec
	applyPortDefaults(gss)
	applySchedulingDefaults(gss)
}

// applyPortDefaults applies default values for all ports
func applyPortDefaults(gss *carrierv1alpha1.GameServerSpec) {
	for i, p := range gss.Ports {
		// basic spec
		if p.PortPolicy == "" {
			gss.Ports[i].PortPolicy = carrierv1alpha1.LoadBalancer
		}

		if p.Protocol == "" {
			gss.Ports[i].Protocol = "UDP"
		}
	}
}

// applySchedulingDefaults set default `MostAllocated`
func applySchedulingDefaults(gss *carrierv1alpha1.GameServerSpec) {
	if gss.Scheduling == "" {
		gss.Scheduling = carrierv1alpha1.MostAllocated
	}
}

// IsDeletable returns false if the server is currently not deletable
func IsDeletable(gs *carrierv1alpha1.GameServer) bool {
	if IsInPlaceUpdating(gs) {
		return false
	}
	return deleteReady(gs)
}

// IsDeletableWithGates returns false if the server is currently not deletable and has deletableGates
func IsDeletableWithGates(gs *carrierv1alpha1.GameServer) bool {
	return len(gs.Spec.DeletableGates) != 0 && IsDeletable(gs)
}

// deleteReady checks if deletable gates in condition are all `True`
func deleteReady(gs *carrierv1alpha1.GameServer) bool {
	condMap := make(map[string]carrierv1alpha1.ConditionStatus, len(gs.Status.Conditions))
	for _, condition := range gs.Status.Conditions {
		condMap[string(condition.Type)] = condition.Status
	}
	for _, gate := range gs.Spec.DeletableGates {
		if v, ok := condMap[gate]; !ok || v != carrierv1alpha1.ConditionTrue {
			return false
		}
	}
	return true
}

// IsDeletableExist checks if deletable gates exits
func IsDeletableExist(gs *carrierv1alpha1.GameServer) bool {
	return len(gs.Spec.DeletableGates) != 0
}

// IsReadinessExist checks if readiness gates exits
func IsReadinessExist(gs *carrierv1alpha1.GameServer) bool {
	return len(gs.Spec.ReadinessGates) != 0
}

// IsBeingDeleted returns true if the server is in the process of being deleted.
func IsBeingDeleted(gs *carrierv1alpha1.GameServer) bool {
	return !gs.DeletionTimestamp.IsZero() || gs.Status.State == carrierv1alpha1.GameServerFailed ||
		gs.Status.State == carrierv1alpha1.GameServerExited
}

// IsStopped returns true if the server is failed or exited
func IsStopped(gs *carrierv1alpha1.GameServer) bool {
	return gs.Status.State == carrierv1alpha1.GameServerFailed ||
		gs.Status.State == carrierv1alpha1.GameServerExited
}

// IsBeforeRunning returns if GameServer is not running.
func IsBeforeRunning(gs *carrierv1alpha1.GameServer) bool {
	if gs.Status.State == "" || gs.Status.State == carrierv1alpha1.GameServerUnknown ||
		gs.Status.State == carrierv1alpha1.GameServerStarting {
		return true
	}
	return false
}

// IsReady returns true if the GameServer Status Condition are all OK
func IsReady(gs *carrierv1alpha1.GameServer) bool {
	condMap := make(map[string]carrierv1alpha1.ConditionStatus, len(gs.Status.Conditions))
	for _, condition := range gs.Status.Conditions {
		condMap[string(condition.Type)] = condition.Status
	}
	for _, gate := range gs.Spec.ReadinessGates {
		if v, ok := condMap[gate]; !ok || v != carrierv1alpha1.ConditionTrue {
			return false
		}
	}
	return true
}

// IsOutOfService checks if a GameServer is marked out of service, and a delete candidate
func IsOutOfService(gs *carrierv1alpha1.GameServer) bool {
	for _, constraint := range gs.Spec.Constraints {
		if constraint.Type != carrierv1alpha1.NotInService {
			continue
		}
		if *constraint.Effective == true {
			return true
		}
	}
	return false
}

// IsInPlaceUpdating checks if a GameServer is inplace updating
func IsInPlaceUpdating(gs *carrierv1alpha1.GameServer) bool {
	if len(gs.Annotations) == 0 {
		return false
	}
	return gs.Annotations[util.GameServerInPlaceUpdatingAnnotation] == "true"
}

// IsDynamicPortAllocated checks if ports allocated
func IsDynamicPortAllocated(gs *carrierv1alpha1.GameServer) bool {
	if len(gs.Annotations) == 0 {
		return false
	}
	return gs.Annotations[util.GameServerDynamicPortAllocated] == "true"
}

// CanInPlaceUpdating checks if a GameServer can inplace updating
func CanInPlaceUpdating(gs *carrierv1alpha1.GameServer) bool {
	if IsBeingDeleted(gs) {
		return false
	}
	if IsBeforeRunning(gs) {
		return true
	}
	return IsInPlaceUpdating(gs) && deleteReady(gs)
}

// SetInPlaceUpdatingStatus set if it is inplace updating
func SetInPlaceUpdatingStatus(gs *carrierv1alpha1.GameServer, status string) {
	if gs.Annotations == nil {
		gs.Annotations = make(map[string]string)
	}
	gs.Annotations[util.GameServerInPlaceUpdatingAnnotation] = status
}

// buildPod build pod according to GameServerSpec
func buildPod(gs *carrierv1alpha1.GameServer) (*corev1.Pod, error) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        gs.Name,
			Namespace:   gs.Namespace,
			Labels:      gs.Spec.Template.Labels,
			Annotations: gs.Spec.Template.Annotations,
		},
		Spec: *gs.Spec.Template.Spec.DeepCopy(),
	}

	podObjectMeta(gs, pod)
	if len(findPorts(gs)) > 0 {
		i, gsContainer, err := FindContainer(&gs.Spec, util.GameServerContainerName)
		if err != nil {
			return pod, err
		}
		klog.V(5).InfoS("Found desired container", "index", i, "GameServer", klog.KObj(gs))
		for _, p := range gs.Spec.Ports {
			if p.ContainerPort != nil {
				cp := corev1.ContainerPort{
					ContainerPort: *p.ContainerPort,
					HostPort:      *p.HostPort,
				}
				if p.Protocol == "TCPUDP" {
					cp.Protocol = corev1.ProtocolTCP
					cpUTP := cp.DeepCopy()
					cpUTP.Protocol = corev1.ProtocolUDP
					gsContainer.Ports = append(gsContainer.Ports, *cpUTP)
				} else {
					cp.Protocol = p.Protocol
				}
				gsContainer.Ports = append(gsContainer.Ports, cp)
			}
			if p.ContainerPortRange != nil && p.HostPortRange != nil {
				for idx := p.ContainerPortRange.MinPort; idx <= p.ContainerPortRange.MaxPort; idx++ {
					cp := corev1.ContainerPort{
						ContainerPort: idx,
						HostPort:      p.HostPortRange.MinPort + (p.HostPortRange.MinPort - idx),
					}
					if p.Protocol == "TCPUDP" {
						cp.Protocol = corev1.ProtocolTCP
						cpUTP := cp.DeepCopy()
						cpUTP.Protocol = corev1.ProtocolUDP
						gsContainer.Ports = append(gsContainer.Ports, *cpUTP)
					} else {
						cp.Protocol = p.Protocol
					}
					gsContainer.Ports = append(gsContainer.Ports, cp)
				}
			}
		}
		klog.V(5).InfoS("Final desired container", "container", gsContainer.String())
		pod.Spec.Containers[i] = gsContainer
	}

	if pod.Annotations == nil {
		pod.Annotations = map[string]string{}
	}
	if pod.Labels == nil {
		pod.Labels = map[string]string{}
	}
	injectPodScheduling(gs, pod)
	injectPodTolerations(pod)
	return pod, nil
}

// podObjectMeta configures the pod ObjectMeta details
func podObjectMeta(gs *carrierv1alpha1.GameServer, pod *corev1.Pod) {
	pod.Labels = util.Merge(gs.Labels, pod.Labels)
	pod.Annotations = util.Merge(gs.Annotations, pod.Annotations)
	pod.Labels[util.RoleLabelKey] = util.GameServerLabelRoleValue
	pod.Labels[util.GameServerPodLabelKey] = gs.Name
	ref := metav1.NewControllerRef(gs, carrierv1alpha1.SchemeGroupVersion.WithKind("GameServer"))
	pod.OwnerReferences = append(pod.OwnerReferences, *ref)
	// Add Carrier version into Pod Annotations
	pod.Annotations[carrier.GroupName] = carrierv1alpha1.SchemeGroupVersion.Version
}

// injectPodScheduling helps inject podAffinity/PodAntiAffinity to podSpec if the policy is `Most/LeastAllocated`
func injectPodScheduling(gs *carrierv1alpha1.GameServer, pod *corev1.Pod) {
	if gs.Spec.Scheduling == carrierv1alpha1.Default {
		return
	}

	if gs.Spec.Scheduling == carrierv1alpha1.LeastAllocated {
		if pod.Spec.Affinity == nil {
			pod.Spec.Affinity = &corev1.Affinity{}
		}
		if pod.Spec.Affinity.PodAntiAffinity == nil {
			pod.Spec.Affinity.PodAntiAffinity = &corev1.PodAntiAffinity{}
		}
		antiAffExection := pod.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution
		term := corev1.WeightedPodAffinityTerm{
			Weight: 100,
			PodAffinityTerm: corev1.PodAffinityTerm{
				TopologyKey: "kubernetes.io/hostname",
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{util.RoleLabelKey: util.GameServerLabelRoleValue},
				},
			},
		}

		pod.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(antiAffExection,
			term)
	}
	if gs.Spec.Scheduling == carrierv1alpha1.MostAllocated {
		if pod.Spec.Affinity == nil {
			pod.Spec.Affinity = &corev1.Affinity{}
		}
		if pod.Spec.Affinity.PodAffinity == nil {
			pod.Spec.Affinity.PodAffinity = &corev1.PodAffinity{}
		}

		affExection := pod.Spec.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution
		term := corev1.WeightedPodAffinityTerm{
			Weight: 100,
			PodAffinityTerm: corev1.PodAffinityTerm{
				TopologyKey: "kubernetes.io/hostname",
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{util.RoleLabelKey: util.GameServerLabelRoleValue},
				},
			},
		}

		pod.Spec.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(affExection, term)
		return
	}
}

// injectPodTolerations helps add tolerations to pod.
// tolerate: NotReady、Unreachable
func injectPodTolerations(pod *corev1.Pod) {
	pod.Spec.Tolerations = append(pod.Spec.Tolerations, []corev1.Toleration{
		{
			Key:      corev1.TaintNodeNotReady,
			Operator: corev1.TolerationOpExists,
			Effect:   corev1.TaintEffectNoExecute,
		},
		{
			Key:      corev1.TaintNodeUnreachable,
			Operator: corev1.TolerationOpExists,
			Effect:   corev1.TaintEffectNoExecute,
		}}...)
}

// isGameServerPod returns if this Pod is a Pod that comes from a GameServer
func isGameServerPod(pod *corev1.Pod) bool {
	return pod.Labels[util.GameServerPodLabelKey] != ""
}

// applyGameServerAddressAndPort applys pod ip and node name to GameServer's fields
func applyGameServerAddressAndPort(gs *carrierv1alpha1.GameServer, pod *corev1.Pod) {
	gs.Status.Address = pod.Status.PodIP
	gs.Status.NodeName = pod.Spec.NodeName
	if isHostPortNetwork(&gs.Spec) {
		var ingress []carrierv1alpha1.LoadBalancerIngress
		for _, p := range gs.Spec.Ports {
			ingress = append(ingress, carrierv1alpha1.LoadBalancerIngress{
				IP: pod.Spec.NodeName,
				Ports: []carrierv1alpha1.LoadBalancerPort{
					{
						ContainerPort:      p.ContainerPort,
						ExternalPort:       p.HostPort,
						ContainerPortRange: p.ContainerPortRange,
						ExternalPortRange:  p.HostPortRange,
						Protocol:           p.Protocol,
					},
				},
			})
		}
		gs.Status.LoadBalancerStatus = &carrierv1alpha1.LoadBalancerStatus{
			Ingress: ingress,
		}
	}
}

// isHostPortNetwork checks if pod runs as hostHost
func isHostPortNetwork(gss *carrierv1alpha1.GameServerSpec) bool {
	spec := gss.Template.Spec
	return spec.HostNetwork
}

// FindContainer returns the container specified by the name parameter. Returns the index and the value.
// Returns an error if not found.
func FindContainer(gss *carrierv1alpha1.GameServerSpec, name string) (int, corev1.Container, error) {
	for i, c := range gss.Template.Spec.Containers {
		if c.Name == name {
			return i, c, nil
		}
	}

	return -1, corev1.Container{}, errors.Errorf("Could not find a container named %s", name)
}

// checkNodeTaintByCA checks if node is marked as deletable.
// if true we should add constraints to gs spec.
func checkNodeTaintByCA(node *corev1.Node) bool {
	if len(node.Spec.Taints) == 0 {
		return false
	}
	for _, taint := range node.Spec.Taints {
		if taint.Key == ToBeDeletedTaint {
			return true
		}
	}
	return false
}

// NotInServiceConstraint describe a constraint that gs should not be
// in service again.
func NotInServiceConstraint() carrierv1alpha1.Constraint {
	effective := true
	now := metav1.NewTime(time.Now())
	return carrierv1alpha1.Constraint{
		Type:      carrierv1alpha1.NotInService,
		Effective: &effective,
		Message:   "Carrier controller mark this game server as not in service",
		TimeAdded: &now,
	}
}

// IsLoadBalancerPortExist check if a GameServer requires Load Balancer.
func IsLoadBalancerPortExist(gs *carrierv1alpha1.GameServer) bool {
	for _, port := range gs.Spec.Ports {
		if port.PortPolicy == carrierv1alpha1.LoadBalancer {
			return true
		}
	}
	return false
}

// updatePodSpec update game server spec, include, image and resource.
func updatePodSpec(gs *carrierv1alpha1.GameServer, pod *corev1.Pod) {
	var image string
	if pod.Labels == nil {
		pod.Labels = make(map[string]string)
	}
	pod.Labels[util.GameServerHash] = gs.Labels[util.GameServerHash]
	for _, container := range gs.Spec.Template.Spec.Containers {
		if container.Name != util.GameServerContainerName {
			continue
		}
		image = container.Image
		break
	}
	for i, container := range pod.Spec.Containers {
		if container.Name != util.GameServerContainerName {
			continue
		}
		pod.Spec.Containers[i].Image = image
		break
	}
}

func getOwner(gs *carrierv1alpha1.GameServer) string {
	if gs.Labels[util.SquadNameLabelKey] != "" {
		return gs.Labels[util.SquadNameLabelKey]
	}
	if gs.Labels[util.GameServerSetLabelKey] != "" {
		return gs.Labels[util.GameServerSetLabelKey]
	}
	return string(gs.UID)
}

func findPorts(gs *carrierv1alpha1.GameServer) []int {
	var ports []int
	if len(gs.Spec.Ports) == 0 {
		return ports
	}
	for _, port := range gs.Spec.Ports {
		if port.PortPolicy == carrierv1alpha1.LoadBalancer {
			continue
		}
		if port.HostPort != nil {
			ports = append(ports, int(*port.HostPort))
		}
		hpr := port.HostPortRange
		if port.HostPortRange != nil {
			for i := hpr.MinPort; i <= hpr.MaxPort; i++ {
				ports = append(ports, int(i))
			}
		}
	}
	return ports
}

func findDynamicPortNumber(gs *carrierv1alpha1.GameServer) int {
	count := 0
	if len(gs.Spec.Ports) == 0 {
		return 0
	}
	for _, port := range gs.Spec.Ports {
		if port.PortPolicy != carrierv1alpha1.Dynamic {
			continue
		}
		if port.ContainerPort != nil {
			count++
		}
	}
	return count
}

func findStaticPorts(gs *carrierv1alpha1.GameServer) []int {
	var ports []int
	if len(gs.Spec.Ports) == 0 {
		return nil
	}
	for _, port := range gs.Spec.Ports {
		if port.PortPolicy == carrierv1alpha1.Static {
			continue
		}
		if port.ContainerPort != nil && port.HostPort != nil {
			ports = append(ports, int(*port.HostPort))
		}
	}
	return ports
}

const (
	PortType      = "Port"
	PortRangeType = "PortRange"
)

func getAllocateType(gs *carrierv1alpha1.GameServer) string {
	if len(gs.Spec.Ports) == 0 {
		return ""
	}
	for _, port := range gs.Spec.Ports {
		if port.PortPolicy != carrierv1alpha1.Dynamic {
			continue
		}
		if port.ContainerPort != nil {
			return PortType
		}
		if port.ContainerPortRange != nil {
			return PortRangeType
		}
	}
	return ""
}

func setHostPort(gs *carrierv1alpha1.GameServer, ports []int) {
	for i, port := range gs.Spec.Ports {
		if port.PortPolicy != carrierv1alpha1.Dynamic {
			continue
		}
		if port.ContainerPort != nil {
			gs.Spec.Ports[i].HostPort = new(int32)
			*gs.Spec.Ports[i].HostPort = int32(ports[i])
		}
	}
}

func setHostPortRange(gs *carrierv1alpha1.GameServer, ports []int) {
	for i, port := range gs.Spec.Ports {
		if port.PortPolicy == carrierv1alpha1.LoadBalancer {
			continue
		}
		if port.ContainerPortRange != nil {
			gs.Spec.Ports[i].HostPortRange = &carrierv1alpha1.PortRange{
				MinPort: int32(ports[0]),
				MaxPort: int32(ports[len(ports)-1]),
			}
		}
	}
}

// getPersistentVolumeClaimName gets the name of PersistentVolumeClaim for a Pod with an ordinal index of ordinal. claim
// must be a PersistentVolumeClaim from set's VolumeClaims template.
func getPersistentVolumeClaimName(gs *carrierv1alpha1.GameServer, claim *corev1.PersistentVolumeClaim) string {
	// NOTE: This name format is used by the heuristics for zone spreading in ChooseZoneForVolume
	return fmt.Sprintf("%s-%s", claim.Name, gs.Name)
}

// getPersistentVolumeClaims gets a map of PersistentVolumeClaims to their template names, as defined in set. The
// returned PersistentVolumeClaims are each constructed with a the name specific to the Pod. This name is determined
// by getPersistentVolumeClaimName.
func getPersistentVolumeClaims(gs *carrierv1alpha1.GameServer) map[string]corev1.PersistentVolumeClaim {
	templates := gs.Spec.VolumeClaimTemplates
	claims := make(map[string]corev1.PersistentVolumeClaim, len(templates))
	for i := range templates {
		claim := templates[i]
		claim.Name = getPersistentVolumeClaimName(gs, &claim)
		claim.Namespace = gs.Namespace
		if claim.Labels != nil {
			for key, value := range gs.Labels {
				claim.Labels[key] = value
			}
		} else {
			claim.Labels = gs.Labels
		}
		claims[templates[i].Name] = claim
	}
	return claims
}

// updateStorage updates pod's Volumes to conform with the PersistentVolumeClaim of set's templates. If pod has
// conflicting local Volumes these are replaced with Volumes that conform to the set's templates.
func updateStorage(gs *carrierv1alpha1.GameServer, pod *corev1.Pod) {
	currentVolumes := pod.Spec.Volumes
	claims := getPersistentVolumeClaims(gs)
	newVolumes := make([]corev1.Volume, 0, len(claims))
	for name, claim := range claims {
		newVolumes = append(newVolumes, corev1.Volume{
			Name: name,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: claim.Name,
					// TODO: Use source definition to set this value when we have one.
					ReadOnly: false,
				},
			},
		})
	}
	for i := range currentVolumes {
		if _, ok := claims[currentVolumes[i].Name]; !ok {
			newVolumes = append(newVolumes, currentVolumes[i])
		}
	}
	pod.Spec.Volumes = newVolumes
}

func IsStateful(gs *carrierv1alpha1.GameServer) bool {
	return len(gs.Spec.VolumeClaimTemplates) > 0
}
