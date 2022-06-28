package upgrade_test

import (
	"context"
	"csbbrokerpakgcp/acceptance-tests/helpers/apps"
	"csbbrokerpakgcp/acceptance-tests/helpers/brokers"
	"csbbrokerpakgcp/acceptance-tests/helpers/matchers"
	"csbbrokerpakgcp/acceptance-tests/helpers/random"
	"csbbrokerpakgcp/acceptance-tests/helpers/services"
	"encoding/json"
	"fmt"
	"time"

	trace "cloud.google.com/go/trace/apiv1"
	"google.golang.org/api/option"
	cloudtracepb "google.golang.org/genproto/googleapis/devtools/cloudtrace/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UpgradeStackdrivertraceTest", Label("stackdrivertrace"), func() {
	When("upgrading broker version", func() {
		It("should continue to work", func() {
			By("pushing latest released broker version")
			serviceBroker := brokers.Create(
				brokers.WithPrefix("csb-stackdrivertrace"),
				brokers.WithSourceDir(releasedBuildDir),
				brokers.WithReleaseEnv(),
			)
			defer serviceBroker.Delete()

			By("creating a service instance")
			serviceInstance := services.CreateInstance("csb-google-stackdriver-trace", "default", services.WithBroker(serviceBroker))
			defer serviceInstance.Delete()

			By("pushing the unstarted app")
			appOne := apps.Push(apps.WithApp(apps.StackdriverTrace))
			defer apps.Delete(appOne)

			By("binding the app to the service instance")
			binding := serviceInstance.Bind(appOne)

			By("starting the apps")
			apps.Start(appOne)

			By("checking that the app environment has a credhub reference for credentials")
			Expect(binding.Credential()).To(matchers.HaveCredHubRef)

			By("triggering trace flush")
			customSpan := random.Hexadecimal()
			response := appOne.GET(customSpan)

			var traceResp struct {
				ProjectID string `json:"ProjectId"`
				TraceID   string `json:"TraceId"`
			}
			err := json.Unmarshal([]byte(response), &traceResp)
			Expect(err).NotTo(HaveOccurred())

			By("checking it got persisted in gcp")
			ctx := context.Background()
			traceClient, err := trace.NewClient(ctx, option.WithCredentialsJSON([]byte(metadata.Credentials)))
			Expect(err).NotTo(HaveOccurred())
			defer traceClient.Close()

			req := cloudtracepb.GetTraceRequest{
				ProjectId: traceResp.ProjectID,
				TraceId:   traceResp.TraceID,
			}

			returnedSpanName := func() string {
				resp, err := traceClient.GetTrace(ctx, &req)
				if err != nil {
					return ""
				}
				return resp.Spans[0].Name
			}

			Eventually(returnedSpanName, 6*time.Second).Should(Equal(fmt.Sprintf("/%s", customSpan)))

			By("pushing the development version of the broker")
			serviceBroker.UpdateBroker(developmentBuildDir)

			By("upgrading service instance")
			serviceInstance.Upgrade()

			By("unbinding works")
			binding.Unbind()
		})
	})
})
