package vault

import (
    "strings"
    "errors"
    vault "github.com/hashicorp/vault/api"
    log "github.com/Sirupsen/logrus"
)

type VaultClient struct {
    Client *vault.Client
}

func NewVaultClient(address string, token string) (*VaultClient, error) {

    if address == "" || token == "" {
        log.Errorf("Vault not configured")
        return nil, nil
    }

    client, err := vault.NewClient(nil)
    if err != nil {
        log.Errorf("Vault config failed.")
        return &VaultClient{nil}, err
    }
    return &VaultClient{ Client: client }, err
}

// SealStatus returns true if vault is unsealed
func (c *VaultClient) InitStatus() (initState bool, err error) {

    status, err := c.Client.Sys().InitStatus()
    if err != nil {
        log.Errorf("Error retrieving vault init status")
        return false, err
    } else {
        log.Debug("InitStatus: %v", status)
        return status, err
    }
}

func (c *VaultClient) SealStatus() (sealState bool, err error) {

    status, err := c.Client.Sys().SealStatus()
    if err != nil || status == nil {
        log.Errorf("Error retrieving vault seal status")
        return true, err
    } else {
        log.Debug("SealStatus: %v", status)
        return status.Sealed, err
    }
}

func (c *VaultClient) Unseal(unsealKeys string) (sealState bool, err error) {

    for _, key := range strings.Split(unsealKeys, ",") {
        log.Debugf("Unseal key: %v", key)
        //TODO handle unseal
        resp, err := c.Client.Sys().Unseal(key)
        if err != nil || resp == nil {
            log.Errorf("Error Unsealing: %v", err)
        }
        if resp.Sealed == false {
            log.Infof("Instance unselae")
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
        log.Debug("LeaderStatus: %v", status)
        return status.IsSelf, err
    }

}
