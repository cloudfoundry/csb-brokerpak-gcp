package integration_test

import (
	"fmt"

	testframework "github.com/cloudfoundry/cloud-service-broker/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var customRedisPlans = []map[string]any{
	customRedisPlan,
}

var customRedisPlan = map[string]any{
	"name":         "custom-plan",
	"id":           "9dfa265e-1c4d-40c6-ade6-b341ffd6ccc3",
	"description":  "custom memorystore plan defined by customer",
	"service_tier": "TIER_UNSPECIFIED",
	"metadata": map[string]any{
		"displayName": "custom cloud memorystore service (beta)",
	},
	"labels": map[string]any{
		"label1": "label1",
		"label2": "label2",
	},
}

var _ = Describe("Redis", func() {
	const redisServiceName = "csb-google-redis"

	AfterEach(func() {
		Expect(mockTerraform.Reset()).NotTo(HaveOccurred())
	})

	It("should publish redis in the catalog", func() {
		catalog, err := broker.Catalog()
		Expect(err).NotTo(HaveOccurred())

		service := testframework.FindService(catalog, redisServiceName)
		Expect(service.ID).NotTo(BeNil())
		Expect(service.Name).NotTo(BeNil())
		Expect(service.Tags).To(ConsistOf("gcp", "redis", "beta"))
		Expect(service.Metadata.ImageUrl).NotTo(BeNil())
		Expect(service.Metadata.DisplayName).NotTo(BeNil())
		Expect(service.Plans).To(
			ConsistOf(
				MatchFields(IgnoreExtras, Fields{"Name": Equal("basic")}),
				MatchFields(IgnoreExtras, Fields{"Name": Equal("ha")}),
				MatchFields(IgnoreExtras, Fields{"Name": Equal("custom-plan")}),
			),
		)
	})

	Describe("provisioning", func() {
		It("should provision basic plan", func() {
			instanceID, _ := broker.Provision(redisServiceName, "basic", nil)

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("service_tier", "BASIC"),
					HaveKeyWithValue("memory_size_gb", float64(4)),
					HaveKeyWithValue("region", "us-central1"),
					HaveKeyWithValue("credentials", "broker-gcp-creds"),
					HaveKeyWithValue("project", "broker-gcp-project"),
					HaveKeyWithValue("authorized_network", "default"),
					HaveKeyWithValue("authorized_network_id", ""),
					HaveKeyWithValue("reserved_ip_range", ""),
					HaveKeyWithValue("display_name", ContainSubstring("pcf-sb-")),
					HaveKeyWithValue("instance_id", ContainSubstring("pcf-sb-")),
					HaveKeyWithValue("labels", HaveKeyWithValue("pcf-instance-id", instanceID)),
				),
			)
		})

		It("should allow setting properties do not defined in the plan", func() {
			broker.Provision(redisServiceName, "basic", map[string]any{
				"memory_size_gb":        float64(10),
				"instance_id":           "fake-instance-id",
				"display_name":          "fake-display-name",
				"region":                "asia-northeast1",
				"credentials":           "fake-credentials",
				"project":               "fake-project",
				"authorized_network":    "fake-authorized_network",
				"authorized_network_id": "fake-authorized_network_id",
				"reserved_ip_range":     "192.168.0.0/29",
			})

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("memory_size_gb", float64(10)),
					HaveKeyWithValue("instance_id", "fake-instance-id"),
					HaveKeyWithValue("display_name", "fake-display-name"),
					HaveKeyWithValue("region", "asia-northeast1"),
					HaveKeyWithValue("credentials", "fake-credentials"),
					HaveKeyWithValue("project", "fake-project"),
					HaveKeyWithValue("authorized_network", "fake-authorized_network"),
					HaveKeyWithValue("authorized_network_id", "fake-authorized_network_id"),
					HaveKeyWithValue("reserved_ip_range", "192.168.0.0/29"),
				),
			)
		})

		It("should not allow changing of plan defined properties", func() {
			_, err := broker.Provision(redisServiceName, "basic", map[string]any{"service_tier": "STANDARD_HA"})

			Expect(err).To(MatchError(ContainSubstring("plan defined properties cannot be changed: service_tier")))
		})

		DescribeTable("property constraints",
			func(params map[string]any, expectedErrorMsg string) {
				_, err := broker.Provision(redisServiceName, customRedisPlan["name"].(string), params)

				Expect(err).To(MatchError(ContainSubstring(expectedErrorMsg)))
			},
			Entry(
				"memory_size_gb maximum value is 300",
				map[string]any{"memory_size_gb": 301},
				"memory_size_gb: Must be less than or equal to 300",
			),
			Entry(
				"memory_size_gb minimum value is 1",
				map[string]any{"memory_size_gb": 0},
				"memory_size_gb: Must be greater than or equal to 1",
			),
			Entry(
				"instance_id maximum length is 30 characters",
				map[string]any{"instance_id": stringOfLen(31)},
				"instance_id: String length must be less than or equal to 30",
			),
			Entry(
				"instance_id minimum length is 6 characters",
				map[string]any{"instance_id": stringOfLen(5)},
				"instance_id: String length must be greater than or equal to 6",
			),
			Entry(
				"instance_id invalid characters",
				map[string]any{"instance_id": ".aaaaa"},
				"instance_id: Does not match pattern '^[a-z][a-z0-9-]+$'",
			),
			Entry(
				"display_name maximum length is 80 characters",
				map[string]any{"display_name": stringOfLen(81)},
				"display_name: String length must be less than or equal to 80",
			),
			Entry(
				"region invalid characters",
				map[string]any{"region": "-Asia-northeast1"},
				"region: Does not match pattern '^[a-z][a-z0-9-]+$'",
			),
		)
	})

	Describe("updating instance", func() {
		var instanceID string

		BeforeEach(func() {
			mockTerraform.SetTFState([]testframework.TFStateValue{})
			instanceID, _ = broker.Provision(redisServiceName, customRedisPlan["name"].(string), nil)

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("instance_id", ContainSubstring("pcf-sb-")),
					HaveKeyWithValue("labels", HaveKeyWithValue("pcf-instance-id", instanceID)),
					HaveKeyWithValue("service_tier", "TIER_UNSPECIFIED"),
				),
			)
			_ = mockTerraform.Reset()
		})

		DescribeTable("should allow updating properties not flagged as `prohibit_update` and not specified in the plan",
			func(params map[string]any) {
				err := broker.Update(instanceID, redisServiceName, customRedisPlan["name"].(string), params)

				Expect(err).NotTo(HaveOccurred())
			},
			Entry("update region", map[string]any{"region": "asia-southeast1"}),
			Entry("update credentials", map[string]any{"credentials": "other-credentials"}),
			Entry("update project", map[string]any{"project": "another-project"}),
		)

		DescribeTable("should prevent updating properties flagged as `prohibit_update` because it can result in the recreation of the service instance and lost data",
			func(params map[string]any) {
				err := broker.Update(instanceID, redisServiceName, customRedisPlan["name"].(string), params)

				Expect(err).To(MatchError(
					ContainSubstring(
						"attempt to update parameter that may result in service instance re-creation and data loss",
					),
				))
				Expect(mockTerraform.ApplyInvocations()).To(HaveLen(0))
			},
			Entry("update authorized_network", map[string]any{"authorized_network": "other-authorized_network"}),
			Entry("update authorized_network_id", map[string]any{"authorized_network_id": "another-authorized_network_id"}),
			Entry("update reserved_ip_range", map[string]any{"reserved_ip_range": "192.168.0.0/29"}),
		)

		DescribeTable("should not allow updating properties that are specified in the plan",
			func(key string, value any) {
				err := broker.Update(instanceID, redisServiceName, customRedisPlan["name"].(string), map[string]any{key: value})

				Expect(err).To(
					MatchError(
						ContainSubstring(
							fmt.Sprintf("plan defined properties cannot be changed: %s", key),
						),
					),
				)
			},
			Entry("update service_tier", "service_tier", "BASIC"),
		)

		DescribeTable("should not allow updating additional properties",
			func(key string, value any) {
				err := broker.Update(instanceID, redisServiceName, customRedisPlan["name"].(string), map[string]any{key: value})

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
