package kubernetes

import (
    "strings"
    "fmt"
    log "github.com/Sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
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

func GetPods(label string, namespace string) (pods *corev1.PodList, err error) {

    pods = &corev1.PodList{}

    clientset, err := getClient()
    if err != nil {
        pods, _ := clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{})

        for _, pod := range pods.Items {
            log.Debugf(pod.Name, pod.Status.PodIP)
        }
    }

    return pods, err
}

func GetPodIps(label string, namespace string) (podIps []string, err error) {

    pods, err := GetPods(label, namespace)

    for _, pod := range pods.Items {
        podIps = append(podIps, pod.Status.PodIP)
    }

    return podIps, err
}

func GetSecret(secretName, string, secretKey string, namespace string)(secretValue string) {

    clientset, err := getClient()
    secret, err := clientset.CoreV1().Secrets(namespace).Get(secretName, metav1.GetOptions{})
    if err != nil {
        log.Errorf("Error retrieving secret %v: %v", secretName, err)
        return ""
    }

    for name, data := range secret.Data {
        if name == secretKey {
            s := fmt.Sprint(data)
            secretValue = strings.TrimSpace(s)
            log.Infof("Found secret: %s", name)
        } else {
            secretValue = ""
        }
    }

    return secretValue
}
