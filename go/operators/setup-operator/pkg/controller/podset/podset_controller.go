package podset

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	appv1alpha1 "home/centos/go/operators/setup-operator/pkg/apis/app/v1alpha1"

	"github.com/spf13/pflag"
	"golang.org/x/crypto/ssh"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	kubevirtv1 "kubevirt.io/client-go/api/v1"
	"kubevirt.io/client-go/kubecli"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_podset")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new PodSet Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcilePodSet{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("podset-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource PodSet
	err = c.Watch(&source.Kind{Type: &appv1alpha1.PodSet{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner PodSet
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appv1alpha1.PodSet{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcilePodSet implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcilePodSet{}

// ReconcilePodSet reconciles a PodSet object
type ReconcilePodSet struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

var sshKeyFilePath string = ""
var sshPublicKey string = ""
var sshPrivateKey string = ""

type LustreVMStatus int

const (
	LustreVMStatusNone LustreVMStatus = iota
	LustreVMStatusCreated
	LustreVMStatusRunning
	LustreVMStatusMounted
	LustreVMStatusRecoveryStart
	LustreVMStatusRecoveryLaunching
	LustreVMStatusRecoveryRunning
)

var mgsStatus LustreVMStatus = LustreVMStatusNone
var mdsStatus LustreVMStatus = LustreVMStatusNone
var ossStatus LustreVMStatus = LustreVMStatusNone
var clientStatus LustreVMStatus = LustreVMStatusNone

var mdsIp string = ""
var mgsIp string = ""
var ossIp string = ""
var clientIp string = ""

// Reconcile reads that state of the cluster for a PodSet object and makes changes based on the state read
// and what is in the PodSet.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcilePodSet) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling PodSet")

	// Fetch the PodSet instance
	podSet := &appv1alpha1.PodSet{}
	err := r.client.Get(context.TODO(), request.NamespacedName, podSet)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	clientConfig := kubecli.DefaultClientConfig(&pflag.FlagSet{})
	namespace, _, err := clientConfig.Namespace()
	virtClient, err := kubecli.GetKubevirtClientFromClientConfig(clientConfig)

	fmt.Println("namespace:", namespace)
	nodes, err := virtClient.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	// get hostnames and ip addresses
	fmt.Println("nodes:")
	hostnames := []string{}
	nodeIp := []string{}
	for i := 0; i < len(nodes.Items); i++ {
		ip := nodes.Items[i].Status.Addresses
		fmt.Println(ip[0].Address, nodes.Items[i].GetLabels()["kubernetes.io/hostname"])
		hostnames = append(hostnames, nodes.Items[i].GetLabels()["kubernetes.io/hostname"])
		nodeIp = append(nodeIp, ip[0].Address)
	}

	vmList, err := virtClient.VirtualMachineInstance(namespace).List(&k8smetav1.ListOptions{})
	if err != nil {
		fmt.Println("cannot obtain KubeVirt client: %v\n", err)
	}

	// get lustre vm status
	hasOss, hasClient := false, false
	for _, vm := range vmList.Items {
		var vmIp string = ""
		if len(vm.Status.Interfaces) > 0 {
			vmIp = vm.Status.Interfaces[0].IP
		}
		if vm.Name == `lustre-mds` {
			if len(vmIp) > 0 && strings.Contains(vmIp, "/") {
				mdsIp = strings.Split(vmIp, "/")[0]
				if mdsStatus == LustreVMStatusNone || mdsStatus == LustreVMStatusCreated {
					mdsStatus = LustreVMStatusRunning
				}
			} else {
				mdsStatus = LustreVMStatusCreated
			}
		} else if vm.Name == `lustre-mgs` {
			if len(vmIp) > 0 && strings.Contains(vmIp, "/") {
				mgsIp = strings.Split(vmIp, "/")[0]
				if mgsStatus == LustreVMStatusNone || mgsStatus == LustreVMStatusCreated {
					mgsStatus = LustreVMStatusRunning
				}
			} else {
				mgsStatus = LustreVMStatusCreated
			}
		} else if vm.Name == `lustre-oss` {
			hasOss = true
			if len(vmIp) > 0 && strings.Contains(vmIp, "/") {
				ossIp = strings.Split(vmIp, "/")[0]
				if ossStatus == LustreVMStatusNone || ossStatus == LustreVMStatusCreated {
					ossStatus = LustreVMStatusRunning
				} else if ossStatus == LustreVMStatusRecoveryLaunching {
					ossStatus = LustreVMStatusRecoveryRunning
				}
			} else if ossStatus != LustreVMStatusRecoveryLaunching && ossStatus != LustreVMStatusRecoveryRunning && ossStatus != LustreVMStatusRecoveryStart {
				ossStatus = LustreVMStatusCreated
			}
		} else if vm.Name == `lustre-client` {
			hasClient = true
			if len(vmIp) > 0 && strings.Contains(vmIp, "/") {
				clientIp = strings.Split(vmIp, "/")[0]
				if clientStatus == LustreVMStatusNone || clientStatus == LustreVMStatusCreated {
					clientStatus = LustreVMStatusRunning
				}
			} else {
				clientStatus = LustreVMStatusCreated
			}
		}
	}

	sshKeyFilePath = podSet.Spec.SSHKeyPath
	sshPublicKey = podSet.Spec.SSHPublicKey
	sshPrivateKey = podSet.Spec.SSHPrivateKey

	if !hasOss && ossStatus != LustreVMStatusNone {
		ossStatus = LustreVMStatusRecoveryStart
	}
	if !hasClient {
		clientStatus = LustreVMStatusNone
	}

	if mgsStatus == LustreVMStatusRunning && checkVMMountStatus(mgsIp) {
		mgsStatus = LustreVMStatusMounted
	}
	if mdsStatus == LustreVMStatusRunning && checkVMMountStatus(mdsIp) {
		mdsStatus = LustreVMStatusMounted
	}
	if ossStatus == LustreVMStatusRunning && checkVMMountStatus(ossIp) {
		ossStatus = LustreVMStatusMounted
	}
	if clientStatus == LustreVMStatusRunning && checkVMMountStatus(clientIp) {
		clientStatus = LustreVMStatusMounted
	}

	if mgsStatus == LustreVMStatusNone {
		fmt.Println("Mgs: not created")
	} else if mgsStatus == LustreVMStatusCreated {
		fmt.Println("Mgs: created")
	} else if mgsStatus == LustreVMStatusRunning {
		fmt.Println("Mgs: running")
	} else if mgsStatus == LustreVMStatusMounted {
		fmt.Println("Mgs: mounted")
	}

	if mdsStatus == LustreVMStatusNone {
		fmt.Println("Mds: not created")
	} else if mdsStatus == LustreVMStatusCreated {
		fmt.Println("Mds: created")
	} else if mdsStatus == LustreVMStatusRunning {
		fmt.Println("Mds: running")
	} else if mdsStatus == LustreVMStatusMounted {
		fmt.Println("Mds: mounted")
	}

	if ossStatus == LustreVMStatusNone {
		fmt.Println("Oss: not created")
	} else if ossStatus == LustreVMStatusCreated {
		fmt.Println("Oss: created")
	} else if ossStatus == LustreVMStatusRunning {
		fmt.Println("Oss: running")
	} else if ossStatus == LustreVMStatusMounted {
		fmt.Println("Oss: mounted")
	} else if ossStatus == LustreVMStatusRecoveryStart {
		fmt.Println("Oss: start to recover")
	} else if ossStatus == LustreVMStatusRecoveryLaunching {
		fmt.Println("Oss: launching recovery node")
	} else if ossStatus == LustreVMStatusRecoveryRunning {
		fmt.Println("Oss: recovery is running")
	}

	if clientStatus == LustreVMStatusNone {
		fmt.Println("Client: not created")
	} else if clientStatus == LustreVMStatusCreated {
		fmt.Println("Client: created")
	} else if clientStatus == LustreVMStatusRunning {
		fmt.Println("Client: running")
	} else if clientStatus == LustreVMStatusMounted {
		fmt.Println("Client: mounted")
	}

	if mgsStatus == LustreVMStatusNone {
		fmt.Println("create mgs vm")
		mgsIp = nodeIp[1]
		createPv(mgsIp, `/pvc-data/mgt`, hostnames[1], `8Gi`, `pv-mgs`)
		createPvc(`vol-mgs`, `8Gi`)
		createService()
		createMgsVm()
		mgsStatus = LustreVMStatusCreated
		return reconcile.Result{Requeue: true, RequeueAfter: 240 * time.Second}, nil
	} else if mgsStatus == LustreVMStatusCreated {
		fmt.Println("mgs vm created")
		return reconcile.Result{Requeue: true, RequeueAfter: 60 * time.Second}, nil
	} else if mgsStatus == LustreVMStatusRunning {
		fmt.Println("mgs vm is running")
		if checkVMMountStatus(mgsIp) {
			fmt.Println("Lustre is mounted on mgs vm")
			mgsStatus = LustreVMStatusMounted
			return reconcile.Result{Requeue: true, RequeueAfter: 5 * time.Second}, nil
		} else {
			return reconcile.Result{Requeue: true, RequeueAfter: 60 * time.Second}, nil
		}
	}

	if mdsStatus == LustreVMStatusNone {
		fmt.Println("create mds vm")
		mdsIp = nodeIp[2]
		createPv(mdsIp, `/pvc-data/mdt`, hostnames[2], `8Gi`, `pv-mds`)
		createPvc(`vol-mds`, `8Gi`)
		createMdsVm()
		mdsStatus = LustreVMStatusCreated
		return reconcile.Result{Requeue: true, RequeueAfter: 240 * time.Second}, nil
	} else if mdsStatus == LustreVMStatusCreated {
		fmt.Println("mds vm created")
		return reconcile.Result{Requeue: true, RequeueAfter: 60 * time.Second}, nil
	} else if mdsStatus == LustreVMStatusRunning {
		fmt.Println("mds vm is running")
		if checkVMMountStatus(mdsIp) {
			fmt.Println("Lustre is mounted on mds vm")
			mdsStatus = LustreVMStatusMounted
			return reconcile.Result{Requeue: true, RequeueAfter: 5 * time.Second}, nil
		} else {
			return reconcile.Result{Requeue: true, RequeueAfter: 60 * time.Second}, nil
		}
	}

	if ossStatus == LustreVMStatusNone {
		fmt.Println("create oss vm")
		ossIp = nodeIp[1]
		createPv(ossIp, `/pvc-data/ost1`, hostnames[1], `1Gi`, `pv-oss1`)
		createPv(ossIp, `/pvc-data/ost2`, hostnames[1], `1Gi`, `pv-oss2`)
		createPvc(`vol-oss1`, `1Gi`)
		createPvc(`vol-oss2`, `1Gi`)
		createOssVm(false)
		ossStatus = LustreVMStatusCreated
		return reconcile.Result{Requeue: true, RequeueAfter: 240 * time.Second}, nil
	} else if ossStatus == LustreVMStatusCreated {
		fmt.Println("oss vm created")
		return reconcile.Result{Requeue: true, RequeueAfter: 60 * time.Second}, nil
	} else if ossStatus == LustreVMStatusRunning {
		fmt.Println("oss vm is running")
		if checkVMMountStatus(ossIp) {
			fmt.Println("Lustre is mounted on oss vm")
			ossStatus = LustreVMStatusMounted
			return reconcile.Result{Requeue: true, RequeueAfter: 5 * time.Second}, nil
		} else {
			return reconcile.Result{Requeue: true, RequeueAfter: 60 * time.Second}, nil
		}
	} else if ossStatus == LustreVMStatusRecoveryStart {
		fmt.Println("Start to recover oss")
		createOssVm(true)
		ossStatus = LustreVMStatusRecoveryLaunching
		return reconcile.Result{Requeue: true, RequeueAfter: 240 * time.Second}, nil
	} else if ossStatus == LustreVMStatusRecoveryLaunching {
		fmt.Println("Wait recovery oss running")
		return reconcile.Result{Requeue: true, RequeueAfter: 60 * time.Second}, nil
	} else if ossStatus == LustreVMStatusRecoveryRunning {
		fmt.Println("Oss recovery vm is running")
		// enableOssRecovery(ossIp)
		if checkVMMountStatus(ossIp) {
			fmt.Println("Oss recovery is mounted")
			ossStatus = LustreVMStatusMounted
			return reconcile.Result{Requeue: true, RequeueAfter: 5 * time.Second}, nil
		} else {
			return reconcile.Result{Requeue: true, RequeueAfter: 60 * time.Second}, nil
		}
	}

	if clientStatus == LustreVMStatusNone {
		fmt.Println("create client vm")
		createClientVM()
		clientStatus = LustreVMStatusCreated
		return reconcile.Result{Requeue: true, RequeueAfter: 240 * time.Second}, nil
	} else if clientStatus == LustreVMStatusCreated {
		fmt.Println("client vm created")
		return reconcile.Result{Requeue: true, RequeueAfter: 60 * time.Second}, nil
	} else if clientStatus == LustreVMStatusRunning {
		fmt.Println("client vm is running")
		if checkVMMountStatus(clientIp) {
			fmt.Println("Lustre is mounted on client vm")
			clientStatus = LustreVMStatusMounted
		} else {
			return reconcile.Result{Requeue: true, RequeueAfter: 60 * time.Second}, nil
		}
	}

	fmt.Println("finish")

	// // Define a new Pod object
	// pod := newPodForCR(instance)

	// // Set PodSet instance as the owner and controller
	// if err := controllerutil.SetControllerReference(instance, pod, r.scheme); err != nil {
	// 	return reconcile.Result{}, err
	// }

	// // Check if this Pod already exists
	// found := &corev1.Pod{}
	// err = r.client.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, found)
	// if err != nil && errors.IsNotFound(err) {
	// 	reqLogger.Info("Creating a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
	// 	err = r.client.Create(context.TODO(), pod)
	// 	if err != nil {
	// 		return reconcile.Result{}, err
	// 	}

	// 	// Pod created successfully - don't requeue
	// 	return reconcile.Result{}, nil
	// } else if err != nil {
	// 	return reconcile.Result{}, err
	// }

	// // Pod already exists - don't requeue
	// reqLogger.Info("Skip reconcile: Pod already exists", "Pod.Namespace", found.Namespace, "Pod.Name", found.Name)
	return reconcile.Result{Requeue: true, RequeueAfter: 10 * time.Second}, nil
}

func getFileContent(filename string) string {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	content, err := ioutil.ReadAll(file)
	return string(content)
}

// fork previous semester's repo
func getKeyFile(buf string) (key ssh.Signer, err error) {
	var b []byte = []byte(buf)
	key, err = ssh.ParsePrivateKey(b)
	if err != nil {
		return
	}
	return
}

// fork previous semester's repo
func connectToHost(user, host string, key ssh.Signer) (*ssh.Client, *ssh.Session, error) {
	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
	}
	sshConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()

	client, err := ssh.Dial("tcp", host, sshConfig)
	if err != nil {
		return nil, nil, err
	}

	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, nil, err
	}

	return client, session, nil
}

func createPv(ip string, path string, hostname string, size string, pvName string) {
	key, err := getKeyFile(sshPrivateKey)
	if err != nil {
		panic(err)
	}

	client, session, err := connectToHost("centos", ip+`:22`, key)
	if err != nil {
		fmt.Println(err)
	}

	// create folder
	var b bytes.Buffer
	session.Stdout = &b
	command := `sudo mkdir -p ` + path
	if len(path) > 0 && path != "/" {
		command = `sudo rm -rf ` + path + "; " + command
	}
	err = session.Run(command)
	if err != nil {
		fmt.Println("create pv error:", fmt.Sprintf("ip: %s, path: %s, error:", ip, path), err.Error())
	}
	client.Close()
	session.Close()

	//create pv
	clientConfig := kubecli.DefaultClientConfig(&pflag.FlagSet{})
	virtClient, err := kubecli.GetKubevirtClientFromClientConfig(clientConfig)
	mode := corev1.PersistentVolumeFilesystem
	pv := &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: pvName,
		},
		Spec: corev1.PersistentVolumeSpec{
			Capacity: corev1.ResourceList{
				"storage": resource.MustParse(size),
			},
			AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
			VolumeMode:       &mode,
			StorageClassName: "local-storage",
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				HostPath: &corev1.HostPathVolumeSource{Path: path},
			},
			PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimDelete,
			NodeAffinity: &corev1.VolumeNodeAffinity{
				Required: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{
						{
							MatchExpressions: []corev1.NodeSelectorRequirement{
								{
									Key:      `kubernetes.io/hostname`,
									Operator: corev1.NodeSelectorOpIn,
									Values: []string{
										hostname,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	_, err = virtClient.CoreV1().PersistentVolumes().Create(pv)
	fmt.Println(err)
}

func createPvc(name string, size string) {
	var class *string
	var s string = "local-storage"
	class = &s
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceName(corev1.ResourceStorage): resource.MustParse(size),
				},
			},
			StorageClassName: class,
		},
	}

	clientConfig := kubecli.DefaultClientConfig(&pflag.FlagSet{})
	virtClient, err := kubecli.GetKubevirtClientFromClientConfig(clientConfig)
	if err != nil {
		fmt.Println("cannot obtain KubeVirt client: %v\n", err)
	}
	_, err = virtClient.CoreV1().PersistentVolumeClaims(corev1.NamespaceDefault).Create(pvc)
	fmt.Println(err)
}

func createService() {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "default-lustre",
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"expose": "me",
			},
			ClusterIP: "None",
			Ports: []corev1.ServicePort{
				{
					Name: "foo",
					Port: 1234,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 1234,
					},
				},
			},
		},
	}

	clientConfig := kubecli.DefaultClientConfig(&pflag.FlagSet{})
	virtClient, err := kubecli.GetKubevirtClientFromClientConfig(clientConfig)
	if err != nil {
		fmt.Println("cannot obtain KubeVirt client: %v\n", err)
	}
	_, err = virtClient.CoreV1().Services(corev1.NamespaceDefault).Create(svc)
}

