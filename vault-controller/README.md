# vault-controller
Image for Bitesize vault controller


Reqd vars:
VAULT_TOKEN or VAULT_TOKEN_FILE (VAULT_TOKEN takes precedence)
if VAULT_TOKEN is not a valid uuid and VAULT_TOKEN_FILE exists then VAULT_TOKEN will be exported from its contents

VAULT_UNSEAL_KEYS or VAULT_UNSEAL_KEYS_FILES (VAULT_UNSEAL_KEYS takes precedence)
if VAULT_TOKEN is empty and VAULT_UNSEAL_KEYS_FILES exists then VAULT_UNSEAL_KEYS will be exported from their contents

Optional vars:

VAULT_INSTANCES static, comma-separated list of target hosts. If provided overrides VAULT_LABEL and VAULT_NAMESPACE.
VAULT_LABEL label on target pods. Default = "vault"
VAULT_NAMESPACE namespace for target pods. Default = "kube-system"
VAULT_PORT vault service port. Default = "8243"
VAULT_SCHEME vault transport scheme. Default = "https"
RELOAD_FREQUENCY how often to run controller loop. Default = "5s"

Note:
Building on ubuntu:
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build controller.go 

