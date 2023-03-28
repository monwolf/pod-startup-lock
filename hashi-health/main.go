/*
 * Copyright 2018, Oath Inc.
 * Licensed under the terms of the MIT license. See LICENSE file in the project root for terms.
 */

package main

import (
	"os"

	"github.com/monwolf/pod-startup-lock/hashi-health/config"
	"github.com/monwolf/pod-startup-lock/hashi-health/hashi"
	"github.com/monwolf/pod-startup-lock/hashi-health/healthcheck"
	"github.com/monwolf/pod-startup-lock/hashi-health/service"
	"github.com/monwolf/pod-startup-lock/hashi-health/watcher"
)

var hashiClient *hashi.Client
var conf config.Config

func newClient() error {
	hashiClient = hashi.NewClient(conf)
	return nil
}

func main() {
	conf = config.Parse()
	conf.Validate()

	cli_cert, ok := os.LookupEnv("NOMAD_CLIENT_CERT")
	if ok {
		watcher.Watch(cli_cert, newClient)
	}

	hashiClient = hashi.NewClient(conf)
	endpointChecker := healthcheck.NewHealthChecker(conf, hashiClient)
	srv := service.NewService(conf.Host, conf.Port, endpointChecker.HealthFunction())

	go srv.Run()
	go endpointChecker.Run()

	select {} // Wait forever and let child goroutines run
}
