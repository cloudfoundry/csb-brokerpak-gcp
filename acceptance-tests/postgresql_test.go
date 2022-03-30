package acceptance_test

import (
	"csbbrokerpakgcp/acceptance-tests/helpers/apps"
	"csbbrokerpakgcp/acceptance-tests/helpers/matchers"
	"csbbrokerpakgcp/acceptance-tests/helpers/random"
	"csbbrokerpakgcp/acceptance-tests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("PostgreSQL", Label("postgresql"), func() {
	It("can be accessed by an app", func() {
		By("creating a service instance")
		serviceInstance := services.CreateInstance("csb-google-postgres", "small")
		defer serviceInstance.Delete()

		By("pushing the unstarted app twice")
		appOne := apps.Push(apps.WithApp(apps.PostgreSQL))
		appTwo := apps.Push(apps.WithApp(apps.PostgreSQL))
		defer apps.Delete(appOne, appTwo)

		By("binding the apps to the service instance")
		binding := serviceInstance.Bind(appOne)
		serviceInstance.Bind(appTwo)

		By("starting the apps")
		apps.Start(appOne, appTwo)

		By("checking that the app environment has a credhub reference for credentials")
		Expect(binding.Credential()).To(matchers.HaveCredHubRef)

		By("creating a schema using the first app")
		schema := random.Name(random.WithMaxLength(10))
		appOne.PUT("", schema)

		By("setting a key-value using the first app")
		key := random.Hexadecimal()
		value := random.Hexadecimal()
		appOne.PUT(value, "%s/%s", schema, key)
		
		By("getting the value using the second app")
		got := appTwo.GET("%s/%s", schema, key)
		Expect(got).To(Equal(value))

		By("triggering ownership of schema to pass to provision user")
		binding.Unbind()

		By("getting the value again using the second app")
		got2 := appTwo.GET("%s/%s", schema, key)
		Expect(got2).To(Equal(value))

		By("setting another value using the second app")
		key2 := random.Hexadecimal()
		value2 := random.Hexadecimal()
		appTwo.PUT(value2, "%s/%s", schema, key2)

		By("getting the other value using the second app")
		got3 := appTwo.GET("%s/%s", schema, key2)
		Expect(got3).To(Equal(value2))

		By("dropping the schema using the second app")
		appTwo.DELETE(schema)
	})
})
