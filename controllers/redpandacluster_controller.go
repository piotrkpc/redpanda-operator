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
	"fmt"
	"github.com/go-logr/logr"
	"github.com/piotrkpc/redpanda-operator/pkg/reconciliation"
	"github.com/piotrkpc/redpanda-operator/pkg/reconciliation/pipelines"
	"github.com/piotrkpc/redpanda-operator/pkg/reconciliation/pipelines/k8s"
	"github.com/piotrkpc/redpanda-operator/pkg/reconciliation/result"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	eventstreamv1alpha1 "github.com/piotrkpc/redpanda-operator/api/v1alpha1"
)

// NOTE: Redesign idea
// controller is configurable, it's an implementation detail for now,
// based on the configuration,
// - ReconcileContext is create - it's an Value object that that contains infromation for Concrete object implementing
//   template pattern
// - appropriate instance of concreate Something is created
// That something will implement (?) template pattern
// Concreate something can consist of other objects utilizing composition
//
// TODO: Esablish Template method
// TODO: Create the something object using factory method

// RedPandaClusterReconciler reconciles a RedPandaCluster object
type RedPandaClusterReconciler struct {
	client.Client
	Log             logr.Logger
	Scheme          *runtime.Scheme
	SelectedBackend string
}

// +kubebuilder:rbac:groups=eventstream.vectorized.io,resources=redpandaclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=eventstream.vectorized.io,resources=redpandaclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete;
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete;
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;

func (r *RedPandaClusterReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {

	//// Reconcile successful - don't requeue
	//return ctrl.Result{}, nil
	//// Reconcile failed due to error - requeue
	//return ctrl.Result{}, err
	//// Requeue for any reason other than an error
	//return ctrl.Result{Requeue: true}, nil

	ctx := context.Background()
	log := r.Log.WithValues("redpandacluster", req.NamespacedName)
	coreReconciliationContext := CreateReconciliationContext(&req, r.Client, r.Scheme, r.Log, ctx)
	var pipeline pipelines.PipelineReconciler
	pipeline, err := CreateReconciliationPipeline(r.SelectedBackend, r.Log)
	pipelineContext := pipeline.AddContextTo(coreReconciliationContext)

	reconcileResult := pipeline.Reconcile(pipelineContext)
	switch r := reconcileResult.(type) {
	case result.ContinueReconcile:
		log.Info("pipeline: reconciliation pipeline finished successfully")
	default:
		return r.Output()
	}

	// step 1. Check for CR
	redPandaCluster := &eventstreamv1alpha1.RedPandaCluster{}
	err = r.Get(ctx, req.NamespacedName, redPandaCluster)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("RedPandaCluster resource not found. Ignoring...")
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
	err = r.Get(ctx, types.NamespacedName{Name: redPandaCluster.Name + "base-config", Namespace: redPandaCluster.Namespace}, configMapFound)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("ConfigMap not found. Trying to create one...")
			var rpConfigMap *corev1.ConfigMap
			rpConfigMap = r.configMapFor(redPandaCluster)
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

	// StatefulSet
	statefulSetFound := &appsv1.StatefulSet{}
	err = r.Get(ctx, types.NamespacedName{Namespace: redPandaCluster.Namespace, Name: redPandaCluster.Name}, statefulSetFound)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("StatefulSet not found ")
			var rpStatefulSet *appsv1.StatefulSet
			rpStatefulSet = r.statefulSetFor(redPandaCluster, configMapFound, serviceFound)
			err := r.Create(ctx, rpStatefulSet)
			if err != nil {
				log.Error(err, "Failed to create StatefulSet resource")
				return ctrl.Result{}, err
			}
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, err
	}

	observedPods := &corev1.PodList{}
	err = r.List(ctx, observedPods, &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(redPandaCluster.Labels),
		Namespace:     redPandaCluster.Namespace,
	})
	if err != nil {
		log.Error(err, "unable to fetch PodList resource")
		return ctrl.Result{}, err
	}
	for _, pod := range observedPods.Items {
		redPandaCluster.Status.NodeStatuses[pod.String()] = eventstreamv1alpha1.RedPandaNodeStatus(pod.Status.Conditions[0].Status)
	}
	err = r.Status().Update(ctx, redPandaCluster)
	if err != nil {
		log.Error(err, "Failed to update RedPandaClusterStatus")
		return ctrl.Result{}, err
	}
	log.Info("reconcile loop ends")
	return ctrl.Result{}, nil
}

