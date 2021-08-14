> 🚧 This project is still a work in progress 🚧

# MicroSim
MicroSim is a tool that you can use to quickly create a mock distributed system on the target Kubernetes cluster.

## Futures

- Create service topology using a visual UI.
- Create a connection between services dynamically.
- Mix and match different web frameworks.
    - ex:- five [express.js](https://expressjs.com/) services are consuming one [actix](https://actix.rs/) API
- Load Testing
- Fault Injection via HTTP requests.
- Create abnormal conditions within services via HTTP requests.

## Supported HTTP servers

- [gorilla/mux](https://github.com/gorilla/mux)

## Use cases

- Learn about distributed systems and how they operate.
- Testing cloud-native tooling for distributed systems.
- Evaluating rollout strategies.
- Evaluating best web frameworks for different use-cases.
    - Sometimes it's more effective to use high-level language to quickly create services that don't require a lot of throughputs to save developer time.
- Testing monitoring systems and alert behaviors.
- Understanding different service meshes implementations and their impacts.
- Testing out machine learning models created for distributed systems.
- Training reinforcement learning agents in a dynamic environment.


## Example Request path

### Request
```json
{
  "designation": "http://localhost:8081/",
  "faults": {
    "before": [
      {"type": "latency", "args": {"delay": 600}}
    ],
    "after": [
      {"type": "memory-leak", "args": {"size": 250, "duration": 10000}}
    ]
  },
  "payload": {
    "designation": "http://localhost:8082/",
    "faults": {
      "before": [],
      "after": [
        {"type": "latency", "args": {"delay": 600}}
      ]
    },
    "payload": null
  }
}
```

### Description

Client will request send this request to control plane which looks like the above request,
  - Then control plane will overwrite all the designation with correct ClusterIPs.
    - Then the request will be forwarded to `service_1`.
    - `service_1` will first look at the `fatuls["before"]` sections in the request and execute those faults in order.
    - Then the payload part of the request taken out and send to send server.
      - `service_2` will also look at the `fatuls["before"]` first.
      - Since it's empty it will then try to forward the payload to another service.
      - Because it's also empty, it will execute faults in `fatuls["after"]` in order.
      - After those were done, `service_2` will return, `{'error' : null}` with 200 status code to the `service_1` request.
    - Then `service_2` will execute faults in the `fatuls["after"]` section.
    - Finally, it will also return `{'error' : null}` with 200 status code to the control plane request.
  - This request will take minimum of 1000ms to complete.