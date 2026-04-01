-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS discovery_runs;
DROP TABLE IF EXISTS jobs;
DROP TABLE IF EXISTS review_schedule;
DROP TABLE IF EXISTS node_tags;
DROP TABLE IF EXISTS edges;
DROP TABLE IF EXISTS tags;
DROP TABLE IF EXISTS nodes;

-- Drop extensions
DROP EXTENSION IF EXISTS "pg_trgm";
DROP EXTENSION IF EXISTS "vector";
DROP EXTENSION IF EXISTS "uuid-ossp";
