CREATE DATABASE pl;

CREATE TABLE IF NOT EXISTS issue
(
    id               SERIAL PRIMARY KEY,
    issue_hash       VARCHAR(255) UNIQUE,
    issue_title      VARCHAR(255),
    scenario_id      VARCHAR(255),
    scenario_version VARCHAR(255)
);

CREATE TABLE incident
(
    id                       SERIAL PRIMARY KEY,
    issue_hash               VARCHAR(255),
    trace_id                 VARCHAR(40),
    incident_collection_time TIMESTAMP,
    CONSTRAINT unique_issue UNIQUE (issue_hash, trace_id)
);

CREATE TABLE IF NOT EXISTS span
(
    id               SERIAL PRIMARY KEY,
    trace_id         VARCHAR(255),
    span_id          VARCHAR(255),
    parent_span_id   VARCHAR(255),
    source           VARCHAR(255),
    destination      VARCHAR(255),
    workload_id_list TEXT[],
    status           VARCHAR(255),
    metadata         TEXT,
    latency_ms       FLOAT,
    protocol         VARCHAR(255),
    issue_hash_list  TEXT[],
    time             TIMESTAMP,
    CONSTRAINT unique_span UNIQUE (trace_id, span_id)
);

CREATE TABLE IF NOT EXISTS span_raw_data
(
    id               SERIAL PRIMARY KEY,
    trace_id         VARCHAR(255),
    span_id          VARCHAR(255),
    request_payload  BYTEA,
    response_payload BYTEA,
    CONSTRAINT unique_span_raw_data UNIQUE (trace_id, span_id)
);