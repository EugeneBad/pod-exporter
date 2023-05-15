package main

import (
	"context"
	"strings"
	"time"

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
			Name: "pods_recent_count_current",
			Help: "The total number of pods older or younger than 7 days",
		},
		[]string{"valid"},
	)
)

func (cset *ClientSet) getPods() (*corev1.PodList, error) {
	// read in-cluster a Kubernetes config.
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Errorf("failed to get Kubernetes config: %v", err)
		return nil, err
	}
	// clientset using the incluster config (requires rbac)
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Errorf("failed to create Kubernetes clientset: %v", err)
		return nil, err

	}

	// list all pods in the defined namespace
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
	// initialise pod count
	var recentPods, oldPods int
	for _, pod := range pods.Items {
		log.WithFields(
			log.Fields{
				"pod": pod.Name,
				"rule_evaluation": []map[string]interface{}{
					{"name": "image_prefix", "valid": checkImagePrefix(pod)},
					{"name": "team_label_present", "valid": checkTeamLabel(pod)},
					{"name": "recent_start_time", "valid": checkStartTime(pod)}},
			},
		).Info("pod scrape successful")
		// count pods depending on recency
		if checkStartTime(pod) {
			recentPods++
		} else {
			oldPods++
		}
	}
	// update prometheus metric
	recentPodCount.WithLabelValues("true").Set(float64(recentPods))
	recentPodCount.WithLabelValues("false").Set(float64(oldPods))
	return nil
}

// check the prefix on container image name
// returns false if image name doesn't match for any container in pod
func checkImagePrefix(pod corev1.Pod) bool {
	for _, container := range pod.Spec.Containers {
		imageParts := strings.Split(container.Image, "/")
		if !(len(imageParts) > 1 && imageParts[0] == "bitnami") {
			return false
		}
	}
	return true
}

func checkTeamLabel(pod corev1.Pod) bool {
	if _, exists := pod.ObjectMeta.Labels["team"]; exists {
		return true
	}
	return false
}

func checkStartTime(pod corev1.Pod) bool {
	startTime := pod.ObjectMeta.CreationTimestamp.Time
	// resolution in hours
	age := time.Since(startTime)
	if age > (7 * 24 * time.Hour) {
		return false
	}
	return true
}
