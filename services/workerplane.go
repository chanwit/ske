package services

import (
	"github.com/rancher/rke/hosts"
	"github.com/rancher/types/apis/cluster.cattle.io/v1"
	"github.com/sirupsen/logrus"
)

func RunWorkerPlane(controlHosts []hosts.Host, workerHosts []hosts.Host, workerServices v1.RKEConfigServices) error {
	logrus.Infof("[%s] Building up Worker Plane..", WorkerRole)
	for _, host := range controlHosts {
		// only one master for now
		err := runKubelet(host, workerServices.Kubelet, true)
		if err != nil {
			return err
		}
		err = runKubeproxy(host, workerServices.Kubeproxy)
		if err != nil {
			return err
		}
	}
	for _, host := range workerHosts {
		// run nginx proxy
		err := runNginxProxy(host, controlHosts)
		if err != nil {
			return err
		}
		// run kubelet
		err = runKubelet(host, workerServices.Kubelet, false)
		if err != nil {
			return err
		}
		// run kubeproxy
		err = runKubeproxy(host, workerServices.Kubeproxy)
		if err != nil {
			return err
		}
	}
	logrus.Infof("[%s] Successfully started Worker Plane..", WorkerRole)
	return nil
}

func RemoveWorkerPlane(controlHosts []hosts.Host, workerHosts []hosts.Host) error {
	logrus.Infof("[%s] Tearing down Worker Plane..", WorkerRole)
	for _, host := range controlHosts {
		err := removeKubelet(host)
		if err != nil {
			return err
		}
		err = removeKubeproxy(host)
		if err != nil {
			return err
		}
	}

	for _, host := range workerHosts {
		err := removeKubelet(host)
		if err != nil {
			return err
		}
		err = removeKubeproxy(host)
		if err != nil {
			return err
		}
		err = removeNginxProxy(host)
		if err != nil {
			return err
		}
	}
	logrus.Infof("[%s] Successfully teared down Worker Plane..", WorkerRole)
	return nil
}