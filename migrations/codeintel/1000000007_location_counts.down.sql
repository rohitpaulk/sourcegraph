BEGIN;

ALTER TABLE lsif_data_definitions DROP COLUMN IF EXISTS schema_version;
ALTER TABLE lsif_data_definitions DROP COLUMN IF EXISTS num_locations;
ALTER TABLE lsif_data_references DROP COLUMN IF EXISTS schema_version;
ALTER TABLE lsif_data_references DROP COLUMN IF EXISTS num_locations;

COMMIT;
