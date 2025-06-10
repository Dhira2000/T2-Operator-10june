package controllers

import (
	// "bytes"
	"context"
	"fmt"
	"os/exec"
	// "strconv"
	// "sigs.k8s.io/controller-runtime/pkg/handler"
	// "sigs.k8s.io/controller-runtime/pkg/source"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// securityv1 "github.com/openshift/api/security/v1"
	// securityv1 "github.com/openshift/api/security/v1"
	// policyv1beta1 "k8s.io/api/policy/v1beta1"
	// policyv1beta1 "k8s.io/api/policy/v1beta1"
	// policyv1 "k8s.io/api/policy/v1"
	// policyv1 "k8s.io/api/security/v1"
	// rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	// "k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	// corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	sriovfect2v1 "github.com/amd-raghurao/AMD-T2/api/v1"
)

var (
	setupLog = log.Log.WithName("setup")

	scriptExecuted bool
	// drainSkip       bool
	adminMode       bool
	pciPrefix       string
	uuidTokenGlobal string
	adminPodName    string
	adminCmdStatic  string
	allPciAddresses []string
	pciAddress      string
	dockerImage     string
	myNameSpace     string
	fetchAllPCI     bool
)

// SriovT2CardReconciler reconciles a SriovT2Card object
type SriovT2CardReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *SriovT2CardReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	setupLog := log.FromContext(ctx)

	cr := &sriovfect2v1.SriovT2Card{}
	if err := r.Get(ctx, req.NamespacedName, cr); err != nil {
		if errors.IsNotFound(err) {
			// Clean up all resources when the CR is deleted
			setupLog.Info("Cleaning up all resources related to the SriovT2Card")
			if cleanupErr := r.cleanupResources(ctx, req.NamespacedName.Namespace); cleanupErr != nil {
				setupLog.Error(cleanupErr, "Failed to clean up resources")
				return ctrl.Result{}, cleanupErr
			}
			fmt.Println("ALL CleanUp Done...")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	//12
	//12

	setupLog.Info("Starting reconciliation process for SriovT2Card", "namespace", cr.Namespace)
	// fmt.Println("Adding SCC")
	// Ensure SCC is created and assigned
	// if err := r.ensurePodSecurityPolicy(ctx, cr.Namespace); err != nil {
	// 	setupLog.Error(err, "Failed to ensure SCC")
	// 	return ctrl.Result{}, err
	// }

	fmt.Println("x.........AMD-T2-Card...........x")
	fmt.Println("x.........5.3.8...........x")
	fmt.Println("x.........Prometheus...........x")

	// adminMode = cr.Spec.AdminMode
	adminMode = true
	myNameSpace = cr.Namespace
	fmt.Println("NameSpace: " + myNameSpace)
	// dockerImage = cr.Spec.Template.Spec.containers[0].Image
	dockerImage = cr.Spec.Template.Spec.Containers[0].Image
	fmt.Println("dockerImage: " + dockerImage)
	uuidTokenGlobal = "14d63f20-8445-11ea-8900-1f9ce7d5650d"
	fmt.Println("uuidTokenGlobal: " + uuidTokenGlobal)
	fmt.Println("pciAddress: " + pciAddress)
	pciAddress = cr.Spec.AcceleratorSelector.PciAddress
	fmt.Println("pciAddress: " + pciAddress)
	// fmt.Println("Creating Dynamic Token")
	// time.Sleep(3 * time.Second)
	// // Generate UUID token
	// tokenCmd := "uuidgen"
	// cmdOutput, err := execCommand(ctx, r.Client, tokenCmd)
	// if err != nil {
	// 	fmt.Println("Error generating UUID token:", err)
	// 	// return
	// 	// continue
	// }
	// uuidTokenGlobal = cmdOutput
	// fmt.Println(uuidTokenGlobal)

	//secret token

	// Create Secret
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "t2-card-token",
			Namespace: cr.Namespace,
		},
		StringData: map[string]string{
			"T2_CARD_TOKEN": uuidTokenGlobal,
		},
	}
	if err := r.Client.Create(ctx, secret); err != nil {
		if errors.IsAlreadyExists(err) {
			if err := r.Client.Update(ctx, secret); err != nil {
				setupLog.Error(err, "Failed to update Secret")
				return ctrl.Result{}, err
			}
		} else {
			setupLog.Error(err, "Failed to create Secret")
			return ctrl.Result{}, err
		}
	}

	//secret token

	time.Sleep(5 * time.Second)

	fmt.Println("Basic Setup Started...")
	// Create a DaemonSet to run the necessary commands on all nodes
	ds := generateSetupDaemonSet(cr, r.Client, pciAddress)
	if err := r.Client.Create(ctx, ds); err != nil {
		setupLog.Error(err, "Failed to create DaemonSet")
		return ctrl.Result{}, err
	}
	setupLog.Info("Setup DaemonSet created successfully")
	time.Sleep(5 * time.Second)

	// Clean up the DaemonSet un-comment
	if err := r.Client.Delete(ctx, ds); err != nil {
		setupLog.Error(err, "Failed to delete DaemonSet")
	}
	setupLog.Info("Setup DaemonSet deleted successfully")
	fmt.Println("Basic Setup Completed...")
	time.Sleep(5 * time.Second)
	fmt.Println("Adding Resources To Node Level...")
	// Apply SR-IOV device plugin configuration
	// Ensure SR-IOV device plugin configuration is reapplied
	if err := applySriovDevicePluginConfig(ctx, r.Client, cr.Namespace); err != nil {
		setupLog.Error(err, "Failed to apply SR-IOV device plugin configuration")
		return ctrl.Result{}, err
	}
	setupLog.Info("SR-IOV device plugin configuration applied successfully")
	time.Sleep(5 * time.Second)
	fmt.Println("Adding Resources To Node Level Completed...")
	// Create a debug DaemonSet to trial
	// dsDebug := generateDebugDaemonSet(cr, r.Client, pciAddress)
	// if err := r.Client.Create(ctx, dsDebug); err != nil {
	// 	setupLog.Error(err, "Failed to create debug DaemonSet")
	// 	return ctrl.Result{}, err
	// }
	// setupLog.Info("Debug DaemonSet created successfully")
	time.Sleep(5 * time.Second)
	fmt.Println("Started To Up bbdev Admin App...")
	// time.Sleep(5 * time.Second)
	//123

	// Define the name and namespace of the Admin DaemonSet
	dsName := "sriovt2card-admin"
	dsNamespace := cr.Namespace

	// Variable to hold the generated Admin DaemonSet
	var dsNew *appsv1.DaemonSet

	// Check if the Admin DaemonSet already exists
	existingDS := &appsv1.DaemonSet{}
	err := r.Client.Get(ctx, types.NamespacedName{Name: dsName, Namespace: dsNamespace}, existingDS)
	if err != nil && errors.IsNotFound(err) {
		// DaemonSet does not exist, so create it
		dsNew = generateAdminDaemonSet(cr, r.Client, pciAddress, uuidTokenGlobal)
		if err := r.Client.Create(ctx, dsNew); err != nil {
			setupLog.Error(err, "Failed to create Admin DaemonSet")
			return ctrl.Result{}, err
		}
		setupLog.Info("Admin DaemonSet created successfully")
	} else if err != nil {
		// An error occurred while trying to check for the DaemonSet
		setupLog.Error(err, "Failed to check if Admin DaemonSet exists")
		return ctrl.Result{}, err
	} else {
		// DaemonSet already exists, log and continue
		dsNew = existingDS // Reuse the existing DaemonSet
		setupLog.Info("Admin DaemonSet already exists, skipping creation")
	}
	//123
	// Create an Admin DaemonSet generateDebugDaemonSetNew
	// dsNew := generateAdminDaemonSet(cr, r.Client, pciAddress, uuidTokenGlobal)
	// if err := r.Client.Create(ctx, dsNew); err != nil {
	// 	setupLog.Error(err, "Failed to create Admin DaemonSet")
	// 	return ctrl.Result{}, err
	// }
	setupLog.Info("Admin DaemonSet created successfully")
	fmt.Println("Admin App Up...")

	time.Sleep(5 * time.Second)

	// Retrieve Admin DaemonSet pods and run the command
	podListNew := &corev1.PodList{}
	listOptsNew := []client.ListOption{
		client.InNamespace(cr.Namespace),
		client.MatchingLabels(dsNew.Spec.Template.ObjectMeta.Labels),
	}
	if errNew := r.Client.List(ctx, podListNew, listOptsNew...); errNew != nil {
		setupLog.Error(errNew, "Failed to list pods created by Admin DaemonSet")
		return ctrl.Result{}, errNew
	}

	for _, podNew := range podListNew.Items {
		if podNew.DeletionTimestamp != nil {
			setupLog.Info(fmt.Sprintf("Skipping execution of commands in terminating pod: %s", podNew.Name))
			continue
		}

		adminPodName = podNew.Name
		setupLog.Info(fmt.Sprintf("Running dpdk main Admin commands in pod: %s", podNew.Name))

		if adminMode {
			// setupLog.Info("Ready to run the bbdev test")
			fmt.Println("Ready to run the bbdev test")
		} else {
			// setupLog.Info("Ready to run the Admin and bbdev test")
			fmt.Println("Ready to run the Admin and bbdev test")
		}
	}

	setupLog.Info("Reconciliation process completed successfully...")
	setupLog.Info("Start monitoring the devices and reapplying if needed...")
	// Start monitoring and reapplying if needed
	//r.Client, cr.Namespace
	// type SriovT2CardSpec struct {
	// 	NodeSelector        map[string]string   `json:"nodeSelector"`}
	// nodeName := cr.Spec.NodeSelector["kubernetes.io/hostname"]
	// monitorAndReapply(ctx, r.Client, cr.Namespace, nodeName)
	//12340000
	// monitorAndReapply(ctx, r.Client, cr.Namespace, cr.Spec.NodeSelector["kubernetes.io/hostname"])

	return ctrl.Result{}, nil

	// Requeue after 5 minutes to ensure resources are continuously managed
	// return ctrl.Result{RequeueAfter: 5 * time.Minute}, nil
}