func CreateReconciliationPipeline(backend string, log logr.Logger) (pipelines.PipelineReconciler, error) {
	switch backend {
	case "k8s":
		pipeline, err := k8s.PipelineFactory()
		if err != nil {
			return pipeline, nil
		}
		log.Error(err, "Error creating reconciliation pipeline")
		return nil, err
	default:
		err := BackendNotSupported(backend)
		log.Error(err, "Error creating reconciliation pipeline")
		return nil, err
	}
}

type BackendNotSupported string

func (s BackendNotSupported) Error() string {
	return fmt.Sprintf("reconciler: backend %s not supported", s)
}

func CreateReconciliationContext(req *ctrl.Request, client client.Client, scheme *runtime.Scheme, log logr.Logger, ctx context.Context) *reconciliation.CoreReconciliationContext {

	return &reconciliation.CoreReconciliationContext{
		Request:   req,
		Client:    client,
		Scheme:    scheme,
		ReqLogger: log,
		Ctx:       ctx,
	}
}

func (r *RedPandaClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&eventstreamv1alpha1.RedPandaCluster{}).
		Owns(&appsv1.StatefulSet{}).
		Complete(r)
}

func (r *RedPandaClusterReconciler) statefulSetFor(cluster *eventstreamv1alpha1.RedPandaCluster, configMap *corev1.ConfigMap, service *corev1.Service) *appsv1.StatefulSet {

	rpLabels := map[string]string{"app": "redpanda"}
	replicas := &cluster.Spec.Size
	volMap := r.volumes(configMap)
	volumes := []corev1.Volume{*volMap["configMap"], *volMap["config"]}
	ss := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name,
			Namespace: cluster.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: rpLabels,
			},
			ServiceName: service.Name,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: rpLabels,
				},
				Spec: corev1.PodSpec{
					Volumes: volumes,
					InitContainers: []corev1.Container{
						{
							Name:         "redpanda-configurator",
							Image:        cluster.Spec.Image,
							Command:      r.initCmd(),
							Args:         r.initScript(),
							VolumeMounts: r.initContainerVolumeMounts(volMap),
						},
					},
					Containers: []corev1.Container{
						{
							Name:         "redpanda",
							Image:        cluster.Spec.Image,
							Args:         r.getRPArgs(),
							Ports:        r.getPorts(),
							VolumeMounts: r.getRPVolumeMounts(volMap),
						},
					},
				},
			},
			VolumeClaimTemplates: r.rpVolClaimTemplates(),
		},
	}
	_ = ctrl.SetControllerReference(cluster, ss, r.Scheme)
	return ss
}

func (r *RedPandaClusterReconciler) serviceFor(cluster *eventstreamv1alpha1.RedPandaCluster) *corev1.Service {
	rpLabels := map[string]string{"app": "redpanda"}
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
			Labels:    rpLabels,
		},
		Spec: corev1.ServiceSpec{
			Ports:     []corev1.ServicePort{*rpcServicePort, *kafkaServicePort},
			Selector:  rpLabels,
			ClusterIP: "None",
		},
	}
	return service

}

