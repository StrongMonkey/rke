package services

import (
	"github.com/docker/docker/api/types/container"
	"github.com/rancher/rke/docker"
	"github.com/rancher/rke/hosts"
	"github.com/rancher/rke/pki"
)

func runScheduler(host hosts.Host, schedulerService Scheduler) error {
	imageCfg, hostCfg := buildSchedulerConfig(host, schedulerService)
	return docker.DoRunContainer(host.DClient, imageCfg, hostCfg, SchedulerContainerName, host.Hostname, ControlRole)
}

func buildSchedulerConfig(host hosts.Host, schedulerService Scheduler) (*container.Config, *container.HostConfig) {
	imageCfg := &container.Config{
		Image: schedulerService.Image,
		Cmd: []string{"/hyperkube",
			"scheduler",
			"--v=2",
			"--address=0.0.0.0",
			"--kubeconfig=" + pki.KubeSchedulerConfigPath,
		},
	}
	hostCfg := &container.HostConfig{
		Binds: []string{
			"/etc/kubernetes:/etc/kubernetes",
		},
		RestartPolicy: container.RestartPolicy{Name: "always"},
	}
	return imageCfg, hostCfg
}
