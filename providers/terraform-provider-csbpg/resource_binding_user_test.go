package main_test

import (
	"csbbrokerpakgcp/providers/terraform-provider-csbpg/csbpg"
	"database/sql"
	"fmt"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	_ "embed"

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

//go:embed "testfixtures/ssl_postgres/certs/ca.crt"
var postgresSSLCACert string

//go:embed "testfixtures/ssl_postgres/certs/server.crt"
var postgresSSLServerCert string

//go:embed "testfixtures/ssl_postgres/keys/server.key"
var postgresSSLServerKey string

var _ = Describe("SSL Postgres Bindings", func() {
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
			"--mount", "source=ssl_postgres,destination=/mnt",
			"-t", "postgres",
			"-c", "config_file=/mnt/pgconf/postgresql.conf",
			"-c", "hba_file=/mnt/pgconf/pg_hba.conf",
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
		}).WithTimeout(10 * time.Second).WithPolling(time.Second).Should(Succeed())
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
		
		  sslrootcert = <<EOF
%s
EOF
		  clientcert {
    		cert = <<EOF
%s
EOF
    		key  = <<EOF
%s
EOF
  	      }
		}

		resource "csbpg_binding_user" "binding_user" {
		  username = "%s"
		  password = "%s"
		}
		`, hostname, port, username, password, database, dataOwnerRole,
			postgresSSLCACert, postgresSSLServerCert, postgresSSLServerKey,
			bindingUsername, bindingPassword), func(state *terraform.State) error {
			By("CHECKING RESOURCE CREATE")

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

			return nil
		}, func(state *terraform.State) error {
			By("CHECKING RESOURCE DELETE")
			db, err := sql.Open("postgres", uri)
			Expect(err).NotTo(HaveOccurred())

			By("checking that the data owner role is not deleted")
			rows, err := db.Query(fmt.Sprintf("SELECT FROM pg_catalog.pg_roles WHERE rolname = '%s'", dataOwnerRole))
			Expect(err).NotTo(HaveOccurred())
			Expect(rows.Next()).To(BeTrue(), fmt.Sprintf("role %q has been deleted", dataOwnerRole))

			By("checking that the binding user is deleted")
			rows, err = db.Query(fmt.Sprintf("SELECT FROM pg_catalog.pg_roles WHERE rolname = '%s'", bindingUsername))
			Expect(err).NotTo(HaveOccurred())
			Expect(rows.Next()).To(BeFalse(), fmt.Sprintf("role %q still exists", bindingUsername))

			return nil
		})
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

		  sslrootcert = <<EOF
%s
EOF
		  clientcert {
    		cert = <<EOF
%s
EOF
    		key  = <<EOF
%s
EOF
		  }
		}

		resource "csbpg_binding_user" "binding_user_1" {
		  username = "%s"
		  password = "%s"
		}

		resource "csbpg_binding_user" "binding_user_2" {
		  username = "%s"
		  password = "%s"
		}
		`, hostname, port, username, password, database, dataOwnerRole,
			postgresSSLCACert, postgresSSLServerCert, postgresSSLServerKey,
			bindingUsername1, bindingPassword1, bindingUsername2, bindingPassword2), func(state *terraform.State) error {
			By("CHECKING RESOURCE CREATE")

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
			return nil
		}, func(state *terraform.State) error {
			By("CHECKING RESOURCE DELETE")

			db, err := sql.Open("postgres", uri)
			Expect(err).NotTo(HaveOccurred())

			By("checking that the data owner role is not deleted")
			rows, err := db.Query(fmt.Sprintf("SELECT FROM pg_catalog.pg_roles WHERE rolname = '%s'", dataOwnerRole))
			Expect(err).NotTo(HaveOccurred())
			Expect(rows.Next()).To(BeTrue(), fmt.Sprintf("role %q has been deleted", dataOwnerRole))

			By("checking that both binding users are deleted")
			rows, err = db.Query(fmt.Sprintf("SELECT FROM pg_catalog.pg_roles WHERE rolname = '%s'", bindingUsername1))
			Expect(err).NotTo(HaveOccurred())
			Expect(rows.Next()).To(BeFalse(), fmt.Sprintf("role %q still exists", bindingUsername1))

			rows, err = db.Query(fmt.Sprintf("SELECT FROM pg_catalog.pg_roles WHERE rolname = '%s'", bindingUsername2))
			Expect(err).NotTo(HaveOccurred())
			Expect(rows.Next()).To(BeFalse(), fmt.Sprintf("role %q still exists", bindingUsername2))

			return nil
		})
	})
})

func createVolume(fixtureName string) {
	path := path.Join(getPWD(), "testfixtures", fixtureName)
	mustRun("docker", "volume", "create", fixtureName)
	for _, folder := range []string{"certs", "keys", "pgconf"} {
		mustRun("docker", "run",
			"-v", path+":/fixture",
			"--mount", fmt.Sprintf("source=%s,destination=/mnt", fixtureName),
			"postgres", "rm", "-rf", "/mnt/"+folder)
		mustRun("docker", "run",
			"-v", path+":/fixture",
			"--mount", fmt.Sprintf("source=%s,destination=/mnt", fixtureName),
			"postgres", "cp", "-r", "/fixture/"+folder, "/mnt")
	}
	mustRun("docker", "run",
		"-v", path+":/fixture",
		"--mount", fmt.Sprintf("source=%s,destination=/mnt", fixtureName),
		"postgres", "chmod", "-R", "0600", "/mnt/keys/server.key")
	mustRun("docker", "run",
		"-v", path+":/fixture",
		"--mount", fmt.Sprintf("source=%s,destination=/mnt", fixtureName),
		"postgres", "chown", "-R", "postgres:postgres", "/mnt/keys/server.key")
}

func mustRun(command ...string) {
	start, err := gexec.Start(exec.Command(
		command[0], command[1:]...,
	), GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(start).WithTimeout(30 * time.Second).WithPolling(time.Second).Should(gexec.Exit(0))
}

func getPWD() string {
	_, file, _, _ := runtime.Caller(1)
	return filepath.Dir(file)
}

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

func applyHCL(hcl string, checkOnCreate, checkOnDestroy resource.TestCheckFunc) {
	resource.Test(GinkgoT(), resource.TestCase{
		IsUnitTest: true, // means we don't need to set TF_ACC
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"csbpg": func() (*schema.Provider, error) { return csbpg.Provider(), nil },
		},
		CheckDestroy: checkOnDestroy,
		Steps: []resource.TestStep{{
			ResourceName: "csbpg_shared_role.shared_role",
			Config:       hcl,
			Check:        checkOnCreate,
		}},
	})
}
