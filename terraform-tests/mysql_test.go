package terraformtests

import (
	"os"
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
		defaultVars           map[string]any
		authorizedNetworkID   string
		privateNetworkID      string
	)

	BeforeEach(func() {
		authorizedNetworkID = os.Getenv("GCP_AUTHORIZED_NETWORK_ID")

		privateNetworkID = "https://www.googleapis.com/compute/v1/projects/cloud-service-broker/global/networks/default"
		if authorizedNetworkID != "" {
			privateNetworkID = authorizedNetworkID
		}

		defaultVars = map[string]any{
			"tier":                                   "db-n1-standard-2",
			"storage_gb":                             10,
			"credentials":                            googleCredentials,
			"project":                                googleProject,
			"instance_name":                          "test-instance-name-456",
			"db_name":                                "test-db-name-987",
			"region":                                 "us-central1",
			"authorized_network_id":                  authorizedNetworkID,
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
			"highly_available":                       false,
			"location_preference_zone":               "",
			"location_preference_secondary_zone":     "",
			"allow_insecure_connections":             false,
			"maintenance_day":                        nil,
			"maintenance_hour":                       nil,
		}
	})

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
				MatchAllKeys(Keys{
					"name":                          Equal("test-instance-name-456"),
					"database_version":              Equal("8.0"),
					"region":                        Equal("us-central1"),
					"deletion_protection":           BeFalse(),
					"root_password":                 BeNil(),
					"root_password_wo":              BeNil(),
					"root_password_wo_version":      BeNil(),
					"backupdr_backup":               BeNil(),
					"final_backup_description":      BeNil(),
					"point_in_time_restore_context": BeEmpty(),
					"clone":                         BeEmpty(),
					"timeouts":                      BeNil(),
					"restore_backup_context":        BeEmpty(),
					"settings": ContainElement(
						MatchKeys(IgnoreExtras, Keys{
							"tier":        Equal("db-n1-standard-2"),
							"user_labels": MatchAllKeys(Keys{"label1": Equal("value1")}),
							"ip_configuration": ContainElement(
								MatchKeys(IgnoreExtras, Keys{
									"ipv4_enabled":        BeFalse(),
									"private_network":     Equal(privateNetworkID),
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
							"location_preference": ContainElement(
								MatchKeys(IgnoreExtras, Keys{
									"zone":           BeEmpty(),
									"secondary_zone": BeEmpty(),
								}),
							),
							"availability_type":  Equal("ZONAL"),
							"maintenance_window": BeEmpty(),
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

			Expect(AfterOutput(plan, "allow_insecure_connections")).NotTo(BeNil())
			Expect(AfterOutput(plan, "allow_insecure_connections")).To(BeFalse())
		})
	})

	Context("maintenance_window", func() {
		Specify("enabling maintenance_day", func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{"maintenance_day": 1}))

			Expect(AfterValuesForType(plan, googleSQLDBInstance)).To(
				MatchKeys(IgnoreExtras, Keys{
					"settings": ContainElement(MatchKeys(IgnoreExtras, Keys{
						"maintenance_window": ContainElement(MatchKeys(IgnoreExtras, Keys{
							"day":  BeNumerically("==", 1),
							"hour": BeNil(),
						})),
					})),
				}),
			)
		})

		Specify("enabling maintenance_hour", func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{"maintenance_hour": 20}))

			Expect(AfterValuesForType(plan, googleSQLDBInstance)).To(
				MatchKeys(IgnoreExtras, Keys{
					"settings": ContainElement(MatchKeys(IgnoreExtras, Keys{
						"maintenance_window": BeEmpty(),
					})),
				}),
			)
		})

		Specify("enabling maintenance_day and maintenance_hour", func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{"maintenance_day": 1, "maintenance_hour": 20}))

			Expect(AfterValuesForType(plan, googleSQLDBInstance)).To(
				MatchKeys(IgnoreExtras, Keys{
					"settings": ContainElement(MatchKeys(IgnoreExtras, Keys{
						"maintenance_window": ContainElement(MatchKeys(IgnoreExtras, Keys{
							"day":  BeNumerically("==", 1),
							"hour": BeNumerically("==", 20),
						})),
					})),
				}),
			)
		})
	})

	Context("backups", func() {
		Specify("disabling", func() {
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

		Specify("enabling transaction log", func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{"backups_transaction_log_retention_days": 3}))

			Expect(AfterValuesForType(plan, googleSQLDBInstance)).To(
				MatchKeys(IgnoreExtras, Keys{
					"settings": ContainElement(MatchKeys(IgnoreExtras, Keys{
						"backup_configuration": ContainElement(MatchKeys(IgnoreExtras, Keys{
							"enabled":                        BeTrue(),
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

	Context("TLS", func() {
		It("generates the artefacts", func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{}))

			Expect(AfterValuesForType(plan, "google_sql_ssl_cert")).To(MatchKeys(IgnoreExtras, Keys{
				"instance": Equal("test-instance-name-456"),
			}))
		})

		It("passes the change to allow insecure connections", func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{"allow_insecure_connections": true}))

			Expect(AfterOutput(plan, "allow_insecure_connections")).To(BeTrue())
		})
	})

	Context("High availability", func() {
		Specify("enabling", func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
				"highly_available":                   true,
				"location_preference_zone":           "a",
				"location_preference_secondary_zone": "c",
			}))

			Expect(AfterValuesForType(plan, googleSQLDBInstance)).To(
				MatchKeys(IgnoreExtras, Keys{
					"settings": ContainElement(MatchKeys(IgnoreExtras, Keys{
						"availability_type": Equal("REGIONAL"),
						"location_preference": ContainElement(MatchKeys(IgnoreExtras, Keys{
							"zone":           Equal("us-central1-a"),
							"secondary_zone": Equal("us-central1-c"),
						})),
					})),
				}),
			)
		})
	})

	Context("disk auto resize", func() {
		It("does not set the disk_gb value explicitly when it is enabled", func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
				"storage_gb":            50,
				"disk_autoresize":       true,
				"disk_autoresize_limit": 300,
			}))

			Expect(AfterValuesForType(plan, googleSQLDBInstance)).To(
				MatchKeys(IgnoreExtras, Keys{
					"settings": ContainElement(MatchKeys(IgnoreExtras, Keys{
						"disk_autoresize":       BeTrue(),
						"disk_autoresize_limit": BeNumerically("==", 300),
					})),
				}),
			)
		})

		It("sets the disk_gb value explicitly when it is disabled", func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
				"storage_gb":            50,
				"disk_autoresize":       false,
				"disk_autoresize_limit": 300,
			}))

			Expect(AfterValuesForType(plan, googleSQLDBInstance)).To(
				MatchKeys(IgnoreExtras, Keys{
					"settings": ContainElement(MatchKeys(IgnoreExtras, Keys{
						"disk_size":             BeNumerically("==", 50),
						"disk_autoresize":       BeFalse(),
						"disk_autoresize_limit": BeNumerically("==", 0),
					})),
				}),
			)
		})
	})
})
