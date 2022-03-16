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
	"github.com/google/uuid"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"math/rand"
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

type requestStatus struct {
	Response      map[string]microsimv1alpha1.Responses
	ResponseTimes time.Duration
}

func eventFilter() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Ignore updates to CR status in which case metadata.Generation does not change
			return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration()
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *LoadGeneratorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&microsimv1alpha1.LoadGenerator{}).
		WithEventFilter(eventFilter()).
		Complete(r)
}

//+kubebuilder:rbac:groups=microsim.isala.me,resources=loadgenerators,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=microsim.isala.me,resources=loadgenerators/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=microsim.isala.me,resources=loadgenerators/finalizers,verbs=update
//+kubebuilder:rbac:groups=microsim.isala.me,resources=simulations,verbs=get;list;watch

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

	for _, request := range loadGenerator.Spec.Routes {
		var route microsimv1alpha1.Route
		if err := json.Unmarshal([]byte(request), &route); err != nil {
			logger.Error(err, "error while decoding request spec")
		}
		route = overwriteDesignations(ctx, route)
		// Send all the requests in background thread
		// This is done because some request may take 10+ seconds and that value will be added spec.delayBetween
		go r.processRequests(ctx, req, route)
	}

	logger.V(1).Info(fmt.Sprintf("requeuing in %s", loadGenerator.Spec.BetweenDelay.Duration))
	return ctrl.Result{RequeueAfter: loadGenerator.Spec.BetweenDelay.Duration}, nil
}

func (r *LoadGeneratorReconciler) processRequests(ctx context.Context, req ctrl.Request, payload microsimv1alpha1.Route) {
	logger := log.FromContext(ctx)
	loadGenerator := ctx.Value("loadgenerator").(microsimv1alpha1.LoadGenerator)

	results := make(chan *requestStatus, loadGenerator.Spec.Replicas)
	for i := 0; i < loadGenerator.Spec.Replicas; i++ {
		// Run this on the background
		go r.forwardRequest(ctx, payload, results)
	}

	// Merge the responses
	var responseTime time.Duration
	responses := make(map[string]microsimv1alpha1.Responses)
	for i := 0; i < loadGenerator.Spec.Replicas; i++ {
		if res := <-results; res != nil {
			responseTime += res.ResponseTimes
			for s, m := range res.Response {
				responses[s] = m
			}
		}
	}

	// Fetch new status because one we have might be outdated
	var newLoadGenerator microsimv1alpha1.LoadGenerator
	if err := r.Get(ctx, req.NamespacedName, &newLoadGenerator); err != nil {
		logger.Error(err, "failed fetch load generator", "namespacedName", req.NamespacedName)
		return
	}

	newLoadGenerator.Status.DoneRequests += 1
	newLoadGenerator.Status.TotalResponseTime.Duration += responseTime
	if newLoadGenerator.Status.Responses == nil {
		newLoadGenerator.Status.Responses = responses
	} else {
		for s, m := range responses {
			newLoadGenerator.Status.Responses[s] = m
		}
	}

	// Coz why not
	newLoadGenerator.Status.Replicas = newLoadGenerator.Spec.Replicas

	// Update the status
	// State here is only should be to get rough idea
	// If a race condition was met this update will be ignored
	if err := r.Status().Update(ctx, &newLoadGenerator); err != nil {
		logger.Error(err, "failed to update load generator status")
	}

}

func (r *LoadGeneratorReconciler) forwardRequest(ctx context.Context, route microsimv1alpha1.Route, results chan *requestStatus) {
	logger := log.FromContext(ctx)
	responses := make(map[string]microsimv1alpha1.Responses)
	startedTime := time.Now()

	//for i, r2 := range route.Routes {
	logger.V(1).Info("sending request", "designation", route.Designation)
	reqBody, err := json.Marshal(route)
	if err != nil {
		logger.Error(err, "failed encoded request route #%d", "route", route)
		results <- nil
		return
	}

	httpClient := http.Client{
		Transport: &http.Transport{DisableKeepAlives: true},
	}
	defer httpClient.CloseIdleConnections()
	req, err := http.NewRequest("POST", route.Designation, bytes.NewBuffer(reqBody))
	if err != nil {
		logger.Error(err, "failed send the request to designation", "designation", route.Designation)
		results <- nil
		return
	}
	reqID := uuid.New()
	req.Header = http.Header{
		"Content-Type": []string{"application/json"},
		"X-Request-ID": []string{reqID.String()},
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Error(err, "failed send the request to designation", "designation", route.Designation)
		results <- nil
		return
	}
	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		logger.Error(err, "failed decoded response for route", "route", route)
		results <- nil
		return
	}

	// Store only unique requests and responses
	reqRespHash := GetMD5Hash(append(buf, reqBody...))

	responses[reqRespHash] = microsimv1alpha1.Responses{
		Response: string(buf),
		Request:  string(reqBody),
	}
	//}

	results <- &requestStatus{
		Response:      responses,
		ResponseTimes: time.Now().Sub(startedTime),
	}
}

func overwriteDesignations(ctx context.Context, route microsimv1alpha1.Route) microsimv1alpha1.Route {
	logger := log.FromContext(ctx)
	simulation := ctx.Value("simulation").(microsimv1alpha1.Simulation)
	var newRoutes []microsimv1alpha1.Route

	if !strings.HasPrefix(route.Designation, "http") {
		if svc, ok := simulation.Status.Services[formatServiceName(route.Designation, simulation)]; ok {
			route.Designation = svc.Endpoint
		} else {
			logger.V(-1).Info("service name was not found", "service name", route.Designation)
		}
	}

	for _, p := range route.Routes {
		if rand.Intn(100) <= route.Probability {
			newRoutes = append(newRoutes, overwriteDesignations(ctx, p))
		}
	}
	route.Routes = newRoutes
	return route
}

func GetMD5Hash(input []byte) string {
	hash := md5.Sum(input)
	return hex.EncodeToString(hash[:])
}
