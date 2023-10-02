# snowflake_id

[Snowflake ID](https://en.wikipedia.org/wiki/Snowflake_ID) is an algorithm created by Twitter for generating unique identifiers under distributed systems. This project is meant to create these IDs for applications hosted in the Kubernetes environment. With the help of [AdmissionWebhook](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/), the customized snowflake-id controller can intercept the pod creation/deletion events from Kubernetes API server and verify if the pod is annotated with the label `snowflake-id.io/enabled` with `true` value. If so, the controller will modify the pod spec for you, attaching generated values, `SNOWFLAKE_DATA_CENTER_ID` and `SNOWFLAKE_WORKER_ID`, to the environment fields, and deploy the pod to the designated node.

## System Design

### Terminology Redefinition

10 bits represent a machine ID consisting of 5 bits for worker nodes and 5 bits for pods
- Data Center ID -> worker node
- Worker ID -> pod

### Limitation factor
1. A single microservice can only be deployed to the utmost 32 worker nodes
2. A single microservice can only be deployed to a single worker node for the utmost 32 pods
3. A single microservice can have 1032 pods (32*32)

![snowflake-id](https://github.com/oliwave/snowflake-id/assets/27968072/15c64322-2e79-401c-9984-13cb07097df8)

## Local development

### Prerequisite 

Download Software
- [kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation) - a local Kubernetes cluster
- [Skaffold](https://skaffold.dev/docs/install/) - Local Kubernetes Development.

### Setup kind & Skaffold configuration files
1. clone current project
2. `cd snowflake-id`
3. create the following files
    - ```yaml=
      # kind.yaml
      ---
      kind: Cluster
      apiVersion: kind.x-k8s.io/v1alpha4
      name: snowflake-id-cluster-test
      nodes:
        - role: control-plane
          image: kindest/node:v1.21.1@sha256:69860bda5563ac81e3c0057d654b5253219618a22ec3a346306239bba8cfa1a6
        - role: worker
          image: kindest/node:v1.21.1@sha256:69860bda5563ac81e3c0057d654b5253219618a22ec3a346306239bba8cfa1a6
        - role: worker
          image: kindest/node:v1.21.1@sha256:69860bda5563ac81e3c0057d654b5253219618a22ec3a346306239bba8cfa1a6
        - role: worker
          image: kindest/node:v1.21.1@sha256:69860bda5563ac81e3c0057d654b5253219618a22ec3a346306239bba8cfa1a6
      ```
    - ```yaml=
      # skaffold.yaml
      apiVersion: skaffold/v2beta27
      kind: Config
      metadata:
        name: local-cluster
      build:
        artifacts:
        - image: snowflake_id
          # An alpha feature https://skaffold.dev/docs/pipeline-stages/lifecycle-hooks/
          #
          # WARNING - please create a target file (in this case should be `webhook`) before executing skaffold
          hooks:
            before:
              - command: ["sh", "-c", "./compile.sh"]
          context: .
          docker:
            dockerfile: Dockerfile
      deploy:
        helm:
          releases:
          - name: "snowflake-id-chart"
            artifactOverrides:
              image: snowflake_id # no tag present!
            imageStrategy:
              helm: {}
            chartPath: # the path to your local helm project
            valuesFiles: # the value file of your local helm project
      ```
4. `skaffold dev`
### Happy coding!
