package nginx

import (
	"crypto/tls"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"k8s.io/client-go/1.4/pkg/apis/extensions/v1beta1"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strings"

	"github.com/pearsontechnology/bitesize-controllers/nginx-ingress-vault/monitor"
	vlt "github.com/pearsontechnology/bitesize-controllers/nginx-ingress-vault/vault"
)

type VirtualHost struct {
	Name         string
	Hosts        map[string]bool
	Namespace    string
	Paths        []*Path
	HTTPSEnabled bool
	HTTPEnabled  bool
	Http2        bool
	Scheme       string
	Ingress      v1beta1.Ingress
	Vault        *vlt.VaultReader
	BlueGreen    bool
}

// NewVirtualHost returns virtual host instance
func NewVirtualHost(ingress v1beta1.Ingress) (*VirtualHost, error) {
	name := strings.Replace(ingress.ObjectMeta.Name, "-", "_", -1) +
		"_" +
		strings.Replace(ingress.Namespace, "-", "_", -1)

	vhost := &VirtualHost{
		Name:         name,
		Hosts:        map[string]bool{},
		Namespace:    ingress.Namespace,
		HTTPSEnabled: false,
		HTTPEnabled:  true,
		Http2:        false,
		Scheme:       "http",
		Ingress:      ingress,
		BlueGreen:    false,
		Vault:        nil,
	}

	for _, rule := range ingress.Spec.Rules {
		vhost.Hosts[rule.Host] = true
	}

	//	vhost.Vault = vault

	vhost.applyLabels()
	return vhost, nil

}

func (vhost *VirtualHost) applyLabels() {
	labels := vhost.Ingress.GetLabels()

	for k, v := range labels {
		if k == "ssl" && v == "true" {
			vhost.HTTPSEnabled = true
			monitor.IncSslVHosts()
		}
		if k == "httpsOnly" && v == "true" {
			vhost.HTTPEnabled = false
		}
		if k == "http2" && v == "true" {
			vhost.Http2 = true
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
	for _, r := range vhost.Ingress.Spec.Rules {
		for _, p := range r.HTTP.Paths {
			vhost.appendService(p.Backend.ServiceName, p)
		}
	}
}

func ProcessIngresses(ingresses *v1beta1.IngressList) []*VirtualHost {

	var err error
	var virtualHosts = []*VirtualHost{}

	for _, ingress := range ingresses.Items {

		vhost, _ := NewVirtualHost(ingress)
		monitor.IncVHosts()
		vhost.CollectPaths()

		if err = vhost.Validate(); err != nil {
			log.Errorf("Ingress %s failed validation: %s", vhost.Name, err.Error())
			monitor.IncFailedVHosts()
			continue
		}

		if err = vhost.CheckCerts(); err != nil {
			log.Errorf("%s\n", err.Error())
			vhost.HTTPSEnabled = false
		}
		if len(vhost.Paths) > 0 {
			virtualHosts = append(virtualHosts, vhost)
		}
	}
	return virtualHosts
}

func (vhost *VirtualHost) appendService(serviceName string, ingressPath v1beta1.HTTPIngressPath) {
	p := &Path{
		URI:       ingressPath.Path,
		Service:   serviceName,
		Port:      ingressPath.Backend.ServicePort.IntVal,
		Namespace: vhost.Namespace,
	}

	vhost.Paths = AppendUniquePath(vhost.Paths, p)

}

// CheckCerts checks for a valid key pair in /etc/nginx/certs.
func (vhost *VirtualHost) CheckCerts() error {
	if !vhost.HTTPSEnabled {
		return fmt.Errorf("No SSL for %s", vhost.Name)
	}

	for domainName, _ := range vhost.Hosts {
		certAbsolutePath := ConfigPath + "/certs/" + domainName + ".crt"
		keyAbsolutePath := ConfigPath + "/certs/" + domainName + ".key"

		_, err := os.Stat(certAbsolutePath)
		if os.IsNotExist(err) {
			vhost.HTTPSEnabled = false
			monitor.IncFailedSslVHosts()
			return fmt.Errorf("%s file does not exist", certAbsolutePath)
		}

		_, err = os.Stat(keyAbsolutePath)
		if os.IsNotExist(err) {
			vhost.HTTPSEnabled = false
			monitor.IncFailedSslVHosts()
			return fmt.Errorf("%s file does not exist", keyAbsolutePath)
		}

		// Cert validation
		if _, err := tls.LoadX509KeyPair(certAbsolutePath, keyAbsolutePath); err != nil {
			vhost.HTTPSEnabled = false
			monitor.IncSslVHostsCertFail()
			return fmt.Errorf("Failed to validate certificate for %s", domainName)
		}

	}
	return nil
}

// cops-374 - Get Nginx pod name and pass it to the nginx.conf.tmpl
// to generate the X-Loadbalancer-Id in runtime.

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func (vhost *VirtualHost) ServerNames() string {
	hosts := []string{}
	for k := range vhost.Hosts {
		hosts = append(hosts, k)
	}
	return strings.Join(hosts, " ")
}

func (vhost *VirtualHost) GetPodName() string {
	return getenv("POD_NAME", "nginx-ingress")
}

func (vhost *VirtualHost) GetResolver() string {
	return getenv("RESOLVER", "127.0.0.1")
}

func (vhost *VirtualHost) GetResolverPort() string {
	return getenv("RESOLVER_PORT", "53")
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

func (vhost *VirtualHost) Validate() error {

	schemeRegex, _ := regexp.Compile("^https?$")
	_, _ = regexp.Compile("[a-z\\d+].*?\\.\\w{2,8}$")

	if len(vhost.Hosts) == 0 {
		return fmt.Errorf("No hosts set")
	}

	// I have no idea what are these below, they are very wrong.
	// Golang is a typed language, reflect.TypeOf for a non-interface
	// does not make sense. I will leave for someone else to fix these.

	if reflect.TypeOf(vhost.Name).String() != "string" || vhost.Name == "" {
		return fmt.Errorf("Name must be set")
	}

	if reflect.TypeOf(vhost.Namespace).String() != "string" || vhost.Namespace == "" {
		return fmt.Errorf("Namespace must be set")
	}
	if reflect.TypeOf(vhost.Scheme).String() != "string" || schemeRegex.MatchString(reflect.ValueOf(vhost.Scheme).String()) != true {
		return fmt.Errorf("Scheme must be set")
	}
	if reflect.TypeOf(vhost.HTTPSEnabled).Kind() != reflect.Bool {
		return fmt.Errorf("HTTPSEnabled label must be true; false")
	}
	if reflect.TypeOf(vhost.Http2).Kind() != reflect.Bool {
		return fmt.Errorf("Http2 label must be true; false")
	}
	if (vhost.Http2 == true) && (vhost.HTTPSEnabled != true || vhost.HTTPEnabled != false) {
		return fmt.Errorf("If http2 is enabled then HTTPSEnabled and httpsOnly must be true")
	}
	if reflect.TypeOf(vhost.HTTPEnabled).Kind() != reflect.Bool {
		return fmt.Errorf("HTTPEnabled label must be true; false")
	}
	if reflect.TypeOf(vhost.BlueGreen).Kind() != reflect.Bool {
		return fmt.Errorf("BlueGreen label must be true; false")
	}
	return nil
}

func AppendUniquePath(slice []*Path, p *Path) []*Path {
	if len(slice) == 0 {
		return append(slice, p)
	}
	return slice
}
