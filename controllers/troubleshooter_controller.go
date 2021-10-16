/*
Copyright 2021.

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

package controllers

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	clusteropsv1 "github.com/simeonpoot/troubleshooter-operator/api/v1"
)

// TroubleshooterReconciler reconciles a Troubleshooter object
type TroubleshooterReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	reconTally int
}

//+kubebuilder:rbac:groups=clusterops.simeonpoot.nl,resources=troubleshooters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusterops.simeonpoot.nl,resources=troubleshooters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusterops.simeonpoot.nl,resources=troubleshooters/finalizers,verbs=update
//+kubebuilder:rbac:resources=pods,verbs=get;list;watch;create;update;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Troubleshooter object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *TroubleshooterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// _ = log.FromContext(ctx)
	log := logf.FromContext(ctx).WithName("Reconcile")
	ctx = logf.IntoContext(ctx, log)

	// your logic here
	log.V(1).Info("Start Reconcile", "tally", r.reconTally)
	defer log.V(1).Info("End Reconcile", "tally", r.reconTally)

	cr := &clusteropsv1.Troubleshooter{}
	if err := r.Get(ctx, req.NamespacedName, cr); err != nil {
		log.V(2).Info("unable to get kind Troubleshooter (retried)", "error", err)
		apierrors.IsNotFound(err)
		return ctrl.Result{}, nil
	}
	log.Info("We've got the Troubleshooter CR, creating POD")

	// TODO: Use labelselector to collect/filter pods?
	// TODO: create function for it.
	pods := &corev1.PodList{}
	err := r.List(ctx, pods, &client.ListOptions{
		Namespace: req.Namespace,
		// LabelSelector: r.ClusterLabelSelector,
	})
	if err != nil {
		log.Info("unable to get cluster Secrets (retried)", "error", err)
		return ctrl.Result{}, nil
	}

	for _, i := range pods.Items {
		if i.Name == cr.Spec.PodName {
			log.Info("Pod already present!!", "error", err)
			return ctrl.Result{}, nil
		}
	}

	// TODO: create function for creating Pod / podSpec.
	// TODO: create UID per Pod, to make session unique.
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: cr.Spec.PodName},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				corev1.Container{
					Image: cr.Spec.Image,
					Name:  cr.Spec.Foo}}}}
	pod.Namespace = cr.Namespace

	fmt.Println(pod)
	if err := r.Create(ctx, pod, &client.CreateOptions{}); err != nil {
		log.Info("unable to get create Pod ", "error", err)
		return ctrl.Result{}, nil
	}
	log.Info("CREATED POD")
	fmt.Println(pod)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TroubleshooterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clusteropsv1.Troubleshooter{}).
		Complete(r)
}
