package acceptance_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"csbbrokerpakgcp/acceptance-tests/helpers/apps"
	"csbbrokerpakgcp/acceptance-tests/helpers/gcloud"
	"csbbrokerpakgcp/acceptance-tests/helpers/random"
	"csbbrokerpakgcp/acceptance-tests/helpers/services"
)

var _ = Describe("Storage Migration", Label("storage-migration"), func() {
	It("can migrate instance from the previous legacy broker to the CSB", func() {

		By("creating an Storage bucket using the legacy broker")

		sourceBucketName := random.Name(random.WithPrefix("old"), random.WithMaxLength(15))
		legacyServiceInstance := services.CreateInstance(
			"google-storage",
			"multiregional",
			services.WithBrokerName(legacyBrokerName),
			services.WithParameters(
				map[string]any{
					"name":         sourceBucketName,
					"force_delete": "true",
				},
			),
		)
		defer legacyServiceInstance.Delete()

		By("pushing an unstarted app")
		appOne := apps.Push(apps.WithApp(apps.Storage))
		defer apps.Delete(appOne)

		By("binding to the apps")
		legacyBinding := legacyServiceInstance.BindWithParams(appOne, `{"role":"storage.objectAdmin"}`)

		By("starting the app")
		apps.Start(appOne)

		By("uploading a blob using the app")
		blobNameOne := random.Hexadecimal()
		blobDataOne := random.Hexadecimal()
		appOne.PUT(blobDataOne, blobNameOne)

		By("downloading the blob using the app")
		got := appOne.GET(blobNameOne)
		Expect(got).To(Equal(blobDataOne))

		By("creating a new bucket using the CSB broker")
		destinationBucketName := random.Name(random.WithPrefix("csb"), random.WithMaxLength(15))
		destinationServiceInstance := services.CreateInstance(
			"csb-google-storage-bucket",
			"default",
			services.WithParameters(
				map[string]any{
					"name": destinationBucketName,
				},
			),
		)
		defer destinationServiceInstance.Delete()

		By("syncing up the legacy Storage bucket to the new one")
		_ = gcloud.GCP(
			"storage",
			"cp",
			"--recursive",
			fmt.Sprintf("gs://%s/*", sourceBucketName),
			fmt.Sprintf("gs://%s", destinationBucketName),
		)

		By("switching the app data source from the legacy bucket to the new one")
		legacyBinding.Unbind()
		newBinding := destinationServiceInstance.BindWithParams(appOne, `{"role":"storage.objectAdmin"}`)
		defer newBinding.Unbind()

		apps.Restart(appOne)

		By("checking existing data can be accessed using the new instance")
		Expect(appOne.GET(blobNameOne)).To(Equal(blobDataOne))

		By("deleting the blob from bucket")
		appOne.DELETE(blobNameOne)
	})
})
