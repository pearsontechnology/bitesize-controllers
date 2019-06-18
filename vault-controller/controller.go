package main

import (
    "os"
    "net"
    "time"
    "errors"
    "strings"
    "strconv"
    log "github.com/Sirupsen/logrus"
    vault "github.com/pearsontechnology/bitesize-controllers/vault-controller/vault"
    k8s "github.com/pearsontechnology/bitesize-controllers/vault-controller/kubernetes"
    vaultcs "github.com/pearsontechnology/bitesize-controllers/vault-controller/pkg/client/clientset/versioned"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/rest"
)
const version = "0.1"
const defaultNameSpace = "kube-system"
const defaultSvcTld = ".svc.cluster.local"
const defaultVaultLabelKey = "k8s-app"
const defaultVaultLabelVal = "vault"
const defaultVaultPort = "8243"
const defaultVaultScheme = "https"
const defaultVaultAddress = "https://vault." + defaultNameSpace + defaultSvcTld + ":" + defaultVaultPort
const defaultReloadFrequency = "30s"
const defaultUnsealSecretName = "vault-unseal-keys"
const defaultUnsealSecretKey = "unseal-keys"
const defaultTokenSecretName = "vault-tokens"
const defaultTokenlSecretKey = "root-token"
const crdRefreshInterval = "30s"
const defaultInitShares = 5
const defaultInitThreashold = 3
func init() {

}

func deletePod(name string, namespace string) {
    var err error
    log.Infof("Killing instance: %v", name)
    k8s.DeletePod(name, namespace)
    if err != nil {
        log.Errorf("Error deleting %v: %v", name, err.Error())
    }
}

func initInstance(c *vault.VaultClient, onKubernetes bool, shares int, threshold int) (r *vault.VaultClient, vaultToken string, unsealKeys string, err error) {
    var token string
    var keys []string
    instanceAddress := c.Client.Address()
    if onKubernetes == true {
        unsealSecretName := os.Getenv("VAULT_UNSEAL_SECRET_NAME")
        if unsealSecretName == "" {
            unsealSecretName = defaultUnsealSecretName
        }
        unsealSecretKey := os.Getenv("VAULT_UNSEAL_SECRET_KEY")
        if unsealSecretKey == "" {
            unsealSecretKey = defaultUnsealSecretKey
        }
        tokenSecretName := os.Getenv("VAULT_TOKEN_SECRET_NAME")
        if tokenSecretName == "" {
            tokenSecretName = defaultTokenSecretName
        }
        tokenSecretKey := os.Getenv("VAULT_TOKEN_SECRET_KEY")
        if tokenSecretKey == "" {
            tokenSecretKey = defaultTokenlSecretKey
        }
        vaultNamespace := os.Getenv("VAULT_NAMESPACE")
        if vaultNamespace == "" {
            vaultNamespace = defaultNameSpace
        }
        s := k8s.GetSecret(unsealSecretName, unsealSecretKey, vaultNamespace)
        if len(strings.Split(s, ",")) >= threshold {
            log.Warnf("WARNING: Existing Unseal keys found in secret %v/%v:%v", vaultNamespace,unsealSecretName,unsealSecretKey )
            log.Debugf("Keys: %v", s)
            return c, "", s, errors.New("Instance already inititialised")
        }
        token, keys, err = c.Init(shares, threshold)
        if err != nil {
            log.Errorf("Error Initialise failed: %v", err.Error())
            return c, "", "", err
        }
        var k string
        for _, v := range keys {
            k = k + string(v) + ","
        }
        unsealKeys = strings.Trim(k, ",")
        log.Debugf("Stashing %v Unseal keys in: %v/%v:%v", len(strings.Split(unsealKeys, ",")), vaultNamespace, unsealSecretName, unsealSecretKey)
        k8s.PutSecret(unsealSecretName, unsealSecretKey, unsealKeys, vaultNamespace)
        log.Debugf("Stashing Root Token in: %v/%v:%v", vaultNamespace, tokenSecretName, tokenSecretKey)
        k8s.PutSecret(tokenSecretName, tokenSecretKey, token, vaultNamespace)
    } else {
        token, keys, err = c.Init(shares, threshold)
        if err != nil {
            log.Errorf("Error Initialise failed: %v", err.Error())
            return c, "", "", err
        }
        var k string
        for _, v := range keys {
            k = k + string(v) + ","
        }
        unsealKeys = strings.Trim(k, ",")
        log.Infof("Unseal Keys: %v", unsealKeys)
        log.Infof("Root token: %v", token)
    }
    time.Sleep(5 * time.Second)
    r, err = vault.NewVaultClient(instanceAddress, token)
    return r, token, unsealKeys, err
}

