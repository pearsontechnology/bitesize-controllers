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
const defaultReloadFrequency = "30s"

func init() {

}

func deletePod(name string, namespace string) {
    var err error
    log.Infof("Killing instance: %v", name)
    k8s.DeletePod(name, namespace)
    if err != nil {
        log.Errorf("Error deleting %v: %v", name, err)
    }
}

func main() {
    var err error
    var instanceList map[string]string
    var host string

    // init stuff
    if os.Getenv("DEBUG") == "true" {
        log.SetLevel(log.DebugLevel)
        log.Debugf("DebugLevel on")
    }

	log.Infof("Starting vault controller version: %s", version)

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
    log.Debugf("vaultScheme: %v", vaultScheme)

    // don't default token
    vaultToken := os.Getenv("VAULT_TOKEN")
    log.Debugf("vaultToken: %v", vaultToken)

    onKubernetes := true

    if os.Getenv("KUBERNETES_SERVICE_HOST") == "" {
        log.Info("WARN: NOT running on Kubernetes, pod discovery DISABLED.")
        onKubernetes = false
    }
    log.Debugf("onKubernetes: %v", onKubernetes)

    v := os.Getenv("RELOAD_FREQUENCY")
    reloadFrequency, err := time.ParseDuration(v)
    if err != nil || v  == "" {
        reloadFrequency, _ = time.ParseDuration(defaultReloadFrequency)
    }

    // Controller loop
    for {

        vaultInstances := os.Getenv("VAULT_INSTANCES")

        unsealKeys := os.Getenv("VAULT_UNSEAL_KEYS")
        if unsealKeys == "" {
            log.Errorf("Invalid value for env var VAULT_UNSEAL_KEYS: %v", unsealKeys)
        }

        if vaultInstances == "" && onKubernetes == false {
            log.Errorf("Invalid value for env var VAULT_INSTANCES: %v", vaultInstances)
        } else if vaultInstances == "" && onKubernetes == true {
            log.Infof("Proceeding with pod discovery on %v", vaultLabel)
            instanceList, err = k8s.GetPodIps(vaultLabel, vaultNamespace)
            if err != nil {
                log.Infof("Error retrieving Pod IPs: %v", err )
            }
        } else {
            log.Info("Proceeding with pod discovery on VAULT_INSTANCES: %v", vaultInstances)

            for _, host = range strings.Split(vaultInstances, ",") {
                hostIp, err := net.LookupHost(host)
                if err != nil {
                    log.Infof("Host lookup error for %v: %v", host, err )
                    continue
                }
                log.Debugf("Vault instance: %v IP: %v", host,hostIp[0])
                instanceList[host] = hostIp[0]
            }
        }

        log.Debugf("instanceList: %v", instanceList)

        // Get Status for each instance
        for name, ip := range instanceList {
            log.Debugf("Pod %v IP: %v", name, ip)
                if ip == "error" {
                    if onKubernetes == true {
                        deletePod(name, vaultNamespace)
                        continue
                    }
                }
                if len(ip) <= 0 {
                    log.Debugf("Skipping pod: %v", name)
                    continue
                }
                instanceAddress := vaultScheme + "://" + ip + ":" + vaultPort
                log.Debugf("Connecting to vault at: %v", instanceAddress)
                vaultClient, err := vault.NewVaultClient(instanceAddress, vaultToken)
                if err != nil {
                    log.Debugf("Vault client failed for: %v, %v", name, err)
                    continue
                }
                initState, err := vaultClient.InitStatus()
                if err != nil {
                    log.Errorf("ERROR: Init state unknown: %v: %v", name, err)
                    //TODO handle errors
                    if onKubernetes == true {
                        deletePod(name, vaultNamespace)
                        continue
                    }
                }
                if initState == true {
                    log.Debugf("Instance initialised: %v", name)
                } else {
                    log.Infof("Instance NOT initialised: %v", name)
                    resp, err := vaultClient.Init()
                    if err != nil {
                        log.Errorf("Error Initialise failed: %v", err)
                        continue
                    }
                }

                sealState, err := vaultClient.SealStatus()
                if err != nil {
                    log.Errorf("ERROR: Seal state unknown: %v: %v", name, err)
                    //TODO handle errors
                }
                if sealState == true {
                    log.Infof("Instance Sealed:", name)
                    if unsealKeys != "" {
                        sealState, err = vaultClient.Unseal(unsealKeys)
                    }
                    if err != nil {
                        log.Errorf("Error unsealing: %v",  err)
                    }
                }

                leaderState, err := vaultClient.LeaderStatus()
                if err != nil {
                    log.Errorf("ERROR: Instance state unknown: %v: %v", name, err)
                    //TODO handle errors
                }
                switch leaderState {
                case true:
                    log.Infof("Instance is leader: %v", name)
                    // TODO Do we care ?
                case false:
                    log.Infof("Instance is standby: %v", name)
                default:
                    log.Errorf("ERROR: Instance state unknown: %v", name)
                    if onKubernetes == true {
                        deletePod(name, vaultNamespace)
                        continue
                    }
                }
        }

        time.Sleep(reloadFrequency)

    } //End controller loop
}
