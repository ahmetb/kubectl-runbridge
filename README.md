# kubectl bridge for Cloud Run <sup>experimental</sup>

:warning::warning: This project is heavily experimental and is not ready for prime-time use. :warning::warning:

This is a small binary that lets you work with Cloud Run API
using `kubectl` because Cloud Run is Knative API-compliant.

Start the program locally and keep it running:

```sh
go run .
```

It'll print something like this:

```text
Assuming GCP project id "ahmetb-demo"
started fake kube-apiserver for Cloud Run
Set this environment variable in your shell:
	export KUBECONFIG=/Users/ahmetb/.kube/config.cloudrun
```

## What works

- [x] `kubectl` Discovery API (`/api`, `/apis`, `/apis/serving.knative.dev/v1` etc)
- [x] Listing/getting resources with `kubectl get`
    - [x] Server-side printing (as tables) in `kubectl get`
- [x] Deleting resources with `kubectl delete`
- [x] generating `KUBECONFIG` files automatically
- [ ] `--all-namespaces` while querying resources (would query all regions)

## What doesn’t work

- [ ] Creating resources w/ `kubectl apply` or `kubectl create` (due to lack of /openapi/v2 endpoint)
- Updating resources
  - [ ] `kubectl edit`  (json merge patch not yet supported)
  - [ ] `kubectl apply` (json merge patch not yet supported)
  - [ ] `kubectl patch` (json merge patch not yet supported)
- [ ] `--watch` (probably will never be supported since Cloud Run API doesn't support it)
- [ ] Mock Cloud Run regions as "Kubernetes namespaces" (not "contexts")
- [ ] Configuring a "GCP project" using command-line option

## Examples

It supports various `kubectl` commands to explore the API:

```sh
$ kubectl api-versions
domains.cloudrun.com/v1
serving.knative.dev/v1
```

```sh
$ kubectl api-resources
NAME             SHORTNAMES      APIVERSION                      NAMESPACED   KIND
domainmappings                   domains.cloudrun.com/v1         true         DomainMapping
configurations   config,cfg      serving.knative.dev/v1          true         Configuration
revisions        rev             serving.knative.dev/v1          true         Revision
routes           rt              serving.knative.dev/v1          true         Route
services         kservice,ksvc   serving.knative.dev/v1          true         Service
```

List Cloud Run services:

```
$ kubectl get ksvc
NAME                       URL                                                        LATESTCREATED                        LATESTREADY                          READY   REASON
object-detection           https://object-detection-2wvlk7vg3a-uc.a.run.app           object-detection-00013-nax           object-detection-00013-nax           True
chat                       https://chat-2wvlk7vg3a-uc.a.run.app                       chat-00001-nuc                       chat-00001-nuc                       True
whiteboard                 https://whiteboard-2wvlk7vg3a-uc.a.run.app                 whiteboard-00001-kic                 whiteboard-00001-kic                 True
```

Other Knative CRDs like `Route`, `Configuration` and `Revision` are supported, too:

```sh
$ kubectl get routes
NAME                       SERVICE                    URL                                                        READY   REASON
object-detection           object-detection           https://object-detection-2wvlk7vg3a-uc.a.run.app           True
chat                       chat                       https://chat-2wvlk7vg3a-uc.a.run.app                       True
whiteboard                 whiteboard                 https://whiteboard-2wvlk7vg3a-uc.a.run.app                 True
```

-----

See [LICENSE](./LICENSE). This is not an official Google project.