//12345

// Function to run the loop every 5 minutes
func monitorAndReapply(ctx context.Context, kubeClient client.Client, namespace string, nodeName string) {
	resourcePF := "amd.com/amd_xilinx_t2_pf"
	resourceVF := "amd.com/amd_xilinx_t2_vf"

	for {
		pfAvailable, err := isResourceAvailable(ctx, kubeClient, nodeName, resourcePF)
		if err != nil {
			fmt.Printf("Error checking PF resource availability: %v\n", err)
		}

		vfAvailable, err := isResourceAvailable(ctx, kubeClient, nodeName, resourceVF)
		if err != nil {
			fmt.Printf("Error checking VF resource availability: %v\n", err)
		}

		if !pfAvailable || !vfAvailable {
			setupLog.Info("SR-IOV device plugin configuration applied successfully...")
			err := applySriovDevicePluginConfig(ctx, kubeClient, namespace)
			setupLog.Info("SR-IOV device plugin configuration applied successfully...")
			time.Sleep(2 * time.Second)
			if err != nil {
				fmt.Printf("Error applying SR-IOV device plugin configuration: %v\n", err)
			} else {
				fmt.Println("SR-IOV device plugin configuration applied successfully")
			}
		} else {
			fmt.Println("Resources are available, no action needed")
		}

		time.Sleep(2 * time.Minute)
	}
}

func isResourceAvailable(ctx context.Context, kubeClient client.Client, nodeName string, resourceName string) (bool, error) {
	node := &corev1.Node{}
	err := kubeClient.Get(ctx, types.NamespacedName{Name: nodeName}, node)
	if err != nil {
		return false, fmt.Errorf("failed to get node %s: %v", nodeName, err)
	}

	allocatable, ok := node.Status.Allocatable[corev1.ResourceName(resourceName)]
	if !ok {
		return false, fmt.Errorf("resource %s not found on node %s", resourceName, nodeName)
	}

	if allocatable.Value() > 0 {
		return true, nil
	}

	return false, nil
}

