## Goal

- File Upload allows users to upload images; the system stores the file and provides a URL for sharing or reuse

- Uploads are handled through `POST /api/v1/files` with authentication; the file is sent as multipart form data

- Files are stored based on the configured storage driver (Local, S3, or R2). If Local, files are served under `/storage/...`; if S3/R2, a public URL is returned

- The system generates a safe, unique filename so files don’t overwrite each other

- On upload, the system records file metadata (original name, file type, size, upload time, and uploader) for tracking

- If saving metadata fails, the uploaded file is cleaned up to avoid orphan files

- Users can access files using `GET /api/v1/files/{filename}`; if the file is stored remotely, the user is redirected to the public URL

- To keep quality and safety, uploads are limited to 5MB and only accept JPG, JPEG, PNG, WEBP

- Error states are clear to users:
  - Missing file in request returns a 400 error
  - File not found returns a 404 error
  - Storage or database errors return a 500 error
  - Invalid type or size is rejected

- Configuration is driven by environment variables:
  - `STORAGE_DRIVER` = local | s3 | r2
  - Local: `STORAGE_LOCAL_PATH`
  - S3: `STORAGE_S3_BUCKET`, `STORAGE_S3_REGION`, `STORAGE_S3_ACCESS_KEY`, `STORAGE_S3_SECRET_KEY`
  - R2: `STORAGE_S3_ENDPOINT`, `STORAGE_S3_PUBLIC_URL` (plus bucket and keys)
  - Optional prefix: `STORAGE_S3_PREFIX`

#### Definition Of Done

[x] Upload succeeds and returns a shareable URL

[x] File can be accessed via URL (local `/storage/...` or public S3/R2 URL)

[x] File metadata is stored (original name, file type, size, upload time, uploader)

[x] Size (max 5MB) and file type (JPG/JPEG/PNG/WEBP) validation works

[x] Cleanup happens if metadata save fails

[x] Error responses are returned for bad request, not found, and server errors

[x] Storage configuration variables are documented and used
