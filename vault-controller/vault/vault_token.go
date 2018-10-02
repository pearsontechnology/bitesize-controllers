package vault

import (
    "errors"
    "time"
    vaultapi "github.com/hashicorp/vault/api"
    log "github.com/Sirupsen/logrus"
    "github.com/google/uuid"
)

// validateToken checks the token string is a valid uuid
func ValidateToken(token string) (isValid bool) {
    _, err := uuid.Parse(token)
    if err != nil {
        log.Errorf("Error parsing Token as uuid %v", token)
        return false
    } else {
        return true
    }
}

// getToken retrieves token data for a given token
func (c *VaultClient) getToken(token string) (tokenData *vaultapi.Secret, err error) {
    if !ValidateToken(token) {
        return nil, errors.New("invalid token")
    }
    tokenData, err = c.Client.Auth().Token().Lookup(token) //https://godoc.org/github.com/hashicorp/vault/api#TokenAuth.Lookup
    if err != nil {
        log.Errorf("Error retrieving Token data %v", token)
        return nil, err
    }
    return tokenData, err
}

func (c *VaultClient) isTokenRenewable(token string) (isRenewable bool) {
    tokenData, err := c.getToken(token)
    if err != nil {
        return false
    }
    isRenewable, err = tokenData.TokenIsRenewable()
    if err != nil {
        log.Errorf("Error retrieving TokenIsRenewable for %v", token)
        isRenewable = false
    }
    return isRenewable
}

// getTokenTTL retrieves token ttl for a given tokenn in seconds
func (c *VaultClient) getTokenTTL(token string) (seconds int64, err error) {
    tokenData, err := c.getToken(token)
    if err != nil {
        return 0, err
    }

    ttl, err := tokenData.TokenTTL()
    if err != nil {
        log.Errorf("Error retrieving Token ttl for %v", token)
        return 0, err
    }

    return int64(ttl / time.Second), err
}

// createToken creates a new token for a given set of policy options
func (c *VaultClient) createToken(opts *vaultapi.TokenCreateRequest) (token string, err error) {
    log.Debugf("CreatePolicy token opts: %v", opts)
    tokenData, err := c.Client.Auth().Token().Create(opts) //https://godoc.org/github.com/hashicorp/vault/api#TokenAuth.Create
    if err != nil {
        log.Errorf("Error creating Token %v", opts)
        return "", err
    }
    token = tokenData.Auth.ClientToken
    return token, err
}

// renewToken renews a given token
func (c *VaultClient) renewToken(token string, ttl int) (err error) {
    if !ValidateToken(token) {
        return errors.New("invalid token")
    }
    if ! c.isTokenRenewable(token) {
        log.Errorf("Token is not renewable: %v", token)
    }
    _, err = c.Client.Auth().Token().Renew(token, ttl) //https://godoc.org/github.com/hashicorp/vault/api#TokenAuth.Renew
    if err != nil {
        log.Errorf("Error creating Token %v", token)
    }
    return err
}
