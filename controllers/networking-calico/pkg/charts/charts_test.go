// Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package charts_test

import (
	"fmt"

	calicov1alpha1 "github.com/gardener/gardener-extensions/controllers/networking-calico/pkg/apis/calico/v1alpha1"
	"github.com/gardener/gardener-extensions/controllers/networking-calico/pkg/calico"
	"github.com/gardener/gardener-extensions/controllers/networking-calico/pkg/charts"
	"github.com/gardener/gardener-extensions/controllers/networking-calico/pkg/imagevector"
	mockchartrenderer "github.com/gardener/gardener-extensions/pkg/mock/gardener/chartrenderer"

	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/chartrenderer"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/helm/pkg/manifest"
)

var (
	trueVar = true
)

var _ = Describe("Chart package test", func() {
	var (
		crossSubnet                     = calicov1alpha1.CrossSubnet
		invalid     calicov1alpha1.IPIP = "invalid"
	)

	var (
		network              *extensionsv1alpha1.Network
		networkConfig        *calicov1alpha1.NetworkConfig
		networkConfigAll     *calicov1alpha1.NetworkConfig
		networkConfigInvalid *calicov1alpha1.NetworkConfig
		objectMeta           = metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "bar",
		}
		podCIDR             = calicov1alpha1.CIDR("12.0.0.0/8")
		autodetectionMethod = "interface=eth1"
	)

	BeforeEach(func() {
		network = &extensionsv1alpha1.Network{
			ObjectMeta: objectMeta,
			Spec: extensionsv1alpha1.NetworkSpec{
				ServiceCIDR: "10.0.0.0/8",
				PodCIDR:     string(podCIDR),
			},
		}
		networkConfig = &calicov1alpha1.NetworkConfig{
			Backend: calicov1alpha1.None,
			IPAM: &calicov1alpha1.IPAM{
				CIDR: &podCIDR,
				Type: "host-local",
			},
		}
		networkConfigAll = &calicov1alpha1.NetworkConfig{
			Backend: calicov1alpha1.None,
			IPAM: &calicov1alpha1.IPAM{
				CIDR: &podCIDR,
				Type: "host-local",
			},
			IPIP:                  &crossSubnet,
			IPAutoDetectionMethod: &autodetectionMethod,
		}
		networkConfigInvalid = &calicov1alpha1.NetworkConfig{
			Backend: "invalid",
			IPAM: &calicov1alpha1.IPAM{
				CIDR: &podCIDR,
				Type: "host-local",
			},
			IPIP:                  &invalid,
			IPAutoDetectionMethod: &autodetectionMethod,
		}
	})

	Describe("#ComputeCalicoChartValues", func() {
		It("should correctly compute the calico chart values", func() {
			values := charts.ComputeCalicoChartValues(network, networkConfig)
			Expect(values).To(Equal(map[string]interface{}{
				"images": map[string]interface{}{
					"calico-cni":              imagevector.CalicoCNIImage(),
					"calico-typha":            imagevector.CalicoTyphaImage(),
					"calico-kube-controllers": imagevector.CalicoKubeControllersImage(),
					"calico-node":             imagevector.CalicoNodeImage(),
					"calico-podtodaemon-flex": imagevector.CalicoFlexVolumeDriverImage(),
					"typha-cpa":               imagevector.TyphaClusterProportionalAutoscalerImage(),
					"typha-cpva":              imagevector.TyphaClusterProportionalVerticalAutoscalerImage(),
				},
				"global": map[string]string{
					"podCIDR": network.Spec.PodCIDR,
				},
				"config": map[string]interface{}{
					"backend": networkConfig.Backend,
					"ipam": map[string]interface{}{
						"type":   networkConfig.IPAM.Type,
						"subnet": *networkConfig.IPAM.CIDR,
					},
					"typha": map[string]interface{}{
						"enabled": trueVar,
					},
				},
				"ipip": calicov1alpha1.Always,
			}))
		})
	})

	Describe("#ComputeAllCalicoChartValues", func() {
		It("should correctly compute all of the calico chart values", func() {
			values := charts.ComputeCalicoChartValues(network, networkConfigAll)
			Expect(values).To(Equal(map[string]interface{}{
				"images": map[string]interface{}{
					"calico-cni":              imagevector.CalicoCNIImage(),
					"calico-typha":            imagevector.CalicoTyphaImage(),
					"calico-kube-controllers": imagevector.CalicoKubeControllersImage(),
					"calico-node":             imagevector.CalicoNodeImage(),
					"calico-podtodaemon-flex": imagevector.CalicoFlexVolumeDriverImage(),
					"typha-cpa":               imagevector.TyphaClusterProportionalAutoscalerImage(),
					"typha-cpva":              imagevector.TyphaClusterProportionalVerticalAutoscalerImage(),
				},
				"global": map[string]string{
					"podCIDR": network.Spec.PodCIDR,
				},
				"config": map[string]interface{}{
					"backend": networkConfigAll.Backend,
					"ipam": map[string]interface{}{
						"type":   networkConfigAll.IPAM.Type,
						"subnet": *networkConfigAll.IPAM.CIDR,
					},
					"typha": map[string]interface{}{
						"enabled": trueVar,
					},
				},
				"ipAutodetectionMethod": *networkConfigAll.IPAutoDetectionMethod,
				"ipip":                  *networkConfigAll.IPIP,
			}))
		})
	})

	Describe("#ComputeInvalidCalicoChartValues", func() {
		It("should replace invalid values for calico charts", func() {
			values := charts.ComputeCalicoChartValues(network, networkConfigInvalid)
			Expect(values).To(Equal(map[string]interface{}{
				"images": map[string]interface{}{
					"calico-cni":              imagevector.CalicoCNIImage(),
					"calico-typha":            imagevector.CalicoTyphaImage(),
					"calico-kube-controllers": imagevector.CalicoKubeControllersImage(),
					"calico-node":             imagevector.CalicoNodeImage(),
					"calico-podtodaemon-flex": imagevector.CalicoFlexVolumeDriverImage(),
					"typha-cpa":               imagevector.TyphaClusterProportionalAutoscalerImage(),
					"typha-cpva":              imagevector.TyphaClusterProportionalVerticalAutoscalerImage(),
				},
				"global": map[string]string{
					"podCIDR": network.Spec.PodCIDR,
				},
				"config": map[string]interface{}{
					"backend": calicov1alpha1.Bird,
					"ipam": map[string]interface{}{
						"type":   networkConfigInvalid.IPAM.Type,
						"subnet": *networkConfigInvalid.IPAM.CIDR,
					},
					"typha": map[string]interface{}{
						"enabled": trueVar,
					},
				},
				"ipAutodetectionMethod": *networkConfigInvalid.IPAutoDetectionMethod,
				"ipip":                  calicov1alpha1.Always,
			}))
		})
	})

	Describe("#RenderCalicoChart", func() {
		var (
			ctrl                = gomock.NewController(GinkgoT())
			mockChartRenderer   = mockchartrenderer.NewMockInterface(ctrl)
			testManifestContent = "test-content"
			mkManifest          = func(name string) manifest.Manifest {
				return manifest.Manifest{Name: fmt.Sprintf("test/templates/%s", name), Content: testManifestContent}
			}
		)
		It("Render Calico charts correctly", func() {
			mockChartRenderer.EXPECT().Render(calico.ChartPath, calico.ReleaseName, metav1.NamespaceSystem, gomock.Any()).Return(&chartrenderer.RenderedChart{
				ChartName: "test",
				Manifests: []manifest.Manifest{
					mkManifest(charts.CalicoConfigKey),
				},
			}, nil)

			_, err := charts.RenderCalicoChart(mockChartRenderer, network, networkConfig)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
