package main_test

import (
	"database/sql"
	"fmt"
	"os/exec"
	"terraform-provider-csbpg/csbpg"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	_ "github.com/lib/pq"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

const (
	username = "postgres"
	hostname = "localhost"
)

var _ = Describe("Tests", func() {
	var session *gexec.Session
	var uri, password, database string
	var port int
	BeforeEach(func() {
		var err error
		password = uuid.New().String()
		database = uuid.New().String()
		port = freePort()
		cmd := exec.Command(
			"docker", "run",
			"-e", fmt.Sprintf("POSTGRES_PASSWORD=%s", password),
			"-e", fmt.Sprintf("POSTGRES_DB=%s", database),
			"-p", fmt.Sprintf("%d:5432", port),
			"-t", "postgres",
		)
		session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		uri = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", username, password, hostname, port, database)
		Eventually(func() error {
			db, err := sql.Open("postgres", uri)
			if err != nil {
				return err
			}
			defer db.Close()
			return db.Ping()
		}).WithTimeout(30 * time.Second).WithPolling(time.Second).Should(Succeed())
	})
	AfterEach(func() {
		session.Terminate()
	})

	It("creates a binding user", func() {
		dataOwnerRole := uuid.New().String()
		bindingUsername := uuid.New().String()
		bindingPassword := uuid.New().String()
		applyHCL(fmt.Sprintf(`
		provider "csbpg" {
		  host            = "%s"
		  port            = %d
		  username        = "%s"
		  password        = "%s"
		  database        = "%s"
		  data_owner_role = "%s"
		}

		resource "csbpg_binding_user" "binding_user" {
		  username = "%s"
		  password = "%s"
		}
		`, hostname, port, username, password, database, dataOwnerRole, bindingUsername, bindingPassword))

		db, err := sql.Open("postgres", uri)
		Expect(err).NotTo(HaveOccurred())

		By("checking that the data owner role is created")
		rows, err := db.Query(fmt.Sprintf("SELECT FROM pg_catalog.pg_roles WHERE rolname = '%s'", dataOwnerRole))
		Expect(err).NotTo(HaveOccurred())
		Expect(rows.Next()).To(BeTrue(), fmt.Sprintf("role %q has not been created", dataOwnerRole))

		By("checking that the binding user is created")
		rows, err = db.Query(fmt.Sprintf("SELECT FROM pg_catalog.pg_roles WHERE rolname = '%s'", bindingUsername))
		Expect(err).NotTo(HaveOccurred())
		Expect(rows.Next()).To(BeTrue(), fmt.Sprintf("role %q has not been created", bindingUsername))

		By("checking that the binding user is a member of the data owner role")
		rows, err = db.Query(fmt.Sprintf("SELECT pg_has_role('%s', '%s', 'member')", bindingUsername, dataOwnerRole))
		Expect(err).NotTo(HaveOccurred())
		var result bool
		Expect(rows.Next()).To(BeTrue(), "pg_has_role() query failed")
		Expect(rows.Scan(&result)).To(Succeed())
		Expect(result).To(BeTrue(), "binding user is not a member of the data_owner_role")
	})

	It("can create multiple binding user", func() {
		dataOwnerRole := uuid.New().String()
		bindingUsername1 := uuid.New().String()
		bindingPassword1 := uuid.New().String()

		bindingUsername2 := uuid.New().String()
		bindingPassword2 := uuid.New().String()
		applyHCL(fmt.Sprintf(`
		provider "csbpg" {
		  host            = "%s"
		  port            = %d
		  username        = "%s"
		  password        = "%s"
		  database        = "%s"
		  data_owner_role = "%s"
		}

		resource "csbpg_binding_user" "binding_user_1" {
		  username = "%s"
		  password = "%s"
		}

		resource "csbpg_binding_user" "binding_user_2" {
		  username = "%s"
		  password = "%s"
		}
		`, hostname, port, username, password, database, dataOwnerRole, bindingUsername1, bindingPassword1, bindingUsername2, bindingPassword2))

		db, err := sql.Open("postgres", uri)
		Expect(err).NotTo(HaveOccurred())

		By("checking that the data owner role is created")
		rows, err := db.Query(fmt.Sprintf("SELECT FROM pg_catalog.pg_roles WHERE rolname = '%s'", dataOwnerRole))
		Expect(err).NotTo(HaveOccurred())
		Expect(rows.Next()).To(BeTrue(), fmt.Sprintf("role %q has not been created", dataOwnerRole))

		By("checking that the binding user is created")
		rows, err = db.Query(fmt.Sprintf("SELECT FROM pg_catalog.pg_roles WHERE rolname = '%s'", bindingUsername1))
		Expect(err).NotTo(HaveOccurred())
		Expect(rows.Next()).To(BeTrue(), fmt.Sprintf("role %q has not been created", bindingUsername1))

		By("checking that the binding user is a member of the data owner role")
		rows, err = db.Query(fmt.Sprintf("SELECT pg_has_role('%s', '%s', 'member')", bindingUsername1, dataOwnerRole))
		Expect(err).NotTo(HaveOccurred())
		var result bool
		Expect(rows.Next()).To(BeTrue(), "pg_has_role() query failed")
		Expect(rows.Scan(&result)).To(Succeed())
		Expect(result).To(BeTrue(), "binding user is not a member of the data_owner_role")

		Expect(query(db, fmt.Sprintf("SELECT pg_has_role('%s', '%s', 'member')", bindingUsername1, dataOwnerRole))).To(ConsistOf(true))
	})
})

func query(db *sql.DB, query string) (any, error) {
	var result []any
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var row any
		err = rows.Scan(&row)
		if err != nil {
			return nil, err
		}

		result = append(result, row)
	}
	return result, nil
}

func applyHCL(hcl string) {
	resource.Test(GinkgoT(), resource.TestCase{
		IsUnitTest: true, // means we don't need to set TF_ACC
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"csbpg": func() (*schema.Provider, error) { return csbpg.Provider(), nil },
		},
		Steps: []resource.TestStep{{
			ResourceName: "csbpg_shared_role.shared_role",
			Config:       hcl,
		}},
	})
}
