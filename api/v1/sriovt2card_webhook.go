/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package v1

import (
	"errors"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var sriovt2cardlog = logf.Log.WithName("sriovt2card-resource")

func (r *SriovT2Card) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

var _ webhook.Validator = &SriovT2Card{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *SriovT2Card) ValidateCreate() error {
	fmt.Println("Hello")
	sriovt2cardlog.Info("validate create", "name", r.Name)

	// Additional validation logic
	if r.Spec.AcceleratorSelector.PciAddress == "0000:00:00.0" {
		return errors.New("invalid PCI address")
	}
	if r.Namespace == "" {
		return errors.New("namespace is required")
	}
	if r.Spec.PhysicalFunction.PFDriver != "vfio-pci" {
		return errors.New("pfDriver must be 'vfio-pci'")
	}
	if r.Spec.PhysicalFunction.VFAmount <= 0 {
		return errors.New("vfAmount must be greater than 0")
	}

	// Add more validation rules as needed

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *SriovT2Card) ValidateUpdate(old runtime.Object) error {
	sriovt2cardlog.Info("validate update", "name", r.Name)

	// Additional validation logic
	if r.Spec.AcceleratorSelector.PciAddress == "0000:00:00.0" {
		return errors.New("invalid PCI address")
	}
	if r.Namespace == "" {
		return errors.New("namespace is required")
	}
	if r.Spec.PhysicalFunction.PFDriver != "vfio-pci" {
		return errors.New("pfDriver must be 'vfio-pci'")
	}
	if r.Spec.PhysicalFunction.VFAmount <= 0 {
		return errors.New("vfAmount must be greater than 0")
	}

	// Add more validation rules as needed

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *SriovT2Card) ValidateDelete() error {
	sriovt2cardlog.Info("validate delete", "name", r.Name)

	// Add your validation logic here for object deletion

	return nil
}
