package integration_test

import (
	"fmt"

	testframework "github.com/cloudfoundry/cloud-service-broker/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var customMySQLPlans = []map[string]any{
	customMySQLPlan,
}

var customMySQLPlan = map[string]any{
	"name":          "custom-plan",
	"id":            "9daa07f1-78e8-4bda-9efe-91576102c30d",
	"description":   "custom plan defined by customer",
	"mysql_version": "MYSQL_5_7",
	"require_ssl":   false,
	"metadata": map[string]any{
		"displayName": "custom plan defined by customer (beta)",
	},
}

var _ = Describe("Mysql", func() {
	AfterEach(func() {
		Expect(mockTerraform.Reset()).NotTo(HaveOccurred())
	})

	It("should publish mysql in the catalog", func() {
		catalog, err := broker.Catalog()
		Expect(err).NotTo(HaveOccurred())

		service := testframework.FindService(catalog, "csb-google-mysql")
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
		It("should provision small plan", func() {
			instanceID, _ := broker.Provision("csb-google-mysql", "small", nil)

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("db_name", "csb-db"),
					HaveKeyWithValue("authorized_network", "default"),
					HaveKeyWithValue("authorized_network_id", ""),
					HaveKeyWithValue("cores", float64(2)),
					HaveKeyWithValue("credentials", "broker-gcp-creds"),
					HaveKeyWithValue("database_version", "MYSQL_5_7"),
					HaveKeyWithValue("db_name", "csb-db"),
					HaveKeyWithValue("instance_name", "csb-mysql-"+instanceID),
					HaveKeyWithValue("project", "broker-gcp-project"),
					HaveKeyWithValue("region", "us-central1"),
					HaveKeyWithValue("storage_gb", float64(10)),
				),
			)
		})

		It("should allow setting properties do not defined in the plan", func() {
			broker.Provision("csb-google-mysql", "small", map[string]any{
				"credentials":           "fake-credentials",
				"project":               "fake-project",
				"instance_name":         "fakeinstancename",
				"db_name":               "fake-db_name",
				"region":                "asia-northeast1",
				"authorized_network":    "fake-authorized_network",
				"authorized_network_id": "fake-authorized_network_id",
			})

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("credentials", "fake-credentials"),
					HaveKeyWithValue("project", "fake-project"),
					HaveKeyWithValue("instance_name", "fakeinstancename"),
					HaveKeyWithValue("db_name", "fake-db_name"),
					HaveKeyWithValue("region", "asia-northeast1"),
					HaveKeyWithValue("authorized_network", "fake-authorized_network"),
					HaveKeyWithValue("authorized_network_id", "fake-authorized_network_id"),
				),
			)
		})

		It("should not allow changing of plan defined properties", func() {
			_, err := broker.Provision("csb-google-mysql", "small", map[string]interface{}{"cores": 5})

			Expect(err).To(MatchError(ContainSubstring("plan defined properties cannot be changed: cores")))
		})

		DescribeTable("property constraints",
			func(params map[string]any, expectedErrorMsg string) {
				_, err := broker.Provision("csb-google-mysql", customMySQLPlan["name"].(string), params)

				Expect(err).To(MatchError(ContainSubstring(expectedErrorMsg)))
			},
			Entry(
				"cores maximum value is 64",
				map[string]any{"cores": 65},
				"cores: Must be a multiple of 2; cores: Must be less than or equal to 64",
			),
			Entry(
				"cores minimum value is 1",
				map[string]any{"cores": 0},
				"cores: Must be greater than or equal to 1",
			),
			Entry(
				"cores multiple of 2",
				map[string]any{"cores": 3},
				"cores: Must be a multiple of 2",
			),
			Entry(
				"storage capacity maximum value is 4096",
				map[string]any{"storage_gb": 4097},
				"storage_gb: Must be less than or equal to 4096",
			),
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
				map[string]any{"region": "invalid-region"},
				"region must be one of the following:",
			),
		)
	})

	Describe("updating instance", func() {
		var instanceID string

		BeforeEach(func() {
			instanceID, _ = broker.Provision("csb-google-mysql", customMySQLPlan["name"].(string), nil)

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				HaveKeyWithValue("instance_name", "csb-mysql-"+instanceID),
			)
			_ = mockTerraform.Reset()
		})

		DescribeTable("should allow updating properties not flagged as `prohibit_update` and not specified in the plan",
			func(params map[string]any) {
				err := broker.Update(instanceID, "csb-google-mysql", customMySQLPlan["name"].(string), params)

				Expect(err).NotTo(HaveOccurred())
			},
			Entry("update credentials", map[string]any{"credentials": "other-credentials"}),
			Entry("update project", map[string]any{"project": "another-project"}),
		)

		DescribeTable("should prevent updating properties flagged as `prohibit_update` because it can result in the recreation of the service instance and lost data",
			func(params map[string]any) {
				err := broker.Update(instanceID, "csb-google-mysql", customMySQLPlan["name"].(string), params)

				Expect(err).To(MatchError(
					ContainSubstring(
						"attempt to update parameter that may result in service instance re-creation and data loss",
					),
				))
				Expect(mockTerraform.ApplyInvocations()).To(HaveLen(0))
			},
			Entry("update instance_name", map[string]any{"instance_name": "another-instance-name"}),
			Entry("update db_name", map[string]any{"db_name": "another-db-name"}),
			Entry("update region", map[string]any{"region": "australia-southeast1"}),
			Entry("update authorized_network", map[string]any{"authorized_network": "another-authorized-network"}),
			Entry("update authorized_network_id", map[string]any{"authorized_network_id": "another-authorized-network_id"}),
		)

		DescribeTable("should not allow updating properties that are specified in the plan",
			func(key string, value any) {
				err := broker.Update(instanceID, "csb-google-mysql", customMySQLPlan["name"].(string), map[string]any{key: value})

				Expect(err).To(
					MatchError(
						ContainSubstring(
							fmt.Sprintf("plan defined properties cannot be changed: %s", key),
						),
					),
				)
			},
			Entry("update require_ssl", "require_ssl", true),
			Entry("update mysql_version", "mysql_version", "MYSQL_5_7"),
		)

		DescribeTable("should not allow updating additional properties",
			func(key string, value any) {
				err := broker.Update(instanceID, "csb-google-mysql", customMySQLPlan["name"].(string), map[string]any{key: value})

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

		DescribeTable("should not allow updating additional properties",
			func(key string, value any) {
				err := broker.Update(instanceID, "csb-google-mysql", customMySQLPlan["name"].(string), map[string]any{key: value})

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
