package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	kubeconfig := flag.String("kubeconfig", "", "Path to a kube config. Only required if out-of-cluster.")
	flag.Parse()

	// Create the client config. Use kubeconfig if given, otherwise assume in-cluster.
	config, err := buildConfig(*kubeconfig)
	if err != nil {
		log.Fatal(err)
	}

	// make a new config for our extension's API group, using the first config as a baseline
	appClient, appScheme, err := NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	// start a controller on instances of our custom resource
	controller := AppController{
		AppClient: appClient,
		AppScheme: appScheme,
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	go controller.Run(ctx)

	err = wait.PollInfinite(5*time.Second, func() (bool, error) {
		list := AppList{}
		err := appClient.Get().Resource("apps").Do().Into(&list)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("LIST: %v\n", list)

		return false, nil
	})
}

func buildConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}