//12345

//cleanup

// Cleanup function to remove all resources related to the operator
func (r *SriovT2CardReconciler) cleanupResources(ctx context.Context, namespace string) error {
	setupLog := log.FromContext(ctx)
	setupLog.Info("Cleaning up resources")
	fmt.Println("Cleaning up resources Started...")

	// Set VFs amount to 0
	cr := &sriovfect2v1.SriovT2Card{}
	if err := resetSriovVfs(ctx, r.Client, cr); err != nil {
		setupLog.Error(err, "Failed to reset SR-IOV VFs")
		return err
	}
	setupLog.Info("Successfully reset SR-IOV VFs to 0")
	time.Sleep(3 * time.Second)
	// Delete all DaemonSets created by the operator
	daemonSetList := &appsv1.DaemonSetList{}
	if err := r.Client.List(ctx, daemonSetList, client.InNamespace(namespace)); err != nil {
		return err
	}

	for _, ds := range daemonSetList.Items {
		if err := r.Client.Delete(ctx, &ds); err != nil {
			setupLog.Error(err, "Failed to delete DaemonSet", "DaemonSet", ds.Name)
			return err
		}
		setupLog.Info("Deleted DaemonSet", "DaemonSet", ds.Name)
	}

	// Delete the Secret
	// secret := &corev1.Secret{}
	// err := r.Client.Get(ctx, client.ObjectKey{
	// 	Name:      "t2-card-token",
	// 	Namespace: namespace,
	// }, secret)

	// if err != nil {
	// 	if errors.IsNotFound(err) {
	// 		setupLog.Info("Secret not found, might have been already deleted", "Secret", "t2-card-token")
	// 	} else {
	// 		setupLog.Error(err, "Failed to get Secret before deletion", "Secret", "t2-card-token")
	// 		return err
	// 	}
	// } else {
	// 	setupLog.Info("Secret found, proceeding to delete", "Secret", "t2-card-token")
	// 	if err := r.Client.Delete(ctx, secret); err != nil && !errors.IsNotFound(err) {
	// 		setupLog.Error(err, "Failed to delete Secret during cleanup")
	// 		return err
	// 	}
	// 	setupLog.Info("Secret successfully deleted", "Secret", "t2-card-token")
	// }

	setupLog.Info("Successfully cleaned up all resources")

	return nil
}

// Function to reset SR-IOV VFs to 0
func resetSriovVfs(ctx context.Context, client client.Client, cr *sriovfect2v1.SriovT2Card) error {
	fmt.Println("inside vfs function")
	// Generate the DaemonSet for resetting SR-IOV VFs
	resetDs := generateResetDaemonSet()
	// resetDs := generateResetDaemonSet(cr)
	// Create the DaemonSet
	if err := client.Create(ctx, resetDs); err != nil {
		// return fmt.Errorf("failed to create reset DaemonSet: %v", err)
		// fmt.Println("11 error")
	}
	// Wait for DaemonSet to complete its job
	time.Sleep(5 * time.Second)
	pciAddress = ""
	dockerImage = ""
	myNameSpace = ""
	time.Sleep(5 * time.Second)
	// Delete the DaemonSet
	if err := client.Delete(ctx, resetDs); err != nil {
		// return fmt.Errorf("failed to delete reset DaemonSet: %v", err)
		// fmt.Println("22 error")
	}

	return nil
}

// Function to generate the DaemonSet for resetting SR-IOV VFs
// Function to generate the DaemonSet for resetting SR-IOV VFs
// Function to generate the DaemonSet for resetting SR-IOV VFs with hard-coded values
func generateResetDaemonSet() *appsv1.DaemonSet {
	fmt.Println("Inside the Reset Vfs")

	// Hard-coded values
	pciDevicesPath := "/sys/bus/pci/devices/" + pciAddress
	systemNodePath := "/sys/devices/system/node"
	lib := "/lib/modules"
	headers := "/usr/src"
	driverName := "vfio-pci"
	// pciAddress := "0000:01:00.0"

	if driverName == "" || pciAddress == "" {
		fmt.Println("driverName or pciAddress is empty")
		return nil
	}

	// resetCmd := fmt.Sprintf(`echo %s > /sys/bus/pci/drivers/%s/unbind && modprobe -r %s && echo 0 > /sys/bus/pci/devices/%s/sriov_numvfs`,
	// 	pciAddress, driverName, driverName, pciAddress)

	resetCmd := fmt.Sprintf(`echo %s > /sys/bus/pci/drivers/%s/unbind && echo 0 > /sys/bus/pci/devices/%s/sriov_numvfs`,
		pciAddress, driverName, pciAddress)

	setupLog.Info("Constructed resetCmd command", "cmd", resetCmd)

	// Hard-coded container details
	container := corev1.Container{
		Name:  "reset-yjb",
		Image: dockerImage, // Replace with your container image
		Command: []string{
			"sh", "-c", resetCmd,
		},
		SecurityContext: &corev1.SecurityContext{
			Privileged: boolPtr(true),
		},
		Env: []corev1.EnvVar{
			{
				Name:  "pci-devices",
				Value: pciDevicesPath,
			},
			{
				Name:  "system-node",
				Value: systemNodePath,
			},
			{
				Name:  "lib",
				Value: lib,
			},
			{
				Name:  "headers",
				Value: headers,
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "lib",
				MountPath: "/lib/modules",
			},
			{
				Name:      "headers",
				MountPath: "/usr/src",
			},
		},
		// Resources: corev1.ResourceRequirements{
		// 	Limits: corev1.ResourceList{
		// 		corev1.ResourceCPU:    resource.MustParse("2"),
		// 		corev1.ResourceMemory: resource.MustParse("1Gi"),
		// 	},
		// 	Requests: corev1.ResourceList{
		// 		corev1.ResourceCPU:    resource.MustParse("2"),
		// 		corev1.ResourceMemory: resource.MustParse("2Gi"),
		// 	},
		// },
	}

	ds := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "reset-sriov-vfs",
			Namespace: myNameSpace, // Replace with your namespace
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":  "dpdk",
					"card": "SriovT2Card",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":  "dpdk",
						"card": "SriovT2Card",
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: "t2-operator-controller-manager",
					Containers: []corev1.Container{
						container,
					},
					ImagePullSecrets: []corev1.LocalObjectReference{
						{
							Name: "regcred",
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "pci-devices",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: pciDevicesPath,
								},
							},
						},
						{
							Name: "system-node",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: systemNodePath,
								},
							},
						},
						{
							Name: "lib",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/lib/modules",
								},
							},
						},
						{
							Name: "headers",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/usr/src",
								},
							},
						},
					},
				},
			},
		},
	}
	fmt.Println("Done vfs function from ds")
	return ds
}

