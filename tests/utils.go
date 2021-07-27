//+build e2e

package tests

import (
	"flag"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	lockclient "github.com/petrkotas/k8s-object-lock/pkg/generated/clientset/versioned"
)

var kubeconfig string

func init() {

	if flag.CommandLine.Lookup("kubeconfig") == nil {
		flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	}

}

func parseFlags() {
	if home := homedir.HomeDir(); home != "" {
		flag.StringVar(&kubeconfig, "kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	}
}

func createClientSets(kubeconfig string) (*kubernetes.Clientset, *lockclient.Clientset) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err)
	}

	kubeClientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	lockClientset, err := lockclient.NewForConfig(config)

	return kubeClientset, lockClientset
}
