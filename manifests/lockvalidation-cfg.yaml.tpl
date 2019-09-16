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
          - CREATE
          - UPDATE
          - DELETE
          - CONNECT
        apiGroups:
          - "*"
        apiVersions: 
          - v1
        resources: 
          - "deployments/*"
          - "pods/*"
    namespaceSelector:
      matchLabels:
        lockable: "true"
