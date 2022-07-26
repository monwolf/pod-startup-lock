# Nomad System Jobs health check service
Typically you would like to postpone application startup until all System Jobs on Node (or at least mapped to the host network) are ready.
This util constantly performs DaemonSet heallthcheck and respond with `200 OK` if passed.
 
Just add this service as a dependent endpoint to the Lock service, and lock won't be acquired until System Jobs are ready.  

## How it works

##### 1. Starts and listens for HTTP requests
Responds with `412 Precondition Failed` until healthckeck succeeds.
Binding `host` and `port` are configurable.

##### 2. Uses Nomad API to get a list of System Jobs required on the same node
`NODE_NAME` environment variable must be set.
If ACL system is set in nomad you should provide a valid token, check env vars in nomad webage.

## Required Configuration
Environment variable `NODE_NAME` must be exposed to indicate which node health should be checked.

## Additional Configuration
You may specify additional command line options to override defaults:

| Option        | Default | Description |
| ------------- |---------| ----------- |
| `--host`      | 0.0.0.0 | Address to bind |
| `--port`      | 9999    | Port to bind    |
| `--namespace` | *none*  | Target Nomad namespace where to perform System Jobs healthcheck. Leave blank for all namespaces |
| `--failHc`    | 10      | Pause between healthchecks if the previous check failed, seconds |
| `--passHc`    | 60      | Pause between healthchecks if the previous check succeeded, seconds |
| `--in`        | *none*  | DaemonSet labels to include in healthcheck, Format: `label:value` |
| `--ex`        | *none*  | DaemonSet labels to exclude from healthcheck, Format: `label:value` |

## How to run locally
Example with some command line options:
```bash
export NODE_NAME=10.11.10.11
go run hashi-health/main.go --in xxx 
```

## How to deploy to Nomad
The preferable way is to deploy as a System Job. Sample deployment YAML:
```hcl
job "pod-startup-lock" {
  region      = "global"
  datacenters = ["dc1"]
  type        = "system"
  priority    = 50

  update {
    stagger      = "10s"
    max_parallel = 1
  }

  group "pod-startup-lock" {
    count = 1

    network {
      port "health" {
        static = 9999
      }
      port "lock" {
        static = 8888
      }

    }
    
    update {
      max_parallel     = 1
      min_healthy_time = "30s"
      healthy_deadline = "9m"
      auto_revert      = true
    }

    restart {
      attempts = 10
      interval = "5m"
      delay    = "25s"
      mode     = "fail"
    }

    task "hashi-health" {
      driver = "docker"

      config {
        extra_hosts = []

        security_opt = [
            "no-new-privileges"
        ]
        pids_limit = 200

        image = "<user>/pod-startup-lock:hashi-health-1.0.2"

        force_pull = true

        ports = ["health"]

        args = ["--port", "9999"]

      }
      env{
        NODE_NAME = "${attr.unique.hostname}"
        NOMAD_ADDR = "http://${attr.unique.hostname}:4646"
      } 
      resources {
        memory = 50
      }

      logs {
        max_files     = 10
        max_file_size = 15
      }

      kill_timeout = "20s"
    }

    task "lock" {
      driver = "docker"

      config {
        extra_hosts = []

        security_opt = [
            "no-new-privileges"
        ]
        pids_limit = 200

        image = "<user>/pod-startup-lock:lock-1.0.2"

        force_pull = true

        ports = ["lock"]
        args = ["--port", "8888", "--locks", "1", "--check", "http://${attr.unique.hostname}:${NOMAD_PORT_health}"]

      }
      
      resources {
        memory = 50
      }

      logs {
        max_files     = 10
        max_file_size = 15
      }

      kill_timeout = "20s"
    }



  }
}```
