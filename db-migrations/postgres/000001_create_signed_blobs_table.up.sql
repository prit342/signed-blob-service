-- Signed Blob Storage Service - Simple Database Schema

-- Create extension for UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create table for storing signed blobs
CREATE TABLE IF NOT EXISTS signed_blobs (
    uuid UUID PRIMARY KEY,
    blob TEXT NOT NULL,
    hash VARCHAR(64) NOT NULL,
    -- due to marshalling, we store the timestamp as a string in RFC3339 format
    timestamp TEXT NOT NULL,
    signature BYTEA NOT NULL
);

-- Create indexes for performance
CREATE INDEX idx_signed_blobs_hash ON signed_blobs(hash);
CREATE INDEX idx_signed_blobs_timestamp ON signed_blobs(timestamp);