func createMgsVm() {
	clientConfig := kubecli.DefaultClientConfig(&pflag.FlagSet{})
	virtClient, err := kubecli.GetKubevirtClientFromClientConfig(clientConfig)
	if err != nil {
		fmt.Println("cannot obtain KubeVirt client: %v\n", err)
	}

	vm := kubevirtv1.NewMinimalVMI(`lustre-mgs`)
	vm.Spec.Domain.Devices.Interfaces = []kubevirtv1.Interface{
		kubevirtv1.Interface{
			Name: "default",
			InterfaceBindingMethod: kubevirtv1.InterfaceBindingMethod{
				Bridge: &kubevirtv1.InterfaceBridge{},
			},
		},
	}

	vm.Spec.Domain.Resources = kubevirtv1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceMemory: resource.MustParse("1024M"),
		},
	}

	vm.Spec.Volumes = []kubevirtv1.Volume{
		{
			Name: `vol-mgs`,
			VolumeSource: kubevirtv1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: `vol-mgs`,
				},
			},
		},
		{
			Name: "containerdisk",
			VolumeSource: kubevirtv1.VolumeSource{
				ContainerDisk: &kubevirtv1.ContainerDiskSource{
					Image: "nakulvr/centos:lustre-server",
				},
			},
		},
		{
			Name: "cloudinitdisk",
			VolumeSource: kubevirtv1.VolumeSource{
				CloudInitNoCloud: &kubevirtv1.CloudInitNoCloudSource{
					UserData: `#cloud-config
ssh_authorized_keys:
  - ` + sshPublicKey + `
runcmd:
  - sudo exec /sbin/modprobe -v lnet >/dev/null 2>&1
  - /sbin/lsmod | /bin/grep lustre 1>/dev/null 2>&1
  - sudo /sbin/modprobe -v lustre >/dev/null 2>&1
  - /sbin/lsmod | /bin/grep zfs 1>/dev/null 2>&1
  - sudo /sbin/modprobe -v zfs >/dev/null 2>&1
  - sudo /usr/sbin/mkfs.lustre --mgs /dev/vdb > /dev/null 2>&1
  - sudo /usr/bin/mkdir /mgt
  - sudo /usr/sbin/mount.lustre /dev/vdb /mgt`,
				},
			},
		},
	}

	vm.Spec.Networks = []kubevirtv1.Network{
		kubevirtv1.Network{
			Name: "default",
			NetworkSource: kubevirtv1.NetworkSource{
				Pod: &kubevirtv1.PodNetwork{},
			},
		},
	}

	vm.Spec.Domain.Devices.Disks = []kubevirtv1.Disk{
		{
			Name: "containerdisk",
			DiskDevice: kubevirtv1.DiskDevice{
				Disk: &kubevirtv1.DiskTarget{
					Bus: "virtio",
				},
			},
		},
		{
			Name: `vol-mgs`,
			DiskDevice: kubevirtv1.DiskDevice{
				Disk: &kubevirtv1.DiskTarget{
					Bus: "virtio",
				},
			},
		},
		{
			Name: "cloudinitdisk",
			DiskDevice: kubevirtv1.DiskDevice{
				Disk: &kubevirtv1.DiskTarget{
					Bus: "virtio",
				},
			},
		},
	}

	fetchedVMI, err := virtClient.VirtualMachineInstance(corev1.NamespaceDefault).Create(vm)
	fmt.Println(fetchedVMI, err)
}

