### Milestone 1 - MVP

- service discovery via Marathon SSE event stream
  - [ ] have apps fully fetched upon startup
  - [ ] have tasks being enabled/disabled upon health status change
  - [ ] HTTP apps properly handled (host headers passed to event stream)
  - [ ] properly handle reconnect (including longer outages)
  - [ ] properly handle initial connect failures (first fully fetch apps state,
      then continue watching SSE stream, optionally do a full refresh regularily)
- HTTP service reverse proxy & load balancer
  - [x] support round robing scheduler
  - [ ] support least load scheduler
  - [ ] support chance scheduler
  - [x] reverse proxying
  - [ ] support request retry (if one backend fails, try another; up to N times, then return 503)

### Milestone 2

- [ ] support listening on more than one service discovery engine
- [ ] ability to add/remove service discovery engines at runtime
- [ ] HTTPS termination
- [ ] HTTPS pass-through with SNI-based service selection
- [ ] TCP load balancer (least load)
- [ ] UDP load balancer (round robin)
- [ ] service discovery via Consul
- [ ] service discovery via Mesos natively

