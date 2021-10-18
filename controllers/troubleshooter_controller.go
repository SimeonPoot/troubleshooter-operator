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
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	clusteropsv1 "github.com/simeonpoot/troubleshooter-operator/api/v1"
)

// TroubleshooterReconciler reconciles a Troubleshooter object
type TroubleshooterReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	LabelSelector labels.Selector
	reconTally    int
}

//+kubebuilder:rbac:groups=clusterops.simeonpoot.nl,resources=troubleshooters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusterops.simeonpoot.nl,resources=troubleshooters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusterops.simeonpoot.nl,resources=troubleshooters/finalizers,verbs=update
//+kubebuilder:rbac:resources=pods,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:resources=namespaces,verbs=get;list;watch;create;update;delete

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
	log.Info("We've got the Troubleshooter CR, continuing")

	// Check whether the 'incident-response' Namespace needs to be created / already present
	// TODO: make function, create much needed tests
	requiredLabel, err := labels.NewRequirement("clusterops.simeonpoot/name", selection.Equals, []string{"incident-response"})
	if err != nil {
		log.Info("something went wrong creating labelselector for namespace reconciliation", "error", err)
	}

	ns := &corev1.NamespaceList{}
	if err := r.List(ctx, ns, &client.ListOptions{
		LabelSelector: labels.NewSelector().Add(*requiredLabel),
	}); err != nil {
		log.Info("something went wrong getting namespaces", "error", err)
	}

	// If we have other namespaces with the same 'labels', how to handle this?!

	if len(ns.Items) == 0 {
		fmt.Println("Incident-response namespace is not present, reconciling")
		if cr.Spec.CreateNamespace {
			fmt.Println("creating namespace")
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "incident-response",
					Labels: map[string]string{
						"clusterops.simeonpoot/name": "incident-response",
					},
				},
			}
			fmt.Println("Namespace to create: ", ns)
			err := r.Create(ctx, ns, &client.CreateOptions{})
			if err != nil {
				log.Info("error creating a namespace", "error", err)
				return ctrl.Result{}, nil
			}
		}
		// Need Tests, condition is not needed... if there are
		// else {
		// 	if len(ns.Items) > 0 {
		// 		for _, i := range ns.Items {
		// 			if i.Name != "incident-response" && !cr.Spec.CreateNamespace {
		// 				// return ctrl.Result{}, nil
		// 				break
		// 			}
		// 		}
		// 	}
		// }

		// check if it wanted 'true'
	}

	// TODO: wait on status namespace creation

	fmt.Println(ns)
	fmt.Println(len(ns.Items))
	// TODO: Use labelselector to collect/filter pods?
	// TODO: create function for it.
	pods := &corev1.PodList{}
	err = r.List(ctx, pods, &client.ListOptions{
		Namespace: req.Namespace,
		// LabelSelector: r.ClusterLabelSelector,
	})
	if err != nil {
		log.Info("unable to get Pods", "error", err)
		return ctrl.Result{}, nil
	}

	duration := cr.Spec.Session.Duration
	for _, i := range pods.Items {
		if i.Name == cr.Spec.PodName && i.Namespace == "incident-response" {
			log.Info("Created at: ", i.GetCreationTimestamp().Rfc3339Copy().Time.String(), err)
			// fmt.Println(i.CreationTimestamp == metav1.Now())

			// timeCreated := time.Time.String(i.CreationTimestamp.Time)
			timeCreated := i.GetCreationTimestamp().Rfc3339Copy().Time
			fmt.Println("time created: ", timeCreated)

			timeNow := time.Now()
			fmt.Println("time now:", timeNow)

			fmt.Println("time expected to end session:", timeCreated.Add(time.Minute*time.Duration(duration)))
			timeCreated = timeCreated.Add(time.Minute * time.Duration(duration))

			if timeCreated.Before(timeNow) {
				fmt.Println("TTL Session have been expired. \ntimeCreated + duration window is less than timeNow, so the TTL should be expired, delete Pod")
				// r.Delete(ctx, &i, &client.DeleteOptions{})
				if err := r.Delete(ctx, &i, &client.DeleteOptions{}); err != nil {
					fmt.Println("error deleting the pod!")
					log.Info("error deleting the pod", "error", err)
					// TODO: wait for status!
				}
				return ctrl.Result{}, nil
			} else {
				fmt.Println("still in the durationwindow of the session, requeue on the time-difference")
				//TODO: we need tests; otherwise we can get into a pickle:
				// If reconcile syncPeriod takes longer than the TTL of the session,
				// Reconcile syncPeriod takes presedent, meaning, the session will
				// take longer than the configured TTL.
				// Example: reconcile every 1hr, TTL is on 20m, this means the session should end within
				// 20minutes, but the reconcile will pick it up after an hour.
				// return ctrl.Result{RequeueAfter: time.Duration(duration) * time.Minute}, nil
			}
			log.Info("Pod already present!! session not yet expired", "error", err)
			return ctrl.Result{RequeueAfter: time.Duration(duration) * time.Minute}, nil // TODO: we need to test this!
			// diff := timeCreated - timeNow
			// fmt.Println("diff:", diff)
			// log.Info("Pod already present!!", "error", err)
			// return ctrl.Result{}, nil
		}
	}

	// TODO: create function for creating Pod / podSpec.
	// TODO: create UID per Pod, to make session unique.
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Spec.PodName,
			Namespace: "incident-response"},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				corev1.Container{
					Image: cr.Spec.Image,
					Name:  cr.Spec.Foo}}}}
	// pod.Namespace = cr.Namespace

	fmt.Println("Creating pod: ", pod.Name)
	if err := r.Create(ctx, pod, &client.CreateOptions{}); err != nil {
		log.Info("unable to get create Pod ", "error", err)
		return ctrl.Result{}, nil
	}
	log.Info("CREATED POD")
	fmt.Println(pod)

	return ctrl.Result{}, nil

	// delete
}

// SetupWithManager sets up the controller with the Manager.
func (r *TroubleshooterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clusteropsv1.Troubleshooter{}).
		Complete(r)
}
