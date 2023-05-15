package main

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	log "github.com/sirupsen/logrus"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type ClientSet struct {
	context   context.Context
	namespace string
}

func NewClientSet(ctx context.Context, namespace string) *ClientSet {
	return &ClientSet{
		context:   ctx,
		namespace: namespace,
	}
}

var (
	// Initialise the prometheus gauge metric to count recent pods
	recentPodCount = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pods_recent_count_total",
			Help: "The total number of pods older or younger than 7 days",
		},
		[]string{"valid"},
	)
)

func (cset *ClientSet) getPods() (*corev1.PodList, error) {
	// create a Kubernetes clientset using the default kubeconfig file
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Errorf("failed to get Kubernetes config: %v", err)
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Errorf("failed to create Kubernetes clientset: %v", err)
		return nil, err

	}

	// list all pods in the default namespace
	pods, err := clientset.CoreV1().Pods(cset.namespace).List(cset.context, metav1.ListOptions{})
	if err != nil {
		log.Errorf("failed to list pods: %v", err)
		return nil, err
	}
	return pods, nil

}

func (cset *ClientSet) countPods() error {
	pods, err := cset.getPods()
	if err != nil {
		return err
	}
	for _, pod := range pods.Items {
		fmt.Printf("Pod name: %s\n", pod.Name)
		fmt.Printf("Pod status: %s\n", pod.Status.Phase)
	}
	return nil
}
