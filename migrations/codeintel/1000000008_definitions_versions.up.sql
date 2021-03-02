BEGIN;

CREATE TABLE lsif_data_definitions_schema_versions (
    dump_id integer NOT NULL,
    min_schema_version integer,
    max_schema_version integer
);
ALTER TABLE lsif_data_definitions_schema_versions ADD PRIMARY KEY (dump_id);

COMMENT ON TABLE lsif_data_definitions_schema_versions IS 'Tracks the range of schema_versions for each upload in the lsif_data_definitions table.';
COMMENT ON COLUMN lsif_data_definitions_schema_versions.dump_id IS 'The identifier of the associated dump in the lsif_uploads table.';
COMMENT ON COLUMN lsif_data_definitions_schema_versions.min_schema_version IS 'A lower-bound on the `lsif_data_definitions.schema_version` where `lsif_data_definitions.dump_id = dump_id`.';
COMMENT ON COLUMN lsif_data_definitions_schema_versions.max_schema_version IS 'An upper-bound on the `lsif_data_definitions.schema_version` where `lsif_data_definitions.dump_id = dump_id`.';

-- Ensure that there is a lsif_data_definitions_schema_versions record for each distinct dump_id in
-- lsif_data_definitions. It does not need to be precise, but needs to include all known schema versions
-- so far so that we re-trigger the progress check. This should update itself back to its true count
-- quickly.
INSERT INTO lsif_data_definitions_schema_versions SELECT DISTINCT dump_id AS dump_id, 1 AS min_schema_version, 2 AS max_schema_version FROM lsif_data_definitions;

CREATE OR REPLACE FUNCTION update_lsif_data_definitions_schema_versions_insert() RETURNS trigger AS $$ BEGIN
    INSERT INTO
        lsif_data_definitions_schema_versions
    SELECT
        dump_id,
        MIN(schema_version) as min_schema_version,
        MAX(schema_version) as max_schema_version
    FROM
        newtab
    GROUP BY
        dump_id
    ON CONFLICT (dump_id) DO UPDATE SET
        -- Update with min(old_min, new_min) and max(old_max, new_max)
        min_schema_version = LEAST(lsif_data_definitions_schema_versions.min_schema_version, EXCLUDED.min_schema_version),
        max_schema_version = GREATEST(lsif_data_definitions_schema_versions.max_schema_version, EXCLUDED.max_schema_version);

    RETURN NULL;
END $$ LANGUAGE plpgsql;

-- On every insert into lsif_data_definitions, we need to make sure we have an associated row in the
-- lsif_data_definitions_schema_versions table. We do not currently care about cleaning the table up
-- (we will do this asynchronously).
--
-- We use FOR EACH STATEMENT here because we batch insert into this table. Running the trigger per
-- statement rather than per row will save a ton of extra work. Running over batch inserts lets us
-- do a GROUP BY on the new table and effectively upsert our new ranges.
CREATE TRIGGER lsif_data_definitions_schema_versions_insert
AFTER INSERT ON lsif_data_definitions REFERENCING NEW TABLE AS newtab
FOR EACH STATEMENT EXECUTE PROCEDURE update_lsif_data_definitions_schema_versions_insert();

COMMIT;
