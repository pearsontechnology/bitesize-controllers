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
    "reflect"
    "os"
    "k8s.io/client-go/1.4/kubernetes"
    "k8s.io/client-go/1.4/pkg/api"
    "k8s.io/client-go/1.4/rest"
    "k8s.io/client-go/1.4/pkg/apis/extensions/v1beta1"

    "k8s.io/kubernetes/pkg/util/flowcontrol"
    "github.com/pearsontechnology/bitesize-controllers/nginx-ingress-vault/nginx"
    vlt "github.com/pearsontechnology/bitesize-controllers/nginx-ingress-vault/vault"

    "github.com/quipo/statsd"

    log "github.com/Sirupsen/logrus"
)

const version = "1.8.9"

func main() {

    log.SetFormatter(&log.JSONFormatter{})

    debug := os.Getenv("DEBUG")
    if debug == "true" {
        log.SetLevel(log.DebugLevel)
    }

    config, err := rest.InClusterConfig()
    if err != nil {
        log.Fatalf("Failed to create client: %v", err.Error())
    }

    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err.Error())
    }

    stats := statsd.NewStatsdClient("localhost:8125", "nginx.config.")

    log.Infof("\n Ingress Controller version: %v", version)

    rateLimiter := flowcontrol.NewTokenBucketRateLimiter(0.1, 1)
    known := &v1beta1.IngressList{}

    vault, _ := vlt.NewVaultReader()
    if vault.Enabled {
        go vault.RenewToken()
    }

    // Controller loop
    for {
        rateLimiter.Accept()

        if !vault.Ready() {
            continue
        }

        ingresses, err := clientset.Extensions().Ingresses("").List(api.ListOptions{})

        if err != nil {
            log.Errorf("Error retrieving ingresses: %v", err)
            continue
        }
        if reflect.DeepEqual(ingresses.Items, known.Items) {
            continue
        }
        known = ingresses

        var virtualHosts = []*nginx.VirtualHost{}


        for _, ingress := range ingresses.Items {
            vhost,_ := nginx.NewVirtualHost(ingress, vault)
            vhost.CollectPaths()

            if err = vhost.CreateVaultCerts(); err != nil {
                log.Errorf("%s\n", err.Error() )
            }
            if len(vhost.Paths) > 0 {
                virtualHosts = append(virtualHosts, vhost)
            }
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
        }
    }
}
