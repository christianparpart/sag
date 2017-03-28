# sag - Service Application Gateway

### Motivation

Continuously serve HTTP/HTTPS/TCP/UDP applications from the cloud without configuration,
supporting multiple and different service discovery backends,
reconfiguring the service proxy without a restart.

- *Service Discovery Methods*
  - Marathon SSE Event Stream
  - Mesos
  - Consul
- *Service Gateway Mode*s
  - UDP load balancing
  - TCP load balancing
  - HTTP load balancing
- *Service Routing Mode*s
  - by HTTP request Host header
  - by SSL SNI (Server-Name-Indication) extension
- *Service backends are chosen based on configurable scheduling algorithms*
  - least load
  - round robin

### Command Line Options

```
SAG_SOURCES="marathon1.mesos:8080,marathon2.mesos:8080"

sag [options] -- [list of service discovery endpoints]

    --gateway-http-port=INT
    --gateway-https-port=INT
-v, --verbose

```

### notes
```
- ServiceDiscovery
  - MarathonSD
  - MesosSD
  - ConsulSD
- Service
  - TcpService
  - UdpService
  - HttpService
  - HttpsService
- Backend
```
