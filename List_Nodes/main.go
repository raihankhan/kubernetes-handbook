package main

import (
	"context"
	"flag"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	cu "kmodules.xyz/client-go/client"
)

func main() {
	// parse the .kubeconfig file
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}

	// create the kubebuilder uncached client
	client, err := cu.NewUncachedClient(config)
	if err != nil {
		klog.Error("Failed to create kubebuilder client", err)
		return
	}

	nodeList := &corev1.NodeList{}
	err = client.List(context.TODO(), nodeList)
	if err != nil {
		klog.Error("Failed to Get node list", err)
		return
	}

	fmt.Println("List of nodes in the cluster (Fetched using Kubebuilder Client)")
	for _, node := range nodeList.Items {
		fmt.Printf("%s\n", node.Name)
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Error("Failed to create kubernetes client", err)
		return
	}

	nodeList, err = clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		klog.Error("Failed to Get node list", err)
		return
	}

	fmt.Println("List of nodes in the cluster (Fetched using Kubernetes Client)")
	for _, node := range nodeList.Items {
		fmt.Printf("%s\n", node.Name)
	}

}