func createMdsVm() {
	clientConfig := kubecli.DefaultClientConfig(&pflag.FlagSet{})
	virtClient, err := kubecli.GetKubevirtClientFromClientConfig(clientConfig)
	if err != nil {
		fmt.Println("cannot obtain KubeVirt client: %v\n", err)
	}

	vm := kubevirtv1.NewMinimalVMI(`lustre-mds`)
	vm.Spec.Domain.Devices.Interfaces = []kubevirtv1.Interface{
		kubevirtv1.Interface{
			Name: "default",
			InterfaceBindingMethod: kubevirtv1.InterfaceBindingMethod{
				Bridge: &kubevirtv1.InterfaceBridge{},
			},
		},
	}

	vm.Spec.Domain.Resources = kubevirtv1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceMemory: resource.MustParse("1024M"),
		},
	}

	vm.Spec.Volumes = []kubevirtv1.Volume{
		{
			Name: `vol-mds`,
			VolumeSource: kubevirtv1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: `vol-mds`,
				},
			},
		},
		{
			Name: "containerdisk",
			VolumeSource: kubevirtv1.VolumeSource{
				ContainerDisk: &kubevirtv1.ContainerDiskSource{
					Image: "nakulvr/centos:lustre-server",
				},
			},
		},
		{
			Name: "cloudinitdisk",
			VolumeSource: kubevirtv1.VolumeSource{
				CloudInitNoCloud: &kubevirtv1.CloudInitNoCloudSource{
					UserData: `#cloud-config
ssh_authorized_keys:
  - ` + sshPublicKey + `
runcmd:
  - sudo exec /sbin/modprobe -v lnet >/dev/null 2>&1
  - /sbin/lsmod | /bin/grep lustre 1>/dev/null 2>&1
  - sudo /sbin/modprobe -v lustre >/dev/null 2>&1
  - /sbin/lsmod | /bin/grep zfs 1>/dev/null 2>&1
  - sudo /sbin/modprobe -v zfs >/dev/null 2>&1
  - sudo /usr/sbin/mkfs.lustre --fsname=lustrefs --mgsnode=` + mgsIp + `@tcp0 --mdt --index=0 /dev/vdb > /dev/null 2>&1
  - sudo /usr/bin/mkdir /mdt
  - sudo /usr/sbin/mount.lustre /dev/vdb /mdt`,
				},
			},
		},
	}

	vm.Spec.Networks = []kubevirtv1.Network{
		kubevirtv1.Network{
			Name: "default",
			NetworkSource: kubevirtv1.NetworkSource{
				Pod: &kubevirtv1.PodNetwork{},
			},
		},
	}

	vm.Spec.Domain.Devices.Disks = []kubevirtv1.Disk{
		{
			Name: "containerdisk",
			DiskDevice: kubevirtv1.DiskDevice{
				Disk: &kubevirtv1.DiskTarget{
					Bus: "virtio",
				},
			},
		},
		{
			Name: `vol-mds`,
			DiskDevice: kubevirtv1.DiskDevice{
				Disk: &kubevirtv1.DiskTarget{
					Bus: "virtio",
				},
			},
		},
		{
			Name: "cloudinitdisk",
			DiskDevice: kubevirtv1.DiskDevice{
				Disk: &kubevirtv1.DiskTarget{
					Bus: "virtio",
				},
			},
		},
	}

	fetchedVMI, err := virtClient.VirtualMachineInstance(corev1.NamespaceDefault).Create(vm)
	fmt.Println(fetchedVMI, err)
}

