\set ON_ERROR_STOP on

-- Este fichero se ejecuta como migrator después de SET ROLE owner. Permite a la
-- API operar únicamente los objetos que ya existen; no concede privilegios por
-- defecto sobre tablas o secuencias futuras.
GRANT USAGE ON SCHEMA public TO :"app_role";
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO :"app_role";
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO :"app_role";
