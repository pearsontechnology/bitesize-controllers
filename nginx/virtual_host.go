package nginx

import (
    "fmt"
    "io/ioutil"
    "k8s.io/client-go/1.4/pkg/apis/extensions/v1beta1"

    vlt "k8s.io/contrib/ingress/controllers/nginx-alpha-ssl/vault"
)

type VirtualHost struct {
    Name      string
    Host      string
    Namespace string
    Paths     []*Path
    Ssl       bool
    Nonssl    bool
    Scheme    string
    Ingress   v1beta1.Ingress
    Vault     *vlt.VaultReader
    BlueGreen bool
}

// NewVirtualHost returns virtual host instance
func NewVirtualHost(ingress v1beta1.Ingress) (*VirtualHost, error) {
    name := ingress.ObjectMeta.Name + "_" + ingress.Namespace
    vhost := &VirtualHost{
        Name: name,
        Host: ingress.Spec.Rules[0].Host,
        Namespace: ingress.Namespace,
        Ssl: false,
        Nonssl: true,
        Scheme: "http",
        Ingress: ingress,
        BlueGreen: false,
    }

    vault, _ := vlt.NewVaultReader()
    vhost.Vault = vault

    vhost.applyLabels()

    return vhost, nil

}

func (vhost *VirtualHost) applyLabels() {
    labels := vhost.Ingress.GetLabels()

    for k, v := range(labels) {
        if k == "ssl" && v == "true" {
            vhost.Ssl = true
        }
        if k == "httpsOnly" && v == "true" {
            vhost.Nonssl = false
        }
        if k == "httpsBackend" && v == "true" {
            vhost.Scheme = "https"
        }
        if k == "deployment_method" && v == "bluegreen" {
            vhost.BlueGreen = true
        }
    }
}

// ParsePaths adds  path list to virtual host instance
// (from ingress)
func (vhost *VirtualHost) ParsePaths() {
    for _, r := range(vhost.Ingress.Spec.Rules) {
        for _, p := range(r.HTTP.Paths) {
            l := new(Path)
            l.Location = p.Path
            l.Service = p.Backend.ServiceName
            l.Port = p.Backend.ServicePort.IntVal

            vhost.Paths = append(vhost.Paths, l)
        }
    }
}

// CreateVaultCerts gets certificates (private and crt) from vault
// and writes them to nginx ssl config path. Returns error on failure
func (vhost *VirtualHost) CreateVaultCerts() error {
    if !vhost.Vault.Enabled {
        return fmt.Errorf("Vault disabled for %s", vhost.Host)
    }   
    
    if !vhost.Ssl {
        return fmt.Errorf("No SSL for %s", vhost.Host)
    }

    vhost.Vault.RenewToken()

    key, crt, err := vhost.Vault.GetSecretsForHost(vhost.Host)
    if err != nil {
        vhost.Ssl = false
        return err
    }

    if err := ioutil.WriteFile(ConfigPath + "/certs/" + key.Filename, []byte(key.Secret), 0400); err != nil {
        vhost.Ssl = false
        return fmt.Errorf("Failed to write file %v: %v", key.Filename, err)
    }

    if err := ioutil.WriteFile(ConfigPath + "/certs/" +  crt.Filename, []byte(crt.Secret), 0400); err != nil {
        vhost.Ssl = false
        return fmt.Errorf("failed to write file %v: %v", crt.Filename, err)
    }

    // Needs cert validation

    return nil
}

func (vhost *VirtualHost) DefaultUrl(path Path) string {
    return fmt.Sprintf("%s://%s.%s.svc.cluster.local:%d", vhost.Scheme, path.Service, vhost.Namespace, path.Port)
}

func (vhost *VirtualHost) GreenUrl(path Path) string {
    return fmt.Sprintf("%s://%s-green.%s.svc.cluster.local:%d", vhost.Scheme, path.Service, vhost.Namespace, path.Port)
}

func (vhost *VirtualHost) BlueUrl(path Path) string {
    return fmt.Sprintf("%s://%s-blue.%s.svc.cluster.local:%d", vhost.Scheme, path.Service, vhost.Namespace, path.Port)
}