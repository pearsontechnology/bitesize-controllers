package kubernetes


import (
    "fmt"
    "os"
    "time"
    log "github.com/Sirupsen/logrus"
    "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
)

func getClient()(*kubernetes.Clientset, error) {
    var err error

    config, err := rest.InClusterConfig()
    if err != nil {
        log.Fatalf("Failed to create client: %v", err.Error())
    }

    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err.Error())
    }

    return clientset, err
}

func GetPods(label string, namespace string) (pods *kubernetes.PodList, err error) {

    pods = &v1Core.PodsList{}
    var err error

    clientset, err := getClient()
    if err != nil {
        pods, err := clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{})
        for _, pod := range pods.Items {
            log.Debug(pod.Name, pod.Status.PodIP)
        }
    }

    return pods, err
}

func GetSecret(secretKey string, namespace string)(secretValue string) {

    namespace := os.Getenv("POD_NAMESPACE")

    clientset = getClient()

    secrets, err := clientset.CoreV1().Secrets(namespace).Get(secretKey)

    if err != nil {
        log.Errorf("Error retrieving secrets: %v", err)
        return ""
    }

    for name, data := range secrets.Data {
        //secret[name] = string(data)
        if name == secretKey {
            secretValue = string(data)
            log.Infof("Found secret: %s", name)
        } else {
            secretValue = ""
        }
    }

    return secretValue
}
