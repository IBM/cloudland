# CloudLand
Cloudland, a light weight infrastructure as a service project, plus OpenShift 4 platform as a service deployment engine, is a system framework to manage VM instances, software defined networks (SDN), volumes ..., It can handle over 10 thousand hypervisors in one cluster so it can be a base of large scale public cloud. More over, with multi-tenant and OpenShift 4 cluster deployments on demand, it can be a straightforward alternative for private cloud or as a hyper converged infrastructure (HCI) solution.

Cloudland's main distinguishing features are:
- Able to deploy OpenShift 4 cluster on demand per tenant
- Compatibility with [Openstack API](https://ibm.github.io/cloudland/) (TBD)
- Light weight, no tons of components
- Flat learning curve for both developers and operators
- Excellent performance for inside messages delivery
- Based on HPC architecture, so super scalable
- Self auto fail recovery and stable
- Easy to customize to implement your own feature

## Architecture overview
![](https://raw.githubusercontent.com/wiki/IBM/cloudland/images/architecture.svg?sanitize=true)   
To support ultra-large scale, the hypervisors are organized into a tree hierarchy, the agents (scia are launched on demand)   

![](https://raw.githubusercontent.com/wiki/IBM/cloudland/images/tree.svg?sanitize=true)

For more information, see the [Introduction](https://github.com/IBM/cloudland/blob/master/doc/Introduction.md)

## Installation

Support two ways to install cloudland

1. Install cloudland with rpm package in a quick starter
   - Refer to [Installation](https://github.com/IBM/cloudland/blob/master/doc/Installation.md) to get more details
2. Build source code and install cloudland from end to end
   - Refer to [Build](https://github.com/IBM/cloudland/blob/master/doc/Build.md) and [Installation](https://github.com/IBM/cloudland/blob/master/doc/Installation.md) to get more informations

## User Guide

- Launch Virtual Machine
- Create Openshift Cluster

For more usage, refer to [User Manual](https://github.com/IBM/cloudland/blob/master/doc/Manual.md)

## Reporting Issues

If you encounter any problem with this package, please open an [issue](https://github.com/IBM/cloudland/issues) tracker to us

## Contributing

Refer to [CONTRIBUTING](https://github.com/IBM/cloudland/blob/master/doc/Contribution.md)

## License

Apache License 2.0, see [LICENSE](https://github.com/IBM/cloudland/blob/master/LICENSE).

Visit [doc](https://github.com/IBM/cloudland/tree/master/doc) for full documentation and guide.



