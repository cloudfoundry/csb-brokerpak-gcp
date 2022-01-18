package stackdrivertrace_test

import (
	"acceptancetests/helpers/apps"
	"acceptancetests/helpers/matchers"
	"acceptancetests/helpers/random"
	"acceptancetests/helpers/services"
	"context"
	"encoding/json"
	"fmt"
	"time"

	trace "cloud.google.com/go/trace/apiv1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/api/option"
	cloudtracepb "google.golang.org/genproto/googleapis/devtools/cloudtrace/v1"
)

var _ = Describe("Stackdrivertrace", func() {
	It("can emit app trace", func() {
		By("creating a service instance")
		serviceInstance := services.CreateInstance("csb-google-stackdriver-trace", "default")
		defer serviceInstance.Delete()

		By("pushing the unstarted app")
		appOne := apps.Push(apps.WithApp(apps.StackdriverTraceNode))
		defer apps.Delete(appOne)

		By("binding the app to the service instance")
		binding := serviceInstance.Bind(appOne)

		By("starting the apps")
		apps.Start(appOne)

		By("checking that the app environment has a credhub reference for credentials")
		Expect(binding.Credential()).To(matchers.HaveCredHubRef)

		By("triggering trace flush")
		customSpan := random.Hexadecimal()
		got := appOne.GET(customSpan)
		var traceResp struct {
			ProjectId string `json:"ProjectId"`
			TraceId   string `json:"TraceId"`
		}
		err := json.Unmarshal([]byte(got), &traceResp)
		Expect(err).NotTo(HaveOccurred())

		By("checking it got persisted in gcp")
		ctx := context.Background()
		traceClient, err := trace.NewClient(ctx, option.WithCredentialsJSON([]byte(GCPMetadata.Credentials)))
		Expect(err).NotTo(HaveOccurred())
		defer traceClient.Close()

		req := cloudtracepb.GetTraceRequest{
			ProjectId: traceResp.ProjectId,
			TraceId:   traceResp.TraceId,
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
