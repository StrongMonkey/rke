package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/rancher/rke/cluster"
	"github.com/rancher/rke/pki"
	"github.com/urfave/cli"
	"k8s.io/client-go/util/cert"
)

func ClusterCommand() cli.Command {
	clusterUpFlags := []cli.Flag{
		cli.StringFlag{
			Name:   "cluster-file",
			Usage:  "Specify an alternate cluster YAML file",
			Value:  "cluster.yml",
			EnvVar: "CLUSTER_FILE",
		},
		cli.StringFlag{
			Name:   "auth-type",
			Usage:  "Specify authentication type",
			Value:  "x509",
			EnvVar: "AUTH_TYPE",
		},
	}
	return cli.Command{
		Name:      "cluster",
		ShortName: "cluster",
		Usage:     "Operations on the cluster",
		Flags:     clusterUpFlags,
		Subcommands: []cli.Command{
			cli.Command{
				Name:   "up",
				Usage:  "Bring the cluster up",
				Action: clusterUpFromCli,
				Flags:  clusterUpFlags,
			},
		},
	}
}

func ClusterUp(clusterFile, authType string) (string, string, string, string, error) {
	logrus.Infof("Building Kubernetes cluster")
	var ApiURL, caCrt, clientCert, clientKey string
	kubeCluster, err := cluster.ParseConfig(clusterFile)
	if err != nil {
		return ApiURL, caCrt, clientCert, clientKey, err
	}

	err = kubeCluster.TunnelHosts()
	if err != nil {
		return ApiURL, caCrt, clientCert, clientKey, err
	}

	currentCluster, err := kubeCluster.GetClusterState()
	if err != nil {
		return ApiURL, caCrt, clientCert, clientKey, err
	}

	err = cluster.SetUpAuthentication(kubeCluster, currentCluster, authType)
	if err != nil {
		return ApiURL, caCrt, clientCert, clientKey, err
	}

	err = kubeCluster.SetUpHosts(authType)
	if err != nil {
		return ApiURL, caCrt, clientCert, clientKey, err
	}

	err = kubeCluster.DeployClusterPlanes()
	if err != nil {
		return ApiURL, caCrt, clientCert, clientKey, err
	}

	err = kubeCluster.SaveClusterState(clusterFile)
	if err != nil {
		return ApiURL, caCrt, clientCert, clientKey, err
	}
	ApiURL = fmt.Sprintf("https://" + kubeCluster.ControlPlaneHosts[0].IP + ":6443")
	caCrt = string(cert.EncodeCertPEM(kubeCluster.Certificates[pki.CACertName].Certificate))
	clientCert = string(cert.EncodeCertPEM(kubeCluster.Certificates[pki.KubeAdminCommonName].Certificate))
	clientKey = string(cert.EncodePrivateKeyPEM(kubeCluster.Certificates[pki.KubeAdminCommonName].Key))
	return ApiURL, caCrt, clientCert, clientKey, nil
}

func clusterUpFromCli(ctx *cli.Context) error {
	authType := ctx.String("auth-type")
	clusterFile, err := resolveClusterFile(ctx)
	if err != nil {
		return fmt.Errorf("Failed to resolve cluster file: %v", err)
	}
	_, _, _, _, err = ClusterUp(clusterFile, authType)
	return err
}

func resolveClusterFile(ctx *cli.Context) (string, error) {
	clusterFile := ctx.String("cluster-file")
	fp, err := filepath.Abs(clusterFile)
	if err != nil {
		return "", fmt.Errorf("failed to lookup current directory name: %v", err)
	}
	file, err := os.Open(fp)
	if err != nil {
		return "", fmt.Errorf("Can not find cluster configuration file: %v", err)
	}
	defer file.Close()
	buf, err := ioutil.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}
	clusterFile = string(buf)

	return clusterFile, nil
}
