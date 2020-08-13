package test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
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
			reactor.PrependReactor("get", "pods", func(action testing.Action) (bool, runtime.Object, error) {
				fmt.Printf("XXX matches %v\n", action.Matches("get", "pod"))
				return true, &corev1.Pod{}, fmt.Errorf("REACTOR CALLED")
			})

			pod := &corev1.Pod{}
			key := crclient.ObjectKey{Namespace: "testns", Name: "testpod"}
			err := reactor.Get(context.TODO(), key, pod)
			Expect(err).ShouldNot(BeNil())
			Expect(err.Error()).To(Equal("REACTOR CALLED"))
		})
		It("should return object defined in client", func() {
			reactor.PrependReactor("get", "configmap", func(action testing.Action) (bool, runtime.Object, error) {
				fmt.Printf("XXX matches %v\n", action.Matches("get", "pod"))
				return true, &corev1.ConfigMap{}, fmt.Errorf("REACTOR CALLED")
			})

			pod := &corev1.Pod{}
			key := crclient.ObjectKey{Namespace: "testns", Name: "testpod"}
			err := reactor.Get(context.TODO(), key, pod)
			Expect(err).Should(BeNil())
			Expect(pod.Name).To(Equal("testpod"))
		})
	})

})
