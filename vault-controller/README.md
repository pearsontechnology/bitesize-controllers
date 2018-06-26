# vault-controller
## Image for Bitesize vault controller

If you align to the defaults then in most cases the controller will work without providing any explicit vars.

It will:

- Discover Vault pods on your cluster.
- Ascertain their init status and initialise them if necessary.
- Ascertain their Seal status and unseal them if necessary.
- It uses Kubernetes(k8s) secrets to store Unseal keys and tokens when it creates them and to read them when required.

### Reqd vars:
- **VAULT_TOKEN** or **VAULT_TOKEN_FILE** (**VAULT_TOKEN** takes precedence). See below example.
- If **VAULT_TOKEN** is not a valid uuid and **VAULT_TOKEN_FILE** exists then **VAULT_TOKEN** will be exported from its contents.

- **VAULT_UNSEAL_KEYS** or **VAULT_UNSEAL_KEYS_FILE** (**VAULT_UNSEAL_KEYS** takes precedence). See below example.
- If **VAULT_TOKEN** is empty and **VAULT_UNSEAL_KEYS_FILE** exists then **VAULT_UNSEAL_KEYS** will be exported from their contents.

- **VAULT_UNSEAL_KEYS_FILE** and **VAULT_TOKEN_FILE** would typically point to files mounted in secret volumes on the controller deployment (see below)

### Optional vars:

- **VAULT_LABEL_KEY** label key on target pods. Default = "k8s-app".
- **VAULT_LABEL_VALUE** label value on target pods. Default = "vault".
- **VAULT_NAMESPACE** namespace for target pods. Default = "kube-system".
- **VAULT_INSTANCES** static, comma-separated list of target hosts for local testing in docker. If provided overrides - **VAULT_LABEL** and **VAULT_NAMESPACE**.
- **VAULT_PORT** vault service port. Default = "8243".
- **VAULT_SCHEME** vault transport scheme. Default = "https".
- **VAULT_ADDR** vault address for cluster. Default = "https://vault.kusbe-syetm.svc.cluster.local:8543".
- **VAULT_TOKEN** vault token for access. No default.
- **VAULT_INIT_SHARES** number of shares for unseal. Default = 5.
- **VAULT_INIT_THRESHOLD** threshold for unseal. Default = 3.
- **VAULT_UNSEAL_KEYS** if already initialised, the keys. No default.
- **RELOAD_FREQUENCY** how often to run controller loop. Default = "30s".
- **VAULT_UNSEAL_SECRET_NAME** k8s secret name for unseal keys in. Default = "vault-unseal-keys".
- **VAULT_UNSEAL_SECRET_KEY** k8s secret key for unseal keys. Default = "unseal-keys".
- **VAULT_TOKEN_SECRET_NAME** k8s secret name for root token. Default = "vault-tokens".
- **VAULT_TOKEN_SECRET_KEY** k8s secret key for root token. Default = "root-token".

Note:
Building on Linux:
```
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build controller.go
```

### Example controller deployment:

```
---

apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: vault-controller
  namespace: kube-system
  labels:
    app: vault-controller
    name: vault-controller
spec:
  replicas: 1
  selector:
    matchLabels:
      app: vault-controller
  template:
    metadata:
      labels:
        app: vault-controller
        name: vault-controller
    spec:
      containers:
      - name: vault-controller
        image: pearsontechnology/vault-controller:0.1
        imagePullPolicy: Always
        env:
          - name: "DEBUG"
            value: "false"
          - name: "VAULT_SCHEME"
            value: "https"
          - name: "VAULT_SKIP_VERIFY"
            value: "true"
          - name: "VAULT_PORT"
            value: "8543"
          - name: "SERVICE_DOMAIN"
            value: "svc.cluster.local"
          - name: "VAULT_LABEL"
            value: "vault"
          - name: "VAULT_NAMESPACE"
            value: "kube-system"
          - name: "VAULT_UNSEAL_KEYS_FILE"
            value: "/etc/vault-secret-unseal-keys/unseal-keys"
          - name: "VAULT_TOKEN_FILE"
            value: "/etc/vault-token-file/root-key"
        volumeMounts:
          - name: vault-volume-unseal-keys
            mountPath: /etc/vault-secret-unseal-keys
          - name: vault-token-file
            mountPath: /etc/vault-token-file
      volumes:
        - name: vault-volume-unseal-keys
          secret:
              secretName: vault-unseal-keys
        - name: vault-token-file
          secret:
              secretName: "vault-tokens"

```

### Policies

The controller optionally supports a CRD to handle creating Vault policies and tokens for them, it will stash these tokens in a secret in the namespace the CRD was created in. Definition:

```
---

apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: vaultpolicies.vault.local
spec:
  scope: Namespaced
  group: vault.local
  version: v1
  names:
    kind: VaultPolicy
    plural: vaultpolicies
    singular: vaultpolicy

```

### Example of a policy:

```
---

apiVersion: vault.local/v1
kind: VaultPolicy
metadata:
  name: vault-mysecrets-read-only
  namespace: my-ns
  labels:
    name: vault-mysecrets-read-only
    app: myapp
spec:
  - Path: secret/mysecrets/*
    Permission: read
    Period: 24h

```

In this case a policy 'vault-mysecrets-read-only' would be created and a periodic token for it stashed in my-ns/vault-mysecrets-read-only:vault-mysecrets-read-only. A TTL may be specified in place of Period, the default TTL in this case is 720h. If Period is specified however, TTL is ignored.

## End
