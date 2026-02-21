ALTER TABLE document_versions
    ADD COLUMN IF NOT EXISTS error_message text;
