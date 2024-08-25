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

package client

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	appsv1 "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/listers/batch/v1beta1"
	v1 "k8s.io/client-go/listers/core/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	resourceclient "k8s.io/metrics/pkg/client/clientset/versioned/typed/metrics/v1beta1"
)

var (
	groupVersionResources = []schema.GroupVersionResource{
		{Group: "", Version: "v1", Resource: "pods"},
		{Group: "", Version: "v1", Resource: "nodes"},
		{Group: "apps", Version: "v1", Resource: "deployments"},
		{Group: "apps", Version: "v1", Resource: "statefulsets"},
		{Group: "apps", Version: "v1", Resource: "daemonsets"},
		{Group: "batch", Version: "v1", Resource: "cronjobs"},
	}
	watchResource = map[string]schema.GroupVersionResource{
		"Pod":         {},
		"Deployment":  {},
		"StatefulSet": {},
		"DaemonSet":   {},
		"CronJob":     {},
	}
)

type PixiuInformer struct {
	Shared informers.SharedInformerFactory
	Cancel context.CancelFunc
}

func (p PixiuInformer) NodesLister() v1.NodeLister {
	return p.Shared.Core().V1().Nodes().Lister()
}

func (p PixiuInformer) PodsLister() v1.PodLister {
	return p.Shared.Core().V1().Pods().Lister()
}

func (p PixiuInformer) NamespacesLister() v1.NamespaceLister {
	return p.Shared.Core().V1().Namespaces().Lister()
}

func (p PixiuInformer) DeploymentsLister() appsv1.DeploymentLister {
	return p.Shared.Apps().V1().Deployments().Lister()
}

func (p *PixiuInformer) StatefulSetsLister() appsv1.StatefulSetLister {
	return p.Shared.Apps().V1().StatefulSets().Lister()
}

func (p *PixiuInformer) DaemonSetsLister() appsv1.DaemonSetLister {
	return p.Shared.Apps().V1().DaemonSets().Lister()
}

func (p *PixiuInformer) CronJobsLister() v1beta1.CronJobLister {
	gvr := watchResource["CronJob"]
	group := gvr.Group
	version := gvr.Version

	batchInterface := reflect.ValueOf(p.Shared.Batch())
	versionMethod := batchInterface.MethodByName(strings.Title(version))
	if !versionMethod.IsValid() {
		return nil // Or handle error appropriately
	}

	versionInterface := versionMethod.Call(nil)[0]
	cronJobsMethod := versionInterface.MethodByName("CronJobs")
	if !cronJobsMethod.IsValid() {
		return nil // Or handle error appropriately
	}

	cronJobsInterface := cronJobsMethod.Call(nil)[0]
	listerMethod := cronJobsInterface.MethodByName("Lister")
	if !listerMethod.IsValid() {
		return nil // Or handle error appropriately
	}

	lister := listerMethod.Call(nil)[0].Interface()
	return lister.(v1beta1.CronJobLister)
}

type ClusterSet struct {
	Client   *kubernetes.Clientset
	Config   *restclient.Config
	Metric   *resourceclient.MetricsV1beta1Client
	Informer *PixiuInformer
}

func (cs *ClusterSet) Complete(cfg []byte) error {
	var err error
	if cs.Config, err = clientcmd.RESTConfigFromKubeConfig(cfg); err != nil {
		return err
	}
	if cs.Client, err = kubernetes.NewForConfig(cs.Config); err != nil {
		return err
	}
	if cs.Metric, err = resourceclient.NewForConfig(cs.Config); err != nil {
		return err
	}

	sharedInformer, cancel, err := NewSharedInformers(cs.Config)
	if err != nil {
		return err
	}
	cs.Informer = &PixiuInformer{
		Shared: sharedInformer,
		Cancel: cancel,
	}
	return nil
}

func NewSharedInformers(c *restclient.Config) (informers.SharedInformerFactory, context.CancelFunc, error) {
	// 重新构造客户端
	clientSet, err := kubernetes.NewForConfig(c)
	if err != nil {
		return nil, nil, err
	}
	apiResourcesList, err := clientSet.Discovery().ServerPreferredResources()
	if err != nil {
		return nil, nil, err
	}
	for _, apiResources := range apiResourcesList {
		for _, apiResource := range apiResources.APIResources {
			kind := apiResource.Kind
			if _, ok := watchResource[apiResource.Kind]; ok {
				groupVersion := strings.Split(apiResources.GroupVersion, "/")
				gvr := schema.GroupVersionResource{}
				if len(groupVersion) == 1 {
					gvr.Group = ""
					gvr.Version = groupVersion[0]
				} else {
					gvr.Group = groupVersion[0]
					gvr.Version = groupVersion[1]
				}
				gvr.Resource = apiResource.Name
				watchResource[kind] = gvr
			}
		}
	}
	informerFactory := informers.NewSharedInformerFactory(clientSet, 0)
	for _, gvr := range watchResource {
		fmt.Printf("watch resource: %v\n", gvr)
		if _, err = informerFactory.ForResource(gvr); err != nil {
			return nil, nil, err
		}
	}
	// for _, gvr := range groupVersionResources {
	// 	if _, err = informerFactory.ForResource(gvr); err != nil {
	// 		return nil, nil, err
	// 	}
	// }

	ctx, cancel := context.WithCancel(context.Background())
	// Start all informers.
	informerFactory.Start(ctx.Done())
	// Wait for all caches to sync.
	informerFactory.WaitForCacheSync(ctx.Done())

	return informerFactory, cancel, nil
}

type store map[string]ClusterSet

type Cache struct {
	sync.RWMutex
	store
}

func NewClusterCache() *Cache {
	return &Cache{
		store: make(store),
	}
}

func (s *Cache) Get(name string) (ClusterSet, bool) {
	s.RLock()
	defer s.RUnlock()

	cluster, ok := s.store[name]
	return cluster, ok
}

func (s *Cache) GetConfig(name string) (*restclient.Config, bool) {
	s.RLock()
	defer s.RUnlock()

	clusterSet, ok := s.store[name]
	if !ok {
		return nil, false
	}
	return clusterSet.Config, true
}

func (s *Cache) GetClient(name string) (*kubernetes.Clientset, bool) {
	s.RLock()
	defer s.RUnlock()

	clusterSet, ok := s.store[name]
	if !ok {
		return nil, false
	}

	return clusterSet.Client, true
}

func (s *Cache) Set(name string, cs ClusterSet) {
	s.Lock()
	defer s.Unlock()

	if s.store == nil {
		s.store = store{}
	}
	s.store[name] = cs
}

func (s *Cache) Delete(name string) {
	s.Lock()
	defer s.Unlock()

	// Cancel informer
	cluster, ok := s.store[name]
	if !ok {
		return
	}
	cluster.Informer.Cancel()

	// 从缓存移除集群数据
	delete(s.store, name)
}

func (s *Cache) List() store {
	s.Lock()
	defer s.Unlock()

	return s.store
}

func (s *Cache) Clear() {
	s.Lock()
	defer s.Unlock()

	s.store = store{}
}
