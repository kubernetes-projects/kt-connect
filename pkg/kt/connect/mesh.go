package connect

import (
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Mesh prepare swap deployment
func Mesh(swap string, options *options.DaemonOptions, clientset *kubernetes.Clientset, labels map[string]string) (workload string, err error) {
	workload, podIP, podName, err := createMeshShadown(swap, clientset, labels, options.Namespace, options.Image)
	if err != nil {
		return
	}
	options.RuntimeOptions.Shadow = workload
	remotePortForward(options.MeshOptions.Expose, options.KubeConfig, options.Namespace, podName, podIP, options.Debug)
	return
}

func createMeshShadown(
	swap string,
	clientset *kubernetes.Clientset,
	extraLabels map[string]string,
	namespace string, image string,
) (shadowName string, podIP string, podName string, err error) {
	deploymentsClient := clientset.AppsV1().Deployments(namespace)
	origin, err := deploymentsClient.Get(swap, metav1.GetOptions{})
	if err != nil {
		return
	}

	meshVersion := strings.ToLower(util.RandomString(5))
	shadowName = origin.GetObjectMeta().GetName() + "-kt-" + meshVersion
	labels := map[string]string{
		"kt":           shadowName,
		"kt-component": "mesh",
		"control-by":   "kt",
		"version":      meshVersion,
	}
	for k, v := range extraLabels {
		labels[k] = v
	}
	for k, v := range origin.Spec.Selector.MatchLabels {
		labels[k] = v
	}

	podIP, podName, err = cluster.CreateShadow(clientset, shadowName, labels, namespace, image)
	if err != nil {
		return "", "", "", err
	}

	log.Printf("-----------------------------------------------------------\n")
	log.Printf("|    Mesh Version '%s' You can update Istio rule       |\n", meshVersion)
	log.Printf("-----------------------------------------------------------\n")

	return
}
