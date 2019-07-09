# CloudLand

Cloudland, a light weight infrastructure as a service project, is a system framework to manage VM instances, software defined networks (SDN), volumes ... It can handle over 10 thousand of hypervisors in one cluster so it can be a base of large scale public cloud. It has OpenStack compatible API so it can be a straightforward alternative for private cloud or as a hyper converged infrastructure (HCI) solution.

Cloudland's main distinguishing features as compared to other IaaS systems are:
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
## Install

There are two purposes of installing Cloudland

### For development
Development environment can be installed from an all-in-one node and then extended to multiple nodes of hypervisors.

### For production
Production environment can be self-incubated from a development environment -- the new control plane VMs with high availability are launched from the development environment and they take over the management of the cluster and turn it to production environment.

Refer to [deployment guide](http//github.com/IBM/cloudland/wiki/Deployment) for more details

## Contributing

Refer to [CONTRIBUTING.md](https://github.com/IBM/cloudland/wiki/contribution)

## License

Apache License 2.0, see [LICENSE](https://github.com/IBM/cloudland/blob/master/LICENSE).

Visit [wiki](https://github.com/IBM/cloudland/wiki) for full documentation and guide.
