package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"reflect"
	"time"

	coreapi "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var blackListedNamespaces = []string{
	"openshift-console", // This may remove console deployment
	"openshift-etcd",    // This may kill etcd pod and cause outage
	"openshift-ingress", // This may remove ingress pods and backend would stop responding
}

func inClusterLogin() (*k8s.Clientset, error) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	// Seed random
	rand.Seed(time.Now().Unix())

	// creates the clientset
	return k8s.NewForConfig(config)
}

func getRandomNamespace(c *k8s.Clientset) (string, error) {
	if len(os.Getenv("NAMESPACE")) != 0 {
		namespace := os.Getenv("NAMESPACE")
		log.Println(fmt.Sprintf("Namespace override found: %s", namespace))
		return namespace, nil
	}
	log.Println("Fetching available namespaces")
	nms, err := c.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil || nms.Items == nil || len(nms.Items) == 0 {
		return "", fmt.Errorf("Failed to list namespaces: %v", err)
	}

	namespacesMap := map[string]bool{}
	for _, n := range nms.Items {
		namespacesMap[n.Name] = true
	}

	if len(os.Getenv("NO_BLACKLIST")) != 0 {
		// Remove blacklisted namespaces
		for n := range blackListedNamespaces {
			delete(namespacesMap, blackListedNamespaces[n])
		}
	}

	// Leave CVO alone
	delete(namespacesMap, "openshift-cluster-version")

	// Get a slice of keys
	keys := reflect.ValueOf(namespacesMap).MapKeys()
	randomNamespace := keys[rand.Intn(len(keys))].String()
	log.Println(fmt.Sprintf("random namespace: %v", randomNamespace))
	return randomNamespace, nil
}

func killRandomPod(c *k8s.Clientset) (string, error) {
	log.Println("Killing random pod")
	// Find random namespace
	randomNamespace, err := getRandomNamespace(c)
	if err != nil {
		return "", fmt.Errorf("Failed to fetch available namespaces: %v", err)
	}

	// Find random pod
	pods, err := c.CoreV1().Pods(randomNamespace).List(metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("Failed to list pods: %v", err)
	}
	if pods.Items == nil || len(pods.Items) == 0 {
		return "", fmt.Errorf("No pods fetched")
	}
	randomPod := pods.Items[rand.Intn(len(pods.Items))]

	switch randomPod.Status.Phase {
	case coreapi.PodFailed, coreapi.PodSucceeded, coreapi.PodUnknown:
		return "", fmt.Errorf("Random pod %s is in phase '%s'", randomPod.Name, randomPod.Status.Phase)
	}

	// Kill kill kill
	if err := c.CoreV1().Pods(randomPod.Namespace).Delete(randomPod.Name, &metav1.DeleteOptions{}); err != nil {
		return "", fmt.Errorf("Failed to kill pods: %v", err)
	}
	return fmt.Sprintf("Killed pod %s in namespace %s", randomPod.Name, randomPod.Namespace), nil
}

func killRandomDeployment(c *k8s.Clientset) (string, error) {
	log.Println("Killing random Deployment")

	// Find random namespace
	randomNamespace, err := getRandomNamespace(c)
	if err != nil {
		return "", fmt.Errorf("Failed to fetch available namespaces: %v", err)
	}

	// Find random pod
	deployments, err := c.AppsV1().Deployments(randomNamespace).List(metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("Failed to list deployments: %v", err)
	}
	if deployments.Items == nil || len(deployments.Items) == 0 {
		return "", fmt.Errorf("No deployments fetched")
	}
	randomDeployment := deployments.Items[rand.Intn(len(deployments.Items))]

	// Kill kill kill
	if err := c.AppsV1().Deployments(randomDeployment.Namespace).Delete(randomDeployment.Name, &metav1.DeleteOptions{}); err != nil {
		return "", fmt.Errorf("Failed to delete deployment: %v", err)
	}
	return fmt.Sprintf("Killed deployment %s in namespace %s", randomDeployment.Name, randomDeployment.Namespace), nil
}

func killRandomStatefulSet(c *k8s.Clientset) (string, error) {
	log.Println("Killing random StatefulSet")

	// Find random namespace
	randomNamespace, err := getRandomNamespace(c)
	if err != nil {
		return "", fmt.Errorf("Failed to fetch available namespaces: %v", err)
	}

	// Find random pod
	statefulSets, err := c.AppsV1().StatefulSets(randomNamespace).List(metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("Failed to list statefulsets: %v", err)
	}
	if statefulSets.Items == nil || len(statefulSets.Items) == 0 {
		return "", fmt.Errorf("No Stateful Sets fetched")
	}
	randomStatefulSet := statefulSets.Items[rand.Intn(len(statefulSets.Items))]

	// Kill kill kill
	if err := c.AppsV1().StatefulSets(randomStatefulSet.Namespace).Delete(randomStatefulSet.Name, &metav1.DeleteOptions{}); err != nil {
		return "", fmt.Errorf("Failed to kill statefulset: %v", err)
	}
	return fmt.Sprintf("Killed statefulset %s in namespace %s", randomStatefulSet.Name, randomStatefulSet.Namespace), nil
}
