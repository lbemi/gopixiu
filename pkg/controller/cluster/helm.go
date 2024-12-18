/*
Copyright 2021 The Pixiu Authors.

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

package cluster

import (
	"os"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	"github.com/caoyingjunz/pixiu/pkg/client"
	"github.com/caoyingjunz/pixiu/pkg/db"
)

type HelmInterface interface {
	Releases(namespace string) ReleaseInterface
	ReleasesV2(namespace string) ReleasesInterfaceV2
	Repository() RepositoryInterface
}

type Helm struct {
	cluster           string
	factory           db.ShareDaoFactory
	settings          *cli.EnvSettings
	actionConfig      *action.Configuration
	resetClientGetter *client.HelmRESTClientGetter
	clientSet         *kubernetes.Clientset
}

func (h *Helm) Releases(namespace string) ReleaseInterface {
	h.settings.SetNamespace(namespace)
	if err := h.actionConfig.Init(
		h.resetClientGetter,
		h.settings.Namespace(),
		os.Getenv("HELM_DRIVER"),
		klog.Infof,
	); err != nil {
		klog.Errorf("failed to init helm action config: %v", err)
		return nil
	}
	return newReleases(h.actionConfig, h.settings)
}

func (h *Helm) ReleasesV2(namespace string) ReleasesInterfaceV2 {
	h.settings.SetNamespace(namespace)
	if err := h.actionConfig.Init(
		h.resetClientGetter,
		h.settings.Namespace(),
		os.Getenv("HELM_DRIVER"),
		klog.Infof,
	); err != nil {
		klog.Errorf("failed to init helm action config: %v", err)
		return nil
	}
	return newReleasesV2(h.actionConfig, h.settings, h.clientSet)
}

func (h *Helm) Repository() RepositoryInterface {
	if err := h.actionConfig.Init(
		h.resetClientGetter,
		h.settings.Namespace(),
		os.Getenv("HELM_DRIVER"),
		klog.Infof,
	); err != nil {
		klog.Errorf("failed to init helm action config: %v", err)
		return nil
	}
	return newRepository(h.cluster, h.settings, h.actionConfig, h.factory)
}

func newHelm(kubeConfig *rest.Config, cluster string, factory db.ShareDaoFactory) *Helm {
	settings := cli.New()
	actionConfig := new(action.Configuration)
	resetClientGetter := client.NewHelmRESTClientGetter(kubeConfig)

	clientSet, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		klog.Errorf("failed to create kubernetes client: %v", err)
		return nil
	}
	settings.KubeAPIServer = kubeConfig.Host
	settings.KubeAsUser = kubeConfig.Username
	settings.KubeToken = kubeConfig.BearerToken

	return &Helm{
		settings:          settings,
		actionConfig:      actionConfig,
		resetClientGetter: resetClientGetter,
		cluster:           cluster,
		factory:           factory,
		clientSet:         clientSet,
	}
}
