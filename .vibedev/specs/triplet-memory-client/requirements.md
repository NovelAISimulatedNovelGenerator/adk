# Requirements: Triplet Memory Client

## 1. Introduction

This document outlines the requirements for a Go client that interfaces with a local triplet memory service. This client, named `TripletMemoryService`, will be part of the `pkg/memory` package and will provide a Go-native way to interact with a REST API for storing and retrieving structured data in the form of subject-predicate-object triplets. The implementation will be similar in style and structure to the existing `CustomRagMemoryService`.

## 2. Requirements

### 2.1. Core Client Implementation

*   **User Story:** As a developer, I want a Go client that can securely connect to a local triplet memory service using Basic Authentication, so that I can easily and safely integrate triplet-based memory into my Go applications.

*   **Acceptance Criteria:**
    1.  **[R2.1.1]** The system **shall** provide a `TripletMemoryService` struct that holds the configuration for the client, including the base URL of the triplet memory service, and optional `Username` and `Password` for Basic Authentication.
    2.  **[R2.1.2]** The system **shall** provide a `NewTripletMemoryService(baseURL, username, password string)` constructor function that returns a new instance of the `TripletMemoryService`.
    3.  **[R2.1.3]** The `TripletMemoryService` **shall** use a configurable HTTP client for all requests to the triplet memory service.
    4.  **[R2.1.4]** If `Username` and `Password` are provided, the client **shall** automatically add a `Basic Auth` header to all outgoing HTTP requests.
    5.  **[R2.1.5]** The client **shall** be implemented in a new file named `pkg/memory/triplet_memory_service.go`.
    6.  **[R2.1.6]** All methods that make external HTTP requests **shall** accept a `context.Context` as their first argument and use it to manage request timeouts and cancellations.

### 2.2. Health Check

*   **User Story:** As a developer, I want to be able to check the health of the triplet memory service, so that I can ensure it is available before making other requests.

*   **Acceptance Criteria:**
    1.  **[R2.2.1]** The `TripletMemoryService` **shall** expose a public method `HealthCheck(ctx context.Context) error`.
    2.  **[R2.2.2]** The `HealthCheck` method **shall** send a GET request to a `/health` endpoint on the configured base URL.
    3.  **[R2.2.3]** The `HealthCheck` method **shall** return `nil` if the service returns a 2xx status code, and an error otherwise.

### 2.3. Add Triplet Functionality

*   **User Story:** As a developer, I want to be able to add new triplets to the memory service and get a unique identifier for each, so that I can store and reference structured information.

*   **Acceptance Criteria:**
    1.  **[R2.3.1]** The `TripletMemoryService` **shall** expose a public method `AddTriplet(ctx context.Context, triplet Triplet) (*Triplet, error)`.
    2.  **[R2.3.2]** The system **shall** define a `Triplet` struct with fields for `ID`, `Subject`, `Predicate`, and `Object`. `ID` may be optional when creating a new triplet.
    3.  **[R2.3.3]** The `AddTriplet` method **shall** send a POST request to `/triplets`.
    4.  **[R2.3.4]** The `AddTriplet` method **shall** serialize the `Triplet` struct (without the ID) into a JSON payload for the request body.
    5.  **[R2.3.5]** The `AddTriplet` method **shall** deserialize the JSON response from the service into a `Triplet` struct, which includes the server-generated `ID`.
    6.  **[R2.3.6]** The `AddTriplet` method **shall** return the created `Triplet` and a `nil` error on success.

### 2.4. Search Triplets Functionality

*   **User Story:** As a developer, I want to be able to search for triplets based on their subject, predicate, or object, so that I can retrieve relevant information from the memory.

*   **Acceptance Criteria:**
    1.  **[R2.4.1]** The `TripletMemoryService` **shall** expose a public method `SearchTriplets(ctx context.Context, query TripletSearchQuery) ([]*Triplet, error)`.
    2.  **[R2.4.2]** The system **shall** define a `TripletSearchQuery` struct that allows for querying by `Subject`, `Predicate`, and `Object`. At least one of these fields must be non-empty.
    3.  **[R2.4.3]** The `SearchTriplets` method **shall** send a POST request to `/triplets/search`.
    4.  **[R2.4.4]** The `SearchTriplets` method **shall** serialize the `TripletSearchQuery` struct into a JSON payload for the request body.
    5.  **[R2.4.5]** The `SearchTriplets` method **shall** deserialize the JSON response from the service into a slice of `*Triplet` structs.
    6.  **[R2.4.6]** If the search returns no results, the `SearchTriplets` method **shall** return an empty slice and no error.

### 2.5. Future Extensions (Optional)

*   **User Story:** As a developer, I want the client to be designed in a way that allows for future extensions like delete and batch operations, so that the client can evolve with the service.

*   **Acceptance Criteria:**
    1.  **[R2.5.1]** The design **should** allow for the future addition of a `DeleteTriplet(ctx context.Context, id string) error` method that sends a DELETE request to `/triplets/{id}`.
    2.  **[R2.5.2]** The design **should** allow for the future addition of batch operations, such as `AddTriplets(ctx context.Context, triplets []Triplet) ([]*Triplet, error)` and `SearchTripletsBatch(ctx context.Context, queries []TripletSearchQuery) ([][]*Triplet, error)`.
