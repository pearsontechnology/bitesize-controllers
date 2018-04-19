package kubernetes

import (
    "strings"
    "time"
    "encoding/json"
    "encoding/base64"
    log "github.com/Sirupsen/logrus"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/types"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
)

type PatchSpec struct {
        Op    string `json:"op"`
        Path  string `json:"path"`
        Value string `json:"value"`
}

func listOptions(value string) metav1.ListOptions {
    return metav1.ListOptions{
        LabelSelector: "name=" + value,
    }
}

func getClient()(*kubernetes.Clientset, error) {
    clientset := &kubernetes.Clientset{}

    config, err := rest.InClusterConfig()
    if err != nil {
        log.Fatalf("Failed to create client: %v", err.Error())
    }

    clientset, err = kubernetes.NewForConfig(config)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err.Error())
    }

    return clientset, err
}

func GetPods(label string, namespace string) (pods *corev1.PodList, err error) {
    pods = &corev1.PodList{}

    clientset, err := getClient()
    if err == nil {
        pods, err = clientset.CoreV1().Pods(namespace).List(listOptions(label))
        log.Debugf("GetPods found: %v pods in %v with label %v", len(pods.Items), namespace, label)
        if err != nil {
            log.Infof("Error GetPods: %v", err.Error())
        }
        return pods, err
    } else {
        log.Infof("Error GetPods.getClient: %v", err.Error())
        return nil, err
    }
}

func GetPodIps(label string, namespace string) (instanceList map[string]string, err error) {

    var m = make(map[string]string)
    i := 0
    pods := &corev1.PodList{}

    pods, err = GetPods(label, namespace)
    for _, pod := range pods.Items {
        log.Debugf("Pod found: %v", pod.ObjectMeta.Name)
        switch pod.Status.Phase {
        case "Pending":
            m[pod.ObjectMeta.Name] = ""
        case "Running":
            m[pod.ObjectMeta.Name] = pod.Status.PodIP
            i++
        case "Failed","Unknown":
            m[pod.ObjectMeta.Name] = "error"
        }

    }
    log.Debugf("GetPodIps found: %v", i)
    return m, err
}

func DeletePod(podName string, namespace string) (err error) {

    clientset, err := getClient()
    log.Debugf("Deleting pod: %v", podName)
    //options := metav1.DeleteOptions{}
    err = clientset.CoreV1().Pods(namespace).Delete(podName, nil)
    time.Sleep(30 * time.Second)
    if err != nil {
        log.Errorf("Error deleting pod %v: %v", podName, err.Error())
        return err
    } else {
        return err
    }
}

func GetSecret(secretName string, secretKey string, namespace string) (secretValue string) {
    log.Debugf("GetSecret: %v/%v:%v", namespace,secretName,secretKey)
    clientset, err := getClient()
    secret, err := clientset.CoreV1().Secrets(namespace).Get(secretName, metav1.GetOptions{})
    if err != nil {
        log.Errorf("Error retrieving secret %v: %v", secretName, err.Error())
        return ""
    }
    secretValue = ""
    for name, data := range secret.Data {
        if name == secretKey {
            str := string(data[:])
            secretValue = strings.TrimSpace(str)
            log.Infof("GetSecret found for %v", secretName)
            log.Debugf("Found secret: %s", secretValue)
        } else {
            log.Debugf("GetSecret not matched: %v", secretKey)
        }
    }
    return secretValue
}

func PutSecret(secretName string, secretKey string, secretValue string, namespace string) (err error) {

    //If Decode fails assume it's already Base64
    _, err = base64.StdEncoding.DecodeString(secretValue)
    if err != nil {
        secretValue = base64.StdEncoding.EncodeToString([]byte(secretValue))
    }

    s := map[string]string{
        secretKey: secretValue,
    }

    secretData := &corev1.Secret{
        ObjectMeta: metav1.ObjectMeta{
            Name:      secretName,
            Namespace: namespace,
        },
        StringData: s,
    }

    clientset, err := getClient()

    secrets, err := clientset.CoreV1().Secrets(namespace).List(metav1.ListOptions{})
    if err != nil {
        log.Errorf("Error retrieving secrets %v:", err.Error())
    }
    found := false
    for _, sec := range secrets.Items {
        if sec.ObjectMeta.Name == secretName {
            found = true
        }
    }
    if found == false {
        log.Debugf("Creating secretData: %v", secretData)
        _, err := clientset.CoreV1().Secrets(namespace).Create(secretData)
        if err != nil {
            log.Errorf("Error Creating secret %v:%v", secretName, err.Error())
        }
    } else {
        patchData := make([]PatchSpec, 1)
        patchData[0].Op = "add"
        patchData[0].Path = "/data/" + secretKey
        patchData[0].Value = secretValue
        patchBytes, err := json.Marshal(patchData)
        if err != nil {
            log.Errorf("Error formatting patch %v:%v", patchData, err.Error())
        }
        log.Debugf("Patching secretData: %v, %v, %v", secretName, types.JSONPatchType, secretValue)
        _, err = clientset.CoreV1().Secrets(namespace).Patch(secretName, types.JSONPatchType, patchBytes)
        if err != nil {
            log.Errorf("Error Patching secret %v: %v", secretName, err.Error())
        }
    }

    return err
}
