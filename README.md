# ske

Suranaree Kubernetes Engine is an RKE-based installer for Docker, Kubernetes, and Istio.

## Installation
### Linux
download [ske-v0.1.9-ske-2.tar.gz](https://www.dropbox.com/s/wqjg2b8goyj7cwd/ske-v0.1.9-ske-2.tar.gz?dl=1) file. 


## Getting Started

To start a single-node Kubernetes cluster, save the following yaml into `cluster.yml`.

```
---
nodes:
- address: localhost
  user: ubuntu
  role:
  - controlplane
  - etcd
  - worker

service_mesh:
  provider: istio
```

And run `$ ske up`

## License

Copyright (c) 2018 [Suranaree University of Technology](http://www.sut.ac.th)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
