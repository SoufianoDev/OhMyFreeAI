# Persistent Storage Architecture

Persistent storage is a user-selected directory or supported cloud-backed artifact location. The application stores model/engine artifacts, manifests, checksums and download metadata there.

For Google Drive, treat the mount as a persistence layer only. Do not execute binaries or load weights directly from the mount. Copy to local runtime disk, verify the copy and then use it. Reuse verified artifacts; never redownload solely because a runtime session restarted.

SQLite remains local application metadata storage and is not the model artifact store. Avoid huge file counts in a single Drive root; use structured subdirectories and manifests.
