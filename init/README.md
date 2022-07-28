# Init container for the Lock service

Application repeatedly queries Lock service endpoint until it gets `200 OK` response.
Then exits letting main application to start.

**Designed to be deployed as an Init Container**

## Additional Configuration
You may specify additional command line options to override defaults:

| Option       | Default               | Description |
| ------------ |---------------------- | ----------- |
| `--port`     | 8888                  | Lock Service's HTTP port |
| `--host`     | localhost             | Lock Service's hostname |
| `--pause`    | 10                    | Pause between Lock acquiring attempts, seconds |
| `--duration` | *none*                | Custom lock duration to request, seconds |
| `--jobanme`  | NOMAD_JOB/Hostname    | Custom JobName used to lock service call |


## How to run locally
Example with some command line options:
```bash
go run init/main.go --port 9000 --duration 15
```

## How to deploy to Kubernetes
Should be deployed as an Init Container. Sample deployment YAML:
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: myapp-pod
  labels:
    app: myapp
spec:
  containers:
    - name: myapp-container
      image: busybox
      command: ['sh', '-c', 'echo The app is running! && sleep 3600']
  initContainers:
    - name: startup-lock-init-container
      image: <docker_user>/pod-startup-lock:init-<version>
      args: ["--host", "$(HOST_IP)", "--port", "8888", "--duration", "15"]
      env:
        - name: HOST_IP
          valueFrom:
            fieldRef:
              fieldPath: status.hostIP
```
## How to deploy to nomad
Should be deployed as an Prestart Container. Sample deployment HCL:

```hcl

job "docs" {
  datacenters = ["dc1"]

  group "example" {
    network {
      port "http" {
        static = "5678"
      }
    }
    task "startup-lock-init-container" {
      driver = "docker"

      config {
        image        = "<docker_user>/pod-startup-lock:init-<version>"
        args         = ["--host", "${attr.unique.hostname}", "--port", "8888", "--duration", "15"]
      }

      resources {
        cpu    = 200
        memory = 30
      }

      lifecycle {
        hook    = "prestart"
        sidecar = false
      }
    }
    task "server" {
      driver = "docker"

      config {
        image = "hashicorp/http-echo"
        ports = ["http"]
        args = [
          "-listen",
          ":5678",
          "-text",
          "hello world",
        ]
      }
    }
  }
}




```

