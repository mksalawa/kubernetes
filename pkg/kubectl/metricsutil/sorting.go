/*
Copyright 2016 The Kubernetes Authors.

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

package metricsutil

import (

	"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/heapster/metrics/apis/metrics/v1alpha1"
	"k8s.io/kubernetes/pkg/api/resource"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"strings"
)


type NodeMetricsContainer struct {
	Data v1alpha1.NodeMetrics
}

func NewNodeMetricsContainer(metrics v1alpha1.NodeMetrics) NodeMetricsContainer {
	return NodeMetricsContainer{Data: metrics}
}

func (n NodeMetricsContainer) GetResource(res v1.ResourceName) resource.Quantity {
	return n.Data.Usage[res]
}

func (n NodeMetricsContainer) GetName() string {
	return n.Data.Name
}

func (n NodeMetricsContainer) GetNamespace() string {
	return n.Data.Namespace
}

func (n NodeMetricsContainer) GetTimestamp() unversioned.Time {
	return n.Data.Timestamp
}

type ContainerMetricsContainer struct {
	Data      v1alpha1.ContainerMetrics
	Namespace string
	Timestamp unversioned.Time
}

func (c ContainerMetricsContainer) GetResource(res v1.ResourceName) resource.Quantity {
	return c.Data.Usage[res]
}

func (c ContainerMetricsContainer) GetName() string {
	return c.Data.Name
}

func (c ContainerMetricsContainer) GetNamespace() string {
	return ""
}

func (c ContainerMetricsContainer) GetTimestamp() unversioned.Time {
	return c.Timestamp
}

func NewContainerMetricsContainer(metrics v1alpha1.ContainerMetrics, namespace string, time unversioned.Time) ContainerMetricsContainer {
	return ContainerMetricsContainer{Data: metrics, Namespace: namespace, Timestamp : time}
}

type PodMetricsContainer struct {
	Data       v1alpha1.PodMetrics
	Total      v1.ResourceList
	Containers []GeneralMetricsContainer
}

func (p PodMetricsContainer) GetContainers() []GeneralMetricsContainer {
	return p.Containers
}

func (p PodMetricsContainer) GetResource(res v1.ResourceName) resource.Quantity {
	return p.Total[res]
}

func (p PodMetricsContainer) GetName() string {
	return p.Data.Name
}

func (p PodMetricsContainer) GetNamespace() string {
	return p.Data.Namespace
}

func (p PodMetricsContainer) GetTimestamp() unversioned.Time {
	return p.Data.Timestamp
}

func NewPodMetricsContainer(metrics v1alpha1.PodMetrics) PodMetricsContainer {
	podMetrics := PodMetricsContainer{Data : metrics}
	podMetrics.Total = make(v1.ResourceList, 0)
	podMetrics.Containers = make([]GeneralMetricsContainer, 0)
	for _, res := range MeasuredResources {
		podMetrics.Total[res], _ = resource.ParseQuantity("0")
	}
	for _, c := range metrics.Containers {
		podMetrics.Containers = append(podMetrics.Containers, NewContainerMetricsContainer(c, metrics.Namespace, metrics.Timestamp))
		for _, res := range MeasuredResources {
			quantity := podMetrics.Total[res]
			quantity.Add(c.Usage[res])
			podMetrics.Total[res] = quantity
		}
	}
	return podMetrics
}

type GeneralMetricsContainer interface {
	GetNamespace() string
	GetName() string
	GetResource(res v1.ResourceName) resource.Quantity
	GetTimestamp() unversioned.Time
}

type SortableMetrics struct {
	Data  []GeneralMetricsContainer
	Field v1.ResourceName
}

func NewSortableMetrics(metrics []GeneralMetricsContainer, field v1.ResourceName) SortableMetrics {
	return SortableMetrics{Data: metrics, Field: field}
}

func (s SortableMetrics) GetData() []GeneralMetricsContainer {
	return s.Data
}

func (s SortableMetrics) Len() int {
	return len(s.Data)
}

func (s SortableMetrics) Swap(i, j int) {
	s.Data[i], s.Data[j] = s.Data[j], s.Data[i]
}

func (s SortableMetrics) Less(i, j int) bool {
	if s.Field == "name" {
		iVal := s.Data[i].GetName()
		jVal := s.Data[j].GetName()
		return strings.Compare(iVal, jVal) < 0
	} else {
		iVal := s.Data[i].GetResource(s.Field)
		jVal := s.Data[j].GetResource(s.Field)
		return iVal.Cmp(jVal) < 0
	}
}