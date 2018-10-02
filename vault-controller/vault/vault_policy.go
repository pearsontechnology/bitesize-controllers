package vault

import (
    "strconv"
    vaultapi "github.com/hashicorp/vault/api"
    k8s "github.com/pearsontechnology/bitesize-controllers/vault-controller/kubernetes"
    vaultpolicy "github.com/pearsontechnology/bitesize-controllers/vault-controller/pkg/apis/vault.local/v1"
    log "github.com/Sirupsen/logrus"
)


// getSeconds takes a string e.g. "3m" and attempts to return the number of seconds it represents e.g. "180"
func getSeconds(timeString string) (seconds int64, err error) {
    log.Debugf("getSeconds: %v", timeString)
    seconds = 0
    uz := len(timeString)
    u := string(timeString[len(timeString)-1])
    un, err := strconv.ParseInt(timeString[:uz-1], 10, 64)
    if err != nil {
        log.Errorf("Error converting to seconds: %v", timeString)
        return seconds, err
    }
    switch u {
    case "s":
        seconds = un
    case "m":
        seconds = un*60
    case "h":
        seconds = un*3600
    case "d":
        seconds = un*86400
    default:
        log.Errorf("Time string should end in s, m, h or d: %v", timeString)
    }
    return seconds, err
}

// policyRenewToken processes token for a policy if it has Renew
func (c *VaultClient) PolicyRenewToken(policy vaultpolicy.VaultPolicy, token string) (err error) {
    log.Debugf("policyRenewToken: %v %v", policy, token)
    s, err := getSeconds(policy.Renew)
    if err != nil || s == 0 {
        log.Errorf("Error renewing token: %v", err.Error())
    }

    existingTTL, err := c.getTokenTTL(token)
    if s > existingTTL && err != nil {
        newTTL, _ := getSeconds(policy.Spec[0].TTL)
        err := c.renewToken(token, int(newTTL))
        if err != nil {
            log.Errorf("Error renewing token: %v", err.Error())
        } else {
            log.Debugf("Renewed token: %v", token)
        }
    }
    return err
}

// policyRecreateToken processes token for a policy if it has Recreate
func (c *VaultClient) PolicyRecreateToken(policy vaultpolicy.VaultPolicy, token string) (err error) {
    log.Debugf("policyRecreateToken: %v %v", policy, token)
    s, err := getSeconds(policy.Recreate)
    if err != nil || s == 0 {
        log.Errorf("Error recreating token: %v", err.Error())
    }

    existingTTL, err := c.getTokenTTL(token)
    if s > existingTTL {
        token, err := c.CreatePolicy(policy)
        if err != nil {
            log.Errorf("Error recreating token: %v", err.Error())
        } else {
            k8s.PutSecret(policy.Name, policy.Name, token, policy.Namespace)
            log.Debugf("Recreated token: " + token)
        }
    }
    return err
}

// CreatePolicy creates a policy and a token for that policy
func (c *VaultClient) CreatePolicy(policy vaultpolicy.VaultPolicy) (token string, err error) {
    log.Debugf("CreatePolicy: %v", policy)

    p, err := c.Client.Sys().GetPolicy(policy.Name)
    if p != "" {
        log.Infof("Policy already exists: %v", policy.Name)
        return "", nil
    }

    var rules string
    for _, s := range policy.Spec {
        rules = rules + "path \"" + s.Path + "\" { capabilities = [\"" + s.Permission + "\"] }"
    }
    err = c.Client.Sys().PutPolicy(policy.Name, rules) //https://godoc.org/github.com/hashicorp/vault/api#Sys.PutPolicy
    if err != nil {
        log.Errorf("Error creating Policy %v", policy.Name)
        return "", err
    }
    log.Infof("CreatePolicy created policy: %v", policy.Name)
    var policies []string

    policies = append(policies, policy.Name)

    opts := &vaultapi.TokenCreateRequest{
        Policies: policies,
        DisplayName: policy.Name}

    for _, s := range policy.Spec {
        if len(s.Period) > 0 {
            opts.Period = s.Period
        } else if len(s.TTL) > 0 {
            opts.TTL = s.TTL
        }
    }

    log.Infof("Creating token for policy: %v", policy.Name )

    return c.createToken(opts)
}
