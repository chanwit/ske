package templates

const IstioVersion = "1.0.1"

const AddonIstioTemplate = `
apiVersion: v1
kind: Namespace
metadata:
  name: istio-system
  labels:
    istio-injection: disabled
---
# Source: istio/charts/galley/templates/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: istio-galley-configuration
  namespace: istio-system
  labels:
    app: istio-galley
    chart: galley-1.0.1
    release: RELEASE-NAME
    heritage: Tiller
    istio: mixer
data:
  validatingwebhookconfiguration.yaml: |-
    apiVersion: admissionregistration.k8s.io/v1beta1
    kind: ValidatingWebhookConfiguration
    metadata:
      name: istio-galley
      namespace: istio-system
      labels:
        app: istio-galley
        chart: galley-1.0.1
        release: RELEASE-NAME
        heritage: Tiller
    webhooks:
      - name: pilot.validation.istio.io
        clientConfig:
          service:
            name: istio-galley
            namespace: istio-system
            path: "/admitpilot"
          caBundle: ""
        rules:
          - operations:
            - CREATE
            - UPDATE
            apiGroups:
            - config.istio.io
            apiVersions:
            - v1alpha2
            resources:
            - httpapispecs
            - httpapispecbindings
            - quotaspecs
            - quotaspecbindings
          - operations:
            - CREATE
            - UPDATE
            apiGroups:
            - rbac.istio.io
            apiVersions:
            - "*"
            resources:
            - "*"
          - operations:
            - CREATE
            - UPDATE
            apiGroups:
            - authentication.istio.io
            apiVersions:
            - "*"
            resources:
            - "*"
          - operations:
            - CREATE
            - UPDATE
            apiGroups:
            - networking.istio.io
            apiVersions:
            - "*"
            resources:
            - destinationrules
            - envoyfilters
            - gateways
            # disabled per @costinm's request
            # - serviceentries
            - virtualservices
        failurePolicy: Fail
      - name: mixer.validation.istio.io
        clientConfig:
          service:
            name: istio-galley
            namespace: istio-system
            path: "/admitmixer"
          caBundle: ""
        rules:
          - operations:
            - CREATE
            - UPDATE
            apiGroups:
            - config.istio.io
            apiVersions:
            - v1alpha2
            resources:
            - rules
            - attributemanifests
            - circonuses
            - deniers
            - fluentds
            - kubernetesenvs
            - listcheckers
            - memquotas
            - noops
            - opas
            - prometheuses
            - rbacs
            - servicecontrols
            - solarwindses
            - stackdrivers
            - statsds
            - stdios
            - apikeys
            - authorizations
            - checknothings
            # - kuberneteses
            - listentries
            - logentries
            - metrics
            - quotas
            - reportnothings
            - servicecontrolreports
            - tracespans
        failurePolicy: Fail


---
# Source: istio/charts/grafana/templates/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: istio-grafana-custom-resources
  namespace: istio-system
  labels:
    app: istio-grafana
    chart: grafana-1.0.1
    release: RELEASE-NAME
    heritage: Tiller
    istio: grafana
data:
  custom-resources.yaml: |-
    apiVersion: authentication.istio.io/v1alpha1
    kind: Policy
    metadata:
      name: grafana-ports-mtls-disabled
      namespace: istio-system
    spec:
      targets:
      - name: grafana
        ports:
        - number: 3000
  run.sh: |-
    #!/bin/sh

    set -x

    if [ "$#" -ne "1" ]; then
        echo "first argument should be path to custom resource yaml"
        exit 1
    fi

    pathToResourceYAML=${1}

    /kubectl get validatingwebhookconfiguration istio-galley 2>/dev/null
    if [ "$?" -eq 0 ]; then
        echo "istio-galley validatingwebhookconfiguration found - waiting for istio-galley deployment to be ready"
        while true; do
            /kubectl -n istio-system get deployment istio-galley 2>/dev/null
            if [ "$?" -eq 0 ]; then
                break
            fi
            sleep 1
        done
        /kubectl -n istio-system rollout status deployment istio-galley
        if [ "$?" -ne 0 ]; then
            echo "istio-galley deployment rollout status check failed"
            exit 1
        fi
        echo "istio-galley deployment ready for configuration validation"
    fi
    sleep 5
    /kubectl apply -f ${pathToResourceYAML}


---
# Source: istio/charts/mixer/templates/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: istio-statsd-prom-bridge
  namespace: istio-system
  labels:
    app: istio-statsd-prom-bridge
    chart: mixer-1.0.1
    release: RELEASE-NAME
    heritage: Tiller
    istio: mixer
data:
  mapping.conf: |-

---
# Source: istio/charts/prometheus/templates/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus
  namespace: istio-system
  labels:
    app: prometheus
    chart: prometheus-1.0.1
    release: RELEASE-NAME
    heritage: Tiller
data:
  prometheus.yml: |-
    global:
      scrape_interval: 15s
    scrape_configs:

    - job_name: 'istio-mesh'
      # Override the global default and scrape targets from this job every 5 seconds.
      scrape_interval: 5s

      kubernetes_sd_configs:
      - role: endpoints
        namespaces:
          names:
          - istio-system

      relabel_configs:
      - source_labels: [__meta_kubernetes_service_name, __meta_kubernetes_endpoint_port_name]
        action: keep
        regex: istio-telemetry;prometheus

    - job_name: 'envoy'
      # Override the global default and scrape targets from this job every 5 seconds.
      scrape_interval: 5s
      # metrics_path defaults to '/metrics'
      # scheme defaults to 'http'.

      kubernetes_sd_configs:
      - role: endpoints
        namespaces:
          names:
          - istio-system

      relabel_configs:
      - source_labels: [__meta_kubernetes_service_name, __meta_kubernetes_endpoint_port_name]
        action: keep
        regex: istio-statsd-prom-bridge;statsd-prom

    - job_name: 'istio-policy'
      # Override the global default and scrape targets from this job every 5 seconds.
      scrape_interval: 5s
      # metrics_path defaults to '/metrics'
      # scheme defaults to 'http'.

      kubernetes_sd_configs:
      - role: endpoints
        namespaces:
          names:
          - istio-system


      relabel_configs:
      - source_labels: [__meta_kubernetes_service_name, __meta_kubernetes_endpoint_port_name]
        action: keep
        regex: istio-policy;http-monitoring

    - job_name: 'istio-telemetry'
      # Override the global default and scrape targets from this job every 5 seconds.
      scrape_interval: 5s
      # metrics_path defaults to '/metrics'
      # scheme defaults to 'http'.

      kubernetes_sd_configs:
      - role: endpoints
        namespaces:
          names:
          - istio-system

      relabel_configs:
      - source_labels: [__meta_kubernetes_service_name, __meta_kubernetes_endpoint_port_name]
        action: keep
        regex: istio-telemetry;http-monitoring

    - job_name: 'pilot'
      # Override the global default and scrape targets from this job every 5 seconds.
      scrape_interval: 5s
      # metrics_path defaults to '/metrics'
      # scheme defaults to 'http'.

      kubernetes_sd_configs:
      - role: endpoints
        namespaces:
          names:
          - istio-system

      relabel_configs:
      - source_labels: [__meta_kubernetes_service_name, __meta_kubernetes_endpoint_port_name]
        action: keep
        regex: istio-pilot;http-monitoring

    - job_name: 'galley'
      # Override the global default and scrape targets from this job every 5 seconds.
      scrape_interval: 5s
      # metrics_path defaults to '/metrics'
      # scheme defaults to 'http'.

      kubernetes_sd_configs:
      - role: endpoints
        namespaces:
          names:
          - istio-system

      relabel_configs:
      - source_labels: [__meta_kubernetes_service_name, __meta_kubernetes_endpoint_port_name]
        action: keep
        regex: istio-galley;http-monitoring

    # scrape config for API servers
    - job_name: 'kubernetes-apiservers'
      kubernetes_sd_configs:
      - role: endpoints
        namespaces:
          names:
          - default
      scheme: https
      tls_config:
        ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
      bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
      relabel_configs:
      - source_labels: [__meta_kubernetes_service_name, __meta_kubernetes_endpoint_port_name]
        action: keep
        regex: kubernetes;https

    # scrape config for nodes (kubelet)
    - job_name: 'kubernetes-nodes'
      scheme: https
      tls_config:
        ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
      bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
      kubernetes_sd_configs:
      - role: node
      relabel_configs:
      - action: labelmap
        regex: __meta_kubernetes_node_label_(.+)
      - target_label: __address__
        replacement: kubernetes.default.svc:443
      - source_labels: [__meta_kubernetes_node_name]
        regex: (.+)
        target_label: __metrics_path__
        replacement: /api/v1/nodes/${1}/proxy/metrics

    # Scrape config for Kubelet cAdvisor.
    #
    # This is required for Kubernetes 1.7.3 and later, where cAdvisor metrics
    # (those whose names begin with 'container_') have been removed from the
    # Kubelet metrics endpoint.  This job scrapes the cAdvisor endpoint to
    # retrieve those metrics.
    #
    # In Kubernetes 1.7.0-1.7.2, these metrics are only exposed on the cAdvisor
    # HTTP endpoint; use "replacement: /api/v1/nodes/${1}:4194/proxy/metrics"
    # in that case (and ensure cAdvisor's HTTP server hasn't been disabled with
    # the --cadvisor-port=0 Kubelet flag).
    #
    # This job is not necessary and should be removed in Kubernetes 1.6 and
    # earlier versions, or it will cause the metrics to be scraped twice.
    - job_name: 'kubernetes-cadvisor'
      scheme: https
      tls_config:
        ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
      bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
      kubernetes_sd_configs:
      - role: node
      relabel_configs:
      - action: labelmap
        regex: __meta_kubernetes_node_label_(.+)
      - target_label: __address__
        replacement: kubernetes.default.svc:443
      - source_labels: [__meta_kubernetes_node_name]
        regex: (.+)
        target_label: __metrics_path__
        replacement: /api/v1/nodes/${1}/proxy/metrics/cadvisor

    # scrape config for service endpoints.
    - job_name: 'kubernetes-service-endpoints'
      kubernetes_sd_configs:
      - role: endpoints
      relabel_configs:
      - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scrape]
        action: keep
        regex: true
      - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scheme]
        action: replace
        target_label: __scheme__
        regex: (https?)
      - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_path]
        action: replace
        target_label: __metrics_path__
        regex: (.+)
      - source_labels: [__address__, __meta_kubernetes_service_annotation_prometheus_io_port]
        action: replace
        target_label: __address__
        regex: ([^:]+)(?::\d+)?;(\d+)
        replacement: $1:$2
      - action: labelmap
        regex: __meta_kubernetes_service_label_(.+)
      - source_labels: [__meta_kubernetes_namespace]
        action: replace
        target_label: kubernetes_namespace
      - source_labels: [__meta_kubernetes_service_name]
        action: replace
        target_label: kubernetes_name

    # Example scrape config for pods
    - job_name: 'kubernetes-pods'
      kubernetes_sd_configs:
      - role: pod

      relabel_configs:
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
        action: keep
        regex: true
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
        action: replace
        target_label: __metrics_path__
        regex: (.+)
      - source_labels: [__address__, __meta_kubernetes_pod_annotation_prometheus_io_port]
        action: replace
        regex: ([^:]+)(?::\d+)?;(\d+)
        replacement: $1:$2
        target_label: __address__
      - action: labelmap
        regex: __meta_kubernetes_pod_label_(.+)
      - source_labels: [__meta_kubernetes_namespace]
        action: replace
        target_label: namespace
      - source_labels: [__meta_kubernetes_pod_name]
        action: replace
        target_label: pod_name

---
# Source: istio/charts/security/templates/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: istio-security-custom-resources
  namespace: istio-system
  labels:
    app: istio-security
    chart: security-1.0.1
    release: RELEASE-NAME
    heritage: Tiller
    istio: security
data:
  custom-resources.yaml: |-
  run.sh: |-
    #!/bin/sh

    set -x

    if [ "$#" -ne "1" ]; then
        echo "first argument should be path to custom resource yaml"
        exit 1
    fi

    pathToResourceYAML=${1}

    /kubectl get validatingwebhookconfiguration istio-galley 2>/dev/null
    if [ "$?" -eq 0 ]; then
        echo "istio-galley validatingwebhookconfiguration found - waiting for istio-galley deployment to be ready"
        while true; do
            /kubectl -n istio-system get deployment istio-galley 2>/dev/null
            if [ "$?" -eq 0 ]; then
                break
            fi
            sleep 1
        done
        /kubectl -n istio-system rollout status deployment istio-galley
        if [ "$?" -ne 0 ]; then
            echo "istio-galley deployment rollout status check failed"
            exit 1
        fi
        echo "istio-galley deployment ready for configuration validation"
    fi
    sleep 5
    /kubectl apply -f ${pathToResourceYAML}


---
# Source: istio/templates/configmap.yaml

apiVersion: v1
kind: ConfigMap
metadata:
  name: istio
  namespace: istio-system
  labels:
    app: istio
    chart: istio-1.0.1
    release: RELEASE-NAME
    heritage: Tiller
data:
  mesh: |-
    # Set the following variable to true to disable policy checks by the Mixer.
    # Note that metrics will still be reported to the Mixer.
    disablePolicyChecks: false

    # Set enableTracing to false to disable request tracing.
    enableTracing: true

    # Set accessLogFile to empty string to disable access log.
    accessLogFile: "/dev/stdout"
    #
    # Deprecated: mixer is using EDS
    mixerCheckServer: istio-policy.istio-system.svc.cluster.local:9091
    mixerReportServer: istio-telemetry.istio-system.svc.cluster.local:9091

    # Unix Domain Socket through which envoy communicates with NodeAgent SDS to get
    # key/cert for mTLS. Use secret-mount files instead of SDS if set to empty.
    sdsUdsPath: ""

    # How frequently should Envoy fetch key/cert from NodeAgent.
    sdsRefreshDelay: 15s

    #
    defaultConfig:
      #
      # TCP connection timeout between Envoy & the application, and between Envoys.
      connectTimeout: 10s
      #
      ### ADVANCED SETTINGS #############
      # Where should envoy's configuration be stored in the istio-proxy container
      configPath: "/etc/istio/proxy"
      binaryPath: "/usr/local/bin/envoy"
      # The pseudo service name used for Envoy.
      serviceCluster: istio-proxy
      # These settings that determine how long an old Envoy
      # process should be kept alive after an occasional reload.
      drainDuration: 45s
      parentShutdownDuration: 1m0s
      #
      # The mode used to redirect inbound connections to Envoy. This setting
      # has no effect on outbound traffic: iptables REDIRECT is always used for
      # outbound connections.
      # If "REDIRECT", use iptables REDIRECT to NAT and redirect to Envoy.
      # The "REDIRECT" mode loses source addresses during redirection.
      # If "TPROXY", use iptables TPROXY to redirect to Envoy.
      # The "TPROXY" mode preserves both the source and destination IP
      # addresses and ports, so that they can be used for advanced filtering
      # and manipulation.
      # The "TPROXY" mode also configures the sidecar to run with the
      # CAP_NET_ADMIN capability, which is required to use TPROXY.
      #interceptionMode: REDIRECT
      #
      # Port where Envoy listens (on local host) for admin commands
      # You can exec into the istio-proxy container in a pod and
      # curl the admin port (curl http://localhost:15000/) to obtain
      # diagnostic information from Envoy. See
      # https://lyft.github.io/envoy/docs/operations/admin.html
      # for more details
      proxyAdminPort: 15000
      #
      # Zipkin trace collector
      zipkinAddress: zipkin.istio-system:9411
      #
      # Statsd metrics collector converts statsd metrics into Prometheus metrics.
      statsdUdpAddress: istio-statsd-prom-bridge.istio-system:9125
      #
      # Mutual TLS authentication between sidecars and istio control plane.
      controlPlaneAuthPolicy: NONE
      #
      # Address where istio Pilot service is running
      discoveryAddress: istio-pilot.istio-system:15007

---
# Source: istio/templates/sidecar-injector-configmap.yaml

apiVersion: v1
kind: ConfigMap
metadata:
  name: istio-sidecar-injector
  namespace: istio-system
  labels:
    app: istio
    chart: istio-1.0.1
    release: RELEASE-NAME
    heritage: Tiller
    istio: sidecar-injector
data:
  config: |-
    policy: enabled
    template: |-
      initContainers:
      - name: istio-init
        image: "gcr.io/istio-release/proxy_init:1.0.1"
        args:
        - "-p"
        - [[ .MeshConfig.ProxyListenPort ]]
        - "-u"
        - 1337
        - "-m"
        - [[ or (index .ObjectMeta.Annotations "sidecar.istio.io/interceptionMode") .ProxyConfig.InterceptionMode.String ]]
        - "-i"
        [[ if (isset .ObjectMeta.Annotations "traffic.sidecar.istio.io/includeOutboundIPRanges") -]]
        - "[[ index .ObjectMeta.Annotations "traffic.sidecar.istio.io/includeOutboundIPRanges"  ]]"
        [[ else -]]
        - "*"
        [[ end -]]
        - "-x"
        [[ if (isset .ObjectMeta.Annotations "traffic.sidecar.istio.io/excludeOutboundIPRanges") -]]
        - "[[ index .ObjectMeta.Annotations "traffic.sidecar.istio.io/excludeOutboundIPRanges"  ]]"
        [[ else -]]
        - ""
        [[ end -]]
        - "-b"
        [[ if (isset .ObjectMeta.Annotations "traffic.sidecar.istio.io/includeInboundPorts") -]]
        - "[[ index .ObjectMeta.Annotations "traffic.sidecar.istio.io/includeInboundPorts"  ]]"
        [[ else -]]
        - [[ range .Spec.Containers -]][[ range .Ports -]][[ .ContainerPort -]], [[ end -]][[ end -]][[ end]]
        - "-d"
        [[ if (isset .ObjectMeta.Annotations "traffic.sidecar.istio.io/excludeInboundPorts") -]]
        - "[[ index .ObjectMeta.Annotations "traffic.sidecar.istio.io/excludeInboundPorts" ]]"
        [[ else -]]
        - ""
        [[ end -]]
        imagePullPolicy: IfNotPresent
        securityContext:
          capabilities:
            add:
            - NET_ADMIN
          restartPolicy: Always

      containers:
      - name: istio-proxy
        image: [[ if (isset .ObjectMeta.Annotations "sidecar.istio.io/proxyImage") -]]
        "[[ index .ObjectMeta.Annotations "sidecar.istio.io/proxyImage" ]]"
        [[ else -]]
        gcr.io/istio-release/proxyv2:1.0.1
        [[ end -]]
        args:
        - proxy
        - sidecar
        - --configPath
        - [[ .ProxyConfig.ConfigPath ]]
        - --binaryPath
        - [[ .ProxyConfig.BinaryPath ]]
        - --serviceCluster
        [[ if ne "" (index .ObjectMeta.Labels "app") -]]
        - [[ index .ObjectMeta.Labels "app" ]]
        [[ else -]]
        - "istio-proxy"
        [[ end -]]
        - --drainDuration
        - [[ formatDuration .ProxyConfig.DrainDuration ]]
        - --parentShutdownDuration
        - [[ formatDuration .ProxyConfig.ParentShutdownDuration ]]
        - --discoveryAddress
        - [[ .ProxyConfig.DiscoveryAddress ]]
        - --discoveryRefreshDelay
        - [[ formatDuration .ProxyConfig.DiscoveryRefreshDelay ]]
        - --zipkinAddress
        - [[ .ProxyConfig.ZipkinAddress ]]
        - --connectTimeout
        - [[ formatDuration .ProxyConfig.ConnectTimeout ]]
        - --statsdUdpAddress
        - [[ .ProxyConfig.StatsdUdpAddress ]]
        - --proxyAdminPort
        - [[ .ProxyConfig.ProxyAdminPort ]]
        - --controlPlaneAuthPolicy
        - [[ or (index .ObjectMeta.Annotations "sidecar.istio.io/controlPlaneAuthPolicy") .ProxyConfig.ControlPlaneAuthPolicy ]]
        env:
        - name: POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        - name: INSTANCE_IP
          valueFrom:
            fieldRef:
              fieldPath: status.podIP
        - name: ISTIO_META_POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: ISTIO_META_INTERCEPTION_MODE
          value: [[ or (index .ObjectMeta.Annotations "sidecar.istio.io/interceptionMode") .ProxyConfig.InterceptionMode.String ]]
        imagePullPolicy: IfNotPresent
        securityContext:
          readOnlyRootFilesystem: true
          [[ if eq (or (index .ObjectMeta.Annotations "sidecar.istio.io/interceptionMode") .ProxyConfig.InterceptionMode.String) "TPROXY" -]]
          capabilities:
            add:
            - NET_ADMIN
          runAsGroup: 1337
          [[ else -]]
          runAsUser: 1337
          [[ end -]]
        restartPolicy: Always
        resources:
          [[ if (isset .ObjectMeta.Annotations "sidecar.istio.io/proxyCPU") -]]
          requests:
            cpu: "[[ index .ObjectMeta.Annotations "sidecar.istio.io/proxyCPU" ]]"
            memory: "[[ index .ObjectMeta.Annotations "sidecar.istio.io/proxyMemory" ]]"
        [[ else -]]
          requests:
            cpu: 10m

        [[ end -]]
        volumeMounts:
        - mountPath: /etc/istio/proxy
          name: istio-envoy
        - mountPath: /etc/certs/
          name: istio-certs
          readOnly: true
      volumes:
      - emptyDir:
          medium: Memory
        name: istio-envoy
      - name: istio-certs
        secret:
          optional: true
          [[ if eq .Spec.ServiceAccountName "" -]]
          secretName: istio.default
          [[ else -]]
          secretName: [[ printf "istio.%s" .Spec.ServiceAccountName ]]
          [[ end -]]

---
# Source: istio/charts/galley/templates/serviceaccount.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: istio-galley-service-account
  namespace: istio-system
  labels:
    app: istio-galley
    chart: galley-1.0.1
    heritage: Tiller
    release: RELEASE-NAME

---
# Source: istio/charts/gateways/templates/serviceaccount.yaml

apiVersion: v1
kind: ServiceAccount
metadata:
  name: istio-egressgateway-service-account
  namespace: istio-system
  labels:
    app: egressgateway
    chart: gateways-1.0.1
    heritage: Tiller
    release: RELEASE-NAME
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: istio-ingressgateway-service-account
  namespace: istio-system
  labels:
    app: ingressgateway
    chart: gateways-1.0.1
    heritage: Tiller
    release: RELEASE-NAME
---

---
# Source: istio/charts/grafana/templates/create-custom-resources-job.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: istio-grafana-post-install-account
  namespace: istio-system
  labels:
    app: istio-grafana
    chart: grafana-1.0.1
    heritage: Tiller
    release: RELEASE-NAME
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: istio-grafana-post-install-istio-system
  labels:
    app: istio-grafana
    chart: grafana-1.0.1
    heritage: Tiller
    release: RELEASE-NAME
rules:
- apiGroups: ["authentication.istio.io"] # needed to create default authn policy
  resources: ["*"]
  verbs: ["*"]
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: istio-grafana-post-install-role-binding-istio-system
  labels:
    app: istio-grafana
    chart: grafana-1.0.1
    heritage: Tiller
    release: RELEASE-NAME
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: istio-grafana-post-install-istio-system
subjects:
  - kind: ServiceAccount
    name: istio-grafana-post-install-account
    namespace: istio-system
---
apiVersion: batch/v1
kind: Job
metadata:
  name: istio-grafana-post-install
  namespace: istio-system
  annotations:
    "helm.sh/hook": post-install
    "helm.sh/hook-delete-policy": hook-succeeded
  labels:
    app: istio-grafana
    chart: grafana-1.0.1
    release: RELEASE-NAME
    heritage: Tiller
spec:
  template:
    metadata:
      name: istio-grafana-post-install
      labels:
        app: istio-grafana
        release: RELEASE-NAME
    spec:
      serviceAccountName: istio-grafana-post-install-account
      containers:
        - name: hyperkube
          image: "quay.io/coreos/hyperkube:v1.7.6_coreos.0"
          command: [ "/bin/bash", "/tmp/grafana/run.sh", "/tmp/grafana/custom-resources.yaml" ]
          volumeMounts:
            - mountPath: "/tmp/grafana"
              name: tmp-configmap-grafana
      volumes:
        - name: tmp-configmap-grafana
          configMap:
            name: istio-grafana-custom-resources
      restartPolicy: OnFailure

---
# Source: istio/charts/mixer/templates/serviceaccount.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: istio-mixer-service-account
  namespace: istio-system
  labels:
    app: mixer
    chart: mixer-1.0.1
    heritage: Tiller
    release: RELEASE-NAME

---
# Source: istio/charts/pilot/templates/serviceaccount.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: istio-pilot-service-account
  namespace: istio-system
  labels:
    app: istio-pilot
    chart: pilot-1.0.1
    heritage: Tiller
    release: RELEASE-NAME

---
# Source: istio/charts/prometheus/templates/serviceaccount.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: prometheus
  namespace: istio-system

---
# Source: istio/charts/security/templates/cleanup-secrets.yaml
# The reason for creating a ServiceAccount and ClusterRole specifically for this
# post-delete hooked job is because the citadel ServiceAccount is being deleted
# before this hook is launched. On the other hand, running this hook before the
# deletion of the citadel (e.g. pre-delete) won't delete the secrets because they
# will be re-created immediately by the to-be-deleted citadel.
#
# It's also important that the ServiceAccount, ClusterRole and ClusterRoleBinding
# will be ready before running the hooked Job therefore the hook weights.

apiVersion: v1
kind: ServiceAccount
metadata:
  name: istio-cleanup-secrets-service-account
  namespace: istio-system
  annotations:
    "helm.sh/hook": post-delete
    "helm.sh/hook-delete-policy": hook-succeeded
    "helm.sh/hook-weight": "1"
  labels:
    app: security
    chart: security-1.0.1
    heritage: Tiller
    release: RELEASE-NAME
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: istio-cleanup-secrets-istio-system
  annotations:
    "helm.sh/hook": post-delete
    "helm.sh/hook-delete-policy": hook-succeeded
    "helm.sh/hook-weight": "1"
  labels:
    app: security
    chart: security-1.0.1
    heritage: Tiller
    release: RELEASE-NAME
rules:
- apiGroups: [""]
  resources: ["secrets"]
  verbs: ["list", "delete"]
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: istio-cleanup-secrets-istio-system
  annotations:
    "helm.sh/hook": post-delete
    "helm.sh/hook-delete-policy": hook-succeeded
    "helm.sh/hook-weight": "2"
  labels:
    app: security
    chart: security-1.0.1
    heritage: Tiller
    release: RELEASE-NAME
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: istio-cleanup-secrets-istio-system
subjects:
  - kind: ServiceAccount
    name: istio-cleanup-secrets-service-account
    namespace: istio-system
---
apiVersion: batch/v1
kind: Job
metadata:
  name: istio-cleanup-secrets
  namespace: istio-system
  annotations:
    "helm.sh/hook": post-delete
    "helm.sh/hook-delete-policy": hook-succeeded
    "helm.sh/hook-weight": "3"
  labels:
    app: security
    chart: security-1.0.1
    release: RELEASE-NAME
    heritage: Tiller
spec:
  template:
    metadata:
      name: istio-cleanup-secrets
      labels:
        app: security
        release: RELEASE-NAME
    spec:
      serviceAccountName: istio-cleanup-secrets-service-account
      containers:
        - name: hyperkube
          image: "quay.io/coreos/hyperkube:v1.7.6_coreos.0"
          command:
          - /bin/bash
          - -c
          - >
              kubectl get secret --all-namespaces | grep "istio.io/key-and-cert" |  while read -r entry; do
                ns=$(echo $entry | awk '{print $1}');
                name=$(echo $entry | awk '{print $2}');
                kubectl delete secret $name -n $ns;
              done
      restartPolicy: OnFailure

---
# Source: istio/charts/security/templates/serviceaccount.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: istio-citadel-service-account
  namespace: istio-system
  labels:
    app: security
    chart: security-1.0.1
    heritage: Tiller
    release: RELEASE-NAME

---
# Source: istio/charts/sidecarInjectorWebhook/templates/serviceaccount.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: istio-sidecar-injector-service-account
  namespace: istio-system
  labels:
    app: istio-sidecar-injector
    chart: sidecarInjectorWebhook-1.0.1
    heritage: Tiller
    release: RELEASE-NAME

---
# Source: istio/templates/crds.yaml
#
# these CRDs only make sense when pilot is enabled
#
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: virtualservices.networking.istio.io
  annotations:
    "helm.sh/hook": crd-install
  labels:
    app: istio-pilot
spec:
  group: networking.istio.io
  names:
    kind: VirtualService
    listKind: VirtualServiceList
    plural: virtualservices
    singular: virtualservice
    categories:
    - istio-io
    - networking-istio-io
  scope: Namespaced
  version: v1alpha3
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: destinationrules.networking.istio.io
  annotations:
    "helm.sh/hook": crd-install
  labels:
    app: istio-pilot
spec:
  group: networking.istio.io
  names:
    kind: DestinationRule
    listKind: DestinationRuleList
    plural: destinationrules
    singular: destinationrule
    categories:
    - istio-io
    - networking-istio-io
  scope: Namespaced
  version: v1alpha3
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: serviceentries.networking.istio.io
  annotations:
    "helm.sh/hook": crd-install
  labels:
    app: istio-pilot
spec:
  group: networking.istio.io
  names:
    kind: ServiceEntry
    listKind: ServiceEntryList
    plural: serviceentries
    singular: serviceentry
    categories:
    - istio-io
    - networking-istio-io
  scope: Namespaced
  version: v1alpha3
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: gateways.networking.istio.io
  annotations:
    "helm.sh/hook": crd-install
    "helm.sh/hook-weight": "-5"
  labels:
    app: istio-pilot
spec:
  group: networking.istio.io
  names:
    kind: Gateway
    plural: gateways
    singular: gateway
    categories:
    - istio-io
    - networking-istio-io
  scope: Namespaced
  version: v1alpha3
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: envoyfilters.networking.istio.io
  annotations:
    "helm.sh/hook": crd-install
  labels:
    app: istio-pilot
spec:
  group: networking.istio.io
  names:
    kind: EnvoyFilter
    plural: envoyfilters
    singular: envoyfilter
    categories:
    - istio-io
    - networking-istio-io
  scope: Namespaced
  version: v1alpha3
---
#

# these CRDs only make sense when security is enabled
#

#
kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  annotations:
    "helm.sh/hook": crd-install
  name: httpapispecbindings.config.istio.io
spec:
  group: config.istio.io
  names:
    kind: HTTPAPISpecBinding
    plural: httpapispecbindings
    singular: httpapispecbinding
    categories:
    - istio-io
    - apim-istio-io
  scope: Namespaced
  version: v1alpha2
---
kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  annotations:
    "helm.sh/hook": crd-install
  name: httpapispecs.config.istio.io
spec:
  group: config.istio.io
  names:
    kind: HTTPAPISpec
    plural: httpapispecs
    singular: httpapispec
    categories:
    - istio-io
    - apim-istio-io
  scope: Namespaced
  version: v1alpha2
---
kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  annotations:
    "helm.sh/hook": crd-install
  name: quotaspecbindings.config.istio.io
spec:
  group: config.istio.io
  names:
    kind: QuotaSpecBinding
    plural: quotaspecbindings
    singular: quotaspecbinding
    categories:
    - istio-io
    - apim-istio-io
  scope: Namespaced
  version: v1alpha2
---
kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  annotations:
    "helm.sh/hook": crd-install
  name: quotaspecs.config.istio.io
spec:
  group: config.istio.io
  names:
    kind: QuotaSpec
    plural: quotaspecs
    singular: quotaspec
    categories:
    - istio-io
    - apim-istio-io
  scope: Namespaced
  version: v1alpha2
---

# Mixer CRDs
kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: rules.config.istio.io
  annotations:
    "helm.sh/hook": crd-install
  labels:
    app: mixer
    package: istio.io.mixer
    istio: core
spec:
  group: config.istio.io
  names:
    kind: rule
    plural: rules
    singular: rule
    categories:
    - istio-io
    - policy-istio-io
  scope: Namespaced
  version: v1alpha2
---

kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: attributemanifests.config.istio.io
  annotations:
    "helm.sh/hook": crd-install
  labels:
    app: mixer
    package: istio.io.mixer
    istio: core
spec:
  group: config.istio.io
  names:
    kind: attributemanifest
    plural: attributemanifests
    singular: attributemanifest
    categories:
    - istio-io
    - policy-istio-io
  scope: Namespaced
  version: v1alpha2
---

kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: bypasses.config.istio.io
  annotations:
    "helm.sh/hook": crd-install
  labels:
    app: mixer
    package: bypass
    istio: mixer-adapter
spec:
  group: config.istio.io
  names:
    kind: bypass
    plural: bypasses
    singular: bypass
    categories:
    - istio-io
    - policy-istio-io
  scope: Namespaced
  version: v1alpha2
---

kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: circonuses.config.istio.io
  annotations:
    "helm.sh/hook": crd-install
  labels:
    app: mixer
    package: circonus
    istio: mixer-adapter
spec:
  group: config.istio.io
  names:
    kind: circonus
    plural: circonuses
    singular: circonus
    categories:
    - istio-io
    - policy-istio-io
  scope: Namespaced
  version: v1alpha2
---

kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: deniers.config.istio.io
  annotations:
    "helm.sh/hook": crd-install
  labels:
    app: mixer
    package: denier
    istio: mixer-adapter
spec:
  group: config.istio.io
  names:
    kind: denier
    plural: deniers
    singular: denier
    categories:
    - istio-io
    - policy-istio-io
  scope: Namespaced
  version: v1alpha2
---

kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: fluentds.config.istio.io
  annotations:
    "helm.sh/hook": crd-install
  labels:
    app: mixer
    package: fluentd
    istio: mixer-adapter
spec:
  group: config.istio.io
  names:
    kind: fluentd
    plural: fluentds
    singular: fluentd
    categories:
    - istio-io
    - policy-istio-io
  scope: Namespaced
  version: v1alpha2
---

kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: kubernetesenvs.config.istio.io
  annotations:
    "helm.sh/hook": crd-install
  labels:
    app: mixer
    package: kubernetesenv
    istio: mixer-adapter
spec:
  group: config.istio.io
  names:
    kind: kubernetesenv
    plural: kubernetesenvs
    singular: kubernetesenv
    categories:
    - istio-io
    - policy-istio-io
  scope: Namespaced
  version: v1alpha2
---

kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: listcheckers.config.istio.io
  annotations:
    "helm.sh/hook": crd-install
  labels:
    app: mixer
    package: listchecker
    istio: mixer-adapter
spec:
  group: config.istio.io
  names:
    kind: listchecker
    plural: listcheckers
    singular: listchecker
    categories:
    - istio-io
    - policy-istio-io
  scope: Namespaced
  version: v1alpha2
---

kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: memquotas.config.istio.io
  annotations:
    "helm.sh/hook": crd-install
  labels:
    app: mixer
    package: memquota
    istio: mixer-adapter
spec:
  group: config.istio.io
  names:
    kind: memquota
    plural: memquotas
    singular: memquota
    categories:
    - istio-io
    - policy-istio-io
  scope: Namespaced
  version: v1alpha2
---

kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: noops.config.istio.io
  annotations:
    "helm.sh/hook": crd-install
  labels:
    app: mixer
    package: noop
    istio: mixer-adapter
spec:
  group: config.istio.io
  names:
    kind: noop
    plural: noops
    singular: noop
    categories:
    - istio-io
    - policy-istio-io
  scope: Namespaced
  version: v1alpha2
---

kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: opas.config.istio.io
  annotations:
    "helm.sh/hook": crd-install
  labels:
    app: mixer
    package: opa
    istio: mixer-adapter
spec:
  group: config.istio.io
  names:
    kind: opa
    plural: opas
    singular: opa
    categories:
    - istio-io
    - policy-istio-io
  scope: Namespaced
  version: v1alpha2
---

kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: prometheuses.config.istio.io
  annotations:
    "helm.sh/hook": crd-install
  labels:
    app: mixer
    package: prometheus
    istio: mixer-adapter
spec:
  group: config.istio.io
  names:
    kind: prometheus
    plural: prometheuses
    singular: prometheus
    categories:
    - istio-io
    - policy-istio-io
  scope: Namespaced
  version: v1alpha2
---

kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: rbacs.config.istio.io
  annotations:
    "helm.sh/hook": crd-install
  labels:
    app: mixer
    package: rbac
    istio: mixer-adapter
spec:
  group: config.istio.io
  names:
    kind: rbac
    plural: rbacs
    singular: rbac
    categories:
    - istio-io
    - policy-istio-io
  scope: Namespaced
  version: v1alpha2
---

kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: redisquotas.config.istio.io
  annotations:
    "helm.sh/hook": crd-install
  labels:
    package: redisquota
    istio: mixer-adapter
spec:
  group: config.istio.io
  names:
    kind: redisquota
    plural: redisquotas
    singular: redisquota
  scope: Namespaced
  version: v1alpha2
---

kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: servicecontrols.config.istio.io
  annotations:
    "helm.sh/hook": crd-install
  labels:
    app: mixer
    package: servicecontrol
    istio: mixer-adapter
spec:
  group: config.istio.io
  names:
    kind: servicecontrol
    plural: servicecontrols
    singular: servicecontrol
    categories:
    - istio-io
    - policy-istio-io
  scope: Namespaced
  version: v1alpha2

---

kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: signalfxs.config.istio.io
  annotations:
    "helm.sh/hook": crd-install
  labels:
    app: mixer
    package: signalfx
    istio: mixer-adapter
spec:
  group: config.istio.io
  names:
    kind: signalfx
    plural: signalfxs
    singular: signalfx
    categories:
    - istio-io
    - policy-istio-io
  scope: Namespaced
  version: v1alpha2
---

kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: solarwindses.config.istio.io
  annotations:
    "helm.sh/hook": crd-install
  labels:
    app: mixer
    package: solarwinds
    istio: mixer-adapter
spec:
  group: config.istio.io
  names:
    kind: solarwinds
    plural: solarwindses
    singular: solarwinds
    categories:
    - istio-io
    - policy-istio-io
  scope: Namespaced
  version: v1alpha2
---

kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: stackdrivers.config.istio.io
  annotations:
    "helm.sh/hook": crd-install
  labels:
    app: mixer
    package: stackdriver
    istio: mixer-adapter
spec:
  group: config.istio.io
  names:
    kind: stackdriver
    plural: stackdrivers
    singular: stackdriver
    categories:
    - istio-io
    - policy-istio-io
  scope: Namespaced
  version: v1alpha2
---

kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: statsds.config.istio.io
  annotations:
    "helm.sh/hook": crd-install
  labels:
    app: mixer
    package: statsd
    istio: mixer-adapter
spec:
  group: config.istio.io
  names:
    kind: statsd
    plural: statsds
    singular: statsd
    categories:
    - istio-io
    - policy-istio-io
  scope: Namespaced
  version: v1alpha2
---

kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: stdios.config.istio.io
  annotations:
    "helm.sh/hook": crd-install
  labels:
    app: mixer
    package: stdio
    istio: mixer-adapter
spec:
  group: config.istio.io
  names:
    kind: stdio
    plural: stdios
    singular: stdio
    categories:
    - istio-io
    - policy-istio-io
  scope: Namespaced
  version: v1alpha2
---

kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: apikeys.config.istio.io
  annotations:
    "helm.sh/hook": crd-install
  labels:
    app: mixer
    package: apikey
    istio: mixer-instance
spec:
  group: config.istio.io
  names:
    kind: apikey
    plural: apikeys
    singular: apikey
    categories:
    - istio-io
    - policy-istio-io
  scope: Namespaced
  version: v1alpha2
---

kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: authorizations.config.istio.io
  annotations:
    "helm.sh/hook": crd-install
  labels:
    app: mixer
    package: authorization
    istio: mixer-instance
spec:
  group: config.istio.io
  names:
    kind: authorization
    plural: authorizations
    singular: authorization
    categories:
    - istio-io
    - policy-istio-io
  scope: Namespaced
  version: v1alpha2
---

kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: checknothings.config.istio.io
  annotations:
    "helm.sh/hook": crd-install
  labels:
    app: mixer
    package: checknothing
    istio: mixer-instance
spec:
  group: config.istio.io
  names:
    kind: checknothing
    plural: checknothings
    singular: checknothing
    categories:
    - istio-io
    - policy-istio-io
  scope: Namespaced
  version: v1alpha2
---

kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: kuberneteses.config.istio.io
  annotations:
    "helm.sh/hook": crd-install
  labels:
    app: mixer
    package: adapter.template.kubernetes
    istio: mixer-instance
spec:
  group: config.istio.io
  names:
    kind: kubernetes
    plural: kuberneteses
    singular: kubernetes
    categories:
    - istio-io
    - policy-istio-io
  scope: Namespaced
  version: v1alpha2
---

kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: listentries.config.istio.io
  annotations:
    "helm.sh/hook": crd-install
  labels:
    app: mixer
    package: listentry
    istio: mixer-instance
spec:
  group: config.istio.io
  names:
    kind: listentry
    plural: listentries
    singular: listentry
    categories:
    - istio-io
    - policy-istio-io
  scope: Namespaced
  version: v1alpha2
---

kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: logentries.config.istio.io
  annotations:
    "helm.sh/hook": crd-install
  labels:
    app: mixer
    package: logentry
    istio: mixer-instance
spec:
  group: config.istio.io
  names:
    kind: logentry
    plural: logentries
    singular: logentry
    categories:
    - istio-io
    - policy-istio-io
  scope: Namespaced
  version: v1alpha2
---

kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: edges.config.istio.io
  annotations:
    "helm.sh/hook": crd-install
  labels:
    app: mixer
    package: edge
    istio: mixer-instance
spec:
  group: config.istio.io
  names:
    kind: edge
    plural: edges
    singular: edge
    categories:
    - istio-io
    - policy-istio-io
  scope: Namespaced
  version: v1alpha2
---

kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: metrics.config.istio.io
  annotations:
    "helm.sh/hook": crd-install
  labels:
    app: mixer
    package: metric
    istio: mixer-instance
spec:
  group: config.istio.io
  names:
    kind: metric
    plural: metrics
    singular: metric
    categories:
    - istio-io
    - policy-istio-io
  scope: Namespaced
  version: v1alpha2
---

kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: quotas.config.istio.io
  annotations:
    "helm.sh/hook": crd-install
  labels:
    app: mixer
    package: quota
    istio: mixer-instance
spec:
  group: config.istio.io
  names:
    kind: quota
    plural: quotas
    singular: quota
    categories:
    - istio-io
    - policy-istio-io
  scope: Namespaced
  version: v1alpha2
---

kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: reportnothings.config.istio.io
  annotations:
    "helm.sh/hook": crd-install
  labels:
    app: mixer
    package: reportnothing
    istio: mixer-instance
spec:
  group: config.istio.io
  names:
    kind: reportnothing
    plural: reportnothings
    singular: reportnothing
    categories:
    - istio-io
    - policy-istio-io
  scope: Namespaced
  version: v1alpha2
---

kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: servicecontrolreports.config.istio.io
  annotations:
    "helm.sh/hook": crd-install
  labels:
    app: mixer
    package: servicecontrolreport
    istio: mixer-instance
spec:
  group: config.istio.io
  names:
    kind: servicecontrolreport
    plural: servicecontrolreports
    singular: servicecontrolreport
    categories:
    - istio-io
    - policy-istio-io
  scope: Namespaced
  version: v1alpha2
---

kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: tracespans.config.istio.io
  annotations:
    "helm.sh/hook": crd-install
  labels:
    app: mixer
    package: tracespan
    istio: mixer-instance
spec:
  group: config.istio.io
  names:
    kind: tracespan
    plural: tracespans
    singular: tracespan
    categories:
    - istio-io
    - policy-istio-io
  scope: Namespaced
  version: v1alpha2
---

kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: rbacconfigs.rbac.istio.io
  annotations:
    "helm.sh/hook": crd-install
  labels:
    app: mixer
    package: istio.io.mixer
    istio: rbac
spec:
  group: rbac.istio.io
  names:
    kind: RbacConfig
    plural: rbacconfigs
    singular: rbacconfig
    categories:
    - istio-io
    - rbac-istio-io
  scope: Namespaced
  version: v1alpha1
---

kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: serviceroles.rbac.istio.io
  annotations:
    "helm.sh/hook": crd-install
  labels:
    app: mixer
    package: istio.io.mixer
    istio: rbac
spec:
  group: rbac.istio.io
  names:
    kind: ServiceRole
    plural: serviceroles
    singular: servicerole
    categories:
    - istio-io
    - rbac-istio-io
  scope: Namespaced
  version: v1alpha1
---

kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: servicerolebindings.rbac.istio.io
  annotations:
    "helm.sh/hook": crd-install
  labels:
    app: mixer
    package: istio.io.mixer
    istio: rbac
spec:
  group: rbac.istio.io
  names:
    kind: ServiceRoleBinding
    plural: servicerolebindings
    singular: servicerolebinding
    categories:
    - istio-io
    - rbac-istio-io
  scope: Namespaced
  version: v1alpha1
---
kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: adapters.config.istio.io
  annotations:
    "helm.sh/hook": crd-install
  labels:
    app: mixer
    package: adapter
    istio: mixer-adapter
spec:
  group: config.istio.io
  names:
    kind: adapter
    plural: adapters
    singular: adapter
    categories:
    - istio-io
    - policy-istio-io
  scope: Namespaced
  version: v1alpha2
---
kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: instances.config.istio.io
  annotations:
    "helm.sh/hook": crd-install
  labels:
    app: mixer
    package: instance
    istio: mixer-instance
spec:
  group: config.istio.io
  names:
    kind: instance
    plural: instances
    singular: instance
    categories:
    - istio-io
    - policy-istio-io
  scope: Namespaced
  version: v1alpha2
---
kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: templates.config.istio.io
  annotations:
    "helm.sh/hook": crd-install
  labels:
    app: mixer
    package: template
    istio: mixer-template
spec:
  group: config.istio.io
  names:
    kind: template
    plural: templates
    singular: template
    categories:
    - istio-io
    - policy-istio-io
  scope: Namespaced
  version: v1alpha2
---
kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: handlers.config.istio.io
  annotations:
    "helm.sh/hook": crd-install
  labels:
    app: mixer
    package: handler
    istio: mixer-handler
spec:
  group: config.istio.io
  names:
    kind: handler
    plural: handlers
    singular: handler
    categories:
    - istio-io
    - policy-istio-io
  scope: Namespaced
  version: v1alpha2
---
#
#
---
# Source: istio/charts/galley/templates/clusterrole.yaml
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: istio-galley-istio-system
  labels:
    app: istio-galley
    chart: galley-1.0.1
    heritage: Tiller
    release: RELEASE-NAME
rules:
- apiGroups: ["admissionregistration.k8s.io"]
  resources: ["validatingwebhookconfigurations"]
  verbs: ["*"]
- apiGroups: ["config.istio.io"] # istio mixer CRD watcher
  resources: ["*"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["*"]
  resources: ["deployments"]
  resourceNames: ["istio-galley"]
  verbs: ["get"]
- apiGroups: ["*"]
  resources: ["endpoints"]
  resourceNames: ["istio-galley"]
  verbs: ["get"]

---
# Source: istio/charts/gateways/templates/clusterrole.yaml

apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  labels:
    app: gateways
    chart: gateways-1.0.1
    heritage: Tiller
    release: RELEASE-NAME
  name: istio-egressgateway-istio-system
rules:
- apiGroups: ["extensions"]
  resources: ["thirdpartyresources", "virtualservices", "destinationrules", "gateways"]
  verbs: ["get", "watch", "list", "update"]
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  labels:
    app: gateways
    chart: gateways-1.0.1
    heritage: Tiller
    release: RELEASE-NAME
  name: istio-ingressgateway-istio-system
rules:
- apiGroups: ["extensions"]
  resources: ["thirdpartyresources", "virtualservices", "destinationrules", "gateways"]
  verbs: ["get", "watch", "list", "update"]
---

---
# Source: istio/charts/mixer/templates/clusterrole.yaml
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: istio-mixer-istio-system
  labels:
    app: mixer
    chart: mixer-1.0.1
    heritage: Tiller
    release: RELEASE-NAME
rules:
- apiGroups: ["config.istio.io"] # istio CRD watcher
  resources: ["*"]
  verbs: ["create", "get", "list", "watch", "patch"]
- apiGroups: ["rbac.istio.io"] # istio RBAC watcher
  resources: ["*"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["apiextensions.k8s.io"]
  resources: ["customresourcedefinitions"]
  verbs: ["get", "list", "watch"]
- apiGroups: [""]
  resources: ["configmaps", "endpoints", "pods", "services", "namespaces", "secrets"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["extensions"]
  resources: ["replicasets"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["apps"]
  resources: ["replicasets"]
  verbs: ["get", "list", "watch"]

---
# Source: istio/charts/pilot/templates/clusterrole.yaml
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: istio-pilot-istio-system
  labels:
    app: istio-pilot
    chart: pilot-1.0.1
    heritage: Tiller
    release: RELEASE-NAME
rules:
- apiGroups: ["config.istio.io"]
  resources: ["*"]
  verbs: ["*"]
- apiGroups: ["rbac.istio.io"]
  resources: ["*"]
  verbs: ["get", "watch", "list"]
- apiGroups: ["networking.istio.io"]
  resources: ["*"]
  verbs: ["*"]
- apiGroups: ["authentication.istio.io"]
  resources: ["*"]
  verbs: ["*"]
- apiGroups: ["apiextensions.k8s.io"]
  resources: ["customresourcedefinitions"]
  verbs: ["*"]
- apiGroups: ["extensions"]
  resources: ["thirdpartyresources", "thirdpartyresources.extensions", "ingresses", "ingresses/status"]
  verbs: ["*"]
- apiGroups: [""]
  resources: ["configmaps"]
  verbs: ["create", "get", "list", "watch", "update"]
- apiGroups: [""]
  resources: ["endpoints", "pods", "services"]
  verbs: ["get", "list", "watch"]
- apiGroups: [""]
  resources: ["namespaces", "nodes", "secrets"]
  verbs: ["get", "list", "watch"]

---
# Source: istio/charts/prometheus/templates/clusterrole.yaml
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: prometheus-istio-system
rules:
- apiGroups: [""]
  resources:
  - nodes
  - services
  - endpoints
  - pods
  - nodes/proxy
  verbs: ["get", "list", "watch"]
- apiGroups: [""]
  resources:
  - configmaps
  verbs: ["get"]
- nonResourceURLs: ["/metrics"]
  verbs: ["get"]

---
# Source: istio/charts/security/templates/clusterrole.yaml
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: istio-citadel-istio-system
  labels:
    app: security
    chart: security-1.0.1
    heritage: Tiller
    release: RELEASE-NAME
rules:
- apiGroups: [""]
  resources: ["secrets"]
  verbs: ["create", "get", "watch", "list", "update", "delete"]
- apiGroups: [""]
  resources: ["serviceaccounts"]
  verbs: ["get", "watch", "list"]
- apiGroups: [""]
  resources: ["services"]
  verbs: ["get", "watch", "list"]

---
# Source: istio/charts/sidecarInjectorWebhook/templates/clusterrole.yaml
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: istio-sidecar-injector-istio-system
  labels:
    app: istio-sidecar-injector
    chart: sidecarInjectorWebhook-1.0.1
    heritage: Tiller
    release: RELEASE-NAME
rules:
- apiGroups: ["*"]
  resources: ["configmaps"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["admissionregistration.k8s.io"]
  resources: ["mutatingwebhookconfigurations"]
  verbs: ["get", "list", "watch", "patch"]

---
# Source: istio/charts/galley/templates/clusterrolebinding.yaml
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: istio-galley-admin-role-binding-istio-system
  labels:
    app: istio-galley
    chart: galley-1.0.1
    heritage: Tiller
    release: RELEASE-NAME
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: istio-galley-istio-system
subjects:
  - kind: ServiceAccount
    name: istio-galley-service-account
    namespace: istio-system

---
# Source: istio/charts/gateways/templates/clusterrolebindings.yaml

apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: istio-egressgateway-istio-system
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: istio-egressgateway-istio-system
subjects:
  - kind: ServiceAccount
    name: istio-egressgateway-service-account
    namespace: istio-system
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: istio-ingressgateway-istio-system
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: istio-ingressgateway-istio-system
subjects:
  - kind: ServiceAccount
    name: istio-ingressgateway-service-account
    namespace: istio-system
---

---
# Source: istio/charts/mixer/templates/clusterrolebinding.yaml
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: istio-mixer-admin-role-binding-istio-system
  labels:
    app: mixer
    chart: mixer-1.0.1
    heritage: Tiller
    release: RELEASE-NAME
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: istio-mixer-istio-system
subjects:
  - kind: ServiceAccount
    name: istio-mixer-service-account
    namespace: istio-system

---
# Source: istio/charts/pilot/templates/clusterrolebinding.yaml
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: istio-pilot-istio-system
  labels:
    app: istio-pilot
    chart: pilot-1.0.1
    heritage: Tiller
    release: RELEASE-NAME
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: istio-pilot-istio-system
subjects:
  - kind: ServiceAccount
    name: istio-pilot-service-account
    namespace: istio-system

---
# Source: istio/charts/prometheus/templates/clusterrolebindings.yaml
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: prometheus-istio-system
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: prometheus-istio-system
subjects:
- kind: ServiceAccount
  name: prometheus
  namespace: istio-system

---
# Source: istio/charts/security/templates/clusterrolebinding.yaml
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: istio-citadel-istio-system
  labels:
    app: security
    chart: security-1.0.1
    heritage: Tiller
    release: RELEASE-NAME
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: istio-citadel-istio-system
subjects:
  - kind: ServiceAccount
    name: istio-citadel-service-account
    namespace: istio-system

---
# Source: istio/charts/sidecarInjectorWebhook/templates/clusterrolebinding.yaml
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: istio-sidecar-injector-admin-role-binding-istio-system
  labels:
    app: istio-sidecar-injector
    chart: sidecarInjectorWebhook-1.0.1
    heritage: Tiller
    release: RELEASE-NAME
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: istio-sidecar-injector-istio-system
subjects:
  - kind: ServiceAccount
    name: istio-sidecar-injector-service-account
    namespace: istio-system

---
# Source: istio/charts/galley/templates/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: istio-galley
  namespace: istio-system
  labels:
    istio: galley
spec:
  ports:
  - port: 443
    name: https-validation
  - port: 9093
    name: http-monitoring
  selector:
    istio: galley

---
# Source: istio/charts/gateways/templates/service.yaml

apiVersion: v1
kind: Service
metadata:
  name: istio-egressgateway
  namespace: istio-system
  annotations:
  labels:
    chart: gateways-1.0.1
    release: RELEASE-NAME
    heritage: Tiller
    app: istio-egressgateway
    istio: egressgateway
spec:
  type: ClusterIP
  selector:
    app: istio-egressgateway
    istio: egressgateway
  ports:
    -
      name: http2
      port: 80
    -
      name: https
      port: 443
---
apiVersion: v1
kind: Service
metadata:
  name: istio-ingressgateway
  namespace: istio-system
  annotations:
  labels:
    chart: gateways-1.0.1
    release: RELEASE-NAME
    heritage: Tiller
    app: istio-ingressgateway
    istio: ingressgateway
spec:
  type: LoadBalancer
  selector:
    app: istio-ingressgateway
    istio: ingressgateway
  ports:
    -
      name: http2
      nodePort: 31380
      port: 80
      targetPort: 80
    -
      name: https
      nodePort: 31390
      port: 443
    -
      name: tcp
      nodePort: 31400
      port: 31400
    -
      name: tcp-pilot-grpc-tls
      port: 15011
      targetPort: 15011
    -
      name: tcp-citadel-grpc-tls
      port: 8060
      targetPort: 8060
    -
      name: tcp-dns-tls
      port: 853
      targetPort: 853
    -
      name: http2-prometheus
      port: 15030
      targetPort: 15030
    -
      name: http2-grafana
      port: 15031
      targetPort: 15031
---

---
# Source: istio/charts/grafana/templates/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: grafana
  namespace: istio-system
  annotations:
  labels:
    app: grafana
    chart: grafana-1.0.1
    release: RELEASE-NAME
    heritage: Tiller
spec:
  type: ClusterIP
  ports:
    - port: 3000
      targetPort: 3000
      protocol: TCP
      name: http
  selector:
    app: grafana

---
# Source: istio/charts/mixer/templates/service.yaml

apiVersion: v1
kind: Service
metadata:
  name: istio-policy
  namespace: istio-system
  labels:
    chart: mixer-1.0.1
    release: RELEASE-NAME
    istio: mixer
spec:
  ports:
  - name: grpc-mixer
    port: 9091
  - name: grpc-mixer-mtls
    port: 15004
  - name: http-monitoring
    port: 9093
  selector:
    istio: mixer
    istio-mixer-type: policy
---
apiVersion: v1
kind: Service
metadata:
  name: istio-telemetry
  namespace: istio-system
  labels:
    chart: mixer-1.0.1
    release: RELEASE-NAME
    istio: mixer
spec:
  ports:
  - name: grpc-mixer
    port: 9091
  - name: grpc-mixer-mtls
    port: 15004
  - name: http-monitoring
    port: 9093
  - name: prometheus
    port: 42422
  selector:
    istio: mixer
    istio-mixer-type: telemetry
---

---
# Source: istio/charts/mixer/templates/statsdtoprom.yaml

---
apiVersion: v1
kind: Service
metadata:
  name: istio-statsd-prom-bridge
  namespace: istio-system
  labels:
    chart: mixer-1.0.1
    release: RELEASE-NAME
    istio: statsd-prom-bridge
spec:
  ports:
  - name: statsd-prom
    port: 9102
  - name: statsd-udp
    port: 9125
    protocol: UDP
  selector:
    istio: statsd-prom-bridge

---

apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: istio-statsd-prom-bridge
  namespace: istio-system
  labels:
    chart: mixer-1.0.1
    release: RELEASE-NAME
    istio: mixer
spec:
  template:
    metadata:
      labels:
        istio: statsd-prom-bridge
      annotations:
        sidecar.istio.io/inject: "false"
    spec:
      serviceAccountName: istio-mixer-service-account
      volumes:
      - name: config-volume
        configMap:
          name: istio-statsd-prom-bridge
      containers:
      - name: statsd-prom-bridge
        image: "docker.io/prom/statsd-exporter:v0.6.0"
        imagePullPolicy: IfNotPresent
        ports:
        - containerPort: 9102
        - containerPort: 9125
          protocol: UDP
        args:
        - '-statsd.mapping-config=/etc/statsd/mapping.conf'
        resources:
          requests:
            cpu: 10m

        volumeMounts:
        - name: config-volume
          mountPath: /etc/statsd

---
# Source: istio/charts/pilot/templates/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: istio-pilot
  namespace: istio-system
  labels:
    app: istio-pilot
    chart: pilot-1.0.1
    release: RELEASE-NAME
    heritage: Tiller
spec:
  ports:
  - port: 15010
    name: grpc-xds # direct
  - port: 15011
    name: https-xds # mTLS
  - port: 8080
    name: http-legacy-discovery # direct
  - port: 9093
    name: http-monitoring
  selector:
    istio: pilot

---
# Source: istio/charts/prometheus/templates/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: prometheus
  namespace: istio-system
  annotations:
    prometheus.io/scrape: 'true'
  labels:
    name: prometheus
spec:
  selector:
    app: prometheus
  ports:
  - name: http-prometheus
    protocol: TCP
    port: 9090

---
# Source: istio/charts/security/templates/service.yaml
apiVersion: v1
kind: Service
metadata:
  # we use the normal name here (e.g. 'prometheus')
  # as grafana is configured to use this as a data source
  name: istio-citadel
  namespace: istio-system
  labels:
    app: istio-citadel
spec:
  ports:
    - name: grpc-citadel
      port: 8060
      targetPort: 8060
      protocol: TCP
    - name: http-monitoring
      port: 9093
  selector:
    istio: citadel

---
# Source: istio/charts/servicegraph/templates/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: servicegraph
  namespace: istio-system
  annotations:
  labels:
    app: servicegraph
    chart: servicegraph-1.0.1
    release: RELEASE-NAME
    heritage: Tiller
spec:
  type: ClusterIP
  ports:
    - port: 8088
      targetPort: 8088
      protocol: TCP
      name: http
  selector:
    app: servicegraph

---
# Source: istio/charts/sidecarInjectorWebhook/templates/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: istio-sidecar-injector
  namespace: istio-system
  labels:
    istio: sidecar-injector
spec:
  ports:
  - port: 443
  selector:
    istio: sidecar-injector

---
# Source: istio/charts/galley/templates/deployment.yaml
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: istio-galley
  namespace: istio-system
  labels:
    app: galley
    chart: galley-1.0.1
    release: RELEASE-NAME
    heritage: Tiller
    istio: galley
spec:
  replicas: 1
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  template:
    metadata:
      labels:
        istio: galley
      annotations:
        sidecar.istio.io/inject: "false"
        scheduler.alpha.kubernetes.io/critical-pod: ""
    spec:
      serviceAccountName: istio-galley-service-account
      containers:
        - name: validator
          image: "gcr.io/istio-release/galley:1.0.1"
          imagePullPolicy: IfNotPresent
          ports:
          - containerPort: 443
          - containerPort: 9093
          command:
          - /usr/local/bin/galley
          - validator
          - --deployment-namespace=istio-system
          - --caCertFile=/etc/istio/certs/root-cert.pem
          - --tlsCertFile=/etc/istio/certs/cert-chain.pem
          - --tlsKeyFile=/etc/istio/certs/key.pem
          - --healthCheckInterval=1s
          - --healthCheckFile=/health
          - --webhook-config-file
          - /etc/istio/config/validatingwebhookconfiguration.yaml
          volumeMounts:
          - name: certs
            mountPath: /etc/istio/certs
            readOnly: true
          - name: config
            mountPath: /etc/istio/config
            readOnly: true
          livenessProbe:
            exec:
              command:
                - /usr/local/bin/galley
                - probe
                - --probe-path=/health
                - --interval=10s
            initialDelaySeconds: 5
            periodSeconds: 5
          readinessProbe:
            exec:
              command:
                - /usr/local/bin/galley
                - probe
                - --probe-path=/health
                - --interval=10s
            initialDelaySeconds: 5
            periodSeconds: 5
          resources:
            requests:
              cpu: 10m

      volumes:
      - name: certs
        secret:
          secretName: istio.istio-galley-service-account
      - name: config
        configMap:
          name: istio-galley-configuration
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
                - ppc64le
                - s390x
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - ppc64le
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - s390x

---
# Source: istio/charts/gateways/templates/deployment.yaml

apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: istio-egressgateway
  namespace: istio-system
  labels:
    chart: gateways-1.0.1
    release: RELEASE-NAME
    heritage: Tiller
    app: istio-egressgateway
    istio: egressgateway
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: istio-egressgateway
        istio: egressgateway
      annotations:
        sidecar.istio.io/inject: "false"
        scheduler.alpha.kubernetes.io/critical-pod: ""
    spec:
      serviceAccountName: istio-egressgateway-service-account
      containers:
        - name: istio-proxy
          image: "gcr.io/istio-release/proxyv2:1.0.1"
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 80
            - containerPort: 443
          args:
          - proxy
          - router
          - -v
          - "2"
          - --discoveryRefreshDelay
          - '1s' #discoveryRefreshDelay
          - --drainDuration
          - '45s' #drainDuration
          - --parentShutdownDuration
          - '1m0s' #parentShutdownDuration
          - --connectTimeout
          - '10s' #connectTimeout
          - --serviceCluster
          - istio-egressgateway
          - --zipkinAddress
          - zipkin:9411
          - --statsdUdpAddress
          - istio-statsd-prom-bridge:9125
          - --proxyAdminPort
          - "15000"
          - --controlPlaneAuthPolicy
          - NONE
          - --discoveryAddress
          - istio-pilot:8080
          resources:
            requests:
              cpu: 10m

          env:
          - name: POD_NAME
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: metadata.name
          - name: POD_NAMESPACE
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: metadata.namespace
          - name: INSTANCE_IP
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: status.podIP
          - name: ISTIO_META_POD_NAME
            valueFrom:
              fieldRef:
                fieldPath: metadata.name
          volumeMounts:
          - name: istio-certs
            mountPath: /etc/certs
            readOnly: true
          - name: egressgateway-certs
            mountPath: "/etc/istio/egressgateway-certs"
            readOnly: true
          - name: egressgateway-ca-certs
            mountPath: "/etc/istio/egressgateway-ca-certs"
            readOnly: true
      volumes:
      - name: istio-certs
        secret:
          secretName: istio.istio-egressgateway-service-account
          optional: true
      - name: egressgateway-certs
        secret:
          secretName: "istio-egressgateway-certs"
          optional: true
      - name: egressgateway-ca-certs
        secret:
          secretName: "istio-egressgateway-ca-certs"
          optional: true
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
                - ppc64le
                - s390x
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - ppc64le
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - s390x
---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: istio-ingressgateway
  namespace: istio-system
  labels:
    chart: gateways-1.0.1
    release: RELEASE-NAME
    heritage: Tiller
    app: istio-ingressgateway
    istio: ingressgateway
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: istio-ingressgateway
        istio: ingressgateway
      annotations:
        sidecar.istio.io/inject: "false"
        scheduler.alpha.kubernetes.io/critical-pod: ""
    spec:
      serviceAccountName: istio-ingressgateway-service-account
      containers:
        - name: istio-proxy
          image: "gcr.io/istio-release/proxyv2:1.0.1"
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 80
            - containerPort: 443
            - containerPort: 31400
            - containerPort: 15011
            - containerPort: 8060
            - containerPort: 853
            - containerPort: 15030
            - containerPort: 15031
          args:
          - proxy
          - router
          - -v
          - "2"
          - --discoveryRefreshDelay
          - '1s' #discoveryRefreshDelay
          - --drainDuration
          - '45s' #drainDuration
          - --parentShutdownDuration
          - '1m0s' #parentShutdownDuration
          - --connectTimeout
          - '10s' #connectTimeout
          - --serviceCluster
          - istio-ingressgateway
          - --zipkinAddress
          - zipkin:9411
          - --statsdUdpAddress
          - istio-statsd-prom-bridge:9125
          - --proxyAdminPort
          - "15000"
          - --controlPlaneAuthPolicy
          - NONE
          - --discoveryAddress
          - istio-pilot:8080
          resources:
            requests:
              cpu: 10m

          env:
          - name: POD_NAME
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: metadata.name
          - name: POD_NAMESPACE
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: metadata.namespace
          - name: INSTANCE_IP
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: status.podIP
          - name: ISTIO_META_POD_NAME
            valueFrom:
              fieldRef:
                fieldPath: metadata.name
          volumeMounts:
          - name: istio-certs
            mountPath: /etc/certs
            readOnly: true
          - name: ingressgateway-certs
            mountPath: "/etc/istio/ingressgateway-certs"
            readOnly: true
          - name: ingressgateway-ca-certs
            mountPath: "/etc/istio/ingressgateway-ca-certs"
            readOnly: true
      volumes:
      - name: istio-certs
        secret:
          secretName: istio.istio-ingressgateway-service-account
          optional: true
      - name: ingressgateway-certs
        secret:
          secretName: "istio-ingressgateway-certs"
          optional: true
      - name: ingressgateway-ca-certs
        secret:
          secretName: "istio-ingressgateway-ca-certs"
          optional: true
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
                - ppc64le
                - s390x
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - ppc64le
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - s390x
---

---
# Source: istio/charts/grafana/templates/deployment.yaml
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: grafana
  namespace: istio-system
  labels:
    app: grafana
    chart: grafana-1.0.1
    release: RELEASE-NAME
    heritage: Tiller
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: grafana
      annotations:
        sidecar.istio.io/inject: "false"
        scheduler.alpha.kubernetes.io/critical-pod: ""
    spec:
      containers:
        - name: grafana
          image: "gcr.io/istio-release/grafana:1.0.1"
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 3000
          readinessProbe:
            httpGet:
              path: /login
              port: 3000
          env:
          - name: GRAFANA_PORT
            value: "3000"
          - name: GF_AUTH_BASIC_ENABLED
            value: "false"
          - name: GF_AUTH_ANONYMOUS_ENABLED
            value: "true"
          - name: GF_AUTH_ANONYMOUS_ORG_ROLE
            value: Admin
          - name: GF_PATHS_DATA
            value: /data/grafana
          resources:
            requests:
              cpu: 10m

          volumeMounts:
          - name: data
            mountPath: /data/grafana
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
                - ppc64le
                - s390x
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - ppc64le
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - s390x
      volumes:
      - name: data
        emptyDir: {}

---
# Source: istio/charts/mixer/templates/deployment.yaml

apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: istio-policy
  namespace: istio-system
  labels:
    chart: mixer-1.0.1
    release: RELEASE-NAME
    istio: mixer
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: policy
        istio: mixer
        istio-mixer-type: policy
      annotations:
        sidecar.istio.io/inject: "false"
        scheduler.alpha.kubernetes.io/critical-pod: ""
    spec:
      serviceAccountName: istio-mixer-service-account
      volumes:
      - name: istio-certs
        secret:
          secretName: istio.istio-mixer-service-account
          optional: true
      - name: uds-socket
        emptyDir: {}
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
                - ppc64le
                - s390x
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - ppc64le
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - s390x
      containers:
      - name: mixer
        image: "gcr.io/istio-release/mixer:1.0.1"
        imagePullPolicy: IfNotPresent
        ports:
        - containerPort: 9093
        - containerPort: 42422
        args:
          - --address
          - unix:///sock/mixer.socket
          - --configStoreURL=k8s://
          - --configDefaultNamespace=istio-system
          - --trace_zipkin_url=http://zipkin:9411/api/v1/spans
        resources:
          requests:
            cpu: 10m

        volumeMounts:
        - name: uds-socket
          mountPath: /sock
        livenessProbe:
          httpGet:
            path: /version
            port: 9093
          initialDelaySeconds: 5
          periodSeconds: 5
      - name: istio-proxy
        image: "gcr.io/istio-release/proxyv2:1.0.1"
        imagePullPolicy: IfNotPresent
        ports:
        - containerPort: 9091
        - containerPort: 15004
        args:
        - proxy
        - --serviceCluster
        - istio-policy
        - --templateFile
        - /etc/istio/proxy/envoy_policy.yaml.tmpl
        - --controlPlaneAuthPolicy
        - NONE
        env:
        - name: POD_NAME
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.name
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.namespace
        - name: INSTANCE_IP
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: status.podIP
        resources:
          requests:
            cpu: 10m

        volumeMounts:
        - name: istio-certs
          mountPath: /etc/certs
          readOnly: true
        - name: uds-socket
          mountPath: /sock

---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: istio-telemetry
  namespace: istio-system
  labels:
    chart: mixer-1.0.1
    release: RELEASE-NAME
    istio: mixer
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: telemetry
        istio: mixer
        istio-mixer-type: telemetry
      annotations:
        sidecar.istio.io/inject: "false"
        scheduler.alpha.kubernetes.io/critical-pod: ""
    spec:
      serviceAccountName: istio-mixer-service-account
      volumes:
      - name: istio-certs
        secret:
          secretName: istio.istio-mixer-service-account
          optional: true
      - name: uds-socket
        emptyDir: {}
      containers:
      - name: mixer
        image: "gcr.io/istio-release/mixer:1.0.1"
        imagePullPolicy: IfNotPresent
        ports:
        - containerPort: 9093
        - containerPort: 42422
        args:
          - --address
          - unix:///sock/mixer.socket
          - --configStoreURL=k8s://
          - --configDefaultNamespace=istio-system
          - --trace_zipkin_url=http://zipkin:9411/api/v1/spans
        resources:
          requests:
            cpu: 10m

        volumeMounts:
        - name: uds-socket
          mountPath: /sock
        livenessProbe:
          httpGet:
            path: /version
            port: 9093
          initialDelaySeconds: 5
          periodSeconds: 5
      - name: istio-proxy
        image: "gcr.io/istio-release/proxyv2:1.0.1"
        imagePullPolicy: IfNotPresent
        ports:
        - containerPort: 9091
        - containerPort: 15004
        args:
        - proxy
        - --serviceCluster
        - istio-telemetry
        - --templateFile
        - /etc/istio/proxy/envoy_telemetry.yaml.tmpl
        - --controlPlaneAuthPolicy
        - NONE
        env:
        - name: POD_NAME
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.name
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.namespace
        - name: INSTANCE_IP
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: status.podIP
        resources:
          requests:
            cpu: 10m

        volumeMounts:
        - name: istio-certs
          mountPath: /etc/certs
          readOnly: true
        - name: uds-socket
          mountPath: /sock

---

---
# Source: istio/charts/pilot/templates/deployment.yaml
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: istio-pilot
  namespace: istio-system
  # TODO: default template doesn't have this, which one is right ?
  labels:
    app: istio-pilot
    chart: pilot-1.0.1
    release: RELEASE-NAME
    heritage: Tiller
    istio: pilot
  annotations:
    checksum/config-volume: f8da08b6b8c170dde721efd680270b2901e750d4aa186ebb6c22bef5b78a43f9
spec:
  replicas: 1
  template:
    metadata:
      labels:
        istio: pilot
        app: pilot
      annotations:
        sidecar.istio.io/inject: "false"
        scheduler.alpha.kubernetes.io/critical-pod: ""
    spec:
      serviceAccountName: istio-pilot-service-account
      containers:
        - name: discovery
          image: "gcr.io/istio-release/pilot:1.0.1"
          imagePullPolicy: IfNotPresent
          args:
          - "discovery"
          ports:
          - containerPort: 8080
          - containerPort: 15010
          readinessProbe:
            httpGet:
              path: /ready
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 30
            timeoutSeconds: 5
          env:
          - name: POD_NAME
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: metadata.name
          - name: POD_NAMESPACE
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: metadata.namespace
          - name: PILOT_CACHE_SQUASH
            value: "5"
          - name: GODEBUG
            value: "gctrace=2"
          - name: PILOT_PUSH_THROTTLE_COUNT
            value: "100"
          - name: PILOT_TRACE_SAMPLING
            value: "100"
          resources:
            requests:
              cpu: 500m
              memory: 2048Mi

          volumeMounts:
          - name: config-volume
            mountPath: /etc/istio/config
          - name: istio-certs
            mountPath: /etc/certs
            readOnly: true
        - name: istio-proxy
          image: "gcr.io/istio-release/proxyv2:1.0.1"
          imagePullPolicy: IfNotPresent
          ports:
          - containerPort: 15003
          - containerPort: 15005
          - containerPort: 15007
          - containerPort: 15011
          args:
          - proxy
          - --serviceCluster
          - istio-pilot
          - --templateFile
          - /etc/istio/proxy/envoy_pilot.yaml.tmpl
          - --controlPlaneAuthPolicy
          - NONE
          env:
          - name: POD_NAME
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: metadata.name
          - name: POD_NAMESPACE
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: metadata.namespace
          - name: INSTANCE_IP
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: status.podIP
          resources:
            requests:
              cpu: 10m

          volumeMounts:
          - name: istio-certs
            mountPath: /etc/certs
            readOnly: true
      volumes:
      - name: config-volume
        configMap:
          name: istio
      - name: istio-certs
        secret:
          secretName: istio.istio-pilot-service-account
          optional: true
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
                - ppc64le
                - s390x
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - ppc64le
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - s390x

---
# Source: istio/charts/prometheus/templates/deployment.yaml
# TODO: the original template has service account, roles, etc
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: prometheus
  namespace: istio-system
  labels:
    app: prometheus
    chart: prometheus-1.0.1
    release: RELEASE-NAME
    heritage: Tiller
spec:
  replicas: 1
  selector:
    matchLabels:
      app: prometheus
  template:
    metadata:
      labels:
        app: prometheus
      annotations:
        sidecar.istio.io/inject: "false"
        scheduler.alpha.kubernetes.io/critical-pod: ""
    spec:
      serviceAccountName: prometheus
      containers:
        - name: prometheus
          image: "docker.io/prom/prometheus:v2.3.1"
          imagePullPolicy: IfNotPresent
          args:
            - '--storage.tsdb.retention=6h'
            - '--config.file=/etc/prometheus/prometheus.yml'
          ports:
            - containerPort: 9090
              name: http
          livenessProbe:
            httpGet:
              path: /-/healthy
              port: 9090
          readinessProbe:
            httpGet:
              path: /-/ready
              port: 9090
          resources:
            requests:
              cpu: 10m

          volumeMounts:
          - name: config-volume
            mountPath: /etc/prometheus
      volumes:
      - name: config-volume
        configMap:
          name: prometheus
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
                - ppc64le
                - s390x
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - ppc64le
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - s390x

---
# Source: istio/charts/security/templates/deployment.yaml
# istio CA watching all namespaces
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: istio-citadel
  namespace: istio-system
  labels:
    app: security
    chart: security-1.0.1
    release: RELEASE-NAME
    heritage: Tiller
    istio: citadel
spec:
  replicas: 1
  template:
    metadata:
      labels:
        istio: citadel
      annotations:
        sidecar.istio.io/inject: "false"
        scheduler.alpha.kubernetes.io/critical-pod: ""
    spec:
      serviceAccountName: istio-citadel-service-account
      containers:
        - name: citadel
          image: "gcr.io/istio-release/citadel:1.0.1"
          imagePullPolicy: IfNotPresent
          args:
            - --append-dns-names=true
            - --grpc-port=8060
            - --grpc-hostname=citadel
            - --citadel-storage-namespace=istio-system
            - --custom-dns-names=istio-pilot-service-account.istio-system:istio-pilot.istio-system,istio-ingressgateway-service-account.istio-system:istio-ingress.istio-system
            - --self-signed-ca=true
          resources:
            requests:
              cpu: 10m

      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
                - ppc64le
                - s390x
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - ppc64le
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - s390x

---
# Source: istio/charts/servicegraph/templates/deployment.yaml
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: servicegraph
  namespace: istio-system
  labels:
    app: servicegraph
    chart: servicegraph-1.0.1
    release: RELEASE-NAME
    heritage: Tiller
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: servicegraph
      annotations:
        sidecar.istio.io/inject: "false"
        scheduler.alpha.kubernetes.io/critical-pod: ""
    spec:
      containers:
        - name: servicegraph
          image: "gcr.io/istio-release/servicegraph:1.0.1"
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 8088
          args:
          - --prometheusAddr=http://prometheus:9090
          livenessProbe:
            httpGet:
              path: /graph
              port: 8088
          readinessProbe:
            httpGet:
              path: /graph
              port: 8088
          resources:
            requests:
              cpu: 10m

      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
                - ppc64le
                - s390x
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - ppc64le
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - s390x

---
# Source: istio/charts/sidecarInjectorWebhook/templates/deployment.yaml
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: istio-sidecar-injector
  namespace: istio-system
  labels:
    app: sidecarInjectorWebhook
    chart: sidecarInjectorWebhook-1.0.1
    release: RELEASE-NAME
    heritage: Tiller
    istio: sidecar-injector
spec:
  replicas: 1
  template:
    metadata:
      labels:
        istio: sidecar-injector
      annotations:
        sidecar.istio.io/inject: "false"
        scheduler.alpha.kubernetes.io/critical-pod: ""
    spec:
      serviceAccountName: istio-sidecar-injector-service-account
      containers:
        - name: sidecar-injector-webhook
          image: "gcr.io/istio-release/sidecar_injector:1.0.1"
          imagePullPolicy: IfNotPresent
          args:
            - --caCertFile=/etc/istio/certs/root-cert.pem
            - --tlsCertFile=/etc/istio/certs/cert-chain.pem
            - --tlsKeyFile=/etc/istio/certs/key.pem
            - --injectConfig=/etc/istio/inject/config
            - --meshConfig=/etc/istio/config/mesh
            - --healthCheckInterval=2s
            - --healthCheckFile=/health
          volumeMounts:
          - name: config-volume
            mountPath: /etc/istio/config
            readOnly: true
          - name: certs
            mountPath: /etc/istio/certs
            readOnly: true
          - name: inject-config
            mountPath: /etc/istio/inject
            readOnly: true
          livenessProbe:
            exec:
              command:
                - /usr/local/bin/sidecar-injector
                - probe
                - --probe-path=/health
                - --interval=4s
            initialDelaySeconds: 4
            periodSeconds: 4
          readinessProbe:
            exec:
              command:
                - /usr/local/bin/sidecar-injector
                - probe
                - --probe-path=/health
                - --interval=4s
            initialDelaySeconds: 4
            periodSeconds: 4
          resources:
            requests:
              cpu: 10m

      volumes:
      - name: config-volume
        configMap:
          name: istio
      - name: certs
        secret:
          secretName: istio.istio-sidecar-injector-service-account
      - name: inject-config
        configMap:
          name: istio-sidecar-injector
          items:
          - key: config
            path: config
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
                - ppc64le
                - s390x
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - ppc64le
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - s390x

---
# Source: istio/charts/tracing/templates/deployment.yaml
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: istio-tracing
  namespace: istio-system
  labels:
    app: istio-tracing
    chart: tracing-1.0.1
    release: RELEASE-NAME
    heritage: Tiller
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: jaeger
      annotations:
        sidecar.istio.io/inject: "false"
        scheduler.alpha.kubernetes.io/critical-pod: ""
    spec:
      containers:
        - name: jaeger
          image: "docker.io/jaegertracing/all-in-one:1.5"
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 9411
            - containerPort: 16686
            - containerPort: 5775
              protocol: UDP
            - containerPort: 6831
              protocol: UDP
            - containerPort: 6832
              protocol: UDP
          env:
          - name: POD_NAMESPACE
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: metadata.namespace
          - name: COLLECTOR_ZIPKIN_HTTP_PORT
            value: "9411"
          - name: MEMORY_MAX_TRACES
            value: "50000"
          livenessProbe:
            httpGet:
              path: /
              port: 16686
          readinessProbe:
            httpGet:
              path: /
              port: 16686
          resources:
            requests:
              cpu: 10m

      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
                - ppc64le
                - s390x
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - ppc64le
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - s390x

---
# Source: istio/charts/pilot/templates/gateway.yaml
apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: istio-autogenerated-k8s-ingress
  namespace: istio-system
spec:
  selector:
    istio: ingress
  servers:
  - port:
      number: 80
      protocol: HTTP2
      name: http
    hosts:
    - "*"

---

---
# Source: istio/charts/gateways/templates/autoscale.yaml

apiVersion: autoscaling/v2beta1
kind: HorizontalPodAutoscaler
metadata:
    name: istio-egressgateway
    namespace: istio-system
spec:
    maxReplicas: 5
    minReplicas: 1
    scaleTargetRef:
      apiVersion: apps/v1beta1
      kind: Deployment
      name: istio-egressgateway
    metrics:
    - type: Resource
      resource:
        name: cpu
        targetAverageUtilization: 80
---
apiVersion: autoscaling/v2beta1
kind: HorizontalPodAutoscaler
metadata:
    name: istio-ingressgateway
    namespace: istio-system
spec:
    maxReplicas: 5
    minReplicas: 1
    scaleTargetRef:
      apiVersion: apps/v1beta1
      kind: Deployment
      name: istio-ingressgateway
    metrics:
    - type: Resource
      resource:
        name: cpu
        targetAverageUtilization: 80
---

---
# Source: istio/charts/mixer/templates/autoscale.yaml

apiVersion: autoscaling/v2beta1
kind: HorizontalPodAutoscaler
metadata:
    name: istio-policy
    namespace: istio-system
spec:
    maxReplicas: 5
    minReplicas: 1
    scaleTargetRef:
      apiVersion: apps/v1beta1
      kind: Deployment
      name: istio-policy
    metrics:
    - type: Resource
      resource:
        name: cpu
        targetAverageUtilization: 80
---
apiVersion: autoscaling/v2beta1
kind: HorizontalPodAutoscaler
metadata:
    name: istio-telemetry
    namespace: istio-system
spec:
    maxReplicas: 5
    minReplicas: 1
    scaleTargetRef:
      apiVersion: apps/v1beta1
      kind: Deployment
      name: istio-telemetry
    metrics:
    - type: Resource
      resource:
        name: cpu
        targetAverageUtilization: 80
---

---
# Source: istio/charts/pilot/templates/autoscale.yaml

apiVersion: autoscaling/v2beta1
kind: HorizontalPodAutoscaler
metadata:
    name: istio-pilot
spec:
    maxReplicas: 5
    minReplicas: 1
    scaleTargetRef:
      apiVersion: apps/v1beta1
      kind: Deployment
      name: istio-pilot
    metrics:
    - type: Resource
      resource:
        name: cpu
        targetAverageUtilization: 80
---

---
# Source: istio/charts/tracing/templates/service-jaeger.yaml


apiVersion: v1
kind: List
items:
- apiVersion: v1
  kind: Service
  metadata:
    name: jaeger-query
    namespace: istio-system
    annotations:
    labels:
      app: jaeger
      jaeger-infra: jaeger-service
      chart: tracing-1.0.1
      release: RELEASE-NAME
      heritage: Tiller
  spec:
    ports:
      - name: query-http
        port: 16686
        protocol: TCP
        targetPort: 16686
    selector:
      app: jaeger
- apiVersion: v1
  kind: Service
  metadata:
    name: jaeger-collector
    namespace: istio-system
    labels:
      app: jaeger
      jaeger-infra: collector-service
      chart: tracing-1.0.1
      release: RELEASE-NAME
      heritage: Tiller
  spec:
    ports:
    - name: jaeger-collector-tchannel
      port: 14267
      protocol: TCP
      targetPort: 14267
    - name: jaeger-collector-http
      port: 14268
      targetPort: 14268
      protocol: TCP
    selector:
      app: jaeger
    type: ClusterIP
- apiVersion: v1
  kind: Service
  metadata:
    name: jaeger-agent
    namespace: istio-system
    labels:
      app: jaeger
      jaeger-infra: agent-service
      chart: tracing-1.0.1
      release: RELEASE-NAME
      heritage: Tiller
  spec:
    ports:
    - name: agent-zipkin-thrift
      port: 5775
      protocol: UDP
      targetPort: 5775
    - name: agent-compact
      port: 6831
      protocol: UDP
      targetPort: 6831
    - name: agent-binary
      port: 6832
      protocol: UDP
      targetPort: 6832
    clusterIP: None
    selector:
      app: jaeger



---
# Source: istio/charts/tracing/templates/service.yaml
apiVersion: v1
kind: List
items:
- apiVersion: v1
  kind: Service
  metadata:
    name: zipkin
    namespace: istio-system
    labels:
      app: jaeger
      chart: tracing-1.0.1
      release: RELEASE-NAME
      heritage: Tiller
  spec:
    type: ClusterIP
    ports:
      - port: 9411
        targetPort: 9411
        protocol: TCP
        name: http
    selector:
      app: jaeger
- apiVersion: v1
  kind: Service
  metadata:
    name: tracing
    namespace: istio-system
    annotations:
    labels:
      app: jaeger
      chart: tracing-1.0.1
      release: RELEASE-NAME
      heritage: Tiller
  spec:
    ports:
      - name: http-query
        port: 80
        protocol: TCP
        targetPort: 16686
    selector:
      app: jaeger

---
# Source: istio/charts/sidecarInjectorWebhook/templates/mutatingwebhook.yaml
apiVersion: admissionregistration.k8s.io/v1beta1
kind: MutatingWebhookConfiguration
metadata:
  name: istio-sidecar-injector
  namespace: istio-system
  labels:
    app: istio-sidecar-injector
    chart: sidecarInjectorWebhook-1.0.1
    release: RELEASE-NAME
    heritage: Tiller
webhooks:
  - name: sidecar-injector.istio.io
    clientConfig:
      service:
        name: istio-sidecar-injector
        namespace: istio-system
        path: "/inject"
      caBundle: ""
    rules:
      - operations: [ "CREATE" ]
        apiGroups: [""]
        apiVersions: ["v1"]
        resources: ["pods"]
    failurePolicy: Fail
    namespaceSelector:
      matchLabels:
        istio-injection: enabled


---
# Source: istio/charts/galley/templates/validatingwehookconfiguration.yaml.tpl


---
# Source: istio/charts/grafana/templates/grafana-ports-mtls.yaml


---
# Source: istio/charts/grafana/templates/secret.yaml

---
# Source: istio/charts/pilot/templates/meshexpansion.yaml


---
# Source: istio/charts/security/templates/create-custom-resources-job.yaml


---
# Source: istio/charts/security/templates/enable-mesh-mtls.yaml


---
# Source: istio/charts/security/templates/meshexpansion.yaml


---

---
# Source: istio/charts/servicegraph/templates/ingress.yaml

---
# Source: istio/charts/telemetry-gateway/templates/gateway.yaml


---
# Source: istio/charts/tracing/templates/ingress-jaeger.yaml

---
# Source: istio/charts/tracing/templates/ingress.yaml

---
# Source: istio/templates/install-custom-resources.sh.tpl


---
# Source: istio/charts/mixer/templates/config.yaml
apiVersion: "config.istio.io/v1alpha2"
kind: attributemanifest
metadata:
  name: istioproxy
  namespace: istio-system
spec:
  attributes:
    origin.ip:
      valueType: IP_ADDRESS
    origin.uid:
      valueType: STRING
    origin.user:
      valueType: STRING
    request.headers:
      valueType: STRING_MAP
    request.id:
      valueType: STRING
    request.host:
      valueType: STRING
    request.method:
      valueType: STRING
    request.path:
      valueType: STRING
    request.reason:
      valueType: STRING
    request.referer:
      valueType: STRING
    request.scheme:
      valueType: STRING
    request.total_size:
          valueType: INT64
    request.size:
      valueType: INT64
    request.time:
      valueType: TIMESTAMP
    request.useragent:
      valueType: STRING
    response.code:
      valueType: INT64
    response.duration:
      valueType: DURATION
    response.headers:
      valueType: STRING_MAP
    response.total_size:
          valueType: INT64
    response.size:
      valueType: INT64
    response.time:
      valueType: TIMESTAMP
    source.uid:
      valueType: STRING
    source.user: # DEPRECATED
      valueType: STRING
    source.principal:
      valueType: STRING
    destination.uid:
      valueType: STRING
    destination.principal:
      valueType: STRING
    destination.port:
      valueType: INT64
    connection.event:
      valueType: STRING
    connection.id:
      valueType: STRING
    connection.received.bytes:
      valueType: INT64
    connection.received.bytes_total:
      valueType: INT64
    connection.sent.bytes:
      valueType: INT64
    connection.sent.bytes_total:
      valueType: INT64
    connection.duration:
      valueType: DURATION
    connection.mtls:
      valueType: BOOL
    connection.requested_server_name:
      valueType: STRING
    context.protocol:
      valueType: STRING
    context.timestamp:
      valueType: TIMESTAMP
    context.time:
      valueType: TIMESTAMP
    # Deprecated, kept for compatibility
    context.reporter.local:
      valueType: BOOL
    context.reporter.kind:
      valueType: STRING
    context.reporter.uid:
      valueType: STRING
    api.service:
      valueType: STRING
    api.version:
      valueType: STRING
    api.operation:
      valueType: STRING
    api.protocol:
      valueType: STRING
    request.auth.principal:
      valueType: STRING
    request.auth.audiences:
      valueType: STRING
    request.auth.presenter:
      valueType: STRING
    request.auth.claims:
      valueType: STRING_MAP
    request.auth.raw_claims:
      valueType: STRING
    request.api_key:
      valueType: STRING

---
apiVersion: "config.istio.io/v1alpha2"
kind: attributemanifest
metadata:
  name: kubernetes
  namespace: istio-system
spec:
  attributes:
    source.ip:
      valueType: IP_ADDRESS
    source.labels:
      valueType: STRING_MAP
    source.metadata:
      valueType: STRING_MAP
    source.name:
      valueType: STRING
    source.namespace:
      valueType: STRING
    source.owner:
      valueType: STRING
    source.service:  # DEPRECATED
      valueType: STRING
    source.serviceAccount:
      valueType: STRING
    source.services:
      valueType: STRING
    source.workload.uid:
      valueType: STRING
    source.workload.name:
      valueType: STRING
    source.workload.namespace:
      valueType: STRING
    destination.ip:
      valueType: IP_ADDRESS
    destination.labels:
      valueType: STRING_MAP
    destination.metadata:
      valueType: STRING_MAP
    destination.owner:
      valueType: STRING
    destination.name:
      valueType: STRING
    destination.container.name:
      valueType: STRING
    destination.namespace:
      valueType: STRING
    destination.service: # DEPRECATED
      valueType: STRING
    destination.service.uid:
      valueType: STRING
    destination.service.name:
      valueType: STRING
    destination.service.namespace:
      valueType: STRING
    destination.service.host:
      valueType: STRING
    destination.serviceAccount:
      valueType: STRING
    destination.workload.uid:
      valueType: STRING
    destination.workload.name:
      valueType: STRING
    destination.workload.namespace:
      valueType: STRING
---
apiVersion: "config.istio.io/v1alpha2"
kind: stdio
metadata:
  name: handler
  namespace: istio-system
spec:
  outputAsJson: true
---
apiVersion: "config.istio.io/v1alpha2"
kind: logentry
metadata:
  name: accesslog
  namespace: istio-system
spec:
  severity: '"Info"'
  timestamp: request.time
  variables:
    sourceIp: source.ip | ip("0.0.0.0")
    sourceApp: source.labels["app"] | ""
    sourcePrincipal: source.principal | ""
    sourceName: source.name | ""
    sourceWorkload: source.workload.name | ""
    sourceNamespace: source.namespace | ""
    sourceOwner: source.owner | ""
    destinationApp: destination.labels["app"] | ""
    destinationIp: destination.ip | ip("0.0.0.0")
    destinationServiceHost: destination.service.host | ""
    destinationWorkload: destination.workload.name | ""
    destinationName: destination.name | ""
    destinationNamespace: destination.namespace | ""
    destinationOwner: destination.owner | ""
    destinationPrincipal: destination.principal | ""
    apiClaims: request.auth.raw_claims | ""
    apiKey: request.api_key | request.headers["x-api-key"] | ""
    protocol: request.scheme | context.protocol | "http"
    method: request.method | ""
    url: request.path | ""
    responseCode: response.code | 0
    responseSize: response.size | 0
    requestSize: request.size | 0
    requestId: request.headers["x-request-id"] | ""
    clientTraceId: request.headers["x-client-trace-id"] | ""
    latency: response.duration | "0ms"
    connection_security_policy: conditional((context.reporter.kind | "inbound") == "outbound", "unknown", conditional(connection.mtls | false, "mutual_tls", "none"))
    requestedServerName: connection.requested_server_name | ""
    userAgent: request.useragent | ""
    responseTimestamp: response.time
    receivedBytes: request.total_size | 0
    sentBytes: response.total_size | 0
    referer: request.referer | ""
    httpAuthority: request.headers[":authority"] | request.host | ""
    xForwardedFor: request.headers["x-forwarded-for"] | "0.0.0.0"
    reporter: conditional((context.reporter.kind | "inbound") == "outbound", "source", "destination")
  monitored_resource_type: '"global"'
---
apiVersion: "config.istio.io/v1alpha2"
kind: logentry
metadata:
  name: tcpaccesslog
  namespace: istio-system
spec:
  severity: '"Info"'
  timestamp: context.time | timestamp("2017-01-01T00:00:00Z")
  variables:
    connectionEvent: connection.event | ""
    sourceIp: source.ip | ip("0.0.0.0")
    sourceApp: source.labels["app"] | ""
    sourcePrincipal: source.principal | ""
    sourceName: source.name | ""
    sourceWorkload: source.workload.name | ""
    sourceNamespace: source.namespace | ""
    sourceOwner: source.owner | ""
    destinationApp: destination.labels["app"] | ""
    destinationIp: destination.ip | ip("0.0.0.0")
    destinationServiceHost: destination.service.host | ""
    destinationWorkload: destination.workload.name | ""
    destinationName: destination.name | ""
    destinationNamespace: destination.namespace | ""
    destinationOwner: destination.owner | ""
    destinationPrincipal: destination.principal | ""
    protocol: context.protocol | "tcp"
    connectionDuration: connection.duration | "0ms"
    connection_security_policy: conditional((context.reporter.kind | "inbound") == "outbound", "unknown", conditional(connection.mtls | false, "mutual_tls", "none"))
    requestedServerName: connection.requested_server_name | ""
    receivedBytes: connection.received.bytes | 0
    sentBytes: connection.sent.bytes | 0
    totalReceivedBytes: connection.received.bytes_total | 0
    totalSentBytes: connection.sent.bytes_total | 0
    reporter: conditional((context.reporter.kind | "inbound") == "outbound", "source", "destination")
  monitored_resource_type: '"global"'
---
apiVersion: "config.istio.io/v1alpha2"
kind: rule
metadata:
  name: stdio
  namespace: istio-system
spec:
  match: context.protocol == "http" || context.protocol == "grpc"
  actions:
  - handler: handler.stdio
    instances:
    - accesslog.logentry
---
apiVersion: "config.istio.io/v1alpha2"
kind: rule
metadata:
  name: stdiotcp
  namespace: istio-system
spec:
  match: context.protocol == "tcp"
  actions:
  - handler: handler.stdio
    instances:
    - tcpaccesslog.logentry
---
apiVersion: "config.istio.io/v1alpha2"
kind: metric
metadata:
  name: requestcount
  namespace: istio-system
spec:
  value: "1"
  dimensions:
    reporter: conditional((context.reporter.kind | "inbound") == "outbound", "source", "destination")
    source_workload: source.workload.name | "unknown"
    source_workload_namespace: source.workload.namespace | "unknown"
    source_principal: source.principal | "unknown"
    source_app: source.labels["app"] | "unknown"
    source_version: source.labels["version"] | "unknown"
    destination_workload: destination.workload.name | "unknown"
    destination_workload_namespace: destination.workload.namespace | "unknown"
    destination_principal: destination.principal | "unknown"
    destination_app: destination.labels["app"] | "unknown"
    destination_version: destination.labels["version"] | "unknown"
    destination_service: destination.service.host | "unknown"
    destination_service_name: destination.service.name | "unknown"
    destination_service_namespace: destination.service.namespace | "unknown"
    request_protocol: api.protocol | context.protocol | "unknown"
    response_code: response.code | 200
    connection_security_policy: conditional((context.reporter.kind | "inbound") == "outbound", "unknown", conditional(connection.mtls | false, "mutual_tls", "none"))
  monitored_resource_type: '"UNSPECIFIED"'
---
apiVersion: "config.istio.io/v1alpha2"
kind: metric
metadata:
  name: requestduration
  namespace: istio-system
spec:
  value: response.duration | "0ms"
  dimensions:
    reporter: conditional((context.reporter.kind | "inbound") == "outbound", "source", "destination")
    source_workload: source.workload.name | "unknown"
    source_workload_namespace: source.workload.namespace | "unknown"
    source_principal: source.principal | "unknown"
    source_app: source.labels["app"] | "unknown"
    source_version: source.labels["version"] | "unknown"
    destination_workload: destination.workload.name | "unknown"
    destination_workload_namespace: destination.workload.namespace | "unknown"
    destination_principal: destination.principal | "unknown"
    destination_app: destination.labels["app"] | "unknown"
    destination_version: destination.labels["version"] | "unknown"
    destination_service: destination.service.host | "unknown"
    destination_service_name: destination.service.name | "unknown"
    destination_service_namespace: destination.service.namespace | "unknown"
    request_protocol: api.protocol | context.protocol | "unknown"
    response_code: response.code | 200
    connection_security_policy: conditional((context.reporter.kind | "inbound") == "outbound", "unknown", conditional(connection.mtls | false, "mutual_tls", "none"))
  monitored_resource_type: '"UNSPECIFIED"'
---
apiVersion: "config.istio.io/v1alpha2"
kind: metric
metadata:
  name: requestsize
  namespace: istio-system
spec:
  value: request.size | 0
  dimensions:
    reporter: conditional((context.reporter.kind | "inbound") == "outbound", "source", "destination")
    source_workload: source.workload.name | "unknown"
    source_workload_namespace: source.workload.namespace | "unknown"
    source_principal: source.principal | "unknown"
    source_app: source.labels["app"] | "unknown"
    source_version: source.labels["version"] | "unknown"
    destination_workload: destination.workload.name | "unknown"
    destination_workload_namespace: destination.workload.namespace | "unknown"
    destination_principal: destination.principal | "unknown"
    destination_app: destination.labels["app"] | "unknown"
    destination_version: destination.labels["version"] | "unknown"
    destination_service: destination.service.host | "unknown"
    destination_service_name: destination.service.name | "unknown"
    destination_service_namespace: destination.service.namespace | "unknown"
    request_protocol: api.protocol | context.protocol | "unknown"
    response_code: response.code | 200
    connection_security_policy: conditional((context.reporter.kind | "inbound") == "outbound", "unknown", conditional(connection.mtls | false, "mutual_tls", "none"))
  monitored_resource_type: '"UNSPECIFIED"'
---
apiVersion: "config.istio.io/v1alpha2"
kind: metric
metadata:
  name: responsesize
  namespace: istio-system
spec:
  value: response.size | 0
  dimensions:
    reporter: conditional((context.reporter.kind | "inbound") == "outbound", "source", "destination")
    source_workload: source.workload.name | "unknown"
    source_workload_namespace: source.workload.namespace | "unknown"
    source_principal: source.principal | "unknown"
    source_app: source.labels["app"] | "unknown"
    source_version: source.labels["version"] | "unknown"
    destination_workload: destination.workload.name | "unknown"
    destination_workload_namespace: destination.workload.namespace | "unknown"
    destination_principal: destination.principal | "unknown"
    destination_app: destination.labels["app"] | "unknown"
    destination_version: destination.labels["version"] | "unknown"
    destination_service: destination.service.host | "unknown"
    destination_service_name: destination.service.name | "unknown"
    destination_service_namespace: destination.service.namespace | "unknown"
    request_protocol: api.protocol | context.protocol | "unknown"
    response_code: response.code | 200
    connection_security_policy: conditional((context.reporter.kind | "inbound") == "outbound", "unknown", conditional(connection.mtls | false, "mutual_tls", "none"))
  monitored_resource_type: '"UNSPECIFIED"'
---
apiVersion: "config.istio.io/v1alpha2"
kind: metric
metadata:
  name: tcpbytesent
  namespace: istio-system
spec:
  value: connection.sent.bytes | 0
  dimensions:
    reporter: conditional((context.reporter.kind | "inbound") == "outbound", "source", "destination")
    source_workload: source.workload.name | "unknown"
    source_workload_namespace: source.workload.namespace | "unknown"
    source_principal: source.principal | "unknown"
    source_app: source.labels["app"] | "unknown"
    source_version: source.labels["version"] | "unknown"
    destination_workload: destination.workload.name | "unknown"
    destination_workload_namespace: destination.workload.namespace | "unknown"
    destination_principal: destination.principal | "unknown"
    destination_app: destination.labels["app"] | "unknown"
    destination_version: destination.labels["version"] | "unknown"
    destination_service: destination.service.name | "unknown"
    destination_service_name: destination.service.name | "unknown"
    destination_service_namespace: destination.service.namespace | "unknown"
    connection_security_policy: conditional((context.reporter.kind | "inbound") == "outbound", "unknown", conditional(connection.mtls | false, "mutual_tls", "none"))
  monitored_resource_type: '"UNSPECIFIED"'
---
apiVersion: "config.istio.io/v1alpha2"
kind: metric
metadata:
  name: tcpbytereceived
  namespace: istio-system
spec:
  value: connection.received.bytes | 0
  dimensions:
    reporter: conditional((context.reporter.kind | "inbound") == "outbound", "source", "destination")
    source_workload: source.workload.name | "unknown"
    source_workload_namespace: source.workload.namespace | "unknown"
    source_principal: source.principal | "unknown"
    source_app: source.labels["app"] | "unknown"
    source_version: source.labels["version"] | "unknown"
    destination_workload: destination.workload.name | "unknown"
    destination_workload_namespace: destination.workload.namespace | "unknown"
    destination_principal: destination.principal | "unknown"
    destination_app: destination.labels["app"] | "unknown"
    destination_version: destination.labels["version"] | "unknown"
    destination_service: destination.service.name | "unknown"
    destination_service_name: destination.service.name | "unknown"
    destination_service_namespace: destination.service.namespace | "unknown"
    connection_security_policy: conditional((context.reporter.kind | "inbound") == "outbound", "unknown", conditional(connection.mtls | false, "mutual_tls", "none"))
  monitored_resource_type: '"UNSPECIFIED"'
---
apiVersion: "config.istio.io/v1alpha2"
kind: prometheus
metadata:
  name: handler
  namespace: istio-system
spec:
  metrics:
  - name: requests_total
    instance_name: requestcount.metric.istio-system
    kind: COUNTER
    label_names:
    - reporter
    - source_app
    - source_principal
    - source_workload
    - source_workload_namespace
    - source_version
    - destination_app
    - destination_principal
    - destination_workload
    - destination_workload_namespace
    - destination_version
    - destination_service
    - destination_service_name
    - destination_service_namespace
    - request_protocol
    - response_code
    - connection_security_policy
  - name: request_duration_seconds
    instance_name: requestduration.metric.istio-system
    kind: DISTRIBUTION
    label_names:
    - reporter
    - source_app
    - source_principal
    - source_workload
    - source_workload_namespace
    - source_version
    - destination_app
    - destination_principal
    - destination_workload
    - destination_workload_namespace
    - destination_version
    - destination_service
    - destination_service_name
    - destination_service_namespace
    - request_protocol
    - response_code
    - connection_security_policy
    buckets:
      explicit_buckets:
        bounds: [0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10]
  - name: request_bytes
    instance_name: requestsize.metric.istio-system
    kind: DISTRIBUTION
    label_names:
    - reporter
    - source_app
    - source_principal
    - source_workload
    - source_workload_namespace
    - source_version
    - destination_app
    - destination_principal
    - destination_workload
    - destination_workload_namespace
    - destination_version
    - destination_service
    - destination_service_name
    - destination_service_namespace
    - request_protocol
    - response_code
    - connection_security_policy
    buckets:
      exponentialBuckets:
        numFiniteBuckets: 8
        scale: 1
        growthFactor: 10
  - name: response_bytes
    instance_name: responsesize.metric.istio-system
    kind: DISTRIBUTION
    label_names:
    - reporter
    - source_app
    - source_principal
    - source_workload
    - source_workload_namespace
    - source_version
    - destination_app
    - destination_principal
    - destination_workload
    - destination_workload_namespace
    - destination_version
    - destination_service
    - destination_service_name
    - destination_service_namespace
    - request_protocol
    - response_code
    - connection_security_policy
    buckets:
      exponentialBuckets:
        numFiniteBuckets: 8
        scale: 1
        growthFactor: 10
  - name: tcp_sent_bytes_total
    instance_name: tcpbytesent.metric.istio-system
    kind: COUNTER
    label_names:
    - reporter
    - source_app
    - source_principal
    - source_workload
    - source_workload_namespace
    - source_version
    - destination_app
    - destination_principal
    - destination_workload
    - destination_workload_namespace
    - destination_version
    - destination_service
    - destination_service_name
    - destination_service_namespace
    - connection_security_policy
  - name: tcp_received_bytes_total
    instance_name: tcpbytereceived.metric.istio-system
    kind: COUNTER
    label_names:
    - reporter
    - source_app
    - source_principal
    - source_workload
    - source_workload_namespace
    - source_version
    - destination_app
    - destination_principal
    - destination_workload
    - destination_workload_namespace
    - destination_version
    - destination_service
    - destination_service_name
    - destination_service_namespace
    - connection_security_policy
---
apiVersion: "config.istio.io/v1alpha2"
kind: rule
metadata:
  name: promhttp
  namespace: istio-system
spec:
  match: context.protocol == "http" || context.protocol == "grpc"
  actions:
  - handler: handler.prometheus
    instances:
    - requestcount.metric
    - requestduration.metric
    - requestsize.metric
    - responsesize.metric
---
apiVersion: "config.istio.io/v1alpha2"
kind: rule
metadata:
  name: promtcp
  namespace: istio-system
spec:
  match: context.protocol == "tcp"
  actions:
  - handler: handler.prometheus
    instances:
    - tcpbytesent.metric
    - tcpbytereceived.metric
---

apiVersion: "config.istio.io/v1alpha2"
kind: kubernetesenv
metadata:
  name: handler
  namespace: istio-system
spec:
  # when running from mixer root, use the following config after adding a
  # symbolic link to a kubernetes config file via:
  #
  # $ ln -s ~/.kube/config mixer/adapter/kubernetes/kubeconfig
  #
  # kubeconfig_path: "mixer/adapter/kubernetes/kubeconfig"

---
apiVersion: "config.istio.io/v1alpha2"
kind: rule
metadata:
  name: kubeattrgenrulerule
  namespace: istio-system
spec:
  actions:
  - handler: handler.kubernetesenv
    instances:
    - attributes.kubernetes
---
apiVersion: "config.istio.io/v1alpha2"
kind: rule
metadata:
  name: tcpkubeattrgenrulerule
  namespace: istio-system
spec:
  match: context.protocol == "tcp"
  actions:
  - handler: handler.kubernetesenv
    instances:
    - attributes.kubernetes
---
apiVersion: "config.istio.io/v1alpha2"
kind: kubernetes
metadata:
  name: attributes
  namespace: istio-system
spec:
  # Pass the required attribute data to the adapter
  source_uid: source.uid | ""
  source_ip: source.ip | ip("0.0.0.0") # default to unspecified ip addr
  destination_uid: destination.uid | ""
  destination_port: destination.port | 0
  attribute_bindings:
    # Fill the new attributes from the adapter produced output.
    # $out refers to an instance of OutputTemplate message
    source.ip: $out.source_pod_ip | ip("0.0.0.0")
    source.uid: $out.source_pod_uid | "unknown"
    source.labels: $out.source_labels | emptyStringMap()
    source.name: $out.source_pod_name | "unknown"
    source.namespace: $out.source_namespace | "default"
    source.owner: $out.source_owner | "unknown"
    source.serviceAccount: $out.source_service_account_name | "unknown"
    source.workload.uid: $out.source_workload_uid | "unknown"
    source.workload.name: $out.source_workload_name | "unknown"
    source.workload.namespace: $out.source_workload_namespace | "unknown"
    destination.ip: $out.destination_pod_ip | ip("0.0.0.0")
    destination.uid: $out.destination_pod_uid | "unknown"
    destination.labels: $out.destination_labels | emptyStringMap()
    destination.name: $out.destination_pod_name | "unknown"
    destination.container.name: $out.destination_container_name | "unknown"
    destination.namespace: $out.destination_namespace | "default"
    destination.owner: $out.destination_owner | "unknown"
    destination.serviceAccount: $out.destination_service_account_name | "unknown"
    destination.workload.uid: $out.destination_workload_uid | "unknown"
    destination.workload.name: $out.destination_workload_name | "unknown"
    destination.workload.namespace: $out.destination_workload_namespace | "unknown"

---
# Configuration needed by Mixer.
# Mixer cluster is delivered via CDS
# Specify mixer cluster settings
apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: istio-policy
  namespace: istio-system
spec:
  host: istio-policy.istio-system.svc.cluster.local
  trafficPolicy:
    connectionPool:
      http:
        http2MaxRequests: 10000
        maxRequestsPerConnection: 10000
---
apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: istio-telemetry
  namespace: istio-system
spec:
  host: istio-telemetry.istio-system.svc.cluster.local
  trafficPolicy:
    connectionPool:
      http:
        http2MaxRequests: 10000
        maxRequestsPerConnection: 10000
---
`