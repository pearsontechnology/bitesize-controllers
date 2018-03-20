package main

import (
    "os"
    "net"
    "time"
    "strings"
    log "github.com/Sirupsen/logrus"
    vault "github.com/pearsontechnology/bitesize-controllers/vault-controller/vault"
    k8s "github.com/pearsontechnology/bitesize-controllers/vault-controller/kubernetes"
)
const version = "0.1"
const defaultNameSpace = "kube-system"
const defaultVaultLabel = "vault"
const defaultVaultPort = "8243"
const defaultVaultScheme = "https"
const defaultVaultAddr = "http://localhost:8200"
const defaultReloadFrequency = "5s"

func init() {

}

func main() {
    var err error
    var instanceIps, hostIp []string
    var host string

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

    vaultScheme := os.Getenv("VAULT_SCHEME")
    if vaultScheme == "" {
        vaultScheme = defaultVaultScheme
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
        vaultAddress = defaultVaultAddr
    }
    log.Debugf("vaultAddress: %v", vaultAddress)

    v := os.Getenv("RELOAD_FREQUENCY")
    reloadFrequency, err := time.ParseDuration(v)
    if err != nil || v  == "" {
        reloadFrequency, _ = time.ParseDuration(defaultReloadFrequency)
    }

    // Controller loop
    for {

        vaultInstances := os.Getenv("VAULT_INSTANCES")

        unsealKeys := os.Getenv("UNSEAL_KEYS")
        if unsealKeys == "" {
            log.Errorf("Invalid value for env var UNSEAL_KEYS: %v", unsealKeys)
        }

        if vaultInstances == "" && onKubernetes == false {
            log.Errorf("Invalid value for env var VAULT_INSTANCES: %v", vaultInstances)
        } else if vaultInstances == "" && onKubernetes == true {
            log.Info("Proceeding with pod discovery on %v", vaultLabel)
            instanceIps, err = k8s.GetPodIps(vaultLabel, vaultNamespace)
            if err != nil {
                log.Infof("Error retrieving Pod IPs: %v", err )
            }
        } else {
            log.Info("Proceeding with pod discovery on VAULT_INSTANCES: %v", vaultInstances)

            for _, host = range strings.Split(vaultInstances, ",") {
                hostIp, err = net.LookupHost(host)
                if err != nil {
                    log.Infof("Host lookup error for %v: %v", host, err )
                    continue
                }
                log.Debugf("Vault instance: %v IP: %v", host,hostIp[0])
                instanceIps = append(instanceIps, hostIp[0])
            }
        }

        // Get Status for each instance
        for _, instanceIp := range instanceIps {
            log.Debugf("Pod IP: %v", instanceIp)
            log.Debugf("Connecting to vault at: %v", vaultAddress)
                instanceAddress := vaultScheme + "://" + instanceIp + ":" + vaultPort
                vaultClient, err := vault.NewVaultClient(instanceAddress, vaultToken)
                initState, err := vaultClient.InitStatus()
                if err != nil {
                    log.Errorf("ERROR: Init state unknown: %v: %v", instanceAddress, err)
                    //TODO handle errors
                }
                if initState != true {
                    log.Infof("Instance NOT initialised: %v", instanceAddress)
                    // TODO Do Init
                } else {
                    log.Debugf("Instance initialised: %v", instanceAddress)
                }
                sealState, err := vaultClient.SealStatus()
                if err != nil {
                    log.Errorf("ERROR: Seal state unknown: %v: %v", instanceAddress, err)
                    //TODO handle errors
                }
                if sealState == true {
                    log.Infof("Instance Sealed:", instanceAddress)
                    if unsealKeys != "" {
                        for _, key := range strings.Split(unsealKeys, ",") {
                            log.Debugf("Unseal key: %v", key)
                            //TODO handle unseal
                        }
                    }
                }
                leaderState, err := vaultClient.LeaderStatus()
                if err != nil {
                    log.Errorf("ERROR: Instance state unknown: %v: %v", instanceAddress, err)
                    //TODO handle errors
                }
                switch leaderState {
                case true:
                    log.Infof("Instance is leader: %v", instanceAddress)
                    // TODO Do we care ?
                case false:
                    log.Infof("Instance is standby: %v", instanceAddress)
                default:
                    log.Errorf("ERROR: Instance state unknown: %v", instanceAddress)
                    // TODO is this where we kill it?
                }
        }

        time.Sleep(reloadFrequency)

    } //End controller loop
}
