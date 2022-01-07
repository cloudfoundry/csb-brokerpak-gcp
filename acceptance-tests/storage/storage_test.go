package storage_test

import (
	"acceptancetests/helpers/apps"
	"acceptancetests/helpers/matchers"
	"acceptancetests/helpers/random"
	"acceptancetests/helpers/services"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Storage", func() {
	It("can be accessed by an app", func() {
		By("creating a service instance")
		serviceInstance := services.CreateInstance("csb-google-storage-bucket", "private", services.WithParameters(map[string]interface{}{"role": "storage.objectViewer"}))
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
		got := appTwo.GET(blobName)
		Expect(got).To(Equal(blobData))

		appOne.DELETE(blobName)
	})
})
