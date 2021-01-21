package controllers

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/piotrkpc/redpanda-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"time"
)

var _ = Describe("RedPandaCluster controller", func() {
	const (
		ClusterName          = "redpanda-test"
		ClusterSize          = 3
		ClusterNamespaceName = "default"
		timeout              = time.Second * 10
		interval             = time.Millisecond * 250
	)

	Context("When creating RedPandaCluster", func() {
		It("all nodes should be available", func() {
			redPandaCluster := &v1alpha1.RedPandaCluster{
				TypeMeta: metav1.TypeMeta{
					Kind:       "RedPandaCluster",
					APIVersion: "eventstream.vectorized.io/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "redpanda-test",
					Namespace: ClusterNamespaceName,
				},
				Spec: v1alpha1.RedPandaClusterSpec{
					Image: "vectorized/redpanda",
					Name:  ClusterName,
					Size:  ClusterSize,
				},
			}
			Expect(k8sClient.Create(context.TODO(), redPandaCluster)).Should(Succeed())
			redpandaLookupKey := types.NamespacedName{
				Namespace: ClusterNamespaceName,
				Name:      ClusterName,
			}
			redPandaCluster = &v1alpha1.RedPandaCluster{}
			By("Checking ")
			Eventually(func() bool {
				err := k8sClient.Get(context.TODO(), redpandaLookupKey, redPandaCluster)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
		})

	})

})