// Function to generate the DaemonSet for resetting SR-IOV VFs with hard-coded values
// func generateResetDaemonSet(cr *sriovfect2v1.SriovT2Card) *appsv1.DaemonSet {
// 	fmt.Println("Inside the Reset Vfs")

// 	// Hard-coded values
// 	pciDevicesPath := "/sys/bus/pci/devices/" + pciAddress
// 	systemNodePath := "/sys/devices/system/node"
// 	lib := "/lib/modules"
// 	headers := "/usr/src"
// 	driverName := "vfio-pci"
// 	// pciAddress := "0000:01:00.0"

// 	if driverName == "" || pciAddress == "" {
// 		fmt.Println("driverName or pciAddress is empty")
// 		return nil
// 	}

// 	resetCmd := fmt.Sprintf(`echo %s > /sys/bus/pci/drivers/%s/unbind && modprobe -r %s && echo 0 > /sys/bus/pci/devices/%s/sriov_numvfs`,
// 		pciAddress, driverName, driverName, pciAddress)

// 	container := cr.Spec.Template.Spec.Containers[0]

// 	ds := &appsv1.DaemonSet{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      cr.Name + "-reset",
// 			Namespace: cr.Namespace,
// 		},
// 		Spec: appsv1.DaemonSetSpec{
// 			Selector: &metav1.LabelSelector{
// 				MatchLabels: map[string]string{
// 					"app":  "dpdk",
// 					"card": "SriovT2Card",
// 				},
// 			},
// 			Template: corev1.PodTemplateSpec{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Labels: map[string]string{
// 						"app":  "dpdk",
// 						"card": "SriovT2Card",
// 					},
// 				},
// 				Spec: corev1.PodSpec{
// 					// Add the serviceAccountName field
// 					ServiceAccountName: "t2-operator-controller-manager",
// 					Containers: []corev1.Container{
// 						{
// 							Name:  "reset-yjb",
// 							Image: container.Image,
// 							Command: []string{
// 								"sh", "-c", resetCmd,
// 							},
// 							SecurityContext: &corev1.SecurityContext{
// 								Privileged: boolPtr(true),
// 								// allowPrivilegeEscalation: false,
// 							},
// 							Env: []corev1.EnvVar{
// 								{
// 									Name:  "pci-devices",
// 									Value: pciDevicesPath,
// 								},
// 								{
// 									Name:  "system-node",
// 									Value: systemNodePath,
// 								},
// 								{
// 									Name:  "lib",
// 									Value: lib,
// 								},
// 								{
// 									Name:  "headers",
// 									Value: headers,
// 								},
// 							},
// 							VolumeMounts: []corev1.VolumeMount{
// 								{
// 									Name:      "lib",
// 									MountPath: "/lib/modules",
// 								},
// 								{
// 									Name:      "headers",
// 									MountPath: "/usr/src",
// 								},
// 							},
// 							Resources: corev1.ResourceRequirements{
// 								Limits: corev1.ResourceList{
// 									// corev1.ResourceName("hugepages-2Mi"): resource.MustParse(container.Resources.Limits.Hugepages2Mi),
// 									// corev1.ResourceName("hugepages-1Gi"): resource.MustParse(container.Resources.Limits.Hugepages1Gi),
// 									// corev1.ResourceName("amd.com/amd_xilinx_t2"): resource.MustParse(strconv.Itoa(container.Resources.Limits.AMDXilinxT2)),
// 									corev1.ResourceCPU:    resource.MustParse(container.Resources.Limits.CPU),
// 									corev1.ResourceMemory: resource.MustParse(container.Resources.Limits.Memory),
// 								},
// 								Requests: corev1.ResourceList{
// 									corev1.ResourceCPU:    resource.MustParse(container.Resources.Requests.CPU),
// 									corev1.ResourceMemory: resource.MustParse(container.Resources.Requests.Memory),
// 								},
// 							},
// 						},
// 					},
// 					ImagePullSecrets: []corev1.LocalObjectReference{
// 						{
// 							Name: "regcred",
// 						},
// 					},
// 					Volumes: []corev1.Volume{
// 						{
// 							Name: "pci-devices",
// 							VolumeSource: corev1.VolumeSource{
// 								HostPath: &corev1.HostPathVolumeSource{
// 									Path: pciDevicesPath,
// 								},
// 							},
// 						},
// 						{
// 							Name: "system-node",
// 							VolumeSource: corev1.VolumeSource{
// 								HostPath: &corev1.HostPathVolumeSource{
// 									Path: systemNodePath,
// 								},
// 							},
// 						},
// 						{
// 							Name: "lib",
// 							VolumeSource: corev1.VolumeSource{
// 								HostPath: &corev1.HostPathVolumeSource{
// 									Path: "/lib/modules",
// 								},
// 							},
// 						},
// 						{
// 							Name: "headers",
// 							VolumeSource: corev1.VolumeSource{
// 								HostPath: &corev1.HostPathVolumeSource{
// 									Path: "/usr/src",
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 		},
// 	}

