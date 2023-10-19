package terraformtests

import (
	"path"

	tfjson "github.com/hashicorp/terraform-json"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	. "csbbrokerpakgcp/terraform-tests/helpers"
)

var _ = Describe("postgres", Label("postgres-terraform"), Ordered, func() {
	const googleSQLDBInstance = "google_sql_database_instance"

	var (
		plan                  tfjson.Plan
		terraformProvisionDir string
	)

	defaultVars := map[string]any{
		"tier":                                  "db-n1-standard-2",
		"storage_gb":                            10,
		"credentials":                           googleCredentials,
		"project":                               googleProject,
		"instance_name":                         "test-instance-name-456",
		"db_name":                               "test-db-name-987",
		"region":                                "us-central1",
		"authorized_network":                    "default",
		"authorized_network_id":                 "",
		"authorized_networks_cidrs":             []string{},
		"public_ip":                             false,
		"database_version":                      "POSTGRES_13",
		"labels":                                map[string]string{"label1": "value1"},
		"require_ssl":                           true,
		"backups_start_time":                    "07:00",
		"backups_location":                      "us",
		"backups_retain_number":                 7,
		"backups_point_in_time_log_retain_days": 7,
	}

	BeforeAll(func() {
		terraformProvisionDir = path.Join(workingDir, "cloudsql/postgresql/provision")
		Init(terraformProvisionDir)
	})

	Context("default values", func() {
		BeforeAll(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{}))
		})

		It("maps parameters to corresponding values", func() {
			Expect(AfterValuesForType(plan, googleSQLDBInstance)).To(
				MatchAllKeys(Keys{
					"name":                Equal("test-instance-name-456"),
					"database_version":    Equal("POSTGRES_13"),
					"region":              Equal("us-central1"),
					"deletion_protection": BeFalse(),
					"root_password":       BeNil(),
					"clone":               BeEmpty(),
					"timeouts": MatchAllKeys(Keys{
						"create": Equal("60m"),
						"delete": BeNil(),
						"update": BeNil(),
					}),
					"restore_backup_context": BeEmpty(),
					"settings": ContainElement(
						MatchKeys(IgnoreExtras, Keys{
							"tier":        Equal("db-n1-standard-2"),
							"disk_size":   BeNumerically("==", 10),
							"user_labels": MatchAllKeys(Keys{"label1": Equal("value1")}),
							"ip_configuration": ContainElement(
								MatchKeys(IgnoreExtras, Keys{
									"ipv4_enabled":        BeFalse(),
									"private_network":     Equal("https://www.googleapis.com/compute/v1/projects/cloud-service-broker/global/networks/default"),
									"authorized_networks": BeEmpty(),
								}),
							),
							"disk_autoresize":       BeTrue(),
							"disk_autoresize_limit": BeNumerically("==", 0),
							"backup_configuration": ContainElement(
								MatchKeys(IgnoreExtras, Keys{
									"enabled":    BeTrue(),
									"start_time": Equal("07:00"),
									"backup_retention_settings": ContainElement(
										MatchKeys(IgnoreExtras, Keys{
											"retained_backups": BeNumerically("==", 7),
											"retention_unit":   Equal("COUNT"),
										}),
									),
								}),
							),
						}),
					),
					"project": Equal(googleProject),
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

			Expect(AfterOutput(plan, "allow_insecure_connections")).To(BeNil())
		})
	})
})
