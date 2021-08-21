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
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	"io/ioutil"
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
//+kubebuilder:rbac:groups=microsim.isala.me,resources=simulations,verbs=get;list;watch

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

	if loadGenerator.Status.Payload == nil {
		var payload microsimv1alpha1.Route
		if err := json.Unmarshal([]byte(loadGenerator.Spec.Request), &payload); err != nil {
			logger.Error(err, "error while decoding request spec")
		}
		payload = overwriteDesignations(ctx, payload)
		loadGenerator.Status.Payload = &payload
	}

	// Run this on the background
	go r.forwardRequest(req.NamespacedName, logger, loadGenerator.Status.Payload)

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

func (r *LoadGeneratorReconciler) forwardRequest(namespacedName types.NamespacedName, logger logr.Logger, route *microsimv1alpha1.Route) {
	ctx := context.Background()

	var requests int
	var responseTimes time.Duration
	responses := make(map[string]microsimv1alpha1.Responses)

	for i, r2 := range route.Routes {
		logger.V(1).Info(fmt.Sprintf("sending %d request", i), "designation", route.Designation)
		startedTime := time.Now()
		reqBody, err := json.Marshal(r2)
		if err != nil {
			logger.Error(err, fmt.Sprintf("failed encoded request route #%d", i), "route", r2)
			return
		}

		resp, err := http.Post(route.Designation, "application/json", bytes.NewBuffer(reqBody))
		if err != nil {
			logger.Error(err, "failed send the request to designation", "designation", route.Designation)
			return
		}
		defer resp.Body.Close()

		buf, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			logger.Error(err, fmt.Sprintf("failed decoded response for route #%d", i), "route", r2)
			return
		}

		// Store only unique requests and responses
		reqRespHash := GetMD5Hash(append(buf, reqBody...))

		responses[reqRespHash] = microsimv1alpha1.Responses{
			Response: string(buf),
			Request:  string(reqBody),
		}
		requests += 1
		responseTimes += startedTime.Sub(time.Now())
	}

	// Fetch new status
	var newLoadGenerator microsimv1alpha1.LoadGenerator
	if err := r.Get(ctx, namespacedName, &newLoadGenerator); err != nil {
		logger.Error(err, "failed fetch load generator", "namespacedName", namespacedName)
		return
	}

	// Merge the status
	newLoadGenerator.Status.DoneRequests += requests
	newLoadGenerator.Status.TotalResponseTime.Duration += responseTimes
	if newLoadGenerator.Status.Responses == nil {
		newLoadGenerator.Status.Responses = responses
	} else {
		for s, m := range responses {
			newLoadGenerator.Status.Responses[s] = m
		}
	}

	// Update the status
	if err := r.Status().Update(ctx, &newLoadGenerator); err != nil {
		logger.Error(err, "failed to update load generator status")
		return
	}

	return
}

func GetMD5Hash(input []byte) string {
	hash := md5.Sum(input)
	return hex.EncodeToString(hash[:])
}
