package controller

import (
	"context"
	"reflect"

	mediumv1beta1 "example.com/example/api/v1beta1"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (r *WordpressReconciler) wordpresssvc(operator *mediumv1beta1.Wordpress, Port int32, Type string) corev1.Service {
	if operator.Spec.Wordpress.Network.Type == "LoadBalancer" {
		Type = "LoadBalancer"
		if operator.Spec.Wordpress.Network.Port >= 10 && operator.Spec.Wordpress.Network.Port <= 300 {
			Port = operator.Spec.Wordpress.Network.Port
		} else {
			Port = 80
		}
	} else if operator.Spec.Wordpress.Network.Type == "NodePort" {
		Type = "NodePort"
		if operator.Spec.Wordpress.Network.Port >= 30000 && operator.Spec.Wordpress.Network.Port <= 32767 {
			Port = operator.Spec.Wordpress.Network.Port
		} else {
			Port = 32000
		}
	}
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      operator.Name + "-wordpress",
			Namespace: operator.Namespace,
			Labels: map[string]string{
				"app": operator.Name,
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       80,
					Protocol:   "TCP",
					TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: Port},
				},
			},
			Selector: map[string]string{
				"app":  operator.Name,
				"tier": "frontend",
			},
			Type: corev1.ServiceType(Type),
		},
	}
	return *service
}

func (r *WordpressReconciler) reconcilewordpresssvc(ctx context.Context, operator *mediumv1beta1.Wordpress, l logr.Logger) error {

	Service := &corev1.Service{}
	serviceName := types.NamespacedName{
		Namespace: operator.Namespace,
		Name:      operator.Name + "-wordpress",
	}
	if err := r.Get(ctx, serviceName, Service); err != nil {
		if errors.IsNotFound(err) {
			// ConfigMap not found. Create it.

			service := r.wordpresssvc(operator, operator.Spec.Wordpress.Network.Port, operator.Spec.Wordpress.Network.Type)
			l.Info("creating wordpress svc", "releaseName", operator.Name, "namespace", operator.Namespace)
			if err := r.Create(ctx, &service); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		updatedService := r.wordpresssvc(operator, operator.Spec.Wordpress.Network.Port, operator.Spec.Wordpress.Network.Type)
		if reflect.DeepEqual(Service.Spec, updatedService.Spec) || reflect.DeepEqual(Service.Name, updatedService.Name) {
			// If they are equal, log a message and return without updating

			return nil
		}
		Service.Spec = updatedService.Spec
		if err := r.Update(ctx, Service); err != nil {
			return err
		}

	}
	return nil
}

/*----------------------------------------------------------------------------------------------*/
func (r *WordpressReconciler) servicemysql(operator *mediumv1beta1.Wordpress) corev1.Service {

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      operator.Name + "-mysql",
			Namespace: operator.Namespace,
			Labels: map[string]string{
				"app": operator.Name,
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port:       3306,
					Protocol:   "TCP",
					TargetPort: intstr.FromInt(int(3306)),
				},
			},
			Selector: map[string]string{
				"app": operator.Name + "-mysql",
			},
			Type: corev1.ServiceType("ClusterIP"),
		},
	}

	return *service
}

func (r *WordpressReconciler) reconcileservicemysql(ctx context.Context, operator *mediumv1beta1.Wordpress, l logr.Logger) error {

	Service := &corev1.Service{}
	serviceName := types.NamespacedName{
		Namespace: operator.Namespace,
		Name:      operator.Name + "-mysql",
	}

	if err := r.Get(ctx, serviceName, Service); err != nil {
		if errors.IsNotFound(err) {
			// ConfigMap not found. Create it.
			Service := r.servicemysql(operator)
			l.Info("creating mysqlsvc", "releaseName", operator.Name, "namespace", operator.Namespace)
			if err := r.Create(ctx, &Service); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		updatedService := r.servicemysql(operator)
		if reflect.DeepEqual(Service.Spec, updatedService.Spec) || reflect.DeepEqual(Service.Name, updatedService.Name) {
			// If they are equal, log a message and return without updating

			return nil
		}
		Service.Spec = updatedService.Spec
		if err := r.Update(ctx, Service); err != nil {
			return err
		}

	}
	return nil
}
