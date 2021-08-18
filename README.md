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
    - Sometimes it's more effective to use high-level language to quickly create services that don't require a lot of
      throughput to save developer time.
- Testing monitoring systems and alert behaviors.
- Understanding different service meshes implementations and their impacts.
- Testing out machine learning models created for distributed systems.
- Training reinforcement learning agents in a dynamic environment.

## Example Request path

### Request
<details>
  <summary>Click to view sample request!</summary>

```json
{
  "designation": "service_1",
  "faults": {
    "before": [
      {
        "type": "latency",
        "args": {
          "delay": 600
        }
      }
    ],
    "after": [
      {
        "type": "memory-leak",
        "args": {
          "size": 250,
          "duration": 10000
        }
      }
    ]
  },
  "probability": 100,
  "routes": [
    {
      "designation": "service_2",
      "probability": 50,
      "faults": {
        "before": [],
        "after": [
          {
            "type": "latency",
            "args": {
              "delay": 600
            }
          }
        ]
      },
      "routes": [
        {
          "designation": "service_4",
          "probability": 70,
          "faults": {
            "before": [
              {
                "type": "latency",
                "args": {
                  "delay": 200
                }
              }
            ],
            "after": []
          }
        }
      ]
    },
    {
      "designation": "service_3",
      "probability": 50,
      "faults": {
        "before": [
          {
            "type": "latency",
            "args": {
              "delay": 1000
            }
          }
        ],
        "after": []
      }
    }
  ]
}
```
</details>

<details>
  <summary>Click to view sample response!</summary>

```json
{
  "service": "service_1",
  "address": "http://localhost:8081/",
  "errors": [],
  "response": [
    {
      "service": "service_2",
      "address": "http://localhost:8082/",
      "errors": [],
      "response": [
        {
          "service": "service_4",
          "address": "http://localhost:8084/",
          "errors": [],
          "response": []
        }
      ]
    },
    {
      "service": "service_3",
      "address": "http://localhost:8083/",
      "errors": [],
      "response": []
    }
  ]
}
```
</details>

### Description

Client will request send this request to control plane which looks like the above request,

- Then control plane will overwrite all the designation with correct ClusterIPs.
    - Then the request will be forwarded to `service_1`.
    - `service_1` will first look at the `fatuls["before"]` sections in the request and execute those faults in order.
    - Then the routes part of the request will be taken out.
      - Then for `route` in each route will get called with the content of the `route`
        - On each child service will do the same till it hit the `route` with empty routes list
        - Then it will resolve from last one to the start. 
    - When all the child services resolve `service_1` will execute `fatuls["after"]`
    - Finally, it will also return a response like shown above with 200 status code to the control plane request.
- This request will take minimum of 2400ms to complete since all the requests are queued sequentially.