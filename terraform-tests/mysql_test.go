package terraformtests

import (
	"path"

	tfjson "github.com/hashicorp/terraform-json"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	. "csbbrokerpakgcp/terraform-tests/helpers"
)

var _ = Describe("mysql", Label("mysql-terraform"), Ordered, func() {
	const googleSQLDBInstance = "google_sql_database_instance"

	var (
		plan                  tfjson.Plan
		terraformProvisionDir string
	)

	defaultVars := map[string]any{
		"tier":                                   "db-n1-standard-2",
		"storage_gb":                             10,
		"credentials":                            googleCredentials,
		"project":                                googleProject,
		"instance_name":                          "test-instance-name-456",
		"db_name":                                "test-db-name-987",
		"region":                                 "us-central1",
		"authorized_network":                     "default",
		"authorized_network_id":                  "",
		"authorized_networks_cidrs":              []string{},
		"public_ip":                              false,
		"mysql_version":                          "8.0",
		"labels":                                 map[string]string{"label1": "value1"},
		"disk_autoresize":                        true,
		"disk_autoresize_limit":                  0,
		"deletion_protection":                    false,
		"backups_start_time":                     "07:00",
		"backups_location":                       nil,
		"backups_retain_number":                  7,
		"backups_transaction_log_retention_days": 0,
	}

	BeforeAll(func() {
		terraformProvisionDir = path.Join(workingDir, "cloudsql/mysql/provision")
		Init(terraformProvisionDir)
	})

	Context("default values", func() {
		BeforeAll(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{}))
		})

		It("maps parameters to corresponding values", func() {
			Expect(AfterValuesForType(plan, googleSQLDBInstance)).To(
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
									"ipv4_enabled":        BeFalse(),
									"private_network":     Equal("https://www.googleapis.com/compute/v1/projects/cloud-service-broker/global/networks/default"),
									"authorized_networks": BeEmpty(),
								}),
							),
							"disk_autoresize":       BeTrue(),
							"disk_autoresize_limit": BeNumerically("==", 0),
							"backup_configuration": ContainElement(
								MatchKeys(IgnoreExtras, Keys{
									"enabled":            BeTrue(),
									"start_time":         Equal("07:00"),
									"binary_log_enabled": BeFalse(),
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

	Context("backups", func() {
		Specify("disabling backups", func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{"backups_retain_number": 0}))

			Expect(AfterValuesForType(plan, googleSQLDBInstance)).To(
				MatchKeys(IgnoreExtras, Keys{
					"settings": ContainElement(MatchKeys(IgnoreExtras, Keys{
						"backup_configuration": ContainElement(MatchKeys(IgnoreExtras, Keys{
							"enabled": BeFalse(),
						})),
					})),
				}),
			)
		})

		Specify("enabling transaction log backups", func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{"backups_transaction_log_retention_days": 3}))

			Expect(AfterValuesForType(plan, googleSQLDBInstance)).To(
				MatchKeys(IgnoreExtras, Keys{
					"settings": ContainElement(MatchKeys(IgnoreExtras, Keys{
						"backup_configuration": ContainElement(MatchKeys(IgnoreExtras, Keys{
							"binary_log_enabled":             BeTrue(),
							"transaction_log_retention_days": BeNumerically("==", 3),
						})),
					})),
				}),
			)
		})
	})

	Context("networking", func() {
		Specify("enabling a public IP address", func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{"public_ip": true}))

			Expect(AfterValuesForType(plan, googleSQLDBInstance)).To(
				MatchKeys(IgnoreExtras, Keys{
					"settings": ContainElement(MatchKeys(IgnoreExtras, Keys{
						"ip_configuration": ContainElement(MatchKeys(IgnoreExtras, Keys{
							"ipv4_enabled": BeTrue(),
						})),
					})),
				}),
			)
		})

		Specify("setting authorized network CIDRs", func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{"authorized_networks_cidrs": []string{"one", "two"}}))

			Expect(AfterValuesForType(plan, googleSQLDBInstance)).To(
				MatchKeys(IgnoreExtras, Keys{
					"settings": ContainElement(MatchKeys(IgnoreExtras, Keys{
						"ip_configuration": ContainElement(MatchKeys(IgnoreExtras, Keys{
							"authorized_networks": ConsistOf(
								MatchKeys(IgnoreExtras, Keys{"value": Equal("one")}),
								MatchKeys(IgnoreExtras, Keys{"value": Equal("two")}),
							),
						})),
					})),
				}),
			)
		})
	})
})
