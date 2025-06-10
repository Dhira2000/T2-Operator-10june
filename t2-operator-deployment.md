## SPDX-License-Identifier: Apache-2.0
## Copyright (c) 2022-2024 AMD Corporation

## Technical Requirements and Dependencies

The T2 Operator for Wireless FEC Accelerators has the following requirements:

- [AMD Xilinx T2 Accelerator](https://www.xilinx.com/content/dam/xilinx/publications/product-briefs/xilinx-t2-product-brief.pdf)
- [OpenShift 4.15.x](https://docs.openshift.com/container-platform/4.15/release_notes/ocp-4-15-release-notes.html)
- RT Kernel configured with [Performance Addon Operator](https://access.redhat.com/documentation/en-us/openshift_container_platform/4.6/html/scalability_and_performance/cnf-performance-addon-operator-for-low-latency-nodes).
- Kernel parameters:
    - vfio_pci.enable_sriov=1
    - vfio_pci.disable_idle_d3=1
- BIOS with enabled settings "AMD Virtualization Technology for Directed I/O" (VT-d), "Single Root I/O Virtualization" (SR-IOV), and "Input–Output Memory Management Unit" (IOMMU)
### Grub Setting
Open the GRUB configuration file for editing:
```shell
[user@ctrl1 /home]# sudo nano /etc/default/grub
```
Replace the following line and Save the File:
```shell
GRUB_CMDLINE_LINUX_DEFAULT="quiet splash default_hugepagesz=1GB hugepagesz=1GB hugepages=8 hugepagesz=2MB hugepages=2048 iommu=pt amd_iommu=on"
```
Update GRUB and Reboot The System:
```shell
[user@ctrl1 /home]# sudo update-grub
[user@ctrl1 /home]# sudo reboot
```

### Install the Bundle

To install the T2 Operator for Wireless FEC Accelerators operator bundle, perform the following steps:

Create the project:
```shell
[user@ctrl1 /home]# oc new-project amd-xilinx-t2
```
Run The Operator:
```shell
[user@ctrl1 /home]# operator-sdk run bundle -n amd-xilinx-t2 quay.io/amdaecgt2/amd-t2-bundle:4.2.6 --timeout 100m
```
Or
```shell
[user@ctrl1 /home]# ./operator_run.sh -n amd-xilinx-t2 quay.io/amdaecgt2/amd-t2-bundle:4.2.6
```

Verify the operator status:
```shell
[user@ctrl1 /home]# oc get operatorgroup -n amd-xilinx-t2
[user@ctrl1 /home]# oc get subscription -n amd-xilinx-t2
[user@ctrl1 /home]# oc get csv -n amd-xilinx-t2
[user@ctrl1 /home]# oc get pods -n amd-xilinx-t2
```

### Allocate Resources

Apply the necessary resource configurations:
```shell
[user@ctrl1 /home]# oc apply -f t2-operator-scc.yaml
[user@ctrl1 /home]# oc apply -f sriovfect2_v1_sriovt2card.yaml
```

### Save the Admin Logs to a File

Collect the logs:
```shell
[user@ctrl1 /home]# ./collect_logs.sh -n amd-xilinx-t2
```

### Testing

Run the following command inside the Admin pod:
```shell
[user@ctrl1 /home]# printenv
[user@ctrl1 /home]# echo ${PCIDEVICE_AMD_COM_AMD_XILINX_T2}
[user@ctrl1 /home]# ~/dpdk-stable/build/app/dpdk-test-bbdev -a ${PCIDEVICE_AMD_COM_AMD_XILINX_T2} --vfio-vf-token=${T2_CARD_TOKEN} -- -n 256 -l 4 -c throughput -v ~/ldpc_dec_LD550_K_prime_minus_L_2536_E_3840_BG_2_Q_m_2_SNR_40.00.data -b 256 -t 6
```

### Cleanup Resources

To delete the resource configurations:
```shell
[user@ctrl1 /home]# oc delete -f sriovfect2_v1_sriovt2card.yaml
OR
[user@ctrl1 /home]# oc delete SriovT2Card sriovt2card -n amd-xilinx-t2
```

### Operator Cleanup

Clean up the operator:
```shell
[user@ctrl1 /home]# operator-sdk cleanup t2-operator --delete-all -n amd-xilinx-t2
```

### Extra Commands

```shell
[user@ctrl1 /home]# oc rollout restart deployment t2-operator-controller-manager -n amd-xilinx-t2
[user@ctrl1 /home]# operator-sdk run bundle-upgrade -n amd-xilinx-t2 quay.io/amdaecgt2/amd-t2-bundle:4.2.7 --timeout 100m --install-mode OwnNamespace/AllNamespaces
```

### Sample Output from Operator-SDK Command for Reference

```shell
[user@ctrl1 /home]# operator-sdk run bundle -n test28 quay.io/amdaecgt2/amd-t2-bundle:4.2.6 --timeout 100m

[user@ctrl1 /home]# operator-sdk run bundle -n test28 quay.io/amdaecgt2/amd-t2-bundle:4.2.6 --timeout 100m --install-mode OwnNamespace
INFO[0031] Creating a File-Based Catalog of the bundle "quay.io/amdaecgt2/amd-t2-bundle:4.2.6"
INFO[0035] Generated a valid File-Based Catalog
INFO[0042] Created registry pod: quay-io-amdaecgt2-amd-t2-bundle-4-2-6
INFO[0043] Created CatalogSource: t2-operator-catalog
INFO[0043] OperatorGroup "operator-sdk-og" created
INFO[0043] Created Subscription: t2-operator-v4-2-6-sub
INFO[0073] Approved InstallPlan install-k4xrg for the Subscription: t2-operator-v4-2-6-sub
INFO[0073] Waiting for ClusterServiceVersion "test28/t2-operator.v4.2.6" to reach 'Succeeded' phase
INFO[0073]   Waiting for ClusterServiceVersion "test28/t2-operator.v4.2.6" to appear
INFO[0075]   Found ClusterServiceVersion "test28/t2-operator.v4.2.6" phase: Pending
INFO[0077]   Found ClusterServiceVersion "test28/t2-operator.v4.2.6" phase: Installing
INFO[0098]   Found ClusterServiceVersion "test28/t2-operator.v4.2.6" phase: Succeeded
INFO[0098] OLM has successfully installed "t2-operator.v4.2.6"
```

```shell
[user@ctrl1 /home]# oc get operatorgroup
NAME              AGE
operator-sdk-og   118s

[user@ctrl1 /home]# oc get subscription
NAME                     PACKAGE       SOURCE                CHANNEL
t2-operator-v4-2-6-sub   t2-operator   t2-operator-catalog   operator-sdk-run-bundle

[user@ctrl1 /home]# oc get csv
NAME                 DISPLAY           VERSION   REPLACES   PHASE
t2-operator.v4.2.6   AMD-T2 Operator   4.2.6                Succeeded

[user@ctrl1 /home]# oc get pods
NAME                                                              READY   STATUS      RESTARTS   AGE
dc4b3e367c7a2d5939f0c1a3c77d90d0d7170bdcc77b1e1f33c738e72blxw2v   0/1     Completed   0          2m30s
quay-io-amdaecgt2-amd-t2-bundle-4-2-6                             1/1     Running     0          2m48s
t2-operator-controller-manager-56848b549f-4bv67                   2/2     Running     0          2m10s
```

### Custom Resource Example

Here is an example of a custom resource YAML file (`sriovfect2_v1_sriovt2card.yaml`):

```yaml
apiVersion: sriovfect2.amd.com/v1
kind: SriovT2Card
metadata:
  labels:
    app.kubernetes.io/name: sriovt2card
    app.kubernetes.io/instance: sriovt2card-sample
    app.kubernetes.io/part-of: t2-operator
    app.kubernetes.io/managed-by: kustomize
    app.kubernetes.io/created-by: t2-operator
  name: sriovt2card
  namespace: system
spec:
  nodeSelector:
    kubernetes.io/hostname: xilinx-t2.amd.com
  acceleratorSelector:
    pciAddress: 0000:01:00.0
  physicalFunction:  
    pfDriver: "vfio_pci"
    vfDriver: "vfio_pci"
    vfAmount: 1
  selector:
    matchLabels:
      app: dpdk
  template:
    spec:
      imagePullSecrets:
        - name: regcred
      containers:
        - name: redhat-container
          image:  hitendramhatre/amd-xilinx-t2-heartbeat:v1
          securityContext:
            privileged: false
            capabilities:
              add: ["SYS_ADMIN", "SYS_RAWIO", "SYS_NICE", "NET_ADMIN", "IPC_LOCK", "ALL"]
              # add: ["IPC_LOCK", "SYS_NICE"]
          command: ["sleep", "infinity"]
          volumeMounts:
          - mountPath: /hugepages-2Mi
            name: hugepage-2mi
          - mountPath: /hugepages-1Gi
            name: hugepage-1gi
          resources:
            limits:
              hugepages-2Mi: 512Mi
              hugepages-1Gi: 2Gi
              memory: 10Gi
              cpu: "4"
              amd.com/amd_xilinx_t2: 1
            requests:
              hugepages-2Mi: 1Gi  
              hugepages-1Gi: 2Gi  
              memory: 10Gi
              cpu: "4"
              amd.com/amd_xilinx_t2: 1
      volumes:
      - name: hugepage-2mi
        emptyDir:
          medium: HugePages-2Mi
      - name: hugepage-1gi
        emptyDir:
          medium: HugePages-1Gi
```