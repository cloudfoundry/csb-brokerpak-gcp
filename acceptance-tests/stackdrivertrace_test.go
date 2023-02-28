package acceptance_test

import (
	"context"
	"fmt"
	"time"

	trace "cloud.google.com/go/trace/apiv1"
	cloudtracepb "cloud.google.com/go/trace/apiv1/tracepb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/api/option"

	"csbbrokerpakgcp/acceptance-tests/helpers/apps"
	"csbbrokerpakgcp/acceptance-tests/helpers/matchers"
	"csbbrokerpakgcp/acceptance-tests/helpers/random"
	"csbbrokerpakgcp/acceptance-tests/helpers/services"
)

var _ = Describe("Stackdrivertrace", Label("stackdrivertrace"), func() {
	It("can emit app trace", func() {
		By("creating a service instance")
		serviceInstance := services.CreateInstance("csb-google-stackdriver-trace", "default")
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
		var traceResp struct {
			ProjectID string `json:"ProjectId"`
			TraceID   string `json:"TraceId"`
		}
		appOne.GET(customSpan).Parse(&traceResp)

		By("checking it got persisted in gcp")
		ctx := context.Background()
		traceClient, err := trace.NewClient(ctx, option.WithCredentialsJSON([]byte(GCPMetadata.Credentials)))
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
	})
})
