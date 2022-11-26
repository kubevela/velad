package e2e_test

import (
	"context"
	"fmt"
	"github.com/oam-dev/kubevela/apis/core.oam.dev/common"
	"github.com/oam-dev/kubevela/apis/core.oam.dev/v1beta1"
	"github.com/oam-dev/kubevela/pkg/oam/util"
	"github.com/oam-dev/velad/pkg/apis"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os/exec"
	"runtime"
	"time"
)

var _ = Describe("Ingress Test", func() {
	Context("Test Traefik Ingress", func() {
		It("Test Traefik Ingress", func() {
			By("Create Application with gateway trait")
			ctx := context.Background()
			app := v1beta1.Application{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-ingress",
					Namespace: "default",
				},
				Spec: v1beta1.ApplicationSpec{
					Components: []common.ApplicationComponent{
						{
							Name: "test",
							Type: "webservice",
							Properties: util.Object2RawExtension(map[string]interface{}{
								"image": "crccheck/hello-world",
							}),
							Traits: []common.ApplicationTrait{
								{
									Type: "gateway",
									Properties: util.Object2RawExtension(map[string]interface{}{
										"domain": "testsvc.example.com",
										"http": map[string]interface{}{
											"/": 8000,
										},
									}),
								},
							},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, &app)).Should(Succeed())

			By("Check Ingress works")
			// We can't set Host header in http go client, so we have to use cURL here.
			// See https://github.com/golang/go/issues/7682
			port := "80"
			if runtime.GOOS != apis.GoosLinux {
				port = "8090"
			}
			Eventually(func(g Gomega) {
				curl := exec.Command("curl", "-H", "Host: testsvc.example.com", fmt.Sprintf("http://127.0.0.1:%s", port))
				g.Expect(curl.Args).Should(ContainElement("Host: testsvc.example.com"))
				output, err := curl.Output()
				g.Expect(err).Should(BeNil())
				g.Expect(string(output)).Should(ContainSubstring("Hello World"))
			}, 30*time.Second).Should(Succeed())

			By("Delete Application")
			Expect(k8sClient.Delete(ctx, &app)).Should(Succeed())
		})
	})
})
