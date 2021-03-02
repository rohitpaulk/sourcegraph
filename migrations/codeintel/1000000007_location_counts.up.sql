BEGIN;

-- faster to supply default than manual update
ALTER TABLE lsif_data_definitions ADD COLUMN schema_version int DEFAULT 1 NOT NULL;
ALTER TABLE lsif_data_definitions ADD COLUMN num_locations int DEFAULT 0 NOT NULL;
ALTER TABLE lsif_data_references ADD COLUMN schema_version int DEFAULT 1 NOT NULL;
ALTER TABLE lsif_data_references ADD COLUMN num_locations int DEFAULT 0 NOT NULL;

-- drop default after all existing columns have been set
ALTER TABLE lsif_data_definitions ALTER COLUMN schema_version DROP DEFAULT;
ALTER TABLE lsif_data_definitions ALTER COLUMN num_locations DROP DEFAULT;
ALTER TABLE lsif_data_references ALTER COLUMN schema_version DROP DEFAULT;
ALTER TABLE lsif_data_references ALTER COLUMN num_locations DROP DEFAULT;

COMMENT ON COLUMN lsif_data_definitions.schema_version IS 'The schema version of this row - used to determine presence and encoding of data.';
COMMENT ON COLUMN lsif_data_definitions.num_locations IS 'The number of locations stored in the data field.';
COMMENT ON COLUMN lsif_data_references.schema_version IS 'The schema version of this row - used to determine presence and encoding of data.';
COMMENT ON COLUMN lsif_data_references.num_locations IS 'The number of locations stored in the data field.';

COMMIT;