func createOssVm(recovery bool) {
	clientConfig := kubecli.DefaultClientConfig(&pflag.FlagSet{})
	virtClient, err := kubecli.GetKubevirtClientFromClientConfig(clientConfig)
	if err != nil {
		fmt.Println("cannot obtain KubeVirt client: %v\n", err)
	}

	vm := kubevirtv1.NewMinimalVMI(`lustre-oss`)
	vm.Spec.Domain.Devices.Interfaces = []kubevirtv1.Interface{
		kubevirtv1.Interface{
			Name: "default",
			InterfaceBindingMethod: kubevirtv1.InterfaceBindingMethod{
				Bridge: &kubevirtv1.InterfaceBridge{},
			},
		},
	}

	vm.Spec.Domain.Resources = kubevirtv1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceMemory: resource.MustParse("1024M"),
		},
	}

	vmUserData := ""
	if !recovery {
		vmUserData = `#cloud-config
ssh_authorized_keys:
  - ` + sshPublicKey + `
runcmd:
  - sudo exec /sbin/modprobe -v lnet >/dev/null 2>&1
  - /sbin/lsmod | /bin/grep lustre 1>/dev/null 2>&1
  - sudo /sbin/modprobe -v lustre >/dev/null 2>&1
  - /sbin/lsmod | /bin/grep zfs 1>/dev/null 2>&1
  - sudo /sbin/modprobe -v zfs >/dev/null 2>&1
  - sudo /usr/sbin/mkfs.lustre --ost --fsname=lustrefs --mgsnode=` + mgsIp + `@tcp0 --index=1 /dev/vdb > /dev/null 2>&1
  - sudo /usr/sbin/mkfs.lustre --ost --fsname=lustrefs --mgsnode=` + mgsIp + `@tcp0 --index=2 /dev/vdc > /dev/null 2>&1
  - sudo /usr/bin/mkdir /ost1
  - sudo /usr/bin/mkdir /ost2
  - sudo /usr/sbin/mount.lustre /dev/vdb /ost1
  - sudo /usr/sbin/mount.lustre /dev/vdc /ost2`
	} else {
		vmUserData = `#cloud-config
ssh_authorized_keys:
  - ` + sshPublicKey + `
runcmd:
  - sudo exec /sbin/modprobe -v lnet >/dev/null 2>&1
  - /sbin/lsmod | /bin/grep lustre 1>/dev/null 2>&1
  - sudo /sbin/modprobe -v lustre >/dev/null 2>&1
  - /sbin/lsmod | /bin/grep zfs 1>/dev/null 2>&1
  - sudo /sbin/modprobe -v zfs >/dev/null 2>&1
  - sudo /usr/sbin/mkfs.lustre --ost --fsname=lustrefs --mgsnode=` + mgsIp + `@tcp0 --index=1 --reformat /dev/vdb > /dev/null 2>&1
  - sudo /usr/sbin/mkfs.lustre --ost --fsname=lustrefs --mgsnode=` + mgsIp + `@tcp0 --index=2 --reformat /dev/vdc > /dev/null 2>&1
  - sudo /usr/bin/mkdir /ost1
  - sudo /usr/bin/mkdir /ost2
  - sudo /usr/sbin/mount.lustre /dev/vdb /ost1
  - sudo /usr/sbin/mount.lustre /dev/vdb /ost1
  - sudo /usr/sbin/mount.lustre /dev/vdc /ost2
  - sudo /usr/sbin/mount.lustre /dev/vdc /ost2`
	}

	vm.Spec.Volumes = []kubevirtv1.Volume{
		{
			Name: `vol-oss1`,
			VolumeSource: kubevirtv1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: `vol-oss1`,
				},
			},
		},
		{
			Name: `vol-oss2`,
			VolumeSource: kubevirtv1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: `vol-oss2`,
				},
			},
		},
		{
			Name: "containerdisk",
			VolumeSource: kubevirtv1.VolumeSource{
				ContainerDisk: &kubevirtv1.ContainerDiskSource{
					Image: "nakulvr/centos:lustre-server",
				},
			},
		},
		{
			Name: "cloudinitdisk",
			VolumeSource: kubevirtv1.VolumeSource{
				CloudInitNoCloud: &kubevirtv1.CloudInitNoCloudSource{
					UserData: vmUserData,
				},
			},
		},
	}

	vm.Spec.Networks = []kubevirtv1.Network{
		kubevirtv1.Network{
			Name: "default",
			NetworkSource: kubevirtv1.NetworkSource{
				Pod: &kubevirtv1.PodNetwork{},
			},
		},
	}

	vm.Spec.Domain.Devices.Disks = []kubevirtv1.Disk{
		{
			Name: "containerdisk",
			DiskDevice: kubevirtv1.DiskDevice{
				Disk: &kubevirtv1.DiskTarget{
					Bus: "virtio",
				},
			},
		},
		{
			Name: `vol-oss1`,
			DiskDevice: kubevirtv1.DiskDevice{
				Disk: &kubevirtv1.DiskTarget{
					Bus: "virtio",
				},
			},
		},
		{
			Name: `vol-oss2`,
			DiskDevice: kubevirtv1.DiskDevice{
				Disk: &kubevirtv1.DiskTarget{
					Bus: "virtio",
				},
			},
		},
		{
			Name: "cloudinitdisk",
			DiskDevice: kubevirtv1.DiskDevice{
				Disk: &kubevirtv1.DiskTarget{
					Bus: "virtio",
				},
			},
		},
	}

	fetchedVMI, err := virtClient.VirtualMachineInstance(corev1.NamespaceDefault).Create(vm)
	fmt.Println(fetchedVMI, err)
}

