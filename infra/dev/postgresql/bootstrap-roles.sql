\set ON_ERROR_STOP on

-- Se ejecuta una sola vez sobre una base vacía con el administrador que crea la
-- imagen oficial de PostgreSQL. No contiene datos de aplicación ni se usa para
-- migraciones posteriores.
CREATE ROLE :"owner_role" NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
CREATE ROLE :"migrator_role" LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS PASSWORD :'migrator_password';
CREATE ROLE :"app_role" LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS PASSWORD :'app_password';

-- Solo migrator puede asumir temporalmente la propiedad del esquema.
GRANT :"owner_role" TO :"migrator_role";

ALTER DATABASE :"database_name" OWNER TO :"owner_role";
ALTER SCHEMA public OWNER TO :"owner_role";

REVOKE ALL ON DATABASE :"database_name" FROM PUBLIC;
GRANT CONNECT ON DATABASE :"database_name" TO :"migrator_role";
GRANT CONNECT ON DATABASE :"database_name" TO :"app_role";

REVOKE ALL ON SCHEMA public FROM PUBLIC;
