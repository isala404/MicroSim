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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"net/http"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"strings"
	"time"

	microsimv1alpha1 "github.com/MrSupiri/MicroSim/api/v1alpha1"
)

// LoadGeneratorReconciler reconciles a LoadGenerator object
type LoadGeneratorReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=microsim.isala.me,resources=loadgenerators,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=microsim.isala.me,resources=loadgenerators/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=microsim.isala.me,resources=loadgenerators/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the LoadGenerator object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *LoadGeneratorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info("Reconciling")

	var loadGenerator microsimv1alpha1.LoadGenerator
	if err := r.Get(ctx, req.NamespacedName, &loadGenerator); err != nil {
		return ctrl.Result{Requeue: false}, client.IgnoreNotFound(err)
	}
	ctx = context.WithValue(ctx, "loadgenerator", loadGenerator)

	var simulation microsimv1alpha1.Simulation
	if err := r.Get(ctx, types.NamespacedName{
		Namespace: loadGenerator.Spec.SimulationRef.Namespace,
		Name:      loadGenerator.Spec.SimulationRef.Name,
	}, &simulation); err != nil {
		return ctrl.Result{}, err
	}
	ctx = context.WithValue(ctx, "simulation", simulation)

	var payload microsimv1alpha1.Route
	if err := json.Unmarshal([]byte(loadGenerator.Spec.Request), &payload); err != nil {
		logger.Error(err, "error while decoding request spec")
	}

	// Stop if the request count is met
	if loadGenerator.Spec.RequestCount != nil {
		if loadGenerator.Status.DoneRequests >= *loadGenerator.Spec.RequestCount {
			logger.V(1).Info("request count met, stopping load generator")
			return ctrl.Result{Requeue: false}, nil
		}
	}

	// Stop if time out was passed
	if loadGenerator.Spec.Timeout != nil {
		timeout := loadGenerator.ObjectMeta.CreationTimestamp.Add(loadGenerator.Spec.Timeout.Duration)
		if loadGenerator.ObjectMeta.CreationTimestamp.After(timeout) {
			logger.V(1).Info("timeout reached, stopping load generator")
			return ctrl.Result{Requeue: false}, nil
		}
	}

	payload = overwriteDesignations(ctx, payload)

	if err := r.forwardRequest(ctx, payload); err != nil {
		return ctrl.Result{}, err
	}

	logger.V(1).Info(fmt.Sprintf("requeuing in %s", loadGenerator.Spec.BetweenDelay.Duration))
	return ctrl.Result{RequeueAfter: loadGenerator.Spec.BetweenDelay.Duration}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *LoadGeneratorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&microsimv1alpha1.LoadGenerator{}).
		WithEventFilter(eventFilter()).
		Complete(r)
}

func eventFilter() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Ignore updates to CR status in which case metadata.Generation does not change
			return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration()
		},
	}
}

func overwriteDesignations(ctx context.Context, route microsimv1alpha1.Route) microsimv1alpha1.Route {
	logger := log.FromContext(ctx)
	simulation := ctx.Value("simulation").(microsimv1alpha1.Simulation)

	if !strings.HasPrefix(route.Designation, "http") {
		if svc, ok := simulation.Status.Services[formatServiceName(route.Designation, simulation)]; ok {
			route.Designation = svc.Endpoint
		} else {
			logger.V(-1).Info("service name was not found", "service name", route.Designation)
		}
	}
	for i, p := range route.Routes {
		route.Routes[i] = overwriteDesignations(ctx, p)
	}
	return route
}

func (r *LoadGeneratorReconciler) forwardRequest(ctx context.Context, route microsimv1alpha1.Route) error {
	logger := log.FromContext(ctx)
	loadGenerator := ctx.Value("loadgenerator").(microsimv1alpha1.LoadGenerator)
	startedTime := time.Now()

	logger.V(1).Info("creating the request", "designation", route.Designation)

	reqBody, err := json.Marshal(route.Routes)
	if err != nil {
		return err
	}

	resp, err := http.Post("http://localhost:9090/", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	// TODO: make this sub resource
	loadGenerator.Status.Responses = append(loadGenerator.Status.Responses, microsimv1alpha1.Responses{
		Response: string(buf),
		Request:  string(reqBody),
	})

	avg := time.Duration(startedTime.Add(loadGenerator.Status.AverageResponseTime.Duration).Second()/2) * time.Second
	loadGenerator.Status.DoneRequests += 1
	loadGenerator.Status.AverageResponseTime = metav1.Duration{Duration: avg}

	if err := r.Status().Update(ctx, &loadGenerator); err != nil {
		logger.V(-1).Info("failed to update load generator status")
		return err
	}

	return nil
}
