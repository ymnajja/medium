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

package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	mediumv1beta1 "example.com/example/api/v1beta1"
)

// WordpressReconciler reconciles a Wordpress object
type WordpressReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=medium.example.org,resources=wordpresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=medium.example.org,resources=wordpresses/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=medium.example.org,resources=wordpresses/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=deployments,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Wordpress object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.0/pkg/reconcile
func (r *WordpressReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	op := &mediumv1beta1.Wordpress{}
	if err := r.mysqlpvc(ctx, op, l); err != nil {
		l.Error(err, "error creating mysql pvc")
		return ctrl.Result{}, err
	}
	if err := r.reconciledploymentsmysql(ctx, op, l); err != nil {
		l.Error(err, "error creating mysql deployment")
		return ctrl.Result{}, err
	}
	if err := r.reconcileservicemysql(ctx, op, l); err != nil {
		l.Error(err, "error creating mysql service")
		return ctrl.Result{}, err
	}

	if err := r.reconcilewordpresspvc(ctx, op, l); err != nil {
		l.Error(err, "error creating wordpress pvc")
		return ctrl.Result{}, err
	}

	if err := r.reconciledployments(ctx, op, l); err != nil {
		l.Error(err, "error creating wordpress deployment")
		return ctrl.Result{}, err
	}

	if err := r.reconcilewordpresssvc(ctx, op, l); err != nil {
		l.Error(err, "error creating wordpress service")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *WordpressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mediumv1beta1.Wordpress{}).
		Complete(r)
}
