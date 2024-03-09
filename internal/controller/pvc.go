package controller

import (
	"fmt"

	"context"

	mediumv1beta1 "example.com/example/api/v1beta1"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// wordpress pvc
func (r *WordpressReconciler) pvcwordpress(operator *mediumv1beta1.Wordpress, storageclass string) corev1.PersistentVolumeClaim {
	storageRequest, _ := resource.ParseQuantity(fmt.Sprintf("%sGi", operator.Spec.Wordpress.Storage))
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      operator.Name + "-wordpress",
			Namespace: operator.Namespace,
			Labels: map[string]string{
				"app": operator.Name,
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				"ReadWriteOnce",
			},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					"storage": storageRequest,
				},
			},
			StorageClassName: &storageclass,
		},
	}

	return *pvc
}

func (r *WordpressReconciler) reconcilewordpresspvc(ctx context.Context, operator *mediumv1beta1.Wordpress, l logr.Logger) error {
	wordpresspvc := &corev1.PersistentVolumeClaim{}
	pvcName := types.NamespacedName{
		Namespace: operator.Namespace,
		Name:      operator.Name + "-wordpress",
	}

	if err := r.Get(ctx, pvcName, wordpresspvc); err != nil {
		if errors.IsNotFound(err) {
			wordpresspvc := r.pvcwordpress(operator, operator.Spec.Storageclass)
			l.Info("creating  wordpress", "releaseName", operator.Name, "namespace", operator.Namespace)
			if err := r.Create(ctx, &wordpresspvc); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return nil

}

/**********************************************************************************************************/
// pvcwordpress creates a PersistentVolumeClaim (PVC) for WordPress.
// It takes the WordPress custom resource and the storage class as input.
func (r *WordpressReconciler) pvcmysql(operator *mediumv1beta1.Wordpress, storageclass string) corev1.PersistentVolumeClaim {
	// Parse storage quantity from the WordPress custom resource specification
	storageRequest, _ := resource.ParseQuantity(fmt.Sprintf("%sGi", operator.Spec.Mysql.Storage))

	// Define the PersistentVolumeClaim (PVC) object
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      operator.Name + "-mysql", // Name of the PVC
			Namespace: operator.Namespace,       // Namespace of the operator
			Labels: map[string]string{
				"app": operator.Name, // Label for identifying the application
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				"ReadWriteOnce", // Access mode for the PVC
			},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					"storage": storageRequest, // Requested storage size for the PVC
				},
			},
			StorageClassName: &storageclass, // Storage class for the PVC
		},
	}

	return *pvc // Return the PVC object
}

// reconcilewordpresspvc reconciles the PersistentVolumeClaim (PVC) for WordPress.
// It takes the context, WordPress custom resource, and logger as input.
func (r *WordpressReconciler) mysqlpvc(ctx context.Context, operator *mediumv1beta1.Wordpress, l logr.Logger) error {
	mysqlpvc := &corev1.PersistentVolumeClaim{}
	pvcName := types.NamespacedName{
		Namespace: operator.Namespace,
		Name:      operator.Name + "-mysql",
	}

	// Check if the PVC exists
	if err := r.Get(ctx, pvcName, mysqlpvc); err != nil {
		if errors.IsNotFound(err) { // PVC does not exist
			// Create a new PVC
			mysqlpvc := r.pvcmysql(operator, operator.Spec.Storageclass)
			l.Info("creating  mysql", "releaseName", operator.Name, "namespace", operator.Namespace)
			if err := r.Create(ctx, &mysqlpvc); err != nil {
				return err
			}
		} else { // Error occurred while retrieving PVC
			return err
		}
	}

	return nil // Return nil error
}
