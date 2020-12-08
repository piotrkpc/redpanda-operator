/*


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

package controllers

import (
	"context"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	v1 "k8s.io/apiserver/pkg/apis/example/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	eventstreamv1alpha1 "github.com/piotrkpc/redpanda-operator/api/v1alpha1"
)

// RedPandaClusterReconciler reconciles a RedPandaCluster object
type RedPandaClusterReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=eventstream.vectorized.io,resources=redpandaclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=eventstream.vectorized.io,resources=redpandaclusters/status,verbs=get;update;patch

func (r *RedPandaClusterReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {

	//// Reconcile successful - don't requeue
	//return ctrl.Result{}, nil
	//// Reconcile failed due to error - requeue
	//return ctrl.Result{}, err
	//// Requeue for any reason other than an error
	//return ctrl.Result{Requeue: true}, nil

	ctx := context.Background()
	log := r.Log.WithValues("redpandacluster", req.NamespacedName)

	var redPandaCluster eventstreamv1alpha1.RedPandaCluster
	err := r.Get(ctx, req.NamespacedName, &redPandaCluster)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("RedPandaCluster resource not statefulSetFound. Ignoring...")
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
		log.Error(err, "unable to fetch RedPandaCluster")
		return ctrl.Result{}, err
	}

	serviceFound := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: redPandaCluster.Name, Namespace: redPandaCluster.Namespace}, serviceFound)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Service not found. Trying to create one...")
			var rpService *corev1.Service
			rpService = r.serviceFor(redPandaCluster)
			log.Info("creating service ", "Service.Namespace", rpService.Namespace, "Service.Name", rpService.Name)
			err := r.Create(ctx, rpService)
			if err != nil {
				log.Error(err, "Failed to create new service", "Service.Namespace", rpService.Namespace, "Service.Name", rpService.Name)
				return ctrl.Result{}, err
			}
			return ctrl.Result{Requeue: true}, nil
		}
		log.Info("unable to fetch Service resource")
		return ctrl.Result{}, err
	}

	// ConfigMap
	configMapFound := &corev1.ConfigMap{}
	err = r.Get(ctx, types.NamespacedName{Name: redPandaCluster.Name, Namespace: redPandaCluster.Namespace}, configMapFound)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("ConfigMap not found. Trying to create one...")
			var rpConfigMap *corev1.ConfigMap
			rpConfigMap = r.configMapFor()
			err := r.Create(ctx, rpConfigMap)
			if err != nil {
				log.Error(err, "Failed to create ConfigMap resource")
				return ctrl.Result{}, err
			}
			return ctrl.Result{Requeue: true}, nil
		}
		log.Info("unable to fetch Service resource")
		return ctrl.Result{}, err
	}

	statefulSetFound := &appsv1.StatefulSet{}
	err = r.Get(ctx, types.NamespacedName{Namespace: redPandaCluster.Namespace, Name: redPandaCluster.Name}, statefulSetFound)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("StatefulSet not found ")
			var rpStatefulSet *appsv1.StatefulSet
			rpStatefulSet = r.statefulSetFor(redPandaCluster)
		}
		return ctrl.Result{}, err
	}
	// TODO: finish here

	// Step 1. Create services, config maps. -> reconcile requeue

	// step 2. Create stateful set. -> requeue
	log.Info("reconcile loop ends")
	return ctrl.Result{}, nil
}

func (r *RedPandaClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&eventstreamv1alpha1.RedPandaCluster{}).
		Owns(&appsv1.StatefulSet{}).

		Complete(r)
}

func (r *RedPandaClusterReconciler) statefulSetFor(cluster eventstreamv1alpha1.RedPandaCluster) *appsv1.StatefulSet {

	labels := map[string]{"app": "redpanda"}
	replicas := cluster.Spec.Size
	cm := &corev1.ConfigMap{}
	ss := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name,
			Namespace: cluster.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			ServiceName: cluster.Name,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: v1.PodSpec{
					Volumes: r.getVolumes(),
					InitContainers: []v1.Container{v1.Container{
						Name:         "redpanda-configurator",
						Image:        cluster.Spec.Image,
						Command:      r.getInitCmd(),
						Args:         r.getInitScript(),
						VolumeMounts: r.getVolumeMounts(),
					}},
					Containers: []v1.Container{v1.Container{
						Name:         "redpanda",
						Image:        cluster.Spec.Image,
						Args:         r.getRPArgs(),
						Ports:        r.getPorts(),
						VolumeMounts: r.getRPVolumeMounts(),
					}},
				},
			},
		},
	}
	// TODO: i've stopped here! Checkout the example:
	// https://github.com/operator-framework/operator-sdk/blob/master/testdata/go/memcached-operator/controllers/memcached_controller.go
	ctrl.SetControllerReference(cluster, ss, r.Scheme)
	r.Client
	return ss
}

func (r *RedPandaClusterReconciler) serviceFor(cluster eventstreamv1alpha1.RedPandaCluster) *corev1.Service {
	labels := map[string]string{"app": "redpanda"}
	rpcPort := 33145
	rpcServicePort := &corev1.ServicePort{
		Name:       "rpc",
		Protocol:   "TCP",
		Port:       int32(rpcPort),
		TargetPort: intstr.FromInt(rpcPort),
	}
	kafkaPort := 9092
	kafkaServicePort := &corev1.ServicePort{
		Name:       "kafka",
		Protocol:   "TCP",
		Port:       int32(kafkaPort),
		TargetPort: intstr.FromInt(kafkaPort),
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name,
			Namespace: cluster.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Ports:     []corev1.ServicePort{*rpcServicePort, *kafkaServicePort},
			Selector:  labels,
			ClusterIP: "None",
		},
	}
	return service

}

func (r *RedPandaClusterReconciler) configMapFor() *corev1.ConfigMap {

	configContent := `
config_file: /etc/redpanda/redpanda.yaml
license_key: ""
node_uuid: mf1PqvLg826mCjiYRRKcBb4MjtL88agNyGajk3GoELmgKixYA
organization: ""
redpanda:
  admin:
	address: 0.0.0.0
	port: 9644
  auto_create_topics_enabled: true
  data_directory: /var/lib/redpanda/data
  developer_mode: false
  kafka_api:
	address: 0.0.0.0
	port: 9092
  kafka_api_tls:
	cert_file: ""
	enabled: false
	key_file: ""
	truststore_file: ""
  node_id: 0
  rpc_server:
	address: 0.0.0.0
	port: 33145
  seed_servers:
	- node_id: 0
	  host:
		address: redpanda-0
		port: 33145
rpk:
  coredump_dir: /var/lib/redpanda/coredump
  enable_memory_locking: false
  enable_usage_stats: true
  overprovisioned: false
  tls:
	cert_file: ""
	key_file: ""
	truststore_file: ""
  tune_aio_events: false
  tune_clocksource: false
  tune_coredump: false
  tune_cpu: false
  tune_disk_irq: false
  tune_disk_nomerges: false
  tune_disk_scheduler: false
  tune_fstrim: false
  tune_network: false
  tune_swappiness: false
  tune_transparent_hugepages: false
`
	cm := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name: "base-config",
		},
		Data: map[string]string{"redpanda.yaml": configContent},
	}
	return cm

}

func (r *RedpandaClusterReconciler) getVolumes()
[]v1.Volume{
return []v1.Volume{{Name: "base-config", VolumeSource:}}
}

func (r *RedpandaClusterReconciler) getInitCmd()
[]string{
return []string{"/bin/sh", "-c"}
}

func (r *RedpandaClusterReconciler) getInitScript()
[]string{
return []string{InitContiinitContainerScript}
}

func (r *RedpandaClusterReconciler) getVolumeMounts()
[]v1.VolumeMount{
return []v1.VolumeMount{}
}

func (r *RedpandaClusterReconciler) getRPArgs()
[]string{

}

func (r *RedpandaClusterReconciler) getPorts()
[]v1.ContainerPort{

}

func (r *RedpandaClusterReconciler) getRPVolumeMounts()
[]v1.VolumeMount{
}

}
