/*
Copyright 2015 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"k8s.io/client-go/1.4/pkg/apis/extensions/v1beta1"
	"os"
	"reflect"
	"time"

	k8s "github.com/pearsontechnology/bitesize-controllers/nginx-ingress-vault/kubernetes"
	"github.com/pearsontechnology/bitesize-controllers/nginx-ingress-vault/monitor"
	"github.com/pearsontechnology/bitesize-controllers/nginx-ingress-vault/nginx"
	vlt "github.com/pearsontechnology/bitesize-controllers/nginx-ingress-vault/vault"
	"github.com/pearsontechnology/bitesize-controllers/nginx-ingress-vault/version"

	"github.com/quipo/statsd"

	log "github.com/Sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

func main() {

	log.SetFormatter(&log.JSONFormatter{})

	debug := os.Getenv("DEBUG")
	if debug == "true" {
		log.SetLevel(log.DebugLevel)
	}

	// Prometheus
	prometheus.MustRegister(&monitor.Status)
	http.Handle("/metrics", promhttp.Handler())
	log.Infof("Starting /metrics on port :8080")
	go func() {
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	log.Infof("Ingress Controller version: %v", version.Version)

	v := os.Getenv("RELOAD_FREQUENCY")
	reloadFrequency, err := time.ParseDuration(v)
	if err != nil || v == "" {
		reloadFrequency, _ = time.ParseDuration("5s")
	}

	onKubernetes := true
	if os.Getenv("KUBERNETES_SERVICE_HOST") == "" {
		log.Errorf("WARN: NOT running on Kubernetes, ingress functionality will be DISABLED")
		onKubernetes = false
	}

	stats := statsd.NewStatsdClient("localhost:8125", "nginx.config.")

	known := &v1beta1.IngressList{}

	vault, _ := vlt.NewVaultReader()
	go vault.RenewToken()

	// Controller loop
	for {

		if !vault.Ready() {
			vault, err = vlt.NewVaultReader()

			// Reset existing ingress list to allow pull of ssl from vault
			known = &v1beta1.IngressList{}
			time.Sleep(reloadFrequency)
			continue
		}

		vault, err = vault.CheckSecretToken()

		if err != nil {
			log.Errorf("Error calling CheckSecretToken: %s", err)
			time.Sleep(reloadFrequency)
			continue
		}

		time.Sleep(reloadFrequency)

		ingresses, err := k8s.GetIngresses(onKubernetes)

		if err != nil {
			log.Errorf("Error retrieving ingresses: %v", err)
			continue
		}

		if reflect.DeepEqual(ingresses.Items, known.Items) {
			continue
		}

		// Generating new config starts here

		// Reset prometheus counters
		monitor.Reset()

		virtualHosts := nginx.ProcessIngresses(ingresses, vault)

		if err != nil {
			log.Errorf("Error processing ingresses: %v", err)
			continue
		}

		if len(virtualHosts) == 0 && onKubernetes == true {
			continue
		}

		nginx.WriteConfig(virtualHosts)
		// cops-165 - Generate custom error page per vhost
		nginx.WriteCustomErrorPages(virtualHosts)

		err = nginx.Verify()

		stats.Incr("reload", 1)

		if err != nil {
			log.Errorf("ERR: nginx config failed validation: %v", err)
			log.Infof("Sent config error notification to statsd.")
			stats.Incr("error", 1)
		} else {
			nginx.Start()
			log.Infof("nginx config updated.")
			known = ingresses
		}

	}
}