func startCRD(stopchan chan bool, vaultAddress string, vaultToken string) {

    config, err := rest.InClusterConfig()
    if err != nil {
        log.Errorf("InClusterConfig error: %v", err.Error())
    }
    // Create CRD and client
    log.Infof("Starting VaultPolicy monitor...")
    clientset, err := vaultcs.NewForConfig(config)
    if err != nil {
        log.Errorf("vaultcs client error: %v", err.Error())
    }
    crdVaultClient, err := vault.NewVaultClient(vaultAddress, vaultToken)
    t, _ := time.ParseDuration(crdRefreshInterval)
    crdTicker := time.NewTicker(t)
    defer crdTicker.Stop()
    for _ = range crdTicker.C {
        select {
        case <- stopchan:
            log.Debugf("startCRD stop...")
            return
        default:
            log.Debugf("crdTicker fired...")
            list, err := clientset.VaultPolicyV1().VaultPolicies().List(metav1.ListOptions{})
            if err !=nil {
                log.Errorf("Error listing vaultpolicies: %v", err.Error())
            }
            for _, policy := range list.Items {
                log.Debugf("VaultPolicy CRD object found: %v/%v", policy.Namespace, policy.Name)
                token, err := crdVaultClient.CreatePolicy(policy)
                if err != nil {
                    log.Errorf("Error creating policy %v: %v", policy.Name, err.Error())
                } else if err == nil && token != "" {
                    log.Debugf("Policy %s token generated: %v", policy.Name, token)
                    log.Debugf("Generating secret %v/%v:%v for token: %v", policy.Namespace, policy.Name, policy.Name, token)
                    k8s.PutSecret(policy.Name, policy.Name, token, policy.Namespace)
                }
            }
        }
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

    vaultLabelKey := os.Getenv("VAULT_LABEL_KEY")
    if vaultLabelKey == "" {
        vaultLabelKey = defaultVaultLabelKey
    }
    vaultLabelVal := os.Getenv("VAULT_LABEL_VALUE")
    if vaultLabelVal == "" {
        vaultLabelVal = defaultVaultLabelVal
    }
    log.Debugf("vaultLabelKey: %v", vaultLabelKey)
    log.Debugf("vaultLabelVal: %v", vaultLabelVal)

    vaultNamespace := os.Getenv("VAULT_NAMESPACE")
    if vaultNamespace == "" {
        vaultNamespace = defaultNameSpace
    }
    log.Debugf("vaultNamespace: %v", vaultNamespace)

    vaultAddress := os.Getenv("VAULT_ADDR")
    if vaultAddress == "" {
        vaultAddress = defaultVaultAddress
    }
    log.Debugf("vaultAddress: %v", vaultAddress)

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

    vaultInitShares, err := strconv.Atoi(os.Getenv("VAULT_INIT_SHARES"))
    if err != nil {
        vaultInitShares = defaultInitShares
    }
    log.Debugf("vaultInitShares: %v", vaultInitShares)

    vaultInitThreshold, err := strconv.Atoi(os.Getenv("VAULT_INIT_THRESHOLD"))
    if err != nil {
        vaultInitThreshold = defaultInitThreashold
    }
    log.Debugf("vaultInitThreshold: %v", vaultInitThreshold)

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

    unsealKeys := os.Getenv("VAULT_UNSEAL_KEYS")
    if unsealKeys == "" {
        log.Errorf("Invalid value for env var VAULT_UNSEAL_KEYS: %v", unsealKeys)
    }

    crdStopChan := make(chan bool)

    go startCRD(crdStopChan, vaultAddress, vaultToken)

    // Controller loop
    for {
        log.Debugf("Starting controller loop")
        vaultInstances := os.Getenv("VAULT_INSTANCES")

        if vaultInstances == "" && onKubernetes == false {
            log.Errorf("Invalid value for env var VAULT_INSTANCES: %v", vaultInstances)
        } else if vaultInstances == "" && onKubernetes == true {
            log.Infof("Proceeding with pod discovery on %v/%v", vaultLabelKey, vaultLabelVal)
            instanceList, err = k8s.GetPodIps(vaultLabelKey, vaultLabelVal, vaultNamespace)
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
                    log.Debugf("Vault client failed for: %v, %v", name, err.Error())
                    continue
                }

                sealState, err := vaultClient.SealStatus()
                if err != nil {
                    log.Errorf("ERROR: Seal state unknown: %v: %v", name, err.Error())
                    //TODO handle errors
                }
                if sealState == true {
                    log.Infof("Instance Sealed: %v", name)
                    if unsealKeys != "" {
                        sealState, err = vaultClient.Unseal(unsealKeys)
                    }
                    if err != nil {
                        log.Errorf("Error unsealing: %v",  err.Error())
                    }
                }
                initState, err := vaultClient.InitStatus()
                if err != nil {
                    log.Errorf("ERROR: Init state unknown: %v: %v", name, err.Error())
                    //TODO handle errors
                    if onKubernetes == true {
                        deletePod(name, vaultNamespace)
                        continue
                    }
                }
                if initState == true {
                    log.Debugf("Instance initialised: %v", name)
                } else if initState == false {
                    log.Infof("Instance NOT initialised: %v", name)
                    vaultClient, vaultToken, unsealKeys, err = initInstance(vaultClient, onKubernetes, vaultInitShares, vaultInitThreshold)
                    vaultClient.Unseal(unsealKeys)
                    close(crdStopChan)
                    crdStopChan = make(chan bool)
                    go startCRD(crdStopChan, vaultAddress, vaultToken)
                    if err != nil {
                        log.Errorf("ERROR: init resturned error: %v", err.Error())
                        //TODO handle errors
                    }
                }

        time.Sleep(reloadFrequency)

    } //End controller loop
}
