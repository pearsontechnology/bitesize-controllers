package vault

import (
    "fmt"
    "strconv"
    "os"
    "time"
    "strings"
    "errors"
    vault "github.com/hashicorp/vault/api"
    log "github.com/Sirupsen/logrus"
    "k8s.io/client-go/1.4/kubernetes"
    "k8s.io/client-go/1.4/rest"
    "github.com/google/uuid"
    "github.com/giantswarm/retry-go"
)

const vaultDefaultRetries = 5
const vaultDefaultTimeout = "30s" //cumulative
const vaultDefaultTokenRefreshInterval = 10

type VaultReader struct {
    Enabled bool
    Client *vault.Client
    TokenRefreshInterval *time.Ticker
    Timeout time.Duration
    Retries int
}

type Cert struct {
    Filename string
    Secret string
}

func getToken() (token string, err error) {

    token = os.Getenv("VAULT_TOKEN")
    secretKey := os.Getenv("VAULT_TOKEN_SECRET")
    if token == "" {
        if secretKey == "" {
            return token, nil
        }
    } else {
        return token, nil
    }

    namespace := os.Getenv("POD_NAMESPACE")

    config, err := rest.InClusterConfig()
    if err != nil {
        log.Fatalf("Failed to create client: %v", err.Error())
    }

    clientset, err := kubernetes.NewForConfig(config)

    if err != nil {
        log.Fatalf("Failed to create client: %v", err.Error())
    }

    secrets, err := clientset.Core().Secrets(namespace).Get(secretKey)

    if err != nil {
        log.Errorf("Error retrieving secrets: %v", err)
        return "", nil
    }

    for name, data := range secrets.Data {
        //secret[name] = string(data)
        if name == secretKey {
            token = strings.TrimSpace(string(data))
            log.Debugf("Found VAULT_TOKEN_SECRET secret: %s", name)
        }
    }
    _, err = uuid.Parse(token)
    if err != nil {
        log.Errorf("Error parsing VAULT_TOKEN_SECRET: %v", token)
    }
    return token, err
}

func (r *VaultReader) CompareToken() (compare bool, err error) {
    existingToken := r.Client.Token()
    secretToken, err := getToken()
    //log.Debugf("Comparing existingToken: %s with secretToken: %s ", existingToken, secretToken)
    if err != nil {
        log.Errorf("Error retrieving token from k8s secret: %v", err)
        return true, err
    }

    if existingToken == secretToken {
        return true, nil
    } else {
        return false, nil
    }
}

func (r *VaultReader) CheckSecretToken() (*VaultReader, error) {

    log.Debugf("Comparing client token to k8s secret.")
    same, err := r.CompareToken()
    if err != nil {
        log.Errorf("Error calling CompareToken: %s", err)
        return r, err
    }
    if !same {
        log.Infof("New token detected in k8s secret, reloading client.")
        nr, err := NewVaultReader()
        if err != nil {
            log.Errorf("Error calling CompareToken: %s", err)
            return r, err
        } else {
            r = nr
        }
    }
    return r, err
}

// RenewToken renews vault's token every TokenRefreshInterval
func (r *VaultReader) RenewToken() {
    if r.Enabled {
        log.Debugf("Start RenewToken func.")
        for _ = range r.TokenRefreshInterval.C {
            token, err := getToken()
            //log.Debugf("Renewing token: %s ", r.Client.Token())
            if err == nil {
                r.Client.SetToken(token)
            } else {
                log.Errorf("Error retrieving Vault token %v", err)
            }

            tokenPath := "/auth/token/renew-self"
            tokenData, err := r.Client.Logical().Write(tokenPath, nil)
            if err != nil || tokenData == nil {
                log.Errorf("Error renewing Vault token %v, %v", err, tokenData)
            } else {
                log.Infof("Successfully renewed Vault token.\n")
            }
        }
    }
}

