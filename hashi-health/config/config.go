/*
 * Copyright 2018, Oath Inc.
 * Licensed under the terms of the MIT license. See LICENSE file in the project root for terms.
 */

package config

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

const defaultPort = 9999
const defaultFailTimeout = 10
const defaultPassTimeout = 60

type arrayFlags []string

func (i *arrayFlags) String() string {
	return fmt.Sprint(*i)
}

func (i *arrayFlags) Set(value string) error {
	for _, str := range strings.Split(value, ",") {
		*i = append(*i, str)
	}
	return nil
}

func Parse() Config {
	host := flag.String("host", "", "Host/Ip to bind")
	port := flag.Int("port", defaultPort, "Port to bind")
	baseUrl := flag.String("baseUrl", "", "K8s api base url. For out-of-cluster usage only")
	namespace := flag.String("namespace", "", "K8s Namespace to check DaemonSets in. Blank for all namespaces")
	failTimeout := flag.Int("failHc", defaultFailTimeout, "Pause between DaemonSet health checks if previous failed, sec")
	passTimeout := flag.Int("passHc", defaultPassTimeout, "Pause between DaemonSet health checks if previous succeeded, sec")

	nodeName, _ := os.LookupEnv("NODE_NAME")

	var includeSystemJobs arrayFlags
	flag.Var(&includeSystemJobs, "in", "Include SystemJobs names: job1,job2,...")

	var excludeSystemJobs arrayFlags
	flag.Var(&excludeSystemJobs, "ex", "Exclude SystemJobs labels, job1,job2,...")
	flag.Parse()

	config := Config{
		*host,
		*port,
		*baseUrl,
		*namespace,
		time.Duration(*failTimeout) * time.Second,
		time.Duration(*passTimeout) * time.Second,
		nodeName,
		includeSystemJobs,
		excludeSystemJobs,
	}
	log.Printf("Application config:\n%+v", config)
	config.Validate()
	return config
}

type Config struct {
	Host              string
	Port              int
	K8sApiBaseUrl     string
	Namespace         string
	HealthFailTimeout time.Duration
	HealthPassTimeout time.Duration
	NodeName          string
	IncludeSystemJobs arrayFlags
	ExcludeSystemJobs arrayFlags
}

func (c *Config) Validate() {
	if c.NodeName == "" {
		log.Panic("NODE_NAME not specified")
	}
	if len(c.IncludeSystemJobs) > 0 && len(c.ExcludeSystemJobs) > 0 {
		log.Panic("Cannot specify both Included and Excluded DaemonSet labels, choose one")
	}
}
