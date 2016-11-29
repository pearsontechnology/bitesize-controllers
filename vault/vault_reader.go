package vault

import (
    "fmt"
    "strconv"
    "os"
    "time"
    vault "github.com/hashicorp/vault/api"
)

type VaultReader struct {
    Enabled bool
    Client *vault.Client
    TokenRefreshInterval *time.Ticker
}

type Cert struct {
    Filename string
    Secret string
}

func NewVaultReader() (*VaultReader, error) {
    vaultEnabledFlag := os.Getenv("VAULT_ENABLED")
    vaultAddress := os.Getenv("VAULT_ADDR")
    vaultToken := os.Getenv("VAULT_TOKEN")
    refreshFlag := os.Getenv("VAULT_REFRESH_INTERVAL")

    refreshInterval, err := strconv.Atoi(refreshFlag)
    if err != nil {
        refreshInterval = 10
    }

    vaultEnabled, err := strconv.ParseBool(vaultEnabledFlag)
    if err != nil {
        fmt.Printf("VAULT_ENABLED not set\n")
        return &VaultReader{ Enabled: false}, err
    }

    if vaultAddress == "" || vaultToken == "" {
        fmt.Printf("Vault not configured\n")
        return &VaultReader{ Enabled: false}, nil
    }

    config := vault.DefaultConfig()
    config.Address = vaultAddress


    client, err := vault.NewClient(config)
    if err != nil {
        fmt.Printf("WARN: Vault config failed.\n")
        return &VaultReader{ Enabled: false}, err
    }

    // Needs VaultReady

    return &VaultReader{
        Enabled: vaultEnabled,
        Client: client,
        TokenRefreshInterval: time.NewTicker(time.Minute * time.Duration(refreshInterval)),
    }, nil
}

// Ready returns true if vault is unsealed and
// ready to use
func (r *VaultReader) Ready() bool {
    if ! r.Enabled {
        // always ready if we don't use it :)
        return true
    }
    status, err := r.Client.Sys().SealStatus()
    if err != nil || status == nil {
        fmt.Printf("Error retrieving vault status\n")
        return false
    }

    return !status.Sealed
}

// RenewToken renews vault's token every TokenRefreshInterval
func (r *VaultReader) RenewToken() {
    tokenPath := "/auth/token/renew-self"

    for _ = range r.TokenRefreshInterval.C {
        tokenData, err := r.Client.Logical().Write(tokenPath, nil)

        if err != nil || tokenData == nil {
            fmt.Printf("Error renewing Vault token %v, %v\n", err, tokenData)
        } else {
            fmt.Printf("Successfully renewed Vault token.\n")
        }
    }
}

func (r *VaultReader) GetSecretsForHost(hostname string) (*Cert, *Cert, error) {
    var e error

    vaultPath := "secret/ssl/" + hostname
    
    keySecretData, err := r.Client.Logical().Read(vaultPath)    
    if err != nil || keySecretData == nil {
        e = fmt.Errorf("No secret for %v", hostname)
        return nil, nil, e
    }

    fmt.Printf("Found secret for %s\n", hostname)

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
        e := fmt.Errorf("WARN: No %s found for %v", dataKey, hostname)
        return nil, e
    }
    path := hostname + "." + dataKey
    
    return &Cert{ Secret: secret, Filename: path}, nil
}