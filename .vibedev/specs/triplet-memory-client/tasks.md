# Task Plan: Quad Memory Client

This document breaks down the work required to implement the `QuadMemoryService` client, based on the approved design document.

## Task List

- [x] **Task 1: File and Type Scaffolding**
  - **Description:** Create the necessary files and define all the core data structures required for the service.
  - **Acceptance Criteria:**
    - Create a new file: `pkg/memory/quad_memory_service.go`.
    - In the new file, define the following structs: `QuadMemoryConfig`, `QuadMemoryService`, `Quad`, `AddQuadRequest`, `QuadSearchQuery`, `SearchQuadsRequest`.
    - Create a new test file: `pkg/memory/quad_memory_service_test.go`.

- [ ] **Task 2: Implement Constructor and Health Check**
  - **Description:** Implement the client's constructor and the `HealthCheck` method.
  - **Acceptance Criteria:**
    - Implement the `NewQuadMemoryService(config QuadMemoryConfig)` function.
    - Implement the `HealthCheck(ctx context.Context)` method.
    - Write unit tests in `quad_memory_service_test.go` for the `HealthCheck` method, covering success, failure (4xx/5xx), and context cancellation scenarios.

- [ ] **Task 3: Implement `AddQuad` Method**
  - **Description:** Implement the logic for adding a new quad to the memory service.
  - **Acceptance Criteria:**
    - Implement the `AddQuad(ctx context.Context, tenantID string, quad Quad) (*Quad, error)` method.
    - The method must correctly serialize the request payload, set the `Authorization` and `Tenant-ID` headers, and send the request.
    - The method must handle the response, deserializing the created quad and returning it.
    - The method must implement the retry logic for transient errors as specified in the design.
    - Write unit tests for `AddQuad`, including success, failure, and retry scenarios.

- [ ] **Task 4: Implement `SearchQuads` Method**
  - **Description:** Implement the logic for searching for quads in the memory service.
  - **Acceptance Criteria:**
    - Implement the `SearchQuads(ctx context.Context, tenantID string, query QuadSearchQuery) ([]*Quad, error)` method.
    - The method must correctly serialize the search query, set the necessary headers, and send the request.
    - The method must handle the response, deserializing the array of quads and returning it.
    - The method must implement the same retry logic as `AddQuad`.
    - Write unit tests for `SearchQuads`, covering success, failure (no results found), and retry scenarios.

- [ ] **Task 5: Finalize and Document**
  - **Description:** Add final touches, including code comments and updating the main project task file.
  - **Acceptance Criteria:**
    - Add package and function-level comments to `quad_memory_service.go`.
    - Review and ensure logging is adequate for debugging and monitoring.
    - Run all tests to confirm the feature is complete and correct.
    - Update the root `Task.md` to mark the `triplet-memory-client` task as complete.
