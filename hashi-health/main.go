/*
 * Copyright 2018, Oath Inc.
 * Licensed under the terms of the MIT license. See LICENSE file in the project root for terms.
 */

package main

import (
	"github.com/monwolf/pod-startup-lock/hashi-health/config"
	"github.com/monwolf/pod-startup-lock/hashi-health/hashi"
	"github.com/monwolf/pod-startup-lock/hashi-health/healthcheck"
	"github.com/monwolf/pod-startup-lock/hashi-health/service"
)

func main() {
	conf := config.Parse()
	conf.Validate()

	hashiClient := hashi.NewClient(conf)
	endpointChecker := healthcheck.NewHealthChecker(conf, hashiClient)
	srv := service.NewService(conf.Host, conf.Port, endpointChecker.HealthFunction())

	go srv.Run()
	go endpointChecker.Run()

	select {} // Wait forever and let child goroutines run
}
