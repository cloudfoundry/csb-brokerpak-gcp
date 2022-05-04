-- Creating role with same privilege as Google Super user role: https://cloud.google.com/sql/docs/postgres/users
CREATE ROLE cloudsqlsuperuser WITH LOGIN PASSWORD 'password' NOSUPERUSER CREATEDB CREATEROLE;
GRANT pg_signal_backend TO cloudsqlsuperuser;
GRANT pg_monitor TO cloudsqlsuperuser;
-- GRANT USAGE ON SCHEMA public  TO cloudsqlsuperuser;
-- GRANT  CREATE ON SCHEMA public  TO cloudsqlsuperuser;

-- ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO PUBLIC;
-- ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT INSERT ON TABLES TO webuser;

-- ALTER SCHEMA public OWNER TO cloudsqlsuperuser;