# Kubernetes object lock

[![asciicast](https://asciinema.org/a/gcXZb7x0taJuxyEkLzXmdykfn.svg)](https://asciinema.org/a/gcXZb7x0taJuxyEkLzXmdykfn)

When building complex workflows in Kubernetes, prohibiting users from editing
or deleting existing objects becomes neccessary. For instance, when a Pod runs
a long running task it is important that the pod survises until the task
is finished. Particular example of such long running task is a virtual machine
snapshot started within the [KubeVirt](https://kubevirt.io) project
to save the virtual machine state.
However the concept is generic and can be used for locking any object
in Kubernetes.

Technically, the Lock does only one thing. When update or delete is issued on
existing object, the Lock checks whether that object is locked. If object is locked,
request is rejected.

Although simple, currently there is no direct way of locking objects stored in
Kubernetes. However, by analysing how data gets into the store, a solution
can be found.

![ETCD data flow](https://github.com/petrkotas/k8s-object-lock/raw/master/pics/flow.png "ETCD data flow.")

Admission controll is the exact place, where the Lock can check whether the object
is locked and rejects the request. The remaining question is, how to mark
object locked. There are three options:

![Object lock options](https://github.com/petrkotas/k8s-object-lock/raw/master/pics/lock_options.png "Object lock options.")

1. **Annotation/label on an object**, this is first and most obvious solution.
When the objects has an annotation/label bearing the lock information,
request is rejected. However there is big drawback. The annotation/label is placed
on an object comming from the user. Therefore it is user, who decides the object
is locked. This is not what the Lock is supposed to do.
2. **Annotation/label on an object in a store**, second solution taking the simmilar
approach. However, this time the annotation/label is placed carefully on the existing
object in store, achieved preferably by higher level controller and not
by direct user interaction. However, this solution would not work. The issue is
every edit/delete request have to go through the same pipeline, as illustrated above.
This applies even for controller api calls. Thus, the result is permanently locked
object. Once the annotation/label is in place, the Lock will prohibit every
edit/delete.
3. **Lock object** introduced as [CRD](https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/). The lock object is placed in the same
namespace with the same name as the object being locked. The Lock simply
checks whether lock object exists. If so, the request if rejected.
Unlocking object is a matter of deleting the lock object from the namespace.

The best approach is the third option. Introduce the lock object and the Lock
will check upon its existence.

## Implementation

The implementation is fairly straightforward and follows pretty much the same
path as any Kubernetes extension. Write the controller. In this case
the controller is simple https server providing the validating endpoint.
Register the controller within the Kubernetes. Introduce custom resource.
Profit.

## Components

### Controller

The "brain" component of the Lock is the controller. It is a simple https
server providing single endpoint: `/validate`. To make it work with
the Kubernetes, endpoint has to accept and return a json payload, with
[AdmissionReview](https://github.com/kubernetes/kubernetes/blob/5a16163c87fe2a90916a51b52771a668bcaf2a0d/pkg/apis/admission/types.go#L29)
object.
AdmissionReview contains both [AdmissionRequest](https://github.com/kubernetes/kubernetes/blob/5a16163c87fe2a90916a51b52771a668bcaf2a0d/pkg/apis/admission/types.go#L42)
and [AdmissionResponse](https://github.com/kubernetes/kubernetes/blob/5a16163c87fe2a90916a51b52771a668bcaf2a0d/pkg/apis/admission/types.go#L84),
containing data accordingly to context.

When asked for validation the AdmissionRequest is filled with the data
belonging to the object being validated. The information contained in the
request comprises of the Name and the Namespace of the object, name of the
[operation](https://github.com/kubernetes/kubernetes/blob/5a16163c87fe2a90916a51b52771a668bcaf2a0d/pkg/apis/admission/types.go#L120)
being performed, object Kind, which subresource is demanded,
object resource and, of course, object itself.

For the purpose of a simple lock, the implementation will make use only of
the Name and the Namespace, following the flow shown in following diagram.

![Lock flow](https://github.com/petrkotas/k8s-object-lock/raw/master/pics/lock_flow.png "Lock flow")

The controller checks whether there is lock object with the same name in the same
namespace as the requested object. If lock exists, the request fails with
reponse "Object is locked".

There is one tricky part in the implementation. It is related to Name passed to
the request. The name is only present when it is not expected that it will be
generated by Kubernetes. Name generation is commonly used for Pods created by
deployements, but it applies also to other higher level controllers.
Luckily this is not an issue for the Lock, because locking only applies to
already existing objects. Name generation applies only to new objects. Therefore
to solve the missing name, simply skipping the "CREATE" operation is sufficient.

### Configuration

When the controller is done, it is time to register it in the Kubernetes.
The registration is straightforward. It follows the same pattern as everything
else. First the controller has to be deployed to the cluster.
Second the controller have to be registered in the Kubernetes API.

The overall process is fully described in the [official documentation](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/).
I will focus only on the murky parts of the configuration.

#### Validating webhook configuration

The configuration is as ussual a YAML. The main parts are:

* `clientConfig` - defines where the request are going to be delivered. There are
[two options](https://github.com/kubernetes/kubernetes/blob/bf3b5e55634cac542e7cce16ba5844b067018bb4/pkg/apis/admissionregistration/types.go#L239).
Either set the URL, which can run everywhere, or set the service and configure it
to run in the cluster. When service is configured, the accompanyning `caBundle`
has to be set properly. It contains the server certificate to verify the
communication. Only https is allowed to communicate with the Kubernetes API.
The right value for the `caBundle` is the base64 encoded PEM server certificate.
When in doubt, you can read great post about tool from [CloudFlare](https://blog.cloudflare.com/introducing-cfssl/)
how they handle certificates.
* `rules` - contains the rules used to filter which resources will be delivered
for the validation. `operations` limits for which action the resource will be validated.
For the Lock, only "UPDATE" and "DELETE" are relevant. `apiGroups` limits for
which API group to validate, can be for instance `kotas.tech` group. `apiVersions`
allows to limit validation only for a specific version, which is handy when testing
new stuff. Finally `resources` enables limiting only for those resources of
interest.
* [`namespaceSelector`](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/) -
is just a label selector and allows limiting validation only for desired namespace.

```yaml
apiVersion: admissionregistration.k8s.io/v1beta1
kind: ValidatingWebhookConfiguration
metadata:
  name: lockvalidation-cfg
  labels:
    app: lockvalidation
webhooks:
  - name: lockvalidation.kotas.tech
    clientConfig:
      service:
        name: lockvalidation-svc
        namespace: default
        path: "/validate"
      caBundle: {{ CA_BUNDLE }}
    rules:
      - operations:
          - UPDATE
          - DELETE
        apiGroups:
          - ""
        apiVersions:
          - v1
        resources:
          - "pods"
    namespaceSelector:
      matchLabels:
        lockable: "true"
```

The `caBundle` is filled by the helper script, that generates Kubernetes secret
to store the certificates for the server `hack/gen_certs.sh`.

#### Deployment

Deployment is exactly the same as any other Kubernetes deplyment. The one
murky part is `serviceAccountName`, which attaches account to the deplyment.
This is required, since the controller communicates with the Kubernetes API.
Without proper account, the calls would be rejected.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: lockvalidation-dpl
  labels:
    app: lockvalidation
spec:
  replicas: 1
  selector:
    matchLabels:
      app: lockvalidation
  template:
    metadata:
      labels:
        app: lockvalidation
    spec:
      serviceAccountName: lockvalidation-sa
      containers:
        - name: lockvalidation
          image: pkotas/lockvalidation:v1
          imagePullPolicy: IfNotPresent
          resources:
            requests:
              memory: "128Mi"
              cpu: "250m"
            limits:
              memory: "256Mi"
              cpu: "500m"
          volumeMounts:
            - name: webhook-certs
              mountPath: /etc/lockvalidation/cert
              readOnly: true
      volumes:
        - name: webhook-certs
          secret:
            secretName: lockvalidation-crt
```

To enable https, proper certificate and key has to be added to the server.
This is done via Kubernetes secret. It mounts the certificate to the pod.
The secret is generated via helper script in `hack/gen_certs.sh`.

Also, the service to expose the controller to the cluster is required.
It allows to connect to single stable endpoint, while the pod may change in time.

```yaml
apiVersion: v1
kind: Service
metadata:
  name: lockvalidation-svc
  labels:
    app: lockvalidation
spec:
  ports:
  - port: 443
    targetPort: 443
  selector:
    app: lockvalidation
```

#### Custom Resource Definition

The custom resource introducing the lock object is not complicated. It follows
the example given in the [documentation](https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/).
It does not have to be complicated, since it is not required to carry additional
information. Only its mere presence is sufficient to lock the object.

```yaml
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: locks.kotas.tech
spec:
  group: kotas.tech
  versions:
    - name: v1
      served: true
      storage: true
  scope: Namespaced
  names:
    plural: locks
    singular: lock
    kind: Lock
    shortNames:
    - l
```

#### Cluster role and service account

To enable controller to access the Kubernetes API, the cluster role has to be
created. It is as simple as allowing to access cluster resources such as pods
and resources belonging to the API group `kotas.tech`.

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: lockvalidation-cr
  labels:
    app: lockvalidation
rules:
- apiGroups:
  - kotas.tech
  resources:
  - "*"
  verbs:
  - "*"
- apiGroups:
  - ""
  resources:
  - pods
  verbs:
  - "*"
```

The cluster role has to be bounded to service account via cluster role binding.