// 	fmt.Println("Done vfs function from ds")
// 	return ds
// }

//cleanup end

// Function to execute a command inside a pod and return the output
func execCommand(ctx context.Context, client interface{}, command string) (string, error) {
	fmt.Println("Inside The execCommand")
	fmt.Println("cmd: ", command)

	// cmd := exec.Command("sh", "-c", command)
	cmd := exec.Command(command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error executing command: %v", err)
	}

	// Trim any leading or trailing spaces from the output
	outputString := strings.TrimSpace(string(output))
	fmt.Println("Output of command:", outputString)
	return outputString, nil
}

// func (r *SriovT2CardReconciler) ensureSCC(ctx context.Context, namespace string) error {
// 	fmt.Println("Inside SCC...")
// 	// Define the SCC
// 	scc := &securityv1.SecurityContextConstraints{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name: "t2-operator-custom-scc",
// 		},
// 		AllowPrivilegedContainer: true,
// 		AllowHostDirVolumePlugin: true,
// 		AllowHostNetwork:         true,
// 		AllowHostPID:             true,
// 		AllowHostIPC:             true,
// 		AllowPrivilegeEscalation: pointer.BoolPtr(true),
// 		AllowedCapabilities:      []corev1.Capability{"*"},
// 		FSGroup:                  securityv1.FSGroupStrategyOptions{Type: securityv1.FSGroupStrategyRunAsAny},
// 		RunAsUser:                securityv1.RunAsUserStrategyOptions{Type: securityv1.RunAsUserStrategyRunAsAny},
// 		SELinuxContext:           securityv1.SELinuxContextStrategyOptions{Type: securityv1.SELinuxStrategyRunAsAny},
// 		SeccompProfiles:          []string{"*"},
// 		SupplementalGroups:       securityv1.SupplementalGroupsStrategyOptions{Type: securityv1.SupplementalGroupsStrategyRunAsAny},
// 		Volumes:                  []securityv1.FSType{securityv1.FSTypeAll},
// 		Users:                    []string{fmt.Sprintf("system:serviceaccount:%s:t2-operator-sa", namespace)},
// 	}

// 	// Create or update the SCC
// 	if err := r.Client.Create(ctx, scc); err != nil && !errors.IsAlreadyExists(err) {
// 		return err
// 	} else if errors.IsAlreadyExists(err) {
// 		if err := r.Client.Update(ctx, scc); err != nil {
// 			return err
// 		}
// 	}

// 	// Create RoleBinding to use the SCC
// 	roleBinding := &rbacv1.RoleBinding{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      "use-scc-rolebinding",
// 			Namespace: namespace,
// 		},
// 		Subjects: []rbacv1.Subject{
// 			{
// 				Kind:      "ServiceAccount",
// 				Name:      "t2-operator-sa",
// 				Namespace: namespace,
// 			},
// 		},
// 		RoleRef: rbacv1.RoleRef{
// 			APIGroup: "rbac.authorization.k8s.io",
// 			Kind:     "Role",
// 			Name:     "use-scc-role",
// 		},
// 	}

// 	if err := r.Client.Create(ctx, roleBinding); err != nil && !errors.IsAlreadyExists(err) {
// 		return err
// 	} else if errors.IsAlreadyExists(err) {
// 		if err := r.Client.Update(ctx, roleBinding); err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

func generateSetupDaemonSet(cr *sriovfect2v1.SriovT2Card, c client.Client, pfAddresses string) *appsv1.DaemonSet {

	fmt.Println("Inside The Setup Daemonset...")
	pciDevicesPath := "/sys/bus/pci/devices/" + pfAddresses
	systemNodePath := "/sys/devices/system/node"
	lib := "/lib/modules"
	headers := "/usr/src"

	driverName := cr.Spec.PhysicalFunction.PFDriver
	vfAmount := cr.Spec.PhysicalFunction.VFAmount
	pciAddress := cr.Spec.AcceleratorSelector.PciAddress

	setupLog.Info("Driver name", "driverName", driverName)
	setupLog.Info("VF amount", "vfAmount", vfAmount)
	setupLog.Info("PCI address", "pciAddress", pciAddress)

	// dynamicCmd := fmt.Sprintf(`modprobe -r %s && modprobe %s && echo 1 | tee /sys/module/%s/parameters/enable_sriov && dpdk-stable/usertools/dpdk-devbind.py -b vfio-pci %s && echo %d | tee /sys/bus/pci/devices/%s/sriov_numvfs && sleep infinity`,
	// 	driverName, driverName, driverName, pciAddress, vfAmount, pciAddress)

	// dynamicCmd1 := fmt.Sprintf(`modprobe %s && echo 1 | tee /sys/module/%s/parameters/enable_sriov && dpdk-stable/usertools/dpdk-devbind.py -b vfio-pci %s && echo %d | tee /sys/bus/pci/devices/%s/sriov_numvfs && sleep infinity`,
	// 	driverName, driverName, pciAddress, vfAmount, pciAddress)

	// setupLog.Info("Constructed dynamic command", "cmd", dynamicCmd)

	//rst
	// dynamicCmd := fmt.Sprintf(`modprobe %s && echo 1 | tee /sys/module/%s/parameters/enable_sriov && dpdk-stable/usertools/dpdk-devbind.py -b vfio-pci %s && echo 0 | tee /sys/bus/pci/devices/%s/sriov_numvfs && echo %d | tee /sys/bus/pci/devices/%s/sriov_numvfs && echo "" | sudo tee /sys/class/vfio-dev/vfio0/device/reset_method && sleep infinity`,
	// 	driverName, driverName, pciAddress, pciAddress, vfAmount, pciAddress)

	dynamicCmd := fmt.Sprintf(`modprobe %s && echo 1 | tee /sys/module/%s/parameters/enable_sriov && dpdk-stable/usertools/dpdk-devbind.py -b vfio-pci %s && echo 0 | tee /sys/bus/pci/devices/%s/sriov_numvfs && echo %d | tee /sys/bus/pci/devices/%s/sriov_numvfs && sleep infinity`,
		driverName, driverName, pciAddress, pciAddress, vfAmount, pciAddress)
	setupLog.Info("Constructed dynamic command", "cmd", dynamicCmd)

	container := cr.Spec.Template.Spec.Containers[0]

	ds := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-setup",
			Namespace: cr.Namespace,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":  "dpdk",
					"card": "SriovT2Card",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":  "dpdk",
						"card": "SriovT2Card",
					},
				},
				Spec: corev1.PodSpec{
					// Add the serviceAccountName field
					ServiceAccountName: "t2-operator-controller-manager",
					Containers: []corev1.Container{
						{
							Name:  "setup-container",
							Image: container.Image,
							Command: []string{
								"sh", "-c", dynamicCmd,
							},
							SecurityContext: &corev1.SecurityContext{
								Privileged: boolPtr(true),
								// allowPrivilegeEscalation: false,
							},
							Env: []corev1.EnvVar{
								{
									Name:  "pci-devices",
									Value: pciDevicesPath,
								},
								{
									Name:  "system-node",
									Value: systemNodePath,
								},
								{
									Name:  "lib",
									Value: lib,
								},
								{
									Name:  "headers",
									Value: headers,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "lib",
									MountPath: "/lib/modules",
								},
								{
									Name:      "headers",
									MountPath: "/usr/src",
								},
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceName("hugepages-2Mi"): resource.MustParse(container.Resources.Limits.Hugepages2Mi),
									corev1.ResourceName("hugepages-1Gi"): resource.MustParse(container.Resources.Limits.Hugepages1Gi),
									// corev1.ResourceName("amd.com/amd_xilinx_t2"): resource.MustParse(strconv.Itoa(container.Resources.Limits.AMDXilinxT2)),
									corev1.ResourceCPU:    resource.MustParse(container.Resources.Limits.CPU),
									corev1.ResourceMemory: resource.MustParse(container.Resources.Limits.Memory),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse(container.Resources.Requests.CPU),
									corev1.ResourceMemory: resource.MustParse(container.Resources.Requests.Memory),
								},
							},
						},
					},
					ImagePullSecrets: []corev1.LocalObjectReference{
						{
							Name: "regcred",
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "pci-devices",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: pciDevicesPath,
								},
							},
						},
						{
							Name: "system-node",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: systemNodePath,
								},
							},
						},
						{
							Name: "lib",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/lib/modules",
								},
							},
						},
						{
							Name: "headers",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/usr/src",
								},
							},
						},
					},
				},
			},
		},
	}
	time.Sleep(2 * time.Second)
	return ds
}

