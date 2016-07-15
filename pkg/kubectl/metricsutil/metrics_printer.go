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
	"fmt"
	"io"

	"k8s.io/kubernetes/pkg/api/resource"
	"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/kubectl"
	"time"
)

var (
	MeasuredResources = []v1.ResourceName{
		v1.ResourceCPU,
		v1.ResourceMemory,
		v1.ResourceStorage,
	}
	NodeColumns = []string{"NAME", "CPU", "MEMORY", "STORAGE", "TIMESTAMP"}
	PodColumns  = []string{"NAMESPACE", "NAME", "CPU", "MEMORY", "STORAGE", "TIMESTAMP"}
)

type ResourceMetricsInfo struct {
	Namespace string
	Name      string
	Metrics   v1.ResourceList
	Timestamp string
}

type TopCmdPrinter struct {
	out io.Writer
}

func NewTopCmdPrinter(out io.Writer) *TopCmdPrinter {
	return &TopCmdPrinter{out: out}
}

func (printer *TopCmdPrinter) PrintNodeMetrics(metrics []GeneralMetricsContainer) error {
	if len(metrics) == 0 {
		return nil
	}
	w := kubectl.GetNewTabWriter(printer.out)
	defer w.Flush()

	PrintColumnNames(w, NodeColumns)
	for _, m := range metrics {
		PrintMetricsLine(w, m, false, true)
	}
	return nil
}

func (printer *TopCmdPrinter) PrintPodMetrics(metrics []GeneralMetricsContainer, printContainers bool) error {
	if len(metrics) == 0 {
		return nil
	}
	w := kubectl.GetNewTabWriter(printer.out)
	defer w.Flush()

	PrintColumnNames(w, PodColumns)
	for _, m := range metrics {
		podMetrics := m.(PodMetricsContainer)
		PrintSinglePodMetrics(w, &podMetrics, printContainers)
	}
	return nil
}

func PrintColumnNames(out io.Writer, names []string) {
	for _, name := range names {
		PrintValue(out, name)
	}
	fmt.Fprintf(out, "\n")
}

func PrintSinglePodMetrics(out io.Writer, m *PodMetricsContainer, printContainers bool) {
	PrintMetricsLine(out, m, true, true)
	if printContainers {
		for _, c := range m.GetContainers() {
			PrintMetricsLine(out, c, true, false)
		}
	}
}

func PrintMetricsLine(out io.Writer, metrics GeneralMetricsContainer, withNamespace bool, withTimestamp bool) {
	if withNamespace {
		PrintValue(out, metrics.GetNamespace())
	}
	PrintValue(out, metrics.GetName())
	PrintAllResourceUsages(out, metrics)
	if withTimestamp {
		PrintValue(out, metrics.GetTimestamp().Time.Format(time.RFC1123Z))
	}
	fmt.Fprintf(out, "\n")
}

func PrintValue(out io.Writer, value interface{}) {
	fmt.Fprintf(out, "%v\t", value)
}

func PrintAllResourceUsages(out io.Writer, metrics GeneralMetricsContainer) {
	for _, res := range MeasuredResources {
		quantity := metrics.GetResource(res)
		PrintSingleResourceUsage(out, res, quantity)
		fmt.Fprintf(out, "\t")
	}
}

func PrintSingleResourceUsage(out io.Writer, resourceType v1.ResourceName, quantity resource.Quantity) {
	switch resourceType {
	case v1.ResourceCPU:
		fmt.Fprintf(out, "%vm", quantity.MilliValue())
	case v1.ResourceMemory, v1.ResourceStorage:
		fmt.Fprintf(out, "%v Mi", quantity.Value()/(1024*1024))
	default:
		fmt.Fprintf(out, "%v", quantity.Value())
	}
}
