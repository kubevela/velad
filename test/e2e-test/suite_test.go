package e2e_test

import (
	"github.com/oam-dev/kubevela/pkg/utils/common"
	"github.com/oam-dev/velad/pkg/utils"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	k8sClient client.Client
)

func TestE2eTest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2eTest Suite")
}

var _ = BeforeSuite(func() {
	By("bootstrapping test environment")
	configPath := utils.GetDefaultVelaDKubeconfigPath()
	_ = os.Setenv(clientcmd.RecommendedConfigPathEnvVar, configPath)
	cfg, err := config.GetConfig()
	Expect(err).Should(BeNil())

	scheme := common.Scheme

	k8sClient, err = client.New(cfg, client.Options{
		Scheme: scheme,
	})
	Expect(err).Should(BeNil())
})