func NewVaultReader() (*VaultReader, error) {
    address := os.Getenv("VAULT_ADDR")
    refreshFlag := os.Getenv("VAULT_REFRESH_INTERVAL")
    enabled, err := strconv.ParseBool(os.Getenv("VAULT_ENABLED"))
    if err != nil {
        enabled = true
    }

    token, err := getToken()

    if err == nil {
        enabled = true
    } else {
        enabled = false
    }

    vr := os.Getenv("VAULT_RETRIES")
    retries, err := strconv.Atoi(vr)
    if err != nil {
        retries = vaultDefaultRetries
    }

    vt := os.Getenv("VAULT_TIMEOUT")
    timeout, err := time.ParseDuration(vt)
    if err != nil {
        timeout, _ = time.ParseDuration(vaultDefaultTimeout)
    }

    refreshInterval, err := strconv.Atoi(refreshFlag)
    if err != nil {
        refreshInterval = vaultDefaultTokenRefreshInterval
    }

    if address == "" || token == "" {
        log.Infof("Vault not configured")
        err := errors.New("Address or Token null.")
        return &VaultReader{ Enabled: false}, err
    }

    client, err := vault.NewClient(nil)
    if err != nil {
        fmt.Errorf("Vault config failed.")
        return &VaultReader{ Enabled: false}, err
    }

    // Needs VaultReady
    client.SetToken(token)
    return &VaultReader{
        Enabled: enabled,
        Client: client,
        TokenRefreshInterval: time.NewTicker(time.Minute * time.Duration(refreshInterval)),
        Retries: retries,
        Timeout: timeout,
    }, nil
}

// Ready returns true if vault is unsealed and
// ready to use
func (r *VaultReader) Ready() bool {
    if r == nil || r.Client == nil {
        return false
    }

    var err error
    var status *vault.SealStatusResponse

    getStatus := func() error {
		status, err = r.Client.Sys().SealStatus()
        return err
	}

    errcheck := func(err error) bool {
        if err != nil {
            log.Errorf("Retrying vault status: %d", err.Error())
            return true
        }
        return false
	}

    retry.Do(getStatus,
        retry.Sleep(1 * time.Second),
        retry.MaxTries(r.Retries),
        retry.Timeout(r.Timeout),
        retry.RetryChecker(errcheck),
    )

    if err != nil || status == nil {
        log.Errorf("Error retrieving vault status: %v, %v", status, err.Error())
        return false
    }
    return !status.Sealed
}

func (r *VaultReader) GetSecretsForHost(hostname string) (*Cert, *Cert, error) {
    var e, err error

    vaultPath := "secret/ssl/" + hostname

    var keySecretData *vault.Secret

    getData := func() error {
       keySecretData, err = r.Client.Logical().Read(vaultPath)
       return err
	}

    errcheck := func(err error) bool {
        if err != nil {
            log.Errorf("retrying retrieve secrets: %d", err.Error())
            return true
        }
        return false
	}

    retry.Do(getData,
        retry.Sleep(1 * time.Second),
        retry.MaxTries(r.Retries),
        retry.Timeout(r.Timeout),
        retry.RetryChecker(errcheck),
    )

    if err != nil {
        e = fmt.Errorf("Error retrieving secret for %v: %v", hostname, err.Error())
        return nil, nil, e
    }

    if keySecretData == nil {
        e = fmt.Errorf("No secret for %v", hostname)
        return nil, nil, e
    }

    log.Infof("Found secret for %s", hostname)

    key, err := getCertFromData(keySecretData, "key", hostname)
    if err != nil {
        return nil, nil, err
    }

    crt, err := getCertFromData(keySecretData, "crt", hostname)
    if err != nil {
        return nil, nil, err
    }

    return key, crt,  nil
}

func getCertFromData(data *vault.Secret, dataKey string, hostname string) (*Cert, error) {

    secret := fmt.Sprintf("%v", data.Data[dataKey])
    if secret == "" {
        e := fmt.Errorf("No %s found for %v", dataKey, hostname)
        return nil, e
    }
    path := hostname + "." + dataKey

    return &Cert{ Secret: secret, Filename: path}, nil
}
