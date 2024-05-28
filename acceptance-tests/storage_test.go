package acceptance_test

import (
	"csbbrokerpakgcp/acceptance-tests/helpers/apps"
	"csbbrokerpakgcp/acceptance-tests/helpers/matchers"
	"csbbrokerpakgcp/acceptance-tests/helpers/random"
	"csbbrokerpakgcp/acceptance-tests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Storage", Label("storage"), func() {
	It("can be accessed by an app", func() {
		By("creating a service instance")
		serviceInstance := services.CreateInstance("csb-google-storage-bucket", "default")
		defer serviceInstance.Delete()

		By("pushing the unstarted app twice")
		appOne := apps.Push(apps.WithApp(apps.Storage))
		appTwo := apps.Push(apps.WithApp(apps.Storage))
		defer apps.Delete(appOne, appTwo)

		By("binding the apps to the storage service instance")
		binding := serviceInstance.BindWithParams(appOne, `{"role":"storage.objectAdmin"}`)
		serviceInstance.BindWithParams(appTwo, `{"role":"storage.objectAdmin"}`)

		By("starting the apps")
		apps.Start(appOne, appTwo)

		By("checking that the app environment has a credhub reference for credentials")
		Expect(binding.Credential()).To(matchers.HaveCredHubRef)

		By("uploading a blob using the first app")
		blobName := random.Hexadecimal()
		blobData := random.Hexadecimal()
		appOne.PUT(blobData, blobName)

		By("downloading the blob using the second app")
		got := appTwo.GET(blobName).String()
		Expect(got).To(Equal(blobData))

		appOne.DELETE(blobName)
	})

	It("works with Spring and  GCP Cloud libraries", Label("spring"), func() {
		By("creating a service instance")
		bucketName := random.Name(random.WithPrefix("csb"))
		serviceInstance := services.CreateInstance(
			"csb-google-storage-bucket",
			"default",
			services.WithParameters(map[string]any{
				"name": bucketName,
			}),
		)
		defer serviceInstance.Delete()

		By("pushing the unstarted app twice")
		appOne := apps.Push(apps.WithApp(apps.SpringStorageApp), apps.WithTestAppManifest(apps.StorageTestAppManifest))
		appTwo := apps.Push(apps.WithApp(apps.SpringStorageApp), apps.WithTestAppManifest(apps.StorageTestAppManifest))
		defer apps.Delete(appOne, appTwo)

		By("binding the apps to the storage service instance")
		binding := serviceInstance.BindWithParams(appOne, `{"role":"storage.objectAdmin"}`)
		serviceInstance.BindWithParams(appTwo, `{"role":"storage.objectAdmin"}`)

		By("starting the apps")
		apps.Start(appOne, appTwo)

		By("checking that the app environment has a credhub reference for credentials")
		Expect(binding.Credential()).To(matchers.HaveCredHubRef)

		By("uploading a blob using the first app")
		blobName := random.Hexadecimal()
		blobData := random.Hexadecimal()
		appOne.POST(blobData, "/storage/write?bucketName=%s&objectName=%s", bucketName, blobName)

		By("downloading the blob using the second app")
		got := appTwo.GET("/storage/read?bucketName=%s&objectName=%s", bucketName, blobName).String()
		Expect(got).To(Equal(blobData))

		appOne.DELETE("/storage/delete?bucketName=%s&objectName=%s", bucketName, blobName)
	})
})
