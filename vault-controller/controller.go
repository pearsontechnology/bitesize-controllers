package main

import (
    "os"
    log "github.com/Sirupsen/logrus"
    vault "github.com/pearsontechnology/bitesize-controllers/vault-controller/vault"
    k8s "github.com/pearsontechnology/bitesize-controllers/vault-controller/kubernetes"
)
const version = "0.1"

func init() {

}

const defaultNameSpace = "kube-system"
const defaultVaultLabel = "vault"
const defaultVaultPort = "8243"

func main() {
    var err error

    if os.Getenv("DEBUG") == "true" {
        log.SetLevel(log.DebugLevel)
    }

	  log.Infof("Starting vault controller version: %s", version)

    serviceDomain := os.Getenv("SERVICE_DOMAIN")
    if serviceDomain == "" {
        serviceDomain = "svc.cluster.local"
    }

    vaultLabel := os.Getenv("VAULT_LABEL")
    if vaultLabel == "" {
        vaultLabel = defaultVaultLabel
    }

    vaultNamespace := os.Getenv("VAULT_NAMESPACE")
    if vaultNamespace == "" {
        vaultNamespace = defaultNameSpace
    }

    vaultPort := os.Getenv("VAULT_PORT")
    if vaultPort == "" {
        vaultPort = defaultVaultPort
    }
    vaultToken := os.Getenv("VAULT_TOKEN")

    onKubernetes := true

    if os.Getenv("KUBERNETES_SERVICE_HOST") == "" {
        log.Info("WARN: NOT running on Kubernetes, pod discovery DISABLED.")
        onKubernetes = false
    }

    vaultAddress := os.Getenv("VAULT_ADDR")
    if vaultAddress == "" && onKubernetes == true {
        vaultAddress = "https://vault." + vaultNamespace + "." + serviceDomain + ":" + vaultPort
    }

    if vaultAddress == "" {
        vaultAddress = "http://localhost:8200"
    }

    vaultInstances := os.Getenv("VAULT_INSTANCES")

    if vaultInstances == "" and onKubernetes == false {
        log.Errorf("Invalid value for env var VAULT_INSTANCES: %v", vaultInstances)
    } else if vaultInstances == "" and onKubernetes == true {
        log.Info("Proceeding with pod discovery on %v", vaultLabel)
        pods, err := k8s.GetPods(onKubernetes, vaultLabel)
        for _, pod := range pods.Items {
            log.Debug(pod.Name, pod.Status.PodIP)
        }
    } else {
        log.Info("Proceeding with pod discovery on %v and VAULT_INSTANCES: %v", vaultLabel,vaultInstances)
        pods, err := k8s.GetPods(onKubernetes, vaultLabel
    }

    unsealKeys := os.Getenv("UNSEAL_KEYS")
    if unsealKeys == "" {
        log.Errorf("Invalid value for env var UNSEAL_KEYS: %v", unsealKeys)
    }

    log.Debug("Connecting to vault at: %v", vaultAddress)
    vaultClient, err := vault.NewVaultClient(vaultAddress, vaultToken)

    if err != nil {
        log.Errorf("Error connecting to vault at: %v", vaultAddress)
    }



    }

}
