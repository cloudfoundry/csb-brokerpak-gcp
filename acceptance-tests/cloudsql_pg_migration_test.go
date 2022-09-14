package acceptance_test

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/go-pg/pg/v10"
	"github.com/lib/pq"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"csbbrokerpakgcp/acceptance-tests/helpers/apps"
	"csbbrokerpakgcp/acceptance-tests/helpers/brokers"
	"csbbrokerpakgcp/acceptance-tests/helpers/gsql"
	"csbbrokerpakgcp/acceptance-tests/helpers/legacybindings"
	"csbbrokerpakgcp/acceptance-tests/helpers/random"
	"csbbrokerpakgcp/acceptance-tests/helpers/services"
)

const (
	legacyBrokerName               = "gcp-service-broker"
	grantBindingUserGroupStatement = `create function temp_fn__grant_role(username pg_catalog.name) returns void as
$$
begin
    raise notice 'granting binding_user_group to %', username;
    execute format('grant binding_user_group to %I', username);
end;
$$ language plpgsql;
select temp_fn__grant_role(usename)
from pg_catalog.pg_user
where usename not like 'cloud%'
  and usename != 'postgres'
  and usename not like 'sb%';
drop function temp_fn__grant_role;
`
	legacyDbTier = "db-f1-micro"
)

var _ = Describe("Postgres service instance migration", func() {

	FIt("retains data", func() {
		By("creating the original service instance")
		sourceServiceOffering := "google-cloudsql-postgres-vpc"
		sourceServicePlan := "default"

		sourceServiceInstance := services.CreateInstance(
			sourceServiceOffering,
			sourceServicePlan,
			services.WithBroker(&brokers.Broker{Name: legacyBrokerName}),
			services.WithParameters(map[string]string{
				"tier":            legacyDbTier,
				"private_network": os.Getenv("GCP_PAS_NETWORK"),
			}),
		)
		defer sourceServiceInstance.Delete()

		By("binding an app to the source service instance")
		sourceApp := apps.Push(apps.WithApp(apps.PostgreSQL))
		defer sourceApp.Delete()

		sourceInstanceBinding := sourceServiceInstance.Bind(sourceApp)
		sourceApp.Start()
		defer sourceApp.DELETETestTable()

		By("creating a schema and adding some data in the source database")
		schema := random.Name(random.WithMaxLength(8))
		sourceApp.PUT("", schema)
		defer sourceApp.DELETE("/")

		key := random.Hexadecimal()
		value := random.Hexadecimal()
		sourceApp.PUT(value, "%s/%s", schema, key)

		By("creating a new service instance with the same version and database name as the original instance")
		credentials := sourceInstanceBinding.Credential()
		legacyBinding, err := legacybindings.ExtractPostgresBinding(credentials)
		Expect(err).NotTo(HaveOccurred())

		currentIPAddress := getCurrentIPAddress()
		targetServiceInstance := services.CreateInstance("csb-google-postgres", "default",
			services.WithParameters(map[string]any{
				"postgres_version":          "POSTGRES_11",
				"db_name":                   legacyBinding.DatabaseName,
				"authorized_networks_cidrs": []string{fmt.Sprintf("%s/32", currentIPAddress)},
				"public_ip":                 true,
			}))
		defer targetServiceInstance.Delete()

		backupId := gsql.CreateBackup(legacyBinding.InstanceName)
		defer gsql.DeleteBackup(legacyBinding.InstanceName, backupId)

		By("creating a service key for the new service instance")
		serviceKey := targetServiceInstance.CreateServiceKey()

		var serviceKeyData struct {
			Hostname    string `json:"hostname"`
			SSLKey      []byte `json:"sslkey"`
			SSLCert     []byte `json:"sslcert"`
			SSLRootCert []byte `json:"sslrootcert"`
		}
		serviceKey.Get(&serviceKeyData)

		By("restoring the backup onto the new service instance")
		gsql.RestoreBackup(fmt.Sprintf("csb-postgres-%v", targetServiceInstance.GUID()), legacyBinding.InstanceName, backupId)

		By("restoring the tf postgres user via updating the target service instance")
		targetServiceInstance.Update("backups_start_time", "12:34")

		// We're unable to delete the service key before the tf postgres user is restored
		defer serviceKey.Delete()

		By("executing SQL against the new service instance")
		certChain := append(serviceKeyData.SSLCert, '\n')
		certChain = append(certChain, serviceKeyData.SSLRootCert...)
		cert, err := tls.X509KeyPair(certChain, serviceKeyData.SSLKey)
		Expect(err).NotTo(HaveOccurred())

		cfg := &tls.Config{Certificates: []tls.Certificate{cert}}
		db := pg.Connect(&pg.Options{
			Addr:      serviceKeyData.Hostname,
			User:      legacyBinding.Username,
			Password:  legacyBinding.Password,
			Database:  legacyBinding.DatabaseName,
			TLSConfig: cfg,
		})

		statements := []string{
			"CREATE ROLE binding_user_group WITH NOLOGIN",
			fmt.Sprintf("GRANT ALL PRIVILEGES ON DATABASE %s TO binding_user_group", pq.QuoteIdentifier(legacyBinding.DatabaseName)),
			grantBindingUserGroupStatement,
			"GRANT binding_user_group TO CURRENT_USER",
			"REASSIGN OWNED BY CURRENT_USER TO binding_user_group",
		}
		for _, stmt := range statements {
			_, err := db.Exec(stmt)
			Expect(err).NotTo(HaveOccurred())
		}

		By("binding an app to the target service instance")
		targetApp := apps.Push(apps.WithApp(apps.PostgreSQL))
		defer targetApp.Delete()
		_ = targetServiceInstance.Bind(targetApp)
		targetApp.Start()

		By("reading the data")
		got := targetApp.GET("%s/%s", schema, key)
		Expect(got).To(Equal(value))
	})
})

func getCurrentIPAddress() string {
	resp, err := http.Get("https://ifconfig.me/")
	Expect(err).NotTo(HaveOccurred())
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	body, err := io.ReadAll(resp.Body)
	return string(body)
}
