# Advancing-Kubernetes-and-HPC

## MOC-UI Project Proposal

Advancing Kubernetes and High Performance Computing

| Team Mentor        | Email                                           |
| ------------------ | ----------------------------------------------- |
| Dan Lambright      | [dlambrig@gmail.com](mailto:dlambrig@gmail.com) |
| **Team Members**   | **Email**                                       |
| Vedaj Jeeja Padman | jeejapadman.v@northeastern.edu                  |
| Yumeng Wu          | wu.yume@northeastern.edu                        |
| Jiayao Li          | li.jiayao@northeastern.edu                      |



## Project Description

## 1. Vision and Goals Of The Project:

The project is to continue the work of last year's students from BU in the cloud computing courses. It aims to enabling Lustre to be a well-behaved microservice in Kubernetes. Luster is a distributed file system used in high performance computing (HPC) and can be used in Kubernetes to be extended microservices, so that scientists only need to learn Kubernetes command line to operate Lustre.

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



## 3. Scope and Features Of The Project:

Environement setting up based on the instructions of last year's student group:

- instal, set up, initializations and execute the cloud, Kubernetes Clusters, Kubervirt, Containerized Lustre and operators
- set up Freeflow TCP

Go code to create "operators" that manipulate the instance, cluster and node of Luster with Kubernetes:

- Generate a new instance when an instance of Lustre crashes
- Install Lustre with various Kubernetes configurations in a user-friendly way
- upgrade a new Luster with simple with simple command line
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