# e2e Tests

e2e tests for the Civo Cloud Contorller Manager.

These tests can be run:

```bash
cd e2e
API_KEY=.... go test --timeout 30 -v .

```

## PreRequisites

A civo.com account is needed, and you'll need to get your API key from the web front end.

## Tests

The general flow for each test is as follows:

1. Provision a new cluster
2. Wait for the cluster to be provisioned
2. Scale down pre-deployed CCM in the cluster
4. Run the local copy of CCM 
5. Run e2e Tests
6. Delete the cluster

### Node

> WIP

### Loadbalancers

> WIP
