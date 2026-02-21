# Document Service Contract Requirements

## Purpose
This document defines the document intake contract validated by tests in the document package. It captures the current expectations for document creation, versioning, and storage behavior.

## Requirements

### Versioning Behavior
- Each upload creates a new document version.
- Re-uploading the same document path within the same knowledge base increments the version number.
- Re-uploads update the existing document record instead of creating a new document ID.

### Processing Status
- Supported document types result in `STORED` processing status after raw content is persisted.
- Unsupported document types result in `SKIPPED_UNSUPPORTED` processing status and do not proceed to downstream processing.

### Object Storage
- When `raw_content_uri` is not provided, the service uploads raw content to the object store.
- When `raw_content_uri` is provided, the service skips object store uploads and preserves the provided URI.

## Tests (Implemented)

### `TestUploadCreatesNewVersion`
- Verifies version numbers increment for repeated uploads.
- Verifies the document ID remains stable for the same KB/path.
- Verifies supported types transition to `STORED` and call object store `Put`.

### `TestUploadUnsupportedTypeSkipsProcessing`
- Verifies unsupported types result in `SKIPPED_UNSUPPORTED` status.

### `TestUploadWithRawContentURISkipsPut`
- Verifies provided `raw_content_uri` is preserved.
- Verifies object store `Put` is skipped when `raw_content_uri` is provided.
