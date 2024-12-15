package cluster

import (
	"context"
	"fmt"
	"io"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/release"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	pixiuTypes "github.com/caoyingjunz/pixiu/pkg/types"
)

const HELM_TOOL_BOX = "harbor.cloud.pixiuio.com/pixiuio/helm-toolbox:v3.9.0"

type IReleasesV2 interface {
	GetRelease(ctx context.Context, name string) (*release.Release, error)
	ListRelease(ctx context.Context) ([]*release.Release, error)
	InstallRelease(ctx context.Context, form *pixiuTypes.ReleaseForm) (*v1.Job, error)
	UninstallRelease(ctx context.Context, name string) (*release.UninstallReleaseResponse, error)
	UpgradeRelease(ctx context.Context, form *pixiuTypes.ReleaseForm) (*release.Release, error)
	GetReleaseHistory(ctx context.Context, name string) ([]*release.Release, error)
	RollbackRelease(ctx context.Context, name string, toVersion int) error
}

type ReleasesV2 struct {
	settings     *cli.EnvSettings
	actionConfig *action.Configuration
	clientSet    *kubernetes.Clientset
}

func newReleasesV2(actionCofnig *action.Configuration, settings *cli.EnvSettings, clientSet *kubernetes.Clientset) *ReleasesV2 {
	return &ReleasesV2{
		actionConfig: actionCofnig,
		settings:     settings,
		clientSet:    clientSet,
	}
}

var _ IReleasesV2 = &ReleasesV2{}

func (r *ReleasesV2) GetRelease(ctx context.Context, name string) (*release.Release, error) {
	client := action.NewGet(r.actionConfig)
	return client.Run(name)
}

func (r *ReleasesV2) ListRelease(ctx context.Context) ([]*release.Release, error) {
	client := action.NewList(r.actionConfig)
	return client.Run()
}

// install release
func (r *ReleasesV2) InstallRelease(ctx context.Context, form *pixiuTypes.ReleaseForm) (*v1.Job, error) {
	jobName := fmt.Sprintf("%s-%s", form.Name, "pixiu")
	cmd := []string{
		"helm",
		"install",
		"--repo", "https://charts.apiseven.com",
		form.Name,
		"apisix",
		"--namespace", r.settings.Namespace(),
		"--version", form.Version,
		"--wait",
		"--timeout", "600s",
		"--atomic",
		"--create-namespace",
		"--debug",
		"--wait-for-jobs",
	}

	return r.creteJob(jobName, cmd)
}

func (r *ReleasesV2) UninstallRelease(ctx context.Context, name string) (*release.UninstallReleaseResponse, error) {
	client := action.NewUninstall(r.actionConfig)
	return client.Run(name)
}

// upgrade release
func (r *ReleasesV2) UpgradeRelease(ctx context.Context, form *pixiuTypes.ReleaseForm) (*release.Release, error) {
	client := action.NewUpgrade(r.actionConfig)
	client.Namespace = r.settings.Namespace()
	client.DryRun = form.Preview
	if client.DryRun {
		client.Description = "server"
	}

	chrt, err := r.locateChart(client.ChartPathOptions, form.Chart, r.settings)
	if err != nil {
		return nil, err
	}

	out, err := client.Run(form.Name, chrt, form.Values)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *ReleasesV2) GetReleaseHistory(ctx context.Context, name string) ([]*release.Release, error) {
	client := action.NewHistory(r.actionConfig)
	return client.Run(name)
}

func (r *ReleasesV2) RollbackRelease(ctx context.Context, name string, toVersion int) error {
	klog.Error("version: ", toVersion)
	_, err := r.GetRelease(ctx, name)
	if err != nil {
		return err
	}

	client := action.NewRollback(r.actionConfig)
	client.Version = toVersion
	return client.Run(name)
}

func (r *ReleasesV2) locateChart(pathOpts action.ChartPathOptions, chart string, settings *cli.EnvSettings) (*chart.Chart, error) {
	// from cmd/helm/install.go and cmd/helm/upgrade.go
	cp, err := pathOpts.LocateChart(chart, settings)
	if err != nil {
		return nil, err
	}

	p := getter.All(settings)

	// Check chart dependencies to make sure all are present in /charts
	chartRequested, err := loader.Load(cp)
	if err != nil {
		return nil, err
	}

	if err := checkIfInstallable(chartRequested); err != nil {
		return nil, err
	}

	registryClient, err := registry.NewClient(
		registry.ClientOptDebug(false),
		//registry.ClientOptWriter(out),
		registry.ClientOptCredentialsFile(settings.RegistryConfig),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to crete helm config object %v", err)
	}

	if req := chartRequested.Metadata.Dependencies; req != nil {
		// If CheckDependencies returns an error, we have unfulfilled dependencies.
		// As of Helm 2.4.0, this is treated as a stopping condition:
		// https://github.com/helm/helm/issues/2209
		if err := action.CheckDependencies(chartRequested, req); err != nil {
			err = fmt.Errorf("an error occurred while checking for chart dependencies. You may need to run `helm dependency build` to fetch missing dependencies: %v", err)
			if true { // client.DependencyUpdate
				man := &downloader.Manager{
					Out:              io.Discard,
					ChartPath:        cp,
					Keyring:          pathOpts.Keyring,
					SkipUpdate:       false,
					Getters:          p,
					RepositoryConfig: settings.RepositoryConfig,
					RepositoryCache:  settings.RepositoryCache,
					Debug:            settings.Debug,
					RegistryClient:   registryClient, // added on top of Helm code
				}
				if err := man.Update(); err != nil {
					return nil, err
				}
				// Reload the chart with the updated Chart.lock file.
				if chartRequested, err = loader.Load(cp); err != nil {
					return nil, fmt.Errorf("failed reloading chart after repo update : %v", err)
				}
			} else {
				return nil, err
			}
		}
	}

	return chartRequested, nil
}

func (r *ReleasesV2) creteJob(name string, cmd []string) (*v1.Job, error) {
	job := &v1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: r.settings.Namespace(),
		},
		Spec: v1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":          name,
						"pixiu-charts": name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  name,
							Image: HELM_TOOL_BOX,
							Env: []corev1.EnvVar{
								{
									Name:  "HELM_KUBEAPISERVER",
									Value: r.settings.KubeAPIServer,
								},
								{
									Name:  "HELM_KUBECAFILE",
									Value: r.settings.KubeCaFile,
								},
								{
									Name:  "HELM_KUBEASUSER",
									Value: r.settings.KubeAsUser,
								},
								{
									Name:  "HELM_KUBETOKEN",
									Value: r.settings.KubeToken,
								},
							},
							Command:         cmd,
							ImagePullPolicy: corev1.PullIfNotPresent,
						},
					},
					RestartPolicy: corev1.RestartPolicyOnFailure,
				},
			},
		},
	}
	return r.clientSet.BatchV1().Jobs(r.settings.Namespace()).Create(context.TODO(), job, metav1.CreateOptions{})
}
