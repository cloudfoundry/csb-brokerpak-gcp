package terraformtests

import (
	"path"

	. "csbbrokerpakgcp/terraform-tests/helpers"

	tfjson "github.com/hashicorp/terraform-json"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("mysql", Label("mysql-terraform"), Ordered, func() {
	var (
		plan                  tfjson.Plan
		terraformProvisionDir string
	)

	defaultVars := map[string]any{
		"tier":                  "db-n1-standard-2",
		"storage_gb":            10,
		"credentials":           googleCredentials,
		"project":               googleProject,
		"instance_name":         "test-instance-name-456",
		"db_name":               "test-db-name-987",
		"region":                "us-central1",
		"authorized_network":    "default",
		"authorized_network_id": "",
		"mysql_version":         "8.0",
		"labels":                map[string]string{"label1": "value1"},
		"disk_autoresize":       true,
		"disk_autoresize_limit": 0,
		"deletion_protection":   false,
	}

	BeforeAll(func() {
		terraformProvisionDir = path.Join(workingDir, "cloudsql/mysql/provision")
		Init(terraformProvisionDir)
	})

	Context("pass through", func() {
		BeforeAll(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{}))
		})

		It("maps parameters to corresponding values", func() {
			Expect(AfterValuesForType(plan, "google_sql_database_instance")).To(
				MatchKeys(0, Keys{
					"name":                   Equal("test-instance-name-456"),
					"database_version":       Equal("8.0"),
					"region":                 Equal("us-central1"),
					"deletion_protection":    BeFalse(),
					"root_password":          BeNil(),
					"clone":                  BeEmpty(),
					"timeouts":               BeNil(),
					"restore_backup_context": BeEmpty(),
					"settings": ContainElement(
						MatchKeys(IgnoreExtras, Keys{
							"tier":        Equal("db-n1-standard-2"),
							"disk_size":   BeNumerically("==", 10),
							"user_labels": MatchKeys(0, Keys{"label1": Equal("value1")}),
							"ip_configuration": ContainElement(
								MatchKeys(IgnoreExtras, Keys{
									"ipv4_enabled":    BeFalse(),
									"private_network": Equal("https://www.googleapis.com/compute/v1/projects/cloud-service-broker/global/networks/default"),
								}),
							),
							"disk_autoresize":       BeTrue(),
							"disk_autoresize_limit": BeNumerically("==", 0),
						}),
					),
				}),
			)

			Expect(AfterValuesForType(plan, "google_sql_database")).To(
				MatchKeys(IgnoreExtras, Keys{
					"name":     Equal("test-db-name-987"),
					"instance": Equal("test-instance-name-456"),
				}),
			)

			Expect(AfterValuesForType(plan, "google_sql_user")).To(
				MatchKeys(IgnoreExtras, Keys{
					"instance": Equal("test-instance-name-456"),
				}),
			)
		})
	})
})
