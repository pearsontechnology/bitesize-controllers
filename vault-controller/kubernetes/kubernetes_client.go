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

func listOptions(value string) metav1.ListOptions {
	return metav1.ListOptions{
		LabelSelector: "name=" + value,
	}
}

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

    clientset, err := getClient()
    if err == nil {
        pods, err := clientset.CoreV1().Pods(namespace).List(listOptions(label))
        log.Debugf("GetPods found: %v pods in %v with label %v", len(pods.Items), namespace, label)
        if err != nil {
            log.Infof("Error GetPods: %v", err)
        }
        return pods, err
    } else {
        log.Infof("Error GetPods.getClient: %v", err)
        return nil, err
    }
}

func GetPodIps(label string, namespace string) (instanceList map[string]string, err error) {

    pods := &corev1.PodList{}

    pods, err = GetPods(label, namespace)
    for _, pod := range pods.Items {
        log.Debugf("Pod found: %v", pod.ObjectMeta.Name)
        instanceList[pod.ObjectMeta.Name] = pod.Status.PodIP
    }
    log.Debugf("GetPodIps found: %v", len(instanceList))
    return instanceList, err
}

func DeletePod(podName string, namespace string) (err error) {

    clientset, err := getClient()
    log.Debugf("Deleting pod: %v", podName)
    //options := metav1.DeleteOptions{}
    err = clientset.CoreV1().Pods(namespace).Delete(podName, nil)
    if err != nil {
        log.Errorf("Error deleting pod %v: %v", podName, err)
        return err
    } else {
        return err
    }
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
    if len(secretValue) > 0 {
        log.Debugf("GetSecret found for %v", secretName)
    } else {
        log.Debugf("GetSecret not found for %v", secretName)
    }
    return secretValue
}