func (r *RedPandaClusterReconciler) configMapFor(cluster *eventstreamv1alpha1.RedPandaCluster) *corev1.ConfigMap {
	// @formatter:off
	configContent := `---
config_file: /etc/redpanda/redpanda.yaml
license_key: ""
node_uuid: mf1PqvLg826mCjiYRRKcBb4MjtL88agNyGajk3GoELmgKixYA
organization: ""
redpanda:
  admin:
    address: "0.0.0.0"
    port: 9644
  auto_create_topics_enabled: true
  data_directory: /var/lib/redpanda/data
  developer_mode: false
  kafka_api:
    address: "0.0.0.0"
    port: 9092
  kafka_api_tls:
    cert_file: ""
    enabled: false
    key_file: ""
    truststore_file: ""
  node_id: 0
  rpc_server:
    address: "0.0.0.0"
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
	// @formatter:on
	cm := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name + "base-config",
			Namespace: cluster.Namespace,
		},
		Data: map[string]string{"redpanda.yaml": configContent},
	}
	return cm

}

func (r *RedPandaClusterReconciler) volumes(cm *corev1.ConfigMap) map[string]*corev1.Volume {
	cmVol := corev1.Volume{
		Name: cm.Name,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: cm.Name},
			},
		},
	}

	config := corev1.Volume{
		Name: "config",
		VolumeSource: corev1.VolumeSource{
			EmptyDir: nil,
		},
	}
	vols := make(map[string]*corev1.Volume)
	vols["configMap"] = &cmVol
	vols["config"] = &config
	return vols
}

func (r *RedPandaClusterReconciler) initCmd() []string {
	return []string{"/bin/sh", "-c"}
}

func (r *RedPandaClusterReconciler) initScript() []string {
	initCmd := `CONFIG=/tmp/config/redpanda.yaml;
NODE_ID=${HOSTNAME##*-};
SERVICE_NAME=${HOSTNAME};
BASE_NAME=${HOSTNAME%-${NODE_ID}}
ROOT_ID=0
echo "Copying ConfigMap to default location $CONFIG"
cp -v /tmp/base-config/redpanda.yaml $CONFIG;
echo "Content of the ConfigMap"
cat ${CONFIG}
rpk --config $CONFIG config set redpanda.node_id $NODE_ID;
if [ "$NODE_ID" = "0" ]; then
rpk --config $CONFIG config set redpanda.seed_servers '[]' --format yaml;
else
SEED_SERVERS="[{node_id: $ROOT_ID, host: {address: \"${BASE_NAME}-${ROOT_ID}.${BASE_NAME}\", port: 33145}}]"
rpk --config $CONFIG config set redpanda.seed_servers "$SEED_SERVERS" --format yaml;
fi;
rpk --config $CONFIG config set redpanda.advertised_rpc_api.address $SERVICE_NAME;
rpk --config $CONFIG config set redpanda.advertised_rpc_api.port 33145;
rpk --config $CONFIG config set redpanda.advertised_kafka_api.address $SERVICE_NAME;
rpk --config $CONFIG config set redpanda.advertised_kafka_api.port 9092;
rpk --config $CONFIG config set rpk.smp 1;
echo "Reconfiguration ends"
cat ${CONFIG}
`
	return []string{initCmd}
}

func (r *RedPandaClusterReconciler) initContainerVolumeMounts(volumes map[string]*corev1.Volume) []corev1.VolumeMount {

	baseConfigVolMount := corev1.VolumeMount{
		Name:      volumes["configMap"].Name,
		MountPath: "/tmp/base-config",
	}

	configVolMount := corev1.VolumeMount{
		Name:      volumes["config"].Name,
		MountPath: "/tmp/config",
	}
	return []corev1.VolumeMount{baseConfigVolMount, configVolMount}

}

func (r *RedPandaClusterReconciler) getRPArgs() []string {
	return []string{"--config /tmp/config/redpanda.yaml", "start", "--", "--default-log-level=debug"}
}

func (r *RedPandaClusterReconciler) getPorts() []corev1.ContainerPort {
	adminContainerPort := corev1.ContainerPort{
		Name:          "admin",
		ContainerPort: 9644,
	}
	kafkaContainerPort := corev1.ContainerPort{
		Name:          "kafka",
		ContainerPort: 9092,
	}
	rpcContainerPort := corev1.ContainerPort{
		Name:          "rpc",
		ContainerPort: 33145,
	}

	return []corev1.ContainerPort{adminContainerPort, kafkaContainerPort, rpcContainerPort}
}

func (r *RedPandaClusterReconciler) getRPVolumeMounts(volMap map[string]*corev1.Volume) []corev1.VolumeMount {

	configMount := corev1.VolumeMount{
		Name:      volMap["config"].Name,
		MountPath: "/tmp/config",
	}

	return []corev1.VolumeMount{configMount}
}

func (r *RedPandaClusterReconciler) rpVolClaimTemplates() []corev1.PersistentVolumeClaim {

	storageClassName := "standard"
	claim := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: "redpanda-data",
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{"ReadWriteOnce"},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{"storage": *resource.NewQuantity(1, "G")},
			},
			StorageClassName: &storageClassName,
		},
	}
	return []corev1.PersistentVolumeClaim{claim}
}
