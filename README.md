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