func generateAdminDaemonSet(cr *sriovfect2v1.SriovT2Card, c client.Client, pfAddresses string, uuidTokenGlobal string) *appsv1.DaemonSet {
	fmt.Println("Inside The Admin Daemonset...")
	pciDevicesPath := "/sys/bus/pci/devices/" + pfAddresses
	systemNodePath := "/sys/devices/system/node"
	lib := "/lib/modules"
	headers := "/usr/src"

	fmt.Println("Dynamic Token: " + uuidTokenGlobal)
	fmt.Println("Admin Mode: ", adminMode)
	// uuidTokenGlobal = "5a7c1e16-3a28-43a4-aedb-d2d581c243e1"

	var adminCmd string
	if adminMode {
		adminCmd = fmt.Sprintf("~/dpdk-stable/build/app/dpdk-admin -a %s --vfio-vf-token=%s --file-prefix PF 2>&1", pfAddresses, uuidTokenGlobal)
		fmt.Println("\033[33mPfAddress:\033[0m", pfAddresses)
		fmt.Println("\033[33mToken:\033[0m", uuidTokenGlobal)
	} else {
		adminCmd = "sleep 3600"
	}

	setupLog.Info("Admin Cmd", "adminCmd", adminCmd)
	fmt.Println("Final adminCmd: ", adminCmd)

	container := cr.Spec.Template.Spec.Containers[0]

	ds := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-admin",
			Namespace: cr.Namespace,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":  "dpdk",
					"card": "SriovT2Card",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":  "dpdk",
						"card": "SriovT2Card",
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: "t2-operator-controller-manager",
					// Add securityContext with fsGroup
					SecurityContext: &corev1.PodSecurityContext{
						FSGroup: int64Ptr(1001),
					},
					Containers: []corev1.Container{
						{
							Name:  "admin-container",
							Image: container.Image,
							Command: []string{
								"sh", "-c", adminCmd,
							},
							SecurityContext: &corev1.SecurityContext{
								Privileged:             boolPtr(false),
								RunAsNonRoot:           boolPtr(true),
								ReadOnlyRootFilesystem: boolPtr(true),
								RunAsUser:              int64Ptr(1001),
								Capabilities: &corev1.Capabilities{
									Add: []corev1.Capability{
										"SYS_ADMIN",
										"DAC_READ_SEARCH",
										"SYS_NICE",
										"IPC_LOCK",
										"SYS_RESOURCE",
										"ALL",
									},
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "pci-devices",
									Value: pciDevicesPath,
								},
								{
									Name:  "system-node",
									Value: systemNodePath,
								},
								{
									Name:  "lib",
									Value: lib,
								},
								{
									Name:  "headers",
									Value: headers,
								},
								{
									Name: "T2_CARD_TOKEN",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "t2-card-token",
											},
											Key: "T2_CARD_TOKEN",
										},
									},
								},
								{
									Name:  "T2_CARD_TOKEN_POD", // Set the global token as an environment variable
									Value: uuidTokenGlobal,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "lib",
									MountPath: "/lib/modules",
								},
								{
									Name:      "headers",
									MountPath: "/usr/src",
								},
								// {
								// 	Name:      "hugepages",
								// 	MountPath: "/dev/hugepages",
								// },
								{
									Name:      "hugepage-2mi",
									MountPath: "/home/nonroot/hugepages-2Mi",
									ReadOnly:  false,
								},
								{
									Name:      "hugepage-1gi",
									MountPath: "/home/nonroot/hugepages-1Gi",
									ReadOnly:  false,
								},
								{
									Name:      "varrun",
									MountPath: "/tmp/dpdk",
									ReadOnly:  false,
								},
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceName("hugepages-2Mi"):            resource.MustParse(container.Resources.Limits.Hugepages2Mi),
									corev1.ResourceName("hugepages-1Gi"):            resource.MustParse(container.Resources.Limits.Hugepages1Gi),
									corev1.ResourceCPU:                              resource.MustParse(container.Resources.Limits.CPU),
									corev1.ResourceMemory:                           resource.MustParse(container.Resources.Limits.Memory),
									corev1.ResourceName("amd.com/amd_xilinx_t2_pf"): resource.MustParse("1"),
									// corev1.ResourceName("amd.com/amd_xilinx_t2_vf"): resource.MustParse(strconv.Itoa(container.Resources.Limits.AMDXilinxT2)),
									// "amd.com/amd_xilinx_t2_pf":           resource.MustParse("1"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:                              resource.MustParse(container.Resources.Requests.CPU),
									corev1.ResourceMemory:                           resource.MustParse(container.Resources.Requests.Memory),
									corev1.ResourceName("amd.com/amd_xilinx_t2_pf"): resource.MustParse("1"),
									// corev1.ResourceName("amd.com/amd_xilinx_t2_vf"): resource.MustParse(strconv.Itoa(container.Resources.Limits.AMDXilinxT2)),
								},
							},
						},
					},
					ImagePullSecrets: []corev1.LocalObjectReference{
						{
							Name: "regcred",
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "pci-devices",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: pciDevicesPath,
								},
							},
						},
						{
							Name: "system-node",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: systemNodePath,
								},
							},
						},
						{
							Name: "lib",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/lib/modules",
								},
							},
						},
						{
							Name: "headers",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/usr/src",
								},
							},
						},
						// {
						// 	Name: "hugepages",
						// 	VolumeSource: corev1.VolumeSource{
						// 		HostPath: &corev1.HostPathVolumeSource{
						// 			Path: "/dev/hugepages",
						// 		},
						// 	},
						// },
						{
							Name: "hugepage-2mi",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{
									Medium: "HugePages-2Mi",
								},
							},
						},
						{
							Name: "hugepage-1gi",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{
									Medium: "HugePages-1Gi",
								},
							},
						},
						{
							Name: "varrun",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}
	return ds
}

