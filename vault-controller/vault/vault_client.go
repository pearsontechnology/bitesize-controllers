package vault

import (
    "strings"
    "errors"
    "time"
    vaultapi "github.com/hashicorp/vault/api"
    log "github.com/Sirupsen/logrus"
)

type VaultClient struct {
    Client *vaultapi.Client
}

const initWait = "5s"

// NewVaultClient returns a vault client with DefaultConfig, address and token
func NewVaultClient(address string, token string) (*VaultClient, error) {

    config := vaultapi.DefaultConfig()
    config.Address = address
    if address == "" {
        log.Errorf("Vault not configured")
        return nil, nil
    }

    client, err := vaultapi.NewClient(config)
    if err != nil {
        log.Errorf("Vault config failed.")
        return &VaultClient{nil}, err
    }

    if token != "" {
        client.SetToken(token)
    }

    return &VaultClient{ Client: client }, err
}

func (c *VaultClient) InitStatus() (initState bool, err error) {

    status, err := c.Client.Sys().InitStatus()
    if err != nil {
        log.Errorf("Error retrieving vault init status")
        return false, err
    } else {
        log.Debugf("InitStatus: %v", status)
        return status, err
    }
}

// Init with defaults
func (c *VaultClient) Init(shares int, threshold int) (token string, keys []string, err error) {
    initReq := &vaultapi.InitRequest {
                SecretShares: shares,
                SecretThreshold: threshold,
            }

    response, err := c.Client.Sys().Init(initReq)
    if err != nil {
        log.Errorf("Error initializing Vault! %v", err.Error())
        var keys []string
        return "", keys, err
    } else {
        log.Infof("Initialised instance %v", c.Client.Address())
        log.Debugf("InitStatus: %v", response)
        w, _ := time.ParseDuration(initWait)
        time.Sleep(w)
    }
    token = response.RootToken
    keys = response.KeysB64
    return token, keys, err
}

// SealStatus returns true if vault is unsealed
func (c *VaultClient) SealStatus() (sealState bool, err error) {

    status, err := c.Client.Sys().SealStatus()
    if err != nil || status == nil {
        log.Errorf("Error retrieving vault seal status")
        return true, err
    } else {
        log.Debugf("SealStatus: %v", status)
        return status.Sealed, err
    }
}

// Unseal unseals instance using a comma-delimited set of unseal keys
func (c *VaultClient) Unseal(unsealKeys string) (sealState bool, err error) {

    for _, key := range strings.Split(unsealKeys, ",") {
        log.Debugf("Unseal key: %v", key)
        if len(key) <= 0 {
            continue
        }
        resp, err := c.Client.Sys().Unseal(key)
        if err != nil || resp == nil {
            log.Errorf("Error Unsealing: %v", err.Error())
            return false, err
        }
        if resp.Sealed == false {
            log.Infof("Instance unsealed")
            return true, nil
        } else {
            log.Infof("Instance seal progress: %v", resp.Progress)
        }
    }
    err = errors.New("Insufficient unseal keys! Instance sealed.")
    return false, err
}

// Ready returns true if vault is unsealed
func (c *VaultClient) LeaderStatus() (leaderState bool, err error) {

    status, err := c.Client.Sys().Leader()
    if err != nil || status == nil {
        log.Errorf("Error retrieving vault leader status")
        return false, err
    } else {
        log.Debugf("LeaderStatus: %v", status)
        return status.IsSelf, err
    }
}
