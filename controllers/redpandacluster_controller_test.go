package controllers

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/piotrkpc/redpanda-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("RedPandaCluster controller", func() {
	const (
		ClusterName = "redpanda-test"
		ClusterSize = 3
	)

	Context("When creating RedPandaCluster", func() {
		It("Should create StatefulSet", func() {
			redPandaCluster := &v1alpha1.RedPandaCluster{
				TypeMeta: metav1.TypeMeta{
					Kind:       "RedPandaCluster",
					APIVersion: "eventstream.vectorized.io/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "redpanda-test",
					Namespace: "default",
				},
				Spec: v1alpha1.RedPandaClusterSpec{
					Image: "vectorized/redpanda",
					Name:  ClusterName,
					Size:  ClusterSize,
				},
			}
			Expect(k8sClient.Create(context.TODO(), redPandaCluster)).Should(Succeed())
		})
	})

})
