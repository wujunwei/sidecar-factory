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
	"encoding/json"
	"github.com/go-logr/logr"
	injectionv1 "github.com/wujunwei/sidecar-factory/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

// SideCarReconciler reconciles a SideCar object
type SideCarReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=injection.wujunwei.io,resources=sidecars,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=injection.wujunwei.io,resources=sidecars/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=injection.wujunwei.io,resources=sidecars/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;update;patch
//+kubebuilder:rbac:groups="",resources=pods/status,verbs=get;update;patch
//+kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;update

// Reconcile For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *SideCarReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("sidecar", req.NamespacedName)
	r.Log.Info("new sidecar come", "namespace-name", req.NamespacedName)
	var res ctrl.Result

	var sidecar injectionv1.SideCar
	err := r.Get(ctx, req.NamespacedName, &sidecar)
	if err != nil {
		return res, client.IgnoreNotFound(err)
	}

	pl, err := r.getPodListBySideCar(sidecar)
	if err != nil || len(pl.Items) == 0 {
		if sidecar.Spec.RetryLimit < sidecar.Status.RetryCount {
			r.Log.Info("pod not found ,retry count has reach the limit")
		} else {
			res.RequeueAfter = 10 * time.Second
			_ = r.updateRetryCount(ctx, &sidecar)
		}
		return res, err
	}

	for _, item := range pl.Items {
		for i, cnt := range item.Spec.Containers {
			if im, ok := sidecar.Spec.Images[cnt.Name]; ok {
				item.Spec.Containers[i].Image = im
			}
		}
		err :=retry.RetryOnConflict(retry.DefaultRetry, func() error {
			return r.Update(ctx, &item)
		})
		if err != nil {
			r.Log.Error(err, "update pod error", "pod name", item.Name)
		}
	}
	if res.Requeue {
		_ = r.updateRetryCount(ctx, &sidecar)
	}
	return res, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SideCarReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&injectionv1.SideCar{}).
		Complete(r)
}

func (r *SideCarReconciler) updateRetryCount(ctx context.Context, sc *injectionv1.SideCar) error {
	mergePatch, _ := json.Marshal(map[string]interface{}{
		"status": map[string]interface{}{
			"retry": sc.Status.RetryCount + 1,
		},
	})
	return r.Patch(ctx, sc, client.RawPatch(types.MergePatchType, mergePatch))
}
func (r *SideCarReconciler) getPodListBySideCar(sc injectionv1.SideCar) (corev1.PodList, error) {
	var podList corev1.PodList
	selector, _ := metav1.LabelSelectorAsSelector(sc.Spec.Selector)
	err := r.List(context.Background(), &podList, client.MatchingLabelsSelector{Selector: selector}, client.InNamespace(sc.Namespace))
	if err != nil {
		return podList, err
	}
	return podList, nil
}
