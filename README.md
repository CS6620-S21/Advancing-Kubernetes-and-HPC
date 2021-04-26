# Advancing-Kubernetes-and-HPC

## MOC-UI Project Proposal

Advancing Kubernetes and High Performance Computing

| Mentor             | Email                                           |
| ------------------ | ----------------------------------------------- |
| Dan Lambright      | [dlambrig@gmail.com](mailto:dlambrig@gmail.com) |
| **Team Members**   | **Email**                                       |
| Jiayao Li          | li.jiayao@northeastern.edu                      |
| Vedaj Jeeja Padman | jeejapadman.v@northeastern.edu                  |
| Yumeng Wu          | wu.yume@northeastern.edu                        |
|                    |                                                 |

## Project Description

The project is a continuation of the work done by the students of the Cloud Computing course acrosss both Boston University and Northeastern University over the past few years.

## 1. Vision and Goals of The Project:

The main vision of the project is to enable Lustre to exist as a well behaved microservice in Kubernetes. Lustre, a distributed parallel filesystem popular in the scientific community, is crucial for dealing with the kind of data associated with the scientific and computational workloads and the the data that these processes in turn, generate.

By realizing Lustre as a Kubernetes microservice, we shall be taking the burden of having to go through innumerable steps of installation and configuration, from its user base. The end goal here is to enable users to simply deploy a container image that will have everything pre-baked, so that Lustre runbs smoothly to cater to its user's needs. 



In order to achieve these goals, we set two major specfictions. First, get more automatic operations working in K8s by adding Go code to create "operators". Second, improve the performance of running Kubervirt in containers by writing C++ code to extend Microsoft's freeflow overlay network.

> 1st meeting with Dan:
>
> - K8s operator definition 
>
> - what we gonna do:
>
>   Kubernetes operator
>
>   Watch the videos from last semester
>
>   Get onto “MOC”
>
>   Need 4 machines at least
>
>   Meet the students from last semester on slack

## 2. Users/Personas Of The Project:

Because of Lustre's wide scalability, high-performance, and high-availability, and Kubernetes' portability and extensibility, cloud-native HPC with Lustre has two major kinds of users:

- Researchers who need to perform HPC tasks with parallel file system.
- Data engineers and analysts who need to analyze massive volumes of data.

When users utilize Lustre, they don't need to have knowledge about file system, perform operations to scale up or to scale down the pods, create a new Lustre instance if one is crashed, or manually setup complex Kubernetes configurations.

## 3. Scope and Features Of The Project:

Environement setting up based on the instructions of last year's student group:

- instal, set up, initializations and execute the cloud, Kubernetes Clusters, Kubervirt, Containerized Lustre and operators
- set up Freeflow TCP

Go code to create "operators" that manipulate the instance, cluster and node of Lustre with Kubernetes:

- Generate a new instance when an instance of Lustre crashes
- Install Lustre with various Kubernetes configurations in a user-friendly way
- upgrade a new Lustre with simple with simple command line
- manage the operations with visualized dashboard

[need modification] C++ implementaion to get RDMA and Freeflow to run with Lustre, explore the usage of DPDK from Intel: 

- do performance test of RDMA, including memory, speed, security and accuracy..
- intergate DPDK with Freeflow to increase the security

## 4. Solution Concept:

**Global Arachitecural Structure of the Project**

**Kubernetes  components:**

- K8s operator definition 

turns distributed storage systems into self-managing, self-scaling, self-healing storage services. It automates the tasks of a storage administrator: deployment, bootstrapping, configuration, provisioning, scaling, upgrading, migration, disaster recovery, monitoring, and resource management.

**Lustre componenets:** 

**Freeflow components:**

>  add figure

## 5. Acceptance criteria:

The minimum acceptance criteria is to ..

- continue previous work, manipulate Lustre components on the MOC
- add more golfing opertor for Kubernetes, including Lustre instance backup, easy configure to install Lustre, dashboard monitoring..
- make Freeflow work inside Kubernetes

## 6. Release Planning:

2/19/2021 **Demo 1**: Setup single instance on MOC

- Follow the instruction of last year's GitHub to create instance on MOC