func generateDebugDaemonSet(cr *sriovfect2v1.SriovT2Card, c client.Client, pfAddresses string) *appsv1.DaemonSet {

	fmt.Println("Inside The Debug Daemonset...")
	pciDevicesPath := "/sys/bus/pci/devices/" + pfAddresses
	systemNodePath := "/sys/devices/system/node"
	lib := "/lib/modules"
	headers := "/usr/src"

	driverName := cr.Spec.PhysicalFunction.PFDriver
	vfAmount := cr.Spec.PhysicalFunction.VFAmount
	pciAddress := cr.Spec.AcceleratorSelector.PciAddress

	setupLog.Info("Driver name", "driverName", driverName)
	setupLog.Info("VF amount", "vfAmount", vfAmount)
	setupLog.Info("PCI address", "pciAddress", pciAddress)

	container := cr.Spec.Template.Spec.Containers[0]

	dsDebug := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-debug",
			Namespace: cr.Namespace,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":  "dpdk-debug",
					"card": "SriovT2Card-debug",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":  "dpdk-debug",
						"card": "SriovT2Card-debug",
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: "t2-operator-controller-manager",
					Containers: []corev1.Container{
						{
							Name:  "debug-container",
							Image: container.Image,
							Command: []string{
								"sh", "-c", "sleep infinity",
							},
							// SecurityContext: &corev1.SecurityContext{
							// 	Privileged: boolPtr(true),
							// },
							Env: []corev1.EnvVar{
								{
									Name:  "pci-devices",
									Value: pciDevicesPath,
								},
								{
									Name:  "system-node",
									Value: systemNodePath,
								},
								{
									Name:  "lib",
									Value: lib,
								},
								{
									Name:  "headers",
									Value: headers,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "lib",
									MountPath: "/lib/modules",
								},
								{
									Name:      "headers",
									MountPath: "/usr/src",
								},
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceName("hugepages-2Mi"): resource.MustParse(container.Resources.Limits.Hugepages2Mi),
									corev1.ResourceName("hugepages-1Gi"): resource.MustParse(container.Resources.Limits.Hugepages1Gi),
									corev1.ResourceCPU:                   resource.MustParse(container.Resources.Limits.CPU),
									corev1.ResourceMemory:                resource.MustParse(container.Resources.Limits.Memory),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse(container.Resources.Requests.CPU),
									corev1.ResourceMemory: resource.MustParse(container.Resources.Requests.Memory),
								},
							},
						},
					},
					ImagePullSecrets: []corev1.LocalObjectReference{
						{
							Name: "regcred",
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "pci-devices",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: pciDevicesPath,
								},
							},
						},
						{
							Name: "system-node",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: systemNodePath,
								},
							},
						},
						{
							Name: "lib",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/lib/modules",
								},
							},
						},
						{
							Name: "headers",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/usr/src",
								},
							},
						},
					},
				},
			},
		},
	}
	return dsDebug
}

