package services

import (
	"github.com/Sirupsen/logrus"
	"github.com/rancher/rke/hosts"
)

func RunControlPlane(controlHosts []hosts.Host, etcdHosts []hosts.Host, controlServices Services) error {
	logrus.Infof("[%s] Building up Controller Plane..", ControlRole)
	for _, host := range controlHosts {
		// run kubeapi
		err := runKubeAPI(host, etcdHosts, controlServices.KubeAPI)
		if err != nil {
			return err
		}
		// run kubecontroller
		err = runKubeController(host, controlServices.KubeController)
		if err != nil {
			return err
		}
		// run scheduler
		err = runScheduler(host, controlServices.Scheduler)
		if err != nil {
			return err
		}
	}
	logrus.Infof("[%s] Successfully started Controller Plane..", ControlRole)
	return nil
}
