# kubectl bridge for Cloud Run

> :warning: :warning: This project is under heavy development.

This is a small binary that offers Cloud Run functionality (as Knative APIs)
using `kubectl`.

Start a local server at port `5555`:

```sh
go run .
```

## Support Table/Roadmap

- [x] `kubectl` Discovery API (`/api`, `/apis`, `/apis/serving.knative.dev/v1` etc)
- [x] Listing/getting resources with `kubectl get`
- Server-side "table" printing for `kubectl get`
    - [x] KService 
    - [ ] Configuration
    - [ ] Route
    - [ ] Revision
    - [ ] DomainMapping
- [x] Deleting resources with `kubectl gelete`
- Updating resources
  - [ ] `kubectl edit`  (json merge patch not yet supported)
  - [ ] `kubectl apply` (json merge patch not yet supported)
  - [ ] `kubectl patch` (json merge patch not yet supported)
- [ ] `--watch` (probably will never be supported)
- [ ] `--all-namespaces` while querying resources (would query all regions)

Roadmap:

- [ ] generating `KUBECONFIG` files automatically
- [ ] showing Cloud Run regions as "Kubernetes namespaces"
- [ ] configuring a "GCP project" using command-line option

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
curl-vpc                   https://curl-vpc-2wvlk7vg3a-uc.a.run.app                   curl-vpc-00004-vad                   curl-vpc-00004-vad                   True
ktest                      https://ktest-2wvlk7vg3a-uc.a.run.app                      ktest-00002-xis                      ktest-00001-biz                      False
web                        https://web-2wvlk7vg3a-uc.a.run.app                        web-00001-deg                        web-00001-deg                        True
example                    https://example-2wvlk7vg3a-uc.a.run.app                    example-vcp54                        example-vcp54                        True
object-detection-emerald   https://object-detection-emerald-2wvlk7vg3a-uc.a.run.app   object-detection-emerald-00003-zam   object-detection-emerald-00003-zam   True
escape-the-sandbox         https://escape-the-sandbox-2wvlk7vg3a-uc.a.run.app         escape-the-sandbox-00014-buy         escape-the-sandbox-00014-buy         True
```

Other Knative CRDs like `Route`, `Configuration` and `Revision` are supported, too:

```sh
$ kubectl get routes
NAME                       AGE
object-detection           35d
chat                       14d
whiteboard                 14d
curl-vpc                   23d
ktest                      105d
```

-----

See [LICENSE](./LICENSE). This is not an official Google project.
