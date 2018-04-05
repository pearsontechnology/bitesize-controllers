package main

import (
    "os"
    "net"
    "time"
    "strings"
    "reflect"
    log "github.com/Sirupsen/logrus"
    vault "github.com/pearsontechnology/bitesize-controllers/vault-controller/vault"
    k8s "github.com/pearsontechnology/bitesize-controllers/vault-controller/kubernetes"
    vaultcs "github.com/pearsontechnology/bitesize-controllers/vault-controller/pkg/client/clientset/versioned"
	apiextcs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
    apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
    apierrors "k8s.io/apimachinery/pkg/api/errors"
)
const version = "0.1"
const defaultNameSpace = "kube-system"
const defaultSvcTld = ".svc.cluster.local"
const defaultVaultLabel = "vault"
const defaultVaultPort = "8243"
const defaultVaultScheme = "https"
const defaultVaultAddress = "https://vault." + defaultNameSpace + defaultSvcTld + ":" + defaultVaultPort
const defaultReloadFrequency = "30s"
const defaultUnsealSecretName = "vault-unseal-keys"
const defaultUnsealSecretKey = "unseal-key"
const defaultTokenSecretName = "vault-tokens"
const defaultTokenlSecretKey = "root-token"
const crdRefreshInterval = "30s"
func init() {

}

func createCRD() err error {

    crd := &apiext.CustomResourceDefinition{
    ObjectMeta: metav1.ObjectMeta{Name: "vaultpolicy"},
    Spec: apiext.CustomResourceDefinitionSpec{
        Group:   "vaultpolicy",
        Version: "v1",
        Scope:   apiext.Cluster,
        Names:   apiext.CustomResourceDefinitionNames{
            Plural: "vaultpolicys",
            Kind:   reflect.TypeOf(Policy{}).Name(),
            },
        },
    }

    log.Debugf("Creating CRD")
    config, err := rest.InClusterConfig()
    if err != nil {
        log.Errorf("createCRD config error: %v", err.Error())
    }
    clientset, err := apiextcs.NewForConfig(config)
    if err != nil {
        log.Errorf("createCRD client error: %v", err.Error())
    }
    _, err := apiextcs.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd)
    if apierrors.IsAlreadyExists(err) {
        log.Infof("CRD already exists")
        return nil
    } else if err != nil {
        log.Errorf("createCRD create error: %v", err.Error())
    }
    return err
}

func deletePod(name string, namespace string) {
    var err error
    log.Infof("Killing instance: %v", name)
    k8s.DeletePod(name, namespace)
    if err != nil {
        log.Errorf("Error deleting %v: %v", name, err.Error())
    }
}

func initInstance(c *vault.VaultClient, onKubernetes bool) (r *vault.VaultClient, unsealKeys string, err error) {
    instanceAddress := c.Client.Address()
    token, keys, err := c.Init()
    if err != nil {
        log.Errorf("Error Initialise failed: %v", err.Error())
        return c, "", err
    } else {
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
            var k string
            for _, v := range keys {
                k = k + string(v) + ","
            }
            unsealKeys = strings.Trim(k, ",")
            log.Debugf("Stashing %v Unseal keys in: %v/%v", len(strings.Split(unsealKeys, ",")), vaultNamespace, unsealSecretName)
            k8s.PutSecret(unsealSecretName, unsealSecretKey, unsealKeys, vaultNamespace)
            log.Debugf("Stashing Root Token in: %v/%v", vaultNamespace, unsealSecretName)
            k8s.PutSecret(tokenSecretName, tokenSecretKey, token, vaultNamespace)
        }
         r, err := vault.NewVaultClient(instanceAddress, token)
         return r, unsealKeys, err
    }
}

func startCRD() {

        config, err := rest.InClusterConfig()
        if err != nil {
            log.Errorf("InClusterConfig error: %v", err.Error())
        }
        // Create CRD and client
        log.Debugf("Creating CRD")
        clientset, err := vaultcs.NewForConfig(config)
        if err != nil {
            log.Errorf("vaultcs client error: %v", err.Error())
        }
        err = CreateCRD()
        if err != nil {
            log.Errorf("createCRD error: %v", err.Error())
        }

        crdVaultClient, err := vault.NewVaultClient(vaultAddress, vaultToken)

        crdTicker := time.NewTicker(time.Duration(crdRefreshInterval))

        for _ = range crdTicker.C {
            list, err := clientset.PolicyV1().Policies().List(metav1.ListOptions{})
            for _, policy := range list.Items {
                fmt.Debugf("Policy %s found\n", policy.Name)
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


    vaultAddress := os.Getenv("VAULT_ADDR")
    if vaultAddress == "" {
        vaultAddress = defaultVaultAddress
    }
    log.Debugf("vaultAddr: %v", vaultAddr)

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

    unsealKeys := os.Getenv("VAULT_UNSEAL_KEYS")
    if unsealKeys == "" {
        log.Errorf("Invalid value for env var VAULT_UNSEAL_KEYS: %v", unsealKeys)
    }

    // Controller loop
    for {

        vaultInstances := os.Getenv("VAULT_INSTANCES")

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
                    log.Debugf("Vault client failed for: %v, %v", name, err.Error())
                    continue
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
                } else {
                    log.Infof("Instance NOT initialised: %v", name)
                    vaultClient, unsealKeys, err = initInstance(vaultClient, onKubernetes)
                    if err != nil {
                        log.Errorf("ERROR: init resturned error: %v", err.Error())
                        //TODO handle errors
                    }
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

                leaderState, err := vaultClient.LeaderStatus()
                if err != nil {
                    log.Errorf("ERROR: Instance state unknown: %v: %v", name, err.Error())
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
