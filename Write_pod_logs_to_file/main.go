// Note: the example only works with the code within the same release/branch.
package main

import (
	"bufio"
	"context"
	"flag"
	"io"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"log"
	"os"
	"path/filepath"
)

const (
	// set namespace and label
	namespace = "demo"
	label     = "broker=set"
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

	// use the current context in kubeconfig
	ctx := context.TODO()
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Println(err, "Failed to build config from flags")
		return
	}

	err = collectApplicationLogs(ctx, config, "/home/raka/logs.txt")
	if err != nil {
		log.Println(err, "Failed to collect logs")
	}

}

func collectApplicationLogs(ctx context.Context, config *rest.Config, filename string) error {
	// create the clientset
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Println("Failed to create clientset from the given config")
		return err
	}
	// get the pods as ListItems
	pods, err := clientSet.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: label,
	})
	if err != nil {
		log.Println(err, "Failed to get pods")
		return err
	}
	// If the file doesn't exist, create it or append to the file
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	// get the pod lists first
	// then get the podLogs from each of the pods
	// write to files concurrently
	// use channel for blocking reasons
	ch := make(chan bool)
	podItems := pods.Items
	for i := 0; i < len(podItems); i++ {
		podLogs, err := clientSet.CoreV1().Pods(namespace).GetLogs(podItems[i].Name, &v1.PodLogOptions{
			Follow: true,
		}).Stream(ctx)
		if err != nil {
			return err
		}
		buffer := bufio.NewReader(podLogs)
		go writeLogs(buffer, file, ch)
	}
	<-ch
	return nil
}

func writeLogs(buffer *bufio.Reader, file *os.File, ch chan bool) {
	defer func() {
		ch <- true
	}()
	for {
		str, readErr := buffer.ReadString('\n')
		if readErr == io.EOF {
			break
		}
		_, err := file.Write([]byte(str))
		if err != nil {
			return
		}
	}
}
