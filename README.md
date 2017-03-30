# sag - Service Application Gateway

### Motivation

Continuously serve HTTP/HTTPS/TCP/UDP applications from the cloud without
reconfiguration, supporting multiple and different service discovery backends,
reconfiguring the service proxy without a restart.

- **Service Discovery Methods**
  - Marathon SSE Event Stream
  - Mesos
  - Consul
- **Service Gateway Modes**
  - UDP load balancing
  - TCP load balancing
  - HTTP load balancing
- **Service Routing Modes**
  - by HTTP request Host header on a well known TCP port (80)
  - by SSL SNI (Server-Name-Indication) extension on a well known TCP port (443)
  - by HTTP request path, emulating Linkerd-routing, on a well known TCP port (9991)
  - by service ports (HTTP/HTTPS/TCP/UDP), as provided from service descovery (such as Marathon)
- **Service backends are chosen based on configurable scheduling algorithms**
  - least load scheduler
  - chance scheduler
  - round robin scheduler

### Command Line Options

here be dragons
