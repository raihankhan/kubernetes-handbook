package main

import (
	"flag"
	"fmt"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"
	"path/filepath"
	"time"
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

	// create config from the kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}

	// create the clientset
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	// stop signal for the informer
	stopper := make(chan struct{})
	defer close(stopper)

	namespace := ""
	// create shared informers for resources in all known API group versions with a reSync period and namespace
	factory := informers.NewSharedInformerFactoryWithOptions(clientSet, 10*time.Second, informers.WithNamespace(namespace))
	podInformer := factory.Apps().V1().Deployments().Informer()

	defer runtime.HandleCrash()

	// start informer ->
	go factory.Start(stopper)

	// start to sync and call list
	if !cache.WaitForCacheSync(stopper, podInformer.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    onAdd, // register add eventhandler
		UpdateFunc: onUpdate,
		DeleteFunc: onDelete,
	})

	// block the main go routine from exiting
	<-stopper
}

func onAdd(obj interface{}) {
	depl := obj.(*v1.Deployment)
	klog.Infof("POD CREATED: %s/%s", depl.Namespace, depl.Name)
}

//func onUpdate(oldObj interface{}, newObj interface{}) {
//	oldDepl := oldObj.(*v1.Deployment)
//	newDepl := newObj.(*v1.Deployment)
//
//	if oldDepl.Spec.Template.Spec.Containers[0].Image != newDepl.Spec.Template.Spec.Containers[0].Image {
//		klog.Infof(
//			"CONTAINER IMAGE UPDATED FROM %s to %s",
//			oldDepl.Spec.Template.Spec.Containers[0].Image, newDepl.Spec.Template.Spec.Containers[0].Image,
//		)
//	}
//}

func onDelete(obj interface{}) {
	depl := obj.(*v1.Deployment)
	klog.Infof("POD DELETED: %s/%s", depl.Namespace, depl.Name)
}

func onUpdate(oldObj interface{}, newObj interface{}) {
	oldDepl := oldObj.(*v1.Deployment)
	newDepl := newObj.(*v1.Deployment)

	for oldContainerID := range oldDepl.Spec.Template.Spec.Containers {
		for newContainerID := range newDepl.Spec.Template.Spec.Containers {
			if oldDepl.Spec.Template.Spec.Containers[oldContainerID].Name == newDepl.Spec.Template.Spec.Containers[newContainerID].Name {
				if oldDepl.Spec.Template.Spec.Containers[oldContainerID].Image != newDepl.Spec.Template.Spec.Containers[newContainerID].Image {
					klog.Infof(
						"CONTAINER IMAGE UPDATED FROM %s to %s",
						oldDepl.Spec.Template.Spec.Containers[oldContainerID].Image, newDepl.Spec.Template.Spec.Containers[newContainerID].Image,
					)
				}
			}
		}
	}
}
