package buildlog

import (
	"fmt"
	"net/url"

	kapi "github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/apiserver"
	kclient "github.com/GoogleCloudPlatform/kubernetes/pkg/client"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/labels"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/runtime"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/api/errors"
	"github.com/openshift/origin/pkg/build/api"
	"github.com/openshift/origin/pkg/build/registry/build"
	"github.com/openshift/origin/pkg/cmd/server/kubernetes"
)

// REST is an implementation of RESTStorage for the api server.
type REST struct {
	BuildRegistry build.Registry
	PodControl    PodControlInterface
}

type PodControlInterface interface {
	getPod(namespace, name string) (*kapi.Pod, error)
}

type RealPodControl struct {
	podsNamspacer kclient.PodsNamespacer
}

func (r RealPodControl) getPod(namespace, name string) (*kapi.Pod, error) {
	return r.podsNamspacer.Pods(namespace).Get(name)
}

// NewREST creates a new REST for BuildLog
// Takes build registry and pod client to get neccessary attibutes to assamble
// URL to which the request shall be redirected in order to get build logs.
func NewREST(b build.Registry, pn kclient.PodsNamespacer) apiserver.RESTStorage {
	return &REST{
		BuildRegistry: b,
		PodControl:    RealPodControl{pn},
	}
}

// Redirector implementation
func (r *REST) ResourceLocation(ctx kapi.Context, id string) (string, error) {
	build, err := r.BuildRegistry.GetBuild(ctx, id)
	if err != nil {
		return "", errors.NewFieldNotFound("Build", id)
	}

	if len(build.PodName) == 0 {
		return "", errors.NewFieldRequired("Build.PodName", build.PodName)
	}

	pod, err := r.PodControl.getPod(build.Namespace, build.PodName)
	if err != nil {
		return "", errors.NewFieldNotFound("Pod.Name", build.PodName)
	}

	buildPodID := build.PodName
	buildPodHost := pod.Status.Host
	buildPodNamespace := pod.Namespace
	// Build will take place only in one container
	buildContainerName := pod.Spec.Containers[0].Name
	location := &url.URL{
		Scheme: kubernetes.NodeScheme,
		Host:   fmt.Sprintf("%s:%d", buildPodHost, kubernetes.NodePort),
		Path:   fmt.Sprintf("/containerLogs/%s/%s/%s", buildPodNamespace, buildPodID, buildContainerName),
	}

	if pod.Status.Phase != kapi.PodRunning {
		return "", errors.NewFieldInvalid("Pod.Status", pod.Status.Phase, "must be Running")
	}

	switch build.Status {
	case api.BuildStatusRunning:
		location.RawQuery = url.Values{"follow": []string{"1"}}.Encode()
	case api.BuildStatusComplete, api.BuildStatusFailed:
		// Do not follow the Complete and Failed logs as the streaming already
		// finished.
	case api.BuildStatusPending:
		return "", fmt.Errorf("unable to retrieve logs, the Build is in 'Pending' state")
	default:
		return "", errors.NewFieldInvalid("build.Status", build.Status, "must be Running")
	}

	if err != nil {
		return "", err
	}
	return location.String(), nil
}

func (r *REST) Get(ctx kapi.Context, id string) (runtime.Object, error) {
	return nil, fmt.Errorf("BuildLog can't be retrieved")
}

func (r *REST) New() runtime.Object {
	return nil
}

func (*REST) NewList() runtime.Object {
	return nil
}

func (r *REST) List(ctx kapi.Context, selector, fields labels.Selector) (runtime.Object, error) {
	return nil, fmt.Errorf("BuildLog can't be listed")
}

func (r *REST) Delete(ctx kapi.Context, id string) (<-chan apiserver.RESTResult, error) {
	return nil, fmt.Errorf("BuildLog can't be deleted")
}

func (r *REST) Create(ctx kapi.Context, obj runtime.Object) (<-chan apiserver.RESTResult, error) {
	return nil, fmt.Errorf("BuildLog can't be created")
}

func (r *REST) Update(ctx kapi.Context, obj runtime.Object) (<-chan apiserver.RESTResult, error) {
	return nil, fmt.Errorf("BuildLog can't be updated")
}
