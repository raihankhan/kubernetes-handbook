package main

import (
	"context"
	"flag"
	"fmt"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/utils/pointer"
	"log"
	"path/filepath"
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

	// create the clientset
	//clientSet, err := kubernetes.NewForConfig(config)
	//if err != nil {
	//	log.Println("Failed to create clientset from the given config")
	//	return
	//}
	// 1. Create custom crdClientSet
	// here restConfig is your .kube/config file
	crdClientSet, err := clientset.NewForConfig(config)
	if err != nil {
		return
	}

	// 2. List down all the existing crd in the cluster
	crdList, err := crdClientSet.ApiextensionsV1().CustomResourceDefinitions().List(ctx, metav1.ListOptions{})
	if err != nil {
		return
	}

	// 3. get new empty schema holder
	scheme := runtime.NewScheme()

	// 4. loop over all the crd and add to the schema
	for _, crd := range crdList.Items {
		for _, v := range crd.Spec.Versions {
			//fmt.Printf("GROUP = %s        VERSION = %s     KIND = %s\n", crd.Spec.Group, v.Name, crd.Spec.Names.Kind)
			scheme.AddKnownTypeWithName(
				schema.GroupVersionKind{
					Group:   crd.Spec.Group,
					Version: v.Name,
					Kind:    crd.Spec.Names.Kind,
				},
				&unstructured.Unstructured{},
			)
		}
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		fmt.Println("Failed to get dynamic client")
		return
	}

	dbGVR := schema.GroupVersionResource{
		Group:    "mysql.presslabs.org",
		Version:  "v1alpha1",
		Resource: "mysqlclusters",
	}

	db := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "mysql.presslabs.org/v1alpha1",
			"kind":       "MysqlCluster",
			"metadata": map[string]interface{}{
				"name":      "mysql-sample",
				"namespace": "db",
			},
			"spec": map[string]interface{}{
				"replicas":     pointer.Int32(3),
				"secretName":   "mysql-cred",
				"mysqlVersion": "8.0",
			},
		},
	}

	//db := v1alpha1.MysqlCluster{
	//	TypeMeta: metav1.TypeMeta{
	//		Kind:       "MysqlCluster",
	//		APIVersion: "mysql.presslabs.org/v1alpha1",
	//	},
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name:      "mysql-sample",
	//		Namespace: "db",
	//	},
	//	Spec: v1alpha1.MysqlClusterSpec{
	//		Replicas:     pointer.Int32(3),
	//		SecretName:   "mysql-cred",
	//		MysqlVersion: "8.0", // 8.0.32
	//		//Image:                          "",
	//		//InitBucketURL:                  "",
	//		//InitBucketURI:                  "",
	//		//InitBucketSecretName:           "",
	//		//MinAvailable:                   "",
	//		//BackupSchedule:                 "",
	//		//BackupURL:                      "",
	//		//BackupRemoteDeletePolicy:       "",
	//		//BackupSecretName:               "",
	//		//BackupScheduleJobsHistoryLimit: nil,
	//		//MysqlConf:                      nil,
	//		//PodSpec:                        v1alpha1.PodSpec{},
	//		//VolumeSpec:                     v1alpha1.VolumeSpec{},
	//		//TmpfsSize:                      nil,
	//		//MaxSlaveLatency:                nil,
	//		//QueryLimits:                    nil,
	//		//ReadOnly:                       false,
	//		//ServerIDOffset:                 nil,
	//		//BackupCompressCommand:          nil,
	//		//BackupDecompressCommand:        nil,
	//		//MetricsExporterExtraArgs:       nil,
	//		//RcloneExtraArgs:                nil,
	//		//XbstreamExtraArgs:              nil,
	//		//XtrabackupExtraArgs:            nil,
	//		//XtrabackupPrepareExtraArgs:     nil,
	//		//XtrabackupTargetDir:            "",
	//		//InitFileExtraSQL:               nil,
	//	},
	//	Status: v1alpha1.MysqlClusterStatus{},
	//}

	//dbRaw, err := json.Marshal(db)
	//if err != nil {
	//	fmt.Println("Failed to convert db object to raw json")
	//	return
	//}
	//
	//var mapObj map[string]interface{}
	//err = json.Unmarshal(dbRaw, &mapObj)
	//if err != nil {
	//	fmt.Println("Failed to convert raw object to map")
	//	return
	//}
	//
	//dbUnstructured := unstructured.Unstructured{
	//	Object: mapObj,
	//}
	//
	//fmt.Println(dbUnstructured)

	_, err = dynamicClient.Resource(dbGVR).Namespace("db").Create(context.TODO(), db, metav1.CreateOptions{})
	if err != nil {
		fmt.Println("Failed to create db")
		fmt.Println(err)
		return
	}

	fmt.Println("Successfully created db")
}
