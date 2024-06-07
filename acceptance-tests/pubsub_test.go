package acceptance_test

import (
	"context"
	"csbbrokerpakgcp/acceptance-tests/helpers/apps"
	"csbbrokerpakgcp/acceptance-tests/helpers/matchers"
	"csbbrokerpakgcp/acceptance-tests/helpers/random"
	"csbbrokerpakgcp/acceptance-tests/helpers/services"
	"fmt"
	"log"
	"time"

	dataflow "cloud.google.com/go/dataflow/apiv1beta3"
	"cloud.google.com/go/dataflow/apiv1beta3/dataflowpb"
	"google.golang.org/api/option"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("PubSub", Label("pubsub"), func() {
	It("can be accessed by an app", func() {
		By("creating a service instance")
		serviceInstance := services.CreateInstance(
			"csb-google-pubsub",
			"default",
			services.WithParameters(map[string]any{"subscription_name": random.Name()}))
		defer serviceInstance.Delete()

		By("pushing the unstarted app twice")
		publisherApp := apps.Push(apps.WithApp(apps.PubSubApp))
		subscriberApp := apps.Push(apps.WithApp(apps.PubSubApp))
		defer apps.Delete(publisherApp, subscriberApp)

		By("binding the apps to the pubsub service instance")
		binding := serviceInstance.BindWithParams(publisherApp, `{"role":"pubsub.editor"}`)
		serviceInstance.BindWithParams(subscriberApp, `{"role":"pubsub.editor"}`)

		By("starting the apps")
		apps.Start(publisherApp, subscriberApp)

		By("checking that the app environment has a credhub reference for credentials")
		Expect(binding.Credential()).To(matchers.HaveCredHubRef)

		By("publishing a message with the publisher app")
		messageData := random.Hexadecimal()
		publisherApp.PUT(messageData, "")

		By("retrieving a message with the subscriber app")
		got := subscriberApp.GET("").String()
		Expect(got).To(Equal(messageData), "Received message matched published message")
	})

	When("migrating from the legacy broker", func() {
		It("can receive legacy topic messages in the new CSB topic", func() {
			By("creating a legacy service instance")
			legacySubscription := random.Name()
			legacyInstance := services.CreateInstance(
				"google-pubsub",
				"default",
				services.WithBrokerName("legacy-gcp-broker"),
				services.WithParameters(map[string]any{"subscription_name": legacySubscription}))
			defer legacyInstance.Delete()

			By("pushing the publisher app")
			publisherApp := apps.Push(apps.WithApp(apps.PubSubApp))
			defer apps.Delete(publisherApp)

			By("binding the publisher app to legacy instance")
			legacyInstance.BindWithParams(publisherApp, `{"role":"pubsub.editor"}`)
			apps.Start(publisherApp)

			By("publishing a message with the publisher app")
			messageData := random.Hexadecimal()
			publisherApp.PUT(messageData, "")

			By("creating a CSB service instance")
			CSBServiceInstance := services.CreateInstance(
				"csb-google-pubsub",
				"default",
				services.WithParameters(map[string]any{"subscription_name": random.Name()}))
			defer CSBServiceInstance.Delete()

			By("pushing a subscriber app")
			subscriberApp := apps.Push(apps.WithApp(apps.PubSubApp))
			defer apps.Delete(subscriberApp)

			By("binding the subscriber app to CSB instance")
			CSBServiceInstance.BindWithParams(subscriberApp, `{"role":"pubsub.editor"}`)
			apps.Start(subscriberApp)

			By("starting a job that moves messages from legacy topic to new CSB topic")
			ctx := context.Background()
			templatesClient, err := dataflow.NewTemplatesClient(ctx, option.WithCredentialsJSON([]byte(GCPMetadata.Credentials)))
			Expect(err).ToNot(HaveOccurred())
			defer templatesClient.Close()

			jobResponse, err := templatesClient.CreateJobFromTemplate(ctx, &dataflowpb.CreateJobFromTemplateRequest{
				ProjectId: GCPMetadata.Project,
				JobName:   "migration-job-" + random.Name(),
				Template: &dataflowpb.CreateJobFromTemplateRequest_GcsPath{
					GcsPath: "gs://dataflow-templates-us-central1/latest/Cloud_PubSub_to_Cloud_PubSub",
				},
				Parameters: map[string]string{
					"inputSubscription": fmt.Sprintf("projects/%s/subscriptions/%s", GCPMetadata.Project, legacySubscription),
					"outputTopic":       fmt.Sprintf("projects/%s/topics/%s", GCPMetadata.Project, fmt.Sprintf("csb-topic-%s", CSBServiceInstance.GUID())),
				},
				Environment: &dataflowpb.RuntimeEnvironment{
					TempLocation:          "gs://test-migrate-pubsub/temp",
					AdditionalExperiments: []string{"streaming_mode_exactly_once"},
				},
				Location: "us-central1",
			})
			Expect(err).ToNot(HaveOccurred())
			log.Printf("Started migration job: %s \n", jobResponse.Name)
			defer setJobToDoneState(ctx, jobResponse.Id)

			// It takes a few minutes (~3min) for the DataFlow job to pick up the message from legacy topic and move it to the new one
			By("retrieving a message with the subscriber app")
			Eventually(func() string {
				return subscriberApp.GET("").String()
			}).WithTimeout(5 * time.Minute).WithPolling(30 * time.Second).Should(Equal(messageData))
		})
	})

})

func setJobToDoneState(context context.Context, jobID string) {
	jobClient, err := dataflow.NewJobsV1Beta3Client(context, option.WithCredentialsJSON([]byte(GCPMetadata.Credentials)))
	Expect(err).ToNot(HaveOccurred())
	defer jobClient.Close()

	_, err = jobClient.UpdateJob(context, &dataflowpb.UpdateJobRequest{
		ProjectId: GCPMetadata.Project,
		JobId:     jobID,
		Job: &dataflowpb.Job{
			Id:             jobID,
			ProjectId:      GCPMetadata.Project,
			RequestedState: dataflowpb.JobState_JOB_STATE_DONE,
		},
		Location: "us-central1",
	})
	Expect(err).ToNot(HaveOccurred())
	_ = jobClient.Close()
}
