package controllers

import (
	"context"
	"encoding/json"
	corev1 "k8s.io/api/core/v1"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/mutate-v1-pod,mutating=true,failurePolicy=ignore,groups="",resources=pods,verbs=create;update,versions=v1,name=pod.wujunwei.io,sideEffects=none,admissionReviewVersions=v1

var _ webhook.AdmissionHandler = &PodSidecar{}

type PodSidecar struct {
	Client  client.Client
	decoder *admission.Decoder
}

var podInjectionLog = logf.Log.WithName("pod injection")

func (a *PodSidecar) Handle(ctx context.Context, req admission.Request) admission.Response {
	podInjectionLog.Info("started !")
	pod := &corev1.Pod{}
	err := a.decoder.Decode(req, pod)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	pod.Spec.Containers = append(pod.Spec.Containers, corev1.Container{
		Name:    "sidecar",
		Image:   "busybox",
		Command: []string{"sleep"},
		Args:    []string{"100000"},
	})
	marshaledPod, err := json.Marshal(pod)
	if err != nil {
		podInjectionLog.Error(err, "json encode error")
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPod)
}
func (a *PodSidecar) InjectDecoder(d *admission.Decoder) error {
	a.decoder = d
	podInjectionLog.Info("injected")
	return nil
}
