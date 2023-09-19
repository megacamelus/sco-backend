# sco-operator

## Run
1. Set up kubernetes cluster (minikube, kind, etc)
1. Set up Camel K
1. `make run`

## Tests
1. `make test`

## e2e Tests
1. Set up kubernetes cluster (minikube, kind, etc)
1. Set up Camel K **OR** just apply the CRDs with `make test/apply-crds`
1. `make test/e2e`

## All checks
1. Set up kubernetes cluster (minikube, kind, etc)
1. Set up Camel K **OR** just apply the CRDs with `make test/apply-crds`
1. `make check/all` 

This runs all the necessary checks, useful as a pre-commit target.