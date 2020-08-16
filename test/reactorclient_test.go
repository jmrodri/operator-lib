package test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	testing "k8s.io/client-go/testing"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("ReactorClient", func() {

	Describe("Get", func() {
		var (
			client  crclient.Client
			reactor ReactorClient
		)
		BeforeEach(func() {
			client = fake.NewFakeClient(
				&corev1.Pod{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Pod",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:              "testpod",
						Namespace:         "testns",
						CreationTimestamp: metav1.Now(),
					},
				})

			reactor = NewReactorClient(client)
		})
		It("should return the error from prependreactor defined", func() {
			reactor.PrependReactor("get", "pods",
				func(action testing.Action) (bool, runtime.Object, error) {
					return true, &corev1.Pod{}, fmt.Errorf("REACTOR CALLED")
				})

			pod := &corev1.Pod{}
			key := crclient.ObjectKey{Namespace: "testns", Name: "testpod"}
			err := reactor.Get(context.TODO(), key, pod)
			Expect(err).ShouldNot(BeNil())
			Expect(err.Error()).To(Equal("REACTOR CALLED"))
		})
		It("should return object defined in client", func() {
			reactor.PrependReactor("get", "configmap",
				func(action testing.Action) (bool, runtime.Object, error) {
					return true, &corev1.ConfigMap{}, fmt.Errorf("REACTOR CALLED")
				})

			pod := &corev1.Pod{}
			key := crclient.ObjectKey{Namespace: "testns", Name: "testpod"}
			err := reactor.Get(context.TODO(), key, pod)
			Expect(err).Should(BeNil())
			Expect(pod.Name).To(Equal("testpod"))
		})
	})
	Describe("Create", func() {
		var (
			client  crclient.Client
			reactor ReactorClient
		)
		BeforeEach(func() {
			client = fake.NewFakeClient()
			reactor = NewReactorClient(client)
		})
		It("should return an error if the reactor matches", func() {
			reactor.PrependReactor("create", "configmaps",
				func(action testing.Action) (bool, runtime.Object, error) {
					return true, &corev1.ConfigMap{}, fmt.Errorf("Create ConfigMap Failed")
				})

			cm := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "reactor-test",
					Namespace: "reactorns",
				},
			}

			err := reactor.Create(context.TODO(), cm)
			Expect(err).ShouldNot(BeNil())
			Expect(err.Error()).Should(Equal("Create ConfigMap Failed"))
		})
		It("should create the object if the reactor does not match", func() {
			reactor.PrependReactor("create", "pods",
				func(action testing.Action) (bool, runtime.Object, error) {
					return true, &corev1.Pod{}, fmt.Errorf("Create Pod Failed")
				})

			cm := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "reactor-test",
					Namespace: "reactorns",
				},
			}

			resourceVersion := cm.GetResourceVersion()

			err := reactor.Create(context.TODO(), cm)
			Expect(err).Should(BeNil())
			Expect(resourceVersion).ShouldNot(Equal(cm.GetResourceVersion()))
		})
	})
	Describe("Delete", func() {
		var (
			client  crclient.Client
			reactor ReactorClient
		)
		BeforeEach(func() {
			client = fake.NewFakeClient()
			reactor = NewReactorClient(client)
		})
		It("should return an error if the reactor matches", func() {
			reactor.PrependReactor("delete", "pods",
				func(action testing.Action) (bool, runtime.Object, error) {
					return true, &corev1.Pod{}, fmt.Errorf("Delete Pod Failed")
				})

			p := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "reactor-test",
					Namespace: "reactorns",
				},
			}

			err := reactor.Delete(context.TODO(), p)
			Expect(err).ShouldNot(BeNil())
			Expect(err.Error()).Should(Equal("Delete Pod Failed"))
		})
		It("should delete the object if the reactor does not match", func() {
			// create a pod to delete
			p := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "reactor-test",
					Namespace: "reactorns",
				},
			}
			reactor.Create(context.TODO(), p)

			// add a delete reactor for configmaps; this should be ignored when
			// trying to delete pods
			reactor.PrependReactor("delete", "configmaps",
				func(action testing.Action) (bool, runtime.Object, error) {
					return true, &corev1.Pod{}, fmt.Errorf("Should be ignored")
				})

			err := reactor.Delete(context.TODO(), p)
			Expect(err).Should(BeNil())

			// see if the pod is gone
			pod := &corev1.Pod{}
			key := crclient.ObjectKey{Namespace: "reactorns", Name: "reactor-test"}
			err = reactor.Get(context.TODO(), key, pod)
			Expect(err).ShouldNot(BeNil())
			Expect(apierrors.IsNotFound(err)).Should(BeTrue())
		})
	})
})