func createClientVM() {
	clientConfig := kubecli.DefaultClientConfig(&pflag.FlagSet{})
	virtClient, err := kubecli.GetKubevirtClientFromClientConfig(clientConfig)
	if err != nil {
		fmt.Println("cannot obtain KubeVirt client: %v\n", err)
	}

	vm := kubevirtv1.NewMinimalVMI(`lustre-client`)
	vm.Spec.Domain.Devices.Interfaces = []kubevirtv1.Interface{
		kubevirtv1.Interface{
			Name: "default",
			InterfaceBindingMethod: kubevirtv1.InterfaceBindingMethod{
				Bridge: &kubevirtv1.InterfaceBridge{},
			},
		},
	}

	vm.Spec.Domain.Resources = kubevirtv1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceMemory: resource.MustParse("1024M"),
		},
	}

	vm.Spec.Volumes = []kubevirtv1.Volume{
		{
			Name: "containerdisk",
			VolumeSource: kubevirtv1.VolumeSource{
				ContainerDisk: &kubevirtv1.ContainerDiskSource{
					Image: "nakulvr/centos:lustre-client",
				},
			},
		},
		{
			Name: "cloudinitdisk",
			VolumeSource: kubevirtv1.VolumeSource{
				CloudInitNoCloud: &kubevirtv1.CloudInitNoCloudSource{
					UserData: `#cloud-config
ssh_authorized_keys:
  - ` + sshPublicKey + `
runcmd:
  - sudo exec /sbin/modprobe -v lnet >/dev/null 2>&1
  - /sbin/lsmod | /bin/grep lustre 1>/dev/null 2>&1
  - sudo /sbin/modprobe -v lustre >/dev/null 2>&1
  - sudo /usr/bin/mkdir /lustrefs
  - sudo /usr/bin/mount -t lustre ` + mgsIp + `@tcp0:/lustrefs /lustrefs`,
				},
			},
		},
	}

	vm.Spec.Networks = []kubevirtv1.Network{
		kubevirtv1.Network{
			Name: "default",
			NetworkSource: kubevirtv1.NetworkSource{
				Pod: &kubevirtv1.PodNetwork{},
			},
		},
	}

	vm.Spec.Domain.Devices.Disks = []kubevirtv1.Disk{
		{
			Name: "containerdisk",
			DiskDevice: kubevirtv1.DiskDevice{
				Disk: &kubevirtv1.DiskTarget{
					Bus: "virtio",
				},
			},
		},
		{
			Name: "cloudinitdisk",
			DiskDevice: kubevirtv1.DiskDevice{
				Disk: &kubevirtv1.DiskTarget{
					Bus: "virtio",
				},
			},
		},
	}

	fetchedVMI, err := virtClient.VirtualMachineInstance(corev1.NamespaceDefault).Create(vm)
	fmt.Println(fetchedVMI, err)
}

func checkVMMountStatus(ip string) bool {
	key, err := getKeyFile(sshPrivateKey)
	if err != nil {
		panic(err)
	}

	client, session, err := connectToHost("centos", ip+`:22`, key)
	if err != nil {
		fmt.Println(err)
		return false
	}

	var b bytes.Buffer
	session.Stdout = &b

	command := "sudo lctl dl"
	err = session.Run(command)
	if err != nil {
		fmt.Println("Failed to run: " + err.Error())
	}
	var res bool = false
	if len(b.String()) > 0 && strings.Contains(b.String(), "UP") {
		res = true
	}
	client.Close()
	session.Close()

	return res
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *appv1alpha1.PodSet) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-pod",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "busybox",
					Image:   "busybox",
					Command: []string{"sleep", "3600"},
				},
			},
		},
	}
}
