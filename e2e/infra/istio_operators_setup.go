package infra

import (
	"fmt"
	"os"

	"github.com/onsi/gomega"

	"github.com/maistra/istio-workspace/test/shell"
)

// BuildOperator builds istio-workspace operator and pushes it to specified registry.
func BuildOperator() (registry string) {
	projectDir := shell.GetProjectDir()
	namespace := setOperatorNamespace()
	registry = SetDockerRegistryExternal()
	setDockerRepository(GetRepositoryName())
	<-shell.Execute(NewProjectCmd(namespace)).Done()
	EnablePullingImages(namespace)
	if RunsOnOpenshift {
		<-shell.ExecuteInDir(".", "bash", "-c", "docker login -u "+user+" -p $(oc whoami -t) "+registry).Done()
	}
	<-shell.ExecuteInDir(projectDir, "make", "docker-build", "docker-push", "bundle", "bundle-build", "bundle-push").Done()

	return
}

func InstallLocalOperator(namespace string) {
	SetDockerRegistryInternal()

	err := os.Setenv("OPERATOR_NAMESPACE", namespace)
	gomega.Expect(err).To(gomega.Not(gomega.HaveOccurred()))
	<-shell.ExecuteInDir(shell.GetProjectDir(), "make", "bundle-run").Done()
}

func setOperatorNamespace() (namespace string) {
	operatorNS := "istio-workspace-operator"
	err := os.Setenv("OPERATOR_NAMESPACE", operatorNS)
	gomega.Expect(err).To(gomega.Not(gomega.HaveOccurred()))

	return operatorNS
}

func GetClusterHost() string {
	if host, found := os.LookupEnv("IKE_CLUSTER_HOST"); found {
		return host
	}

	return "127.0.0.1.nip.io"
}

func GetIstioNamespace() string {
	if istioNs, found := os.LookupEnv("ISTIO_NS"); found {
		return istioNs
	}

	return "istio-system"
}

func GetIstioIngressHostname() string {
	if istioIngress, found := os.LookupEnv("IKE_ISTIO_INGRESS"); found {
		return istioIngress
	}

	cmd := shell.ExecuteInDir(".", "bash", "-c",
		fmt.Sprintf("kubectl get svc istio-ingressgateway -n %v -o jsonpath='{.status.loadBalancer.ingress[0].hostname}'", GetIstioNamespace()))
	<-cmd.Done()
	if cmd.Status().Exit == 0 && len(cmd.Status().Stdout) > 0 {
		return "http://" + cmd.Status().Stdout[0]
	}
	cmd = shell.ExecuteInDir(".", "bash", "-c", fmt.Sprintf("kubectl get route istio-ingressgateway -n %v -o jsonpath='{.spec.host}'", GetIstioNamespace()))
	<-cmd.Done()
	if cmd.Status().Exit == 0 && len(cmd.Status().Stdout) > 0 {
		return "http://" + cmd.Status().Stdout[0]
	}

	return "http://istio-ingressgateway-" + GetIstioNamespace() + "." + GetClusterHost()
}