func applySriovDevicePluginConfig(ctx context.Context, c client.Client, namespace string) error {
	fmt.Println("Inside The Device Plugin Daemonset...10")

	// Create ConfigMap for SR-IOV device plugin
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "t2-operator-sriovdp-config",
			Namespace: namespace,
		},
		Data: map[string]string{
			"config.json": `
			{
				"resourceList": [
					{
						"resourceName": "amd_xilinx_t2_pf",
						"resourcePrefix": "amd.com",
						"deviceType": "accelerator",
						"selectors": {
							"vendors": ["10ee"],
							"devices": ["9048"],
							"drivers": ["vfio-pci"]
						}
					},
					{
						"resourceName": "amd_xilinx_t2_vf",
						"resourcePrefix": "amd.com",
						"deviceType": "accelerator",
						"selectors": {
							"vendors": ["10ee"],
							"devices": ["a048"],
							"drivers": ["vfio-pci"]
						}
					}
				]
			}`,
		},
	}

	// Check if the ConfigMap exists
	existingConfigMap := &corev1.ConfigMap{}
	err := c.Get(ctx, client.ObjectKey{Name: configMap.Name, Namespace: namespace}, existingConfigMap)
	if err == nil {
		// ConfigMap exists, delete it
		if err := c.Delete(ctx, existingConfigMap); err != nil {
			return err
		}
		fmt.Println("Existing ConfigMap deleted.")
	} else if !errors.IsNotFound(err) {
		// Error other than "not found"
		// return err
	}

	// Create the ConfigMap
	if err := c.Create(ctx, configMap); err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	// List nodes and get their labels
	nodeList := &corev1.NodeList{}
	if err := c.List(ctx, nodeList); err != nil {
		return err
	}

	nodeSelector := map[string]string{}
	for _, node := range nodeList.Items {
		// Example: Get label "kubernetes.io/arch" from the first node
		if arch, exists := node.Labels["kubernetes.io/arch"]; exists {
			nodeSelector["kubernetes.io/arch"] = arch
			break
		}
	}
	time.Sleep(5 * time.Second)

	// Create DaemonSet for SR-IOV device plugin
	ds := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sriovt2card-device-plugin",
			Namespace: namespace,
			Labels: map[string]string{
				"tier": "node",
				"app":  "sriovdp",
			},
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"name": "sriov-device-plugin",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"name": "sriov-device-plugin",
						"tier": "node",
						"app":  "sriovdp",
					},
				},
				Spec: corev1.PodSpec{
					HostNetwork: true,
					HostPID:     true,
					// NodeSelector: map[string]string{
					// 	"kubernetes.io/arch": "amd64", // Updated to use the non-deprecated label
					// },
					NodeSelector: nodeSelector,
					Tolerations: []corev1.Toleration{
						{
							Key:      "node-role.kubernetes.io/master",
							Operator: corev1.TolerationOpExists,
							Effect:   corev1.TaintEffectNoSchedule,
						},
					},
					// ServiceAccountName: "sriov-device-plugin",
					ServiceAccountName: "t2-operator-controller-manager",
					Containers: []corev1.Container{
						{
							Name:            "kube-sriovdp",
							Image:           "dhiraj30/device-plugin:ft",
							ImagePullPolicy: corev1.PullAlways,
							Args: []string{
								"--log-dir=sriovdp",
								"--log-level=10",
							},
							SecurityContext: &corev1.SecurityContext{
								Privileged: boolPtr(true),
								// allowPrivilegeEscalation: true,
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "devicesock",
									MountPath: "/var/lib/kubelet/",
									ReadOnly:  false,
								},
								{
									Name:      "log",
									MountPath: "/var/log",
								},
								{
									Name:      "config-volume",
									MountPath: "/etc/pcidp/",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "devicesock",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/var/lib/kubelet/",
								},
							},
						},
						{
							Name: "log",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/var/log",
								},
							},
						},
						{
							Name: "config-volume",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "t2-operator-sriovdp-config",
									},
									Items: []corev1.KeyToPath{
										{
											Key:  "config.json",
											Path: "config.json",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	if err := c.Create(ctx, ds); err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	return nil
}

// Function to ensure SCC (Security Context Constraints)
// func ensureSCC(ctx context.Context, c client.Client, namespace string) error {
// 	fmt.Println("Inside ensureSCC")
// 	// Create ServiceAccount
// 	serviceAccount := &corev1.ServiceAccount{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      "t2-operator-sa",
// 			Namespace: namespace,
// 		},
// 	}

// 	fmt.Println("1")
// 	if err := c.Create(ctx, serviceAccount); err != nil && !errors.IsAlreadyExists(err) {
// 		return err
// 	}
// 	fmt.Println("2")

// 	// Create SCC
// 	scc := &policyv1.PodSecurityPolicy{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name: "t2-operator-custom-scc",
// 		},
// 		// Directly set the values
// 		AllowHostDirVolumePlugin: true,
// 		AllowHostIPC:             true,
// 		AllowHostNetwork:         true,
// 		AllowHostPID:             true,
// 		AllowPrivilegeEscalation: true,
// 		AllowPrivilegedContainer: true,
// 		AllowedCapabilities:      []corev1.Capability{"*"},
// 		FSGroup:                  policyv1.FSGroupStrategyOptions{Type: policyv1.FSGroupStrategyRunAsAny},
// 		RunAsUser:                policyv1.RunAsUserStrategyOptions{Type: policyv1.RunAsUserStrategyRunAsAny},
// 		SELinux:                  &policyv1.SELinuxStrategyOptions{Type: policyv1.SELinuxStrategyRunAsAny},
// 		SeccompProfiles:          []policyv1.SeccompProfileType{"*"},
// 		SupplementalGroups:       policyv1.SupplementalGroupsStrategyOptions{Type: policyv1.SupplementalGroupsStrategyRunAsAny},
// 		Volumes:                  []policyv1.FSType{policyv1.FSTypeAll},
// 		Users:                    []string{"system:serviceaccount:" + namespace + ":t2-operator-sa"},
// 	}

// 	fmt.Println("3")
// 	if err := c.Create(ctx, scc); err != nil && !errors.IsAlreadyExists(err) {
// 		return err
// 	}
// 	fmt.Println("4")

//		return nil
//	}
func boolPtr(b bool) *bool {
	return &b
}

// Helper function to create a pointer to an int64 value
func int64Ptr(i int64) *int64 {
	return &i
}

// SetupWithManager sets up the controller with the Manager.
func (r *SriovT2CardReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&sriovfect2v1.SriovT2Card{}).
		Complete(r)
}
