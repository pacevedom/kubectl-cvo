package client

import (
	"context"
	"fmt"
	"strings"

	apiv1 "github.com/openshift/api/config/v1"
	clientconfigv1 "github.com/openshift/client-go/config/clientset/versioned/typed/config/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func sliceSubtraction(a, b []string) []string {
	mapValues := make(map[string]struct{})
	for _, x := range a {
		mapValues[x] = struct{}{}
	}
	for _, x := range b {
		delete(mapValues, x)
	}
	listValues := make([]string, 0, len(mapValues))
	for k := range mapValues {
		listValues = append(listValues, k)
	}
	return listValues
}

func extractCVNamespaceName(operator string) (string, string, string) {
	cvSplit := strings.Split(operator, ":")
	namespaceNameSplit := strings.Split(cvSplit[1], "/")
	return cvSplit[0], namespaceNameSplit[0], namespaceNameSplit[1]
}

type Client struct {
	clientConfig *clientconfigv1.ConfigV1Client
	clientKube   *kubernetes.Clientset
}

func NewClient(restConfig *rest.Config) (*Client, error) {
	configClient, err := getConfigClient(restConfig)
	if err != nil {
		return nil, err
	}
	kubeClient, err := getKubeClientSet(restConfig)
	if err != nil {
		return nil, err
	}

	return &Client{
		clientConfig: configClient,
		clientKube:   kubeClient,
	}, nil
}

func (c *Client) ListUnmanagedOperators() ([]string, error) {
	l, err := c.clientConfig.ClusterVersions().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	unmanaged := make([]string, 0)
	for _, version := range l.Items {
		for _, override := range version.Spec.Overrides {
			if override.Unmanaged {
				unmanaged = append(unmanaged, fmt.Sprintf("%s:%s/%s", version.Name, override.Namespace, override.Name))
			}
		}
	}
	return unmanaged, nil
}

func (c *Client) ListManagedOperators() ([]string, error) {
	deploymentList, err := c.listOwnedDeployments()
	if err != nil {
		return nil, err
	}
	unmanagedOperators, err := c.ListUnmanagedOperators()
	if err != nil {
		return nil, err
	}
	return sliceSubtraction(deploymentList, unmanagedOperators), err
}

func (c *Client) ManageOperator(operator string) error {
	cv, namespace, name := extractCVNamespaceName(operator)
	clusterVersion, err := c.clientConfig.ClusterVersions().Get(context.Background(), cv, metav1.GetOptions{})
	if err != nil {
		return err
	}
	index := 0
	for i, override := range clusterVersion.Spec.Overrides {
		if override.Kind == "Deployment" && override.Group == "apps" && override.Namespace == namespace && override.Name == name {
			index = i
			break
		}
	}
	clusterVersion.Spec.Overrides = append(clusterVersion.Spec.Overrides[:index], clusterVersion.Spec.Overrides[index+1:]...)
	_, err = c.clientConfig.ClusterVersions().Update(context.Background(), clusterVersion, metav1.UpdateOptions{})
	return err
}

func (c *Client) UnmanageOperator(operator string) error {
	cv, namespace, name := extractCVNamespaceName(operator)
	clusterVersion, err := c.clientConfig.ClusterVersions().Get(context.Background(), cv, metav1.GetOptions{})
	if err != nil {
		return err
	}
	clusterVersion.Spec.Overrides = append(clusterVersion.Spec.Overrides, apiv1.ComponentOverride{
		Kind:      "Deployment",
		Group:     "apps",
		Namespace: namespace,
		Name:      name,
		Unmanaged: true,
	})
	_, err = c.clientConfig.ClusterVersions().Update(context.Background(), clusterVersion, metav1.UpdateOptions{})
	return err
}

func (c *Client) listOwnedDeployments() ([]string, error) {
	deployments, err := c.clientKube.AppsV1().Deployments("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	deploymentList := make([]string, 0)
	for _, deployment := range deployments.Items {
		for _, ownerReference := range deployment.OwnerReferences {
			if ownerReference.APIVersion == "config.openshift.io/v1" && ownerReference.Kind == "ClusterVersion" {
				deploymentList = append(deploymentList, fmt.Sprintf("%s:%s/%s", ownerReference.Name, deployment.Namespace, deployment.Name))
				break
			}
		}
	}
	return deploymentList, nil
}

func getConfigClient(restConfig *rest.Config) (*clientconfigv1.ConfigV1Client, error) {
	return clientconfigv1.NewForConfig(restConfig)
}

func getKubeClientSet(restConfig *rest.Config) (*kubernetes.Clientset, error) {
	return kubernetes.NewForConfig(restConfig)
}
