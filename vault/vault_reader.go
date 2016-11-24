package vault

import (
    "fmt"
    "strconv"
    "os"
    vault "github.com/hashicorp/vault/api"
)

type VaultReader struct {
    Enabled bool
    Client *vault.Client
}

type Cert struct {
    Filename string
    Secret string
}

func NewVaultReader() (*VaultReader, error) {
    vaultEnabledFlag := os.Getenv("VAULT_ENABLED")
    vaultAddress := os.Getenv("VAULT_ADDR")
    vaultToken := os.Getenv("VAULT_TOKEN")

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
    }, nil
}

// RenewToken renews vault's token
func (r *VaultReader) RenewToken() {
    // It's a go routine now. Add ticker
    tokenPath := "/auth/token/renew-self"
    tokenData, err := r.Client.Logical().Write(tokenPath, nil)

    if err != nil || tokenData == nil {
        fmt.Printf("Error renewing Vault token %v, %v\n", err, tokenData)
    } else {
        fmt.Printf("Successfully renewed Vault token.\n")
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