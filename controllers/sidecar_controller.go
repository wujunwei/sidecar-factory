/*


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
	"github.com/go-logr/logr"
	injectionv1 "github.com/wujunwei/sidecar-factory/api/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
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

// +kubebuilder:rbac:groups=injection.wujunwei.io,resources=sidecars,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=injection.wujunwei.io,resources=sidecars/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=pods/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;update

func (r *SideCarReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("sidecar", req.NamespacedName)
	var res ctrl.Result

	var sidecar injectionv1.SideCar
	err := r.Get(ctx, req.NamespacedName, &sidecar)
	if err != nil {
		return res, client.IgnoreNotFound(err)
	}
	pl, err := r.getPodListBySideCar(sidecar)
	if err != nil {
		if sidecar.Spec.RetryLimit < sidecar.Status.RetryCount {
			r.Log.Info("pod not found ,retry count has reach the limit")
		} else {
			res.RequeueAfter = 10 * time.Second
			_ = r.updateRetryCount(ctx, &sidecar)
		}
		return res, err
	}
	var mergePatch, _ = json.Marshal(map[string]interface{}{
		"spec": map[string]interface{}{
			"containers": sidecar.Spec.Containers,
		},
	})
	for _, item := range pl.Items {
		err := r.Patch(ctx, &item, client.RawPatch(types.StrategicMergePatchType, mergePatch))
		if err != nil {
			if apierrors.IsConflict(err) {
				res.Requeue = true
			}
			r.Log.Error(err, "update pod error", "pod name", item.Name)
		}
	}
	if res.Requeue {
		_ = r.updateRetryCount(ctx, &sidecar)
	}
	return res, nil
}

func (r *SideCarReconciler) updateRetryCount(ctx context.Context, sc *injectionv1.SideCar) error {
	mergePatch, _ := json.Marshal(map[string]interface{}{
		"status": map[string]interface{}{
			"retry": sc.Spec.Containers,
		},
	})
	return r.Patch(ctx, sc, client.RawPatch(types.MergePatchType, mergePatch))
}

func (r *SideCarReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&injectionv1.SideCar{}).
		Complete(r)
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
