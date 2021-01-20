# kubectl bridge for Cloud Run

-----

Examples:

```sh
$ kubectl api-versions
domains.cloudrun.com/v1alpha1
serving.knative.dev/v1
```

```sh
$ kubectl api-resources
NAME             SHORTNAMES      APIVERSION                      NAMESPACED   KIND
domainmappings                   domains.cloudrun.com/v1alpha1   true         DomainMapping
configurations   config,cfg      serving.knative.dev/v1          true         Configuration
revisions        rev             serving.knative.dev/v1          true         Revision
routes           rt              serving.knative.dev/v1          true         Route
services         kservice,ksvc   serving.knative.dev/v1          true         Service
```

```
$ kubectl get ksvc
kubectl get ksvc
NAME                       AGE
object-detection           35d
chat                       14d
whiteboard                 14d
```

See [LICENSE].
