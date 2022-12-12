package integration_test

import (
	"fmt"

	testframework "github.com/cloudfoundry/cloud-service-broker/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

const (
	mySQLServiceName    = "csb-google-mysql"
	customMySQLPlanName = "custom-plan"
)

var customMySQLPlans = []map[string]any{
	customMySQLPlan,
}

var customMySQLPlan = map[string]any{
	"name":          customMySQLPlanName,
	"id":            "9daa07f1-78e8-4bda-9efe-91576102c30d",
	"description":   "custom plan defined by customer",
	"mysql_version": "MYSQL_5_7",
	"metadata": map[string]any{
		"displayName": "custom plan defined by customer (beta)",
	},
}

var _ = Describe("Mysql", Label("MySQL"), func() {
	BeforeEach(func() {
		Expect(mockTerraform.SetTFState([]testframework.TFStateValue{})).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(mockTerraform.Reset()).To(Succeed())
	})

	It("should publish mysql in the catalog", func() {
		catalog, err := broker.Catalog()
		Expect(err).NotTo(HaveOccurred())

		service := testframework.FindService(catalog, mySQLServiceName)
		Expect(service.ID).NotTo(BeNil())
		Expect(service.Name).NotTo(BeNil())
		Expect(service.Tags).To(ConsistOf("gcp", "mysql", "beta"))
		Expect(service.Metadata.ImageUrl).NotTo(BeNil())
		Expect(service.Metadata.DisplayName).NotTo(BeNil())
		Expect(service.Plans).To(
			ConsistOf(
				MatchFields(IgnoreExtras, Fields{"Name": Equal("small")}),
				MatchFields(IgnoreExtras, Fields{"Name": Equal("medium")}),
				MatchFields(IgnoreExtras, Fields{"Name": Equal("large")}),
				MatchFields(IgnoreExtras, Fields{"Name": Equal("custom-plan")}),
			),
		)
	})

	Describe("provisioning", func() {
		It("should provision a plan", func() {
			instanceID, err := broker.Provision(mySQLServiceName, customMySQLPlanName, map[string]any{"tier": "db-n1-standard-1"})

			Expect(err).NotTo(HaveOccurred())
			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("db_name", "csb-db"),
					HaveKeyWithValue("authorized_network", "default"),
					HaveKeyWithValue("authorized_network_id", ""),
					HaveKeyWithValue("credentials", "broker-gcp-creds"),
					HaveKeyWithValue("database_version", "MYSQL_5_7"),
					HaveKeyWithValue("db_name", "csb-db"),
					HaveKeyWithValue("instance_name", "csb-mysql-"+instanceID),
					HaveKeyWithValue("project", "broker-gcp-project"),
					HaveKeyWithValue("region", "us-central1"),
					HaveKeyWithValue("storage_gb", BeNumerically("==", 10)),

					HaveKeyWithValue("tier", "db-n1-standard-1"),
					HaveKeyWithValue("disk_autoresize", true),
					HaveKeyWithValue("disk_autoresize_limit", BeNumerically("==", 0)),
					HaveKeyWithValue("deletion_protection", false),
				),
			)
		})

		It("should allow setting properties do not defined in the plan", func() {
			_, err := broker.Provision(mySQLServiceName, customMySQLPlanName, map[string]any{
				"credentials":           "fake-credentials",
				"project":               "fake-project",
				"instance_name":         "fakeinstancename",
				"db_name":               "fake-db_name",
				"region":                "asia-northeast1",
				"authorized_network":    "fake-authorized_network",
				"authorized_network_id": "fake-authorized_network_id",
				"tier":                  "fake-tier",
				"disk_autoresize":       true,
				"disk_autoresize_limit": 400,
				"deletion_protection":   true,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("credentials", "fake-credentials"),
					HaveKeyWithValue("project", "fake-project"),
					HaveKeyWithValue("instance_name", "fakeinstancename"),
					HaveKeyWithValue("db_name", "fake-db_name"),
					HaveKeyWithValue("region", "asia-northeast1"),
					HaveKeyWithValue("authorized_network", "fake-authorized_network"),
					HaveKeyWithValue("authorized_network_id", "fake-authorized_network_id"),
					HaveKeyWithValue("tier", "fake-tier"),
					HaveKeyWithValue("disk_autoresize", true),
					HaveKeyWithValue("disk_autoresize_limit", BeNumerically("==", 400)),
					HaveKeyWithValue("deletion_protection", BeTrue()),
				),
			)
		})

		It("should not allow changing of plan defined properties", func() {
			_, err := broker.Provision(mySQLServiceName, "small", map[string]any{"storage_gb": 44})

			Expect(err).To(MatchError(ContainSubstring("plan defined properties cannot be changed: storage_gb")))
		})

		DescribeTable("property constraints",
			func(params map[string]any, expectedErrorMsg string) {
				_, err := broker.Provision(mySQLServiceName, customMySQLPlanName, params)

				Expect(err).To(MatchError(ContainSubstring(expectedErrorMsg)))
			},
			Entry(
				"storage capacity minimum value is 10",
				map[string]any{"storage_gb": 9},
				"storage_gb: Must be greater than or equal to 10",
			),
			Entry(
				"instance name maximum length is 98 characters",
				map[string]any{"instance_name": stringOfLen(99)},
				"instance_name: String length must be less than or equal to 98",
			),
			Entry(
				"instance name minimum length is 6 characters",
				map[string]any{"instance_name": stringOfLen(5)},
				"instance_name: String length must be greater than or equal to 6",
			),
			Entry(
				"instance name invalid characters",
				map[string]any{"instance_name": ".aaaaa"},
				"instance_name: Does not match pattern '^[a-z][a-z0-9-]+$'",
			),
			Entry(
				"database name maximum length is 64 characters",
				map[string]any{"db_name": stringOfLen(65)},
				"db_name: String length must be less than or equal to 64",
			),
			Entry(
				"invalid region",
				map[string]any{"region": "-Asia-northeast1"},
				"region: Does not match pattern '^[a-z][a-z0-9-]+$'",
			),
			Entry(
				"tier invalid characters",
				map[string]any{"tier": ".aaaaa"},
				"tier: Does not match pattern '^[a-z][a-z0-9-]+$'",
			),
		)
	})

	Describe("updating instance", func() {
		var instanceID string

		BeforeEach(func() {
			var err error
			instanceID, err = broker.Provision(mySQLServiceName, customMySQLPlanName, map[string]any{"tier": "db-n1-standard-1"})

			Expect(err).NotTo(HaveOccurred())
		})

		DescribeTable("should allow updating properties not flagged as `prohibit_update` and not specified in the plan",
			func(params map[string]any) {
				err := broker.Update(instanceID, mySQLServiceName, customMySQLPlanName, params)

				Expect(err).NotTo(HaveOccurred())
			},
			Entry("update credentials", map[string]any{"credentials": "other-credentials"}),
			Entry("update project", map[string]any{"project": "another-project"}),
			Entry("update tier", map[string]any{"tier": "db-n1-standard-16"}),
			Entry("update disk_autoresize", map[string]any{"disk_autoresize": true}),
			Entry("update disk_autoresize_limit", map[string]any{"disk_autoresize_limit": 400}),
			Entry("update storage_gb", map[string]any{"storage_gb": 100}),
			Entry("update deletion_protection", map[string]any{"deletion_protection": true}),
		)

		DescribeTable("should prevent updating properties flagged as `prohibit_update` because it can result in the recreation of the service instance and lost data",
			func(params map[string]any) {
				err := broker.Update(instanceID, mySQLServiceName, customMySQLPlanName, params)

				Expect(err).To(MatchError(
					ContainSubstring(
						"attempt to update parameter that may result in service instance re-creation and data loss",
					),
				))

				const initialProvisionInvocation = 1
				Expect(mockTerraform.ApplyInvocations()).To(HaveLen(initialProvisionInvocation))
			},
			Entry("update instance_name", map[string]any{"instance_name": "another-instance-name"}),
			Entry("update db_name", map[string]any{"db_name": "another-db-name"}),
			Entry("update region", map[string]any{"region": "australia-southeast1"}),
			Entry("update authorized_network", map[string]any{"authorized_network": "another-authorized-network"}),
			Entry("update authorized_network_id", map[string]any{"authorized_network_id": "another-authorized-network_id"}),
		)

		DescribeTable("should not allow updating properties that are specified in the plan",
			func(key string, value any) {
				err := broker.Update(instanceID, mySQLServiceName, customMySQLPlanName, map[string]any{key: value})

				Expect(err).To(
					MatchError(
						ContainSubstring(
							fmt.Sprintf("plan defined properties cannot be changed: %s", key),
						),
					),
				)
			},
			Entry("update mysql_version", "mysql_version", "MYSQL_5_7"),
		)

		DescribeTable("should not allow updating additional properties",
			func(key string, value any) {
				err := broker.Update(instanceID, mySQLServiceName, customMySQLPlanName, map[string]any{key: value})

				Expect(err).To(
					MatchError(
						ContainSubstring(
							fmt.Sprintf("additional properties are not allowed: %s", key),
						),
					),
				)
			},
			Entry("update name", "name", "fake-name"),
			Entry("update id", "id", "fake-id"),
		)
	})

})
