package kubernetes

import (
	log "github.com/Sirupsen/logrus"
	"k8s.io/client-go/1.4/kubernetes"
	"k8s.io/client-go/1.4/pkg/api"
	"k8s.io/client-go/1.4/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/1.4/rest"

	"k8s.io/kubernetes/pkg/util/flowcontrol"
)

func getClient() (*kubernetes.Clientset, error) {
	var err error

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Failed to get cluster config: %v", err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err.Error())
	}

	return clientset, err
}

func GetIngresses(onKubernetes bool) (*v1beta1.IngressList, error) {

	ingresses := &v1beta1.IngressList{}
	var err error

	if onKubernetes == true {
		clientset, _ := getClient()

		rateLimiter := flowcontrol.NewTokenBucketRateLimiter(0.1, 1)

		rateLimiter.Accept()
		ingresses, err = clientset.Extensions().Ingresses("").List(api.ListOptions{})

	}
	return ingresses, err
}
