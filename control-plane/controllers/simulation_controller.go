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
	microsimv1alpha1 "github.com/MrSupiri/MicroSim/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"strings"
)

// SimulationReconciler reconciles a Simulation object
type SimulationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=microsim.isala.me,resources=simulations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=microsim.isala.me,resources=simulations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=microsim.isala.me,resources=simulations/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=services,verbs=get;watch;list;create;delete
//+kubebuilder:rbac:groups="apps",resources=deployments,verbs=get;watch;list;create;delete
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *SimulationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info("Reconciling")

	var simulation microsimv1alpha1.Simulation
	if err := r.Get(ctx, req.NamespacedName, &simulation); err != nil {
		if apierrors.IsNotFound(err) {
			if err := r.CleanUpResources(ctx, req.Name); client.IgnoreNotFound(err) != nil {
				// TODO: write this to event log
				return ctrl.Result{Requeue: true}, err
			}
		} else {
			logger.Error(err, "failed to get the simulation")
		}
		return ctrl.Result{Requeue: false}, client.IgnoreNotFound(err)
	}

	ctx = context.WithValue(ctx, "simulation", simulation)

	if len(simulation.Status.Services) == 0 {
		simulation.Status.Services = map[string]microsimv1alpha1.ServiceStatus{}
	}

	requeue := false
	for name, service := range simulation.Spec.Services {
		name = formatServiceName(name, simulation)
		// Check if the service is already deployed
		if _, ok := simulation.Status.Services[name]; ok {
			continue
		}

		// Create the Deployment and Service if not created before
		if err := r.ProvisionIfNoExists(ctx, name, service); err != nil {
			// TODO: write the error here to CRD events
			requeue = true
			continue
		}
		// Update the status
		simulation.Status.Services[name] = microsimv1alpha1.ServiceStatus{
			Endpoint:  fmt.Sprintf("http://%s.%s.svc/", name, simulation.ObjectMeta.Namespace),
			Language:  service.Language,
			Framework: service.Framework,
		}
	}

	// Write the status to etcd
	if err := r.Status().Update(ctx, &simulation); err != nil {
		logger.Error(err, "failed to update simulation status")
		return ctrl.Result{Requeue: true}, err
	}
	return ctrl.Result{Requeue: requeue}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SimulationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&microsimv1alpha1.Simulation{}).
		Complete(r)
}

func (r *SimulationReconciler) ProvisionIfNoExists(ctx context.Context, name string, service microsimv1alpha1.ServiceSpec) error {
	logger := log.FromContext(ctx)
	simulation := ctx.Value("simulation").(microsimv1alpha1.Simulation)

	labels := map[string]string{
		"app.kubernetes.io/instance":   name,
		"app.kubernetes.io/part-of":    simulation.ObjectMeta.Name,
		"app.kubernetes.io/managed-by": "microsim-simulation",
		"app.kubernetes.io/created-by": "microsim",
	}

	deployment := appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/appsv1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: simulation.ObjectMeta.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{{
						Name:            "service",
						Image:           fmt.Sprintf("ghcr.io/mrsupiri/microsim:%s-%s", service.Language, service.Framework),
						ImagePullPolicy: v1.PullIfNotPresent,
						Env: []v1.EnvVar{
							{
								Name:  "SERVICE_NAME",
								Value: name,
							},
						},
						Ports: []v1.ContainerPort{{
							Name:          "http",
							ContainerPort: 8080,
							Protocol:      v1.ProtocolTCP,
						}},
					}},
				},
			},
		},
	}

	clusterIP := v1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: simulation.ObjectMeta.Namespace,
			Labels:    labels,
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeClusterIP,
			Ports: []v1.ServicePort{{
				Name:       "http",
				Protocol:   v1.ProtocolTCP,
				Port:       80,
				TargetPort: intstr.IntOrString{Type: intstr.String, StrVal: "http"},
			}},
			Selector: labels,
		},
	}

	if err := r.Create(ctx, &deployment); IgnoreAlreadyExist(err) != nil {
		logger.Error(err, fmt.Sprintf("failed to create deployment %s", deployment.ObjectMeta.Name))
		return err
	}

	if err := r.Create(ctx, &clusterIP); IgnoreAlreadyExist(err) != nil {
		logger.Error(err, fmt.Sprintf("failed to create service %s", clusterIP.ObjectMeta.Name))
		return err
	}
	return nil
}

func (r *SimulationReconciler) CleanUpResources(ctx context.Context, name string) (err error) {
	logger := log.FromContext(ctx)

	matchingLabels := &client.MatchingLabels{"app.kubernetes.io/part-of": name}

	var deploymentList appsv1.DeploymentList
	if err = r.List(ctx, &deploymentList, matchingLabels); client.IgnoreNotFound(err) != nil {
		logger.Error(err, fmt.Sprintf("failed to get provisioned %s", deploymentList.Kind))
		return err
	}
	for _, resource := range deploymentList.Items {
		if err = r.Delete(ctx, &resource); client.IgnoreNotFound(err) != nil {
			logger.Error(err, fmt.Sprintf("failed to delete a provisioned %s", resource.Kind), "uuid", resource.GetUID(), "name", resource.GetName())
			return err
		}
		logger.V(1).Info(fmt.Sprintf("%s deleted", resource.Kind), "uuid", resource.GetUID(), "name", resource.GetName())
	}

	var serviceList v1.ServiceList
	if err = r.List(ctx, &serviceList, matchingLabels); client.IgnoreNotFound(err) != nil {
		logger.Error(err, fmt.Sprintf("failed to get provisioned %s", serviceList.Kind))
		return err
	}

	for _, resource := range serviceList.Items {
		if err = r.Delete(ctx, &resource); client.IgnoreNotFound(err) != nil {
			logger.Error(err, fmt.Sprintf("failed to delete a provisioned %s", resource.Kind), "uuid", resource.GetUID(), "name", resource.GetName())
			return err
		}
		logger.V(1).Info(fmt.Sprintf("%s deleted", resource.Kind), "uuid", resource.GetUID(), "name", resource.GetName())
	}
	return err
}

func IgnoreAlreadyExist(err error) error {
	if apierrors.IsAlreadyExists(err) {
		return nil
	}
	return err
}

func formatServiceName(name string, simulation microsimv1alpha1.Simulation) string {
	return fmt.Sprintf("%s-%s", strings.Replace(name, "_", "-", -1), simulation.ObjectMeta.UID[:8])
}
