# devcontainers (oci)

> This is an experiment!

### Login

```bash

make build-linux-amd64 && cd bin/linux/amd64

./devcontainer-oci login ghcr.io -u USERNAME -p <GITHUB_PAT_WITH_PACKAGING>
```

### Resolve feature identifiers to tarballs & lockfile

```bash
$ ./devcontainer-oci resolve ghcr.io/codspace/features/ruby:1,ghcr.io/codspace/features/go:1.0,ghcr.io/codspace/features/python:1.0.10


Fetching reference: ghcr.io/codspace/features/ruby:1
Downloaded  c33008d0dc12 ruby.tgz
Pulled ghcr.io/codspace/features/ruby:1

Fetching reference: ghcr.io/codspace/features/go:1.0
Downloaded  b2006e764719 go.tgz
Pulled ghcr.io/codspace/features/go:1.0

Fetching reference: ghcr.io/codspace/features/python:1.0.10
Downloaded  ef1941092547 python.tgz
Pulled ghcr.io/codspace/features/python:1.0.10


$ ls

devcontainer-oci  devcontainers.lock  go.tgz  python.tgz  ruby.tgz


$ cat devcontainers.lock 

ghcr.io/codspace/features/ruby:1
      1
      ghcr.io
      sha256:c33008d0dc12d0e631734082401bec692da809eae2ac51e24f58c1cac68fc0c9
      ruby.tgz
      application/vnd.devcontainers.layer.v1+tar

ghcr.io/codspace/features/go:1.0
      1.0
      ghcr.io
      sha256:b2006e7647191f7b47222ae48df049c6e21a4c5a04acfad0c4ef614d819de4c5
      go.tgz
      application/vnd.devcontainers.layer.v1+tar

ghcr.io/codspace/features/python:1.0.10
      1.0.10
      ghcr.io
      sha256:ef1941092547ee21c7cedfead12604bcfdd5dc096589c5b7a5a8f49e96c8d5d1
      python.tgz
      application/vnd.devcontainers.layer.v1+tar


$ cat go.tgz | sha256sum 

b2006e7647191f7b47222ae48df049c6e21a4c5a04acfad0c4ef614d819de4c5  -

```

### Fetch metadata for all features under namespace

```bash
$ ./devcontainer-oci metadata ghcr.io/codspace/features


Source Code: https://github.com/codspace/features
Commit:      06d028bfa680823ac1905d51ef0d3e6e626b452a

Available Features:
go
   1.0.9
   1.0
   1
   1.0.10
powershell
   1.0
   1.0.10
   1
   1.0.11
python
   1.0
   1.0.9
   1
   1.0.10
ruby
   1.0.13
   1.0
   1
   1.0.14
rust
   1.0
   1
   1.0.9
   1.0.10

```

## GitHub GHCR UX
See _https://github.com/orgs/codspace/packages?repo_name=features_
