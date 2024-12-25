# CloudLand
Cloudland is a lightweight Infrastructure-as-a-Service (IaaS) framework designed to manage virtual machine instances, software-defined networks (SDN), storage volumes, and more. Capable of supporting over 10,000 hypervisors within a single cluster, Cloudland is well-suited to serve as the foundation for large-scale public cloud environments.   
   
In addition to its built-in multi-tenant support, Cloudland offers seamless integration with third-party authentication and authorization systems, making it an ideal solution for private cloud deployments or hyper-converged infrastructure (HCI) setups.   
    
**Cloudland’s key distinguishing features include:**   
- Lightweight Architecture: Minimal components for a streamlined, efficient design.
- Easy Learning Curve: Simple for both developers and operators to get up to speed quickly.
- High-Performance Messaging: Optimized for fast and reliable internal message delivery.
- HPC-Based Scalability: Built on a High-Performance Computing (HPC) architecture, ensuring exceptional scalability.
- Self-Healing and Stable: Automatic failure recovery with robust stability for continuous operation.
- Customizable: Easily extensible to implement custom features and meet specific needs.

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

Refer to [User Manual](https://github.com/IBM/cloudland/blob/master/doc/Manual.md)

## Reporting Issues

If you encounter any problem with this package, please open an [issue](https://github.com/IBM/cloudland/issues) tracker to us

## Contributing

Refer to [CONTRIBUTING](https://github.com/IBM/cloudland/blob/master/doc/Contribution.md)

## License

Apache License 2.0, see [LICENSE](https://github.com/IBM/cloudland/blob/master/LICENSE).

Visit [doc](https://github.com/IBM/cloudland/tree/master/doc) for full documentation and guide.



