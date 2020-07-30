// Copyright 2020 The Operator-SDK Authors
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

package leader

import (
	"context"
	"errors"
	"fmt"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Leader election", func() {

	Describe("Become", func() {
		var (
			client crclient.Client
		)
		BeforeEach(func() {
			client = fake.NewFakeClient(
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "leader-test",
						Namespace: "testns",
						OwnerReferences: []metav1.OwnerReference{
							{
								APIVersion: "v1",
								Kind:       "Pod",
								Name:       "leader-test",
							},
						},
					},
				},
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "leader-test",
						Namespace: "testns",
						OwnerReferences: []metav1.OwnerReference{
							{
								APIVersion: "v1",
								Kind:       "Pod",
								Name:       "leader-test",
							},
						},
					},
				},
			)
		})
		When("POD_NAME is not set", func() {
			It("should return a required env POD_NAME error", func() {
				os.Unsetenv("POD_NAME")
				readNamespace = func() ([]byte, error) {
					return []byte("testns"), nil
				}
				err := Become(context.TODO(), "leader-test")
				Expect(err).ShouldNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("required env POD_NAME"))
			})
		})
		When("ErrNotExist is returned while reading namespace", func() {
			It("should return an ErrNoNamespace", func() {
				os.Setenv("POD_NAME", "leader-test")
				readNamespace = func() ([]byte, error) {
					return nil, os.ErrNotExist
				}
				err := Become(context.TODO(), "leader-test", WithClient(client))
				Expect(err).ShouldNot(BeNil())
				Expect(err).To(Equal(ErrNoNamespace))
				Expect(errors.Is(err, ErrNoNamespace)).To(Equal(true))
			})
		})
		Context("using an invalid Option", func() {
			It("should fail with an error", func() {
				err := Become(context.TODO(), "leader-test", func(*Config) error {
					return fmt.Errorf("better fail")
				})
				Expect(err).ShouldNot(BeNil())
				Expect(err.Error()).To(Equal("better fail"))
			})
		})
		When("client is nil and no client config found", func() {
			It("should fail with an error", func() {
				os.Setenv("KUBECONFIG", "/tmp/shouldnotexist")
				err := Become(context.TODO(), "leader-test", func(c *Config) error {
					c.Client = nil
					return nil
				})
				Expect(err).ShouldNot(BeNil())
			})
		})
		It("should not return an error", func() {
			os.Setenv("POD_NAME", "leader-test")
			readNamespace = func() ([]byte, error) {
				return []byte("testns"), nil
			}

			err := Become(context.TODO(), "leader-test", WithClient(client))
			Expect(err).Should(BeNil())
		})
	})
	Describe("isPodEvicted", func() {
		var (
			leaderPod *corev1.Pod
		)
		BeforeEach(func() {
			leaderPod = &corev1.Pod{}
		})
		It("should return false with an empty status", func() {
			Expect(isPodEvicted(*leaderPod)).To(Equal(false))
		})
		It("should return false if reason is incorrect", func() {
			leaderPod.Status.Phase = corev1.PodFailed
			leaderPod.Status.Reason = "invalid"
			Expect(isPodEvicted(*leaderPod)).To(Equal(false))
		})
		It("should return false if pod is in the wrong phase", func() {
			leaderPod.Status.Phase = corev1.PodRunning
			Expect(isPodEvicted(*leaderPod)).To(Equal(false))
		})
		It("should return true when Phase and Reason are set", func() {
			leaderPod.Status.Phase = corev1.PodFailed
			leaderPod.Status.Reason = "Evicted"
			Expect(isPodEvicted(*leaderPod)).To(Equal(true))
		})
	})
	Describe("getOperatorNamespace", func() {
		It("should return error when namespace not found", func() {
			readNamespace = func() ([]byte, error) {
				return nil, os.ErrNotExist
			}
			namespace, err := getOperatorNamespace()
			Expect(err).To(Equal(ErrNoNamespace))
			Expect(namespace).To(Equal(""))
		})
		It("should return namespace", func() {
			readNamespace = func() ([]byte, error) {
				return []byte("testnamespace"), nil
			}

			// test
			namespace, err := getOperatorNamespace()
			Expect(err).Should(BeNil())
			Expect(namespace).To(Equal("testnamespace"))
		})
		It("should trim whitespace from namespace", func() {
			readNamespace = func() ([]byte, error) {
				return []byte("   testnamespace    "), nil
			}

			// test
			namespace, err := getOperatorNamespace()
			Expect(err).Should(BeNil())
			Expect(namespace).To(Equal("testnamespace"))
		})
	})
	Describe("myOwnerRef", func() {
		var (
			client crclient.Client
		)
		BeforeEach(func() {
			client = fake.NewFakeClient(
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "mypod",
						Namespace: "testns",
					},
				},
			)
		})
		It("should return an error when POD_NAME is not set", func() {
			os.Unsetenv("POD_NAME")
			_, err := myOwnerRef(context.TODO(), client, "")
			Expect(err).ShouldNot(BeNil())
		})
		It("should return an error if no pod is found", func() {
			os.Setenv("POD_NAME", "thisisnotthepodyourelookingfor")
			_, err := myOwnerRef(context.TODO(), client, "")
			Expect(err).ShouldNot(BeNil())
		})
		It("should return the owner reference without error", func() {
			os.Setenv("POD_NAME", "mypod")
			owner, err := myOwnerRef(context.TODO(), client, "testns")
			Expect(err).Should(BeNil())
			Expect(owner.APIVersion).To(Equal("v1"))
			Expect(owner.Kind).To(Equal("Pod"))
			Expect(owner.Name).To(Equal("mypod"))
		})
	})
	Describe("getPod", func() {
		var (
			client crclient.Client
		)
		BeforeEach(func() {
			client = fake.NewFakeClient(
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "mypod",
						Namespace: "testns",
					},
				},
			)
		})
		It("should return an error when POD_NAME is not set", func() {
			os.Unsetenv("POD_NAME")
			_, err := getPod(context.TODO(), nil, "")
			Expect(err).ShouldNot(BeNil())
		})
		It("should return an error if no pod is found", func() {
			os.Setenv("POD_NAME", "thisisnotthepodyourelookingfor")
			_, err := getPod(context.TODO(), client, "")
			Expect(err).ShouldNot(BeNil())
		})
		It("should return the pod with the given name", func() {
			os.Setenv("POD_NAME", "mypod")
			pod, err := getPod(context.TODO(), client, "testns")
			Expect(err).Should(BeNil())
			Expect(pod).ShouldNot(BeNil())
			Expect(pod.TypeMeta.APIVersion).To(Equal("v1"))
			Expect(pod.TypeMeta.Kind).To(Equal("Pod"))
		})
	})
})
