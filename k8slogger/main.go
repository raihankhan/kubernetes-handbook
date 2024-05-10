package main

import (
	"fmt"
	"k8s.io/klog/v2/textlogger"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func main() {
	log.SetLogger(textlogger.NewLogger(textlogger.NewConfig(textlogger.Verbosity(1))))
	fmt.Println(ctrl.Log.GetV())
	ctrl.Log.Info("Hello")
}
