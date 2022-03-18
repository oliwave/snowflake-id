# snowflake_id

## Local development

### Prerequisite 

1. Download software
    -  [kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation) - a local Kuberentes cluster
    -  [Skaffold](https://skaffold.dev/docs/install/) - Local Kubernetes Development.

2. Clone Helm chart to local
    - `git clone git@gitlab.com:platntsist/devops/app/snowflake-id.git`

### Setup kind & Skaffold conifguration files
1. clone current project
2. `cd snowflake-id`
3. create following files
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
          # WARNING - please create target file (in this case should be `webhook`) before execute skaffold
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