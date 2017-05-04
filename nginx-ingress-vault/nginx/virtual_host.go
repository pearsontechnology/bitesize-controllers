package nginx

import (
    "fmt"
    "os"
    "io/ioutil"
    "net/url"
    "net/http"
    "strings"
    "crypto/tls"
    "reflect"
    "regexp"
    "k8s.io/client-go/1.4/pkg/apis/extensions/v1beta1"

    vlt "github.com/pearsontechnology/bitesize-controllers/nginx-ingress-vault/vault"
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
func NewVirtualHost(ingress v1beta1.Ingress, vault *vlt.VaultReader) (*VirtualHost, error) {
    name := strings.Replace(ingress.ObjectMeta.Name, "-","_",-1) +
            "_" +
            strings.Replace(ingress.Namespace, "-", "_", -1)

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

// CollectPaths adds  path list to virtual host instance
// (from ingress)
func (vhost *VirtualHost) CollectPaths() {
    for _, r := range(vhost.Ingress.Spec.Rules) {
        for _, p := range(r.HTTP.Paths) {
            vhost.appendService(p.Backend.ServiceName, p)
        }
    }
}

func (vhost *VirtualHost) appendService(serviceName string, ingressPath v1beta1.HTTPIngressPath) {
    p := &Path {
        URI: ingressPath.Path,
        Service: serviceName,
        Port: ingressPath.Backend.ServicePort.IntVal,
        Namespace: vhost.Namespace,
    }

    vhost.Paths = append(vhost.Paths, p)
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

    key, crt, err := vhost.Vault.GetSecretsForHost(vhost.Host)
    if err != nil {
        vhost.Ssl = false
        return err
    }

    keyAbsolutePath := ConfigPath + "/certs/" + key.Filename
    if err := ioutil.WriteFile(keyAbsolutePath, []byte(key.Secret), 0400); err != nil {
        vhost.Ssl = false
        return fmt.Errorf("Failed to write file %v: %v", keyAbsolutePath, err)
    }

    certAbsolutePath := ConfigPath + "/certs/" + crt.Filename
    if err := ioutil.WriteFile(certAbsolutePath, []byte(crt.Secret), 0400); err != nil {
        vhost.Ssl = false
        return fmt.Errorf("failed to write file %v: %v", certAbsolutePath, err)
    }

    // Cert validation
    if _, err := tls.LoadX509KeyPair(certAbsolutePath, keyAbsolutePath); err != nil {
        vhost.Ssl = false
        return fmt.Errorf("Failed to validate certificate")
    }

    return nil
}

// cops-374 - Get Nginx pod name and pass it to the nginx.conf.tmpl
// to generate the X-Loadbalancer-Id in runtime.
func (vhost *VirtualHost) GetPodName() string {
    return os.Getenv("POD_NAME")
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

func newHTTPClient(dest *url.URL) *http.Client {
    if strings.ToLower(dest.Scheme) == "https" {
        tr := &http.Transport{
            TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
        }
        return &http.Client{Transport: tr}
    }
    return &http.Client{}
}

func (vhost *VirtualHost) ValidateVirtualHost() error {

    schemeRegex, _ := regexp.Compile("^https?$")
    hostRegex, _ := regexp.Compile("[a-z\\d+].*?\\.\\w{2,8}$")

    if reflect.TypeOf(vhost.Name).String() != "string" || vhost.Name == "" {
        return fmt.Errorf("Name must be set")
    }
    if reflect.TypeOf(vhost.Host).String() != "string" || hostRegex.MatchString(reflect.ValueOf(vhost.Host).String()) != true {
        return fmt.Errorf("Host must be set")
    }
    if reflect.TypeOf(vhost.Namespace).String() != "string" || vhost.Namespace == "" {
        return fmt.Errorf("Namespace must be set")
    }
    if reflect.TypeOf(vhost.Scheme).String() != "string" || schemeRegex.MatchString(reflect.ValueOf(vhost.Scheme).String()) != true {
        return fmt.Errorf("Scheme must be set")
    }
    if reflect.TypeOf(vhost.Ssl).String() != "bool" {
        return fmt.Errorf("Ssl label must be true; false")
    }
    if reflect.TypeOf(vhost.Nonssl).String() != "bool" {
        return fmt.Errorf("Nonssl label must be true; false")
    }
    if reflect.TypeOf(vhost.BlueGreen).String() != "bool" {
        return fmt.Errorf("BlueGreen label must be true; false")
    }
    if reflect.TypeOf(vhost.Paths).String() != "[]*nginx.Path" || vhost.Paths == nil {
        return fmt.Errorf("Paths must be set")
    }
    return nil
}
