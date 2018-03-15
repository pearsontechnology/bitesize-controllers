package main

import (
    "os"
    log "github.com/Sirupsen/logrus"
    vault "github.com/pearsontechnology/bitesize-controllers/vault-controller/vault"
    k8s "github.com/pearsontechnology/bitesize-controllers/vault-controller/kubernetes"
)
const version = "0.1"
const defaultNameSpace = "kube-system"
const defaultVaultLabel = "vault"
const defaultVaultPort = "8243"

func init() {

}

func main() {
    var err error

    // init stuff
    if os.Getenv("DEBUG") == "true" {
        log.SetLevel(log.DebugLevel)
        log.Debugf("DebugLevel on")
    }

	log.Infof("Starting vault controller version: %s", version)

    serviceDomain := os.Getenv("SERVICE_DOMAIN")
    if serviceDomain == "" {
        serviceDomain = "svc.cluster.local"
    }
    log.Debugf("serviceDomain: %v", serviceDomain)

    vaultLabel := os.Getenv("VAULT_LABEL")
    if vaultLabel == "" {
        vaultLabel = defaultVaultLabel
    }
    log.Debugf("vaultLabel: %v", vaultLabel)

    vaultNamespace := os.Getenv("VAULT_NAMESPACE")
    if vaultNamespace == "" {
        vaultNamespace = defaultNameSpace
    }
    log.Debugf("vaultNamespace: %v", vaultNamespace)

    vaultPort := os.Getenv("VAULT_PORT")
    if vaultPort == "" {
        vaultPort = defaultVaultPort
    }
    log.Debugf("vaultPort: %v", vaultPort)

    // don't default token
    vaultToken := os.Getenv("VAULT_TOKEN")
    log.Debugf("vaultToken: %v", vaultToken)

    onKubernetes := true

    if os.Getenv("KUBERNETES_SERVICE_HOST") == "" {
        log.Info("WARN: NOT running on Kubernetes, pod discovery DISABLED.")
        onKubernetes = false
    }
    log.Debugf("onKubernetes: %v", onKubernetes)

    vaultAddress := os.Getenv("VAULT_ADDR")
    if vaultAddress == "" && onKubernetes == true {
        vaultAddress = "https://vault." + vaultNamespace + "." + serviceDomain + ":" + vaultPort
    }

    // worst-case scenario
    if vaultAddress == "" {
        vaultAddress = "http://localhost:8200"
    }
    log.Debugf("vaultAddress: %v", vaultAddress)

    vaultInstances := os.Getenv("VAULT_INSTANCES")

    if vaultInstances == "" and onKubernetes == false {
        log.Errorf("Invalid value for env var VAULT_INSTANCES: %v", vaultInstances)
    } else if vaultInstances == "" and onKubernetes == true {
        log.Info("Proceeding with pod discovery on %v", vaultLabel)
        pods, err := k8s.GetPods(onKubernetes, vaultLabel)
        for _, pod := range pods.Items {
            log.Debugf(pod.Name, pod.Status.PodIP)
        }
    } else {
        log.Info("Proceeding with pod discovery on %v and VAULT_INSTANCES: %v", vaultLabel,vaultInstances)
        pods, err := k8s.GetPods(onKubernetes, vaultLabel
    }

    unsealKeys := os.Getenv("UNSEAL_KEYS")
    if unsealKeys == "" {
        log.Errorf("Invalid value for env var UNSEAL_KEYS: %v", unsealKeys)
    }

    log.Debugf("Connecting to vault at: %v", vaultAddress)
    vaultClient, err := vault.NewVaultClient(vaultAddress, vaultToken)

    if err != nil {
        log.Errorf("Error connecting to vault at: %v", vaultAddress)
    }



    }

}
