package controller

import (
	"context"
	"reflect"

	mediumv1beta1 "example.com/example/api/v1beta1"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (r *WordpressReconciler) deploymentswordpress(operator *mediumv1beta1.Wordpress) appsv1.Deployment {

	deployments := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      operator.Name + "-wordpress",
			Namespace: operator.Namespace,
			Labels: map[string]string{
				"app": operator.Name,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":  operator.Name,
					"tier": "frontend",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":  operator.Name,
						"tier": "frontend",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "wordpress",
							Image: operator.Spec.Wordpress.Image,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 80,
									Name:          "http",
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "WORDPRESS_DB_HOST",
									Value: "wordpress-mysql",
								},
								{
									Name: "WORDPRESS_DB_PASSWORD",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: operator.Name + "-mysql",
											},
											Key: "password",
										},
									},
								},
								{
									Name:  "WORDPRESS_DB_USER",
									Value: operator.Spec.Mysql.User,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "wordpress-persistent-storage",
									MountPath: "/var/www/html",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "wordpress-persistent-storage",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: operator.Name + "-wordpress",
								},
							},
						},
					},
				},
			},
		},
	}

	return *deployments
}
func (r *WordpressReconciler) reconciledployments(ctx context.Context, operator *mediumv1beta1.Wordpress, l logr.Logger) error {
	//checkif root configmap exists
	deployments := &appsv1.Deployment{}
	deploymentsName := types.NamespacedName{
		Namespace: operator.Namespace,
		Name:      operator.Name + "-wordpress",
	}
	if err := r.Get(ctx, deploymentsName, deployments); err != nil {
		if errors.IsNotFound(err) {

			deployments := r.deploymentswordpress(operator)
			l.Info("creating deployments", "releaseName", operator.Name, "namespace", operator.Namespace)
			if err := r.Create(ctx, &deployments); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		updateddeployments := r.deploymentswordpress(operator)
		if reflect.DeepEqual(deployments.Spec, updateddeployments.Spec) && reflect.DeepEqual(deployments.ObjectMeta.Name, updateddeployments.ObjectMeta.Name) {
			// If they are equal, log a message and return without updating

			return nil
		}
		deployments.Spec = updateddeployments.Spec
		if err := r.Update(ctx, deployments); err != nil {
			return err
		}

	}
	return nil
}

func (r *WordpressReconciler) mysql(operator *mediumv1beta1.Wordpress) appsv1.Deployment {
	deployments := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      operator.Name + "-mysql",
			Namespace: operator.Namespace,
			Labels: map[string]string{
				"app": operator.Name,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":  operator.Name,
					"tier": "mysql",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":  operator.Name,
						"tier": "mysql",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "mysql",
							Image: operator.Spec.Mysql.Image,
							Env: []corev1.EnvVar{
								{
									Name: "ROOT_PASSWORD",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: operator.Name + "-mysql",
											},
											Key: "password",
										},
									},
								},
								{
									Name:  "DB_USER",
									Value: operator.Spec.Mysql.User,
								},
							},

							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "mysql-persistent-storage",
									MountPath: "/var/lib/mysql",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "mysql-persistent-storage",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: operator.Name + "-mysql",
								},
							},
						},
					},
				},
			},
		},
	}
	return *deployments
}

func (r *WordpressReconciler) reconciledploymentsmysql(ctx context.Context, operator *mediumv1beta1.Wordpress, l logr.Logger) error {
	deployments := &appsv1.Deployment{}
	deploymentsName := types.NamespacedName{
		Namespace: operator.Namespace,
		Name:      operator.Name + "-mysql",
	}
	if err := r.Get(ctx, deploymentsName, deployments); err != nil {
		if errors.IsNotFound(err) {

			deployments := r.mysql(operator)
			l.Info("creating deployments", "releaseName", operator.Name, "namespace", operator.Namespace)
			if err := r.Create(ctx, &deployments); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		updateddeployments := r.mysql(operator)
		if reflect.DeepEqual(deployments.Spec, updateddeployments.Spec) && reflect.DeepEqual(deployments.ObjectMeta.Name, updateddeployments.ObjectMeta.Name) {
			// If they are equal, log a message and return without updating

			return nil
		}
		deployments.Spec = updateddeployments.Spec
		if err := r.Update(ctx, deployments); err != nil {
			return err
		}

	}
	return nil
}
