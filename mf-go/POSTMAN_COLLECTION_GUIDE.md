# Postman Collection - Endpoints & Scenarios Guide

This document provides a complete overview of all endpoints and test scenarios available in the MasterFabric Go Postman collection.

## Collection Structure

The collection is organized into the following main folders:

1. **Health & Observability** (3 requests)
2. **Authentication (IAM)** (3 requests)
3. **User Management** (4 requests)
4. **Organization (Tenant) Management** (4 requests)
5. **Application Management** (4 requests)
6. **API Key Management** (5 requests)
7. **Endpoint Management (API Gateway)** (11 requests)
8. **RBAC (Role-Based Access Control)** (10 requests)
9. **Invoke Defined Endpoints** (10 requests)
10. **Audit Logs** (3 requests)
11. **Error Scenarios** (5 requests)

**Total: 62+ requests covering all API functionality**

---

## 1. Health & Observability

### Endpoints:
- **GET** `/health/live` - Liveness Probe
- **GET** `/health/ready` - Readiness Probe (checks PostgreSQL & Redis)
- **GET** `/metrics` - Prometheus Metrics

### Scenarios:
- ✅ Health check validation
- ✅ Database connectivity check
- ✅ Metrics endpoint verification

---

## 2. Authentication (IAM)

### Endpoints:
- **POST** `/api/v1/auth/register` - Register User
- **POST** `/api/v1/auth/register` - Register Second User (for testing)
- **POST** `/api/v1/auth/login` - Login
- **POST** `/api/v1/auth/login` - Login (Invalid Password) - Negative test

### Scenarios:
- ✅ User registration with validation
- ✅ JWT token auto-capture
- ✅ Login with valid credentials
- ✅ Login failure handling
- ✅ Token storage in collection variables

---

## 3. User Management

### Endpoints:
- **GET** `/api/v1/users/me` - Get Current User (Me)
- **GET** `/api/v1/users` - List Users
- **GET** `/api/v1/users/{userId}` - Get User by ID
- **GET** `/api/v1/users/{userId}` - Get User (Not Found) - Negative test

### Scenarios:
- ✅ Current user profile retrieval
- ✅ User listing with pagination
- ✅ User lookup by ID
- ✅ 404 error handling

---

## 4. Organization (Tenant) Management

### Endpoints:
- **POST** `/api/v1/organizations` - Create Organization
- **POST** `/api/v1/organizations` - Create Organization (Validation Error) - Negative test
- **GET** `/api/v1/organizations` - List Organizations
- **GET** `/api/v1/organizations/{orgId}` - Get Organization by ID
- **GET** `/api/v1/organizations/{orgId}` - Get Organization (Not Found) - Negative test

### Scenarios:
- ✅ Organization creation
- ✅ Organization ID auto-capture
- ✅ Organization listing
- ✅ Organization retrieval
- ✅ Validation error handling
- ✅ Multi-tenant isolation verification

---

## 5. Application Management

### Endpoints:
- **POST** `/api/v1/organizations/{orgId}/apps` - Create Application
- **GET** `/api/v1/organizations/{orgId}/apps` - List Applications
- **GET** `/api/v1/organizations/{orgId}/apps/{appId}` - Get Application by ID
- **GET** `/api/v1/organizations/{orgId}/apps/{appId}` - Get Application (Not Found) - Negative test

### Scenarios:
- ✅ Application creation under organization
- ✅ Application ID auto-capture
- ✅ Application listing
- ✅ Application retrieval
- ✅ Organization-app relationship validation

---

## 6. API Key Management

### Endpoints:
- **POST** `/api/v1/organizations/{orgId}/apps/{appId}/keys` - Create API Key
- **POST** `/api/v1/organizations/{orgId}/apps/{appId}/keys` - Create API Key (Staging)
- **GET** `/api/v1/organizations/{orgId}/apps/{appId}/keys` - List API Keys
- **GET** `/api/v1/organizations/{orgId}/apps/{appId}/keys/{keyId}` - Get API Key by ID
- **DELETE** `/api/v1/organizations/{orgId}/apps/{appId}/keys/{keyId}` - Revoke API Key

### Scenarios:
- ✅ API key creation (production)
- ✅ API key creation (staging)
- ✅ API key secret auto-capture
- ✅ API key listing
- ✅ API key retrieval
- ✅ API key revocation
- ✅ Secret masking verification

---

## 7. Endpoint Management (API Gateway)

### Endpoints:

#### Define Endpoints:
- **POST** `/api/v1/organizations/{orgId}/apps/{appId}/endpoints` - Define Endpoint (GET /products)
  - `backend_action: "list"`
- **POST** `/api/v1/organizations/{orgId}/apps/{appId}/endpoints` - Define Endpoint (POST /orders)
  - `backend_action: "create"`
- **POST** `/api/v1/organizations/{orgId}/apps/{appId}/endpoints` - Define Endpoint (GET /orders/{id})
  - `backend_action: "get"`
- **POST** `/api/v1/organizations/{orgId}/apps/{appId}/endpoints` - Define Endpoint (PUT /orders/{id})
  - `backend_action: "update"`
- **POST** `/api/v1/organizations/{orgId}/apps/{appId}/endpoints` - Define Endpoint (DELETE /orders/{id})
  - `backend_action: "delete"`
- **POST** `/api/v1/organizations/{orgId}/apps/{appId}/endpoints` - Define Endpoint (Custom Table Name)
  - Example with custom table name in schema

#### Manage Endpoints:
- **GET** `/api/v1/organizations/{orgId}/apps/{appId}/endpoints` - List Endpoints
- **GET** `/api/v1/organizations/{orgId}/apps/{appId}/endpoints/{endpointId}` - Get Endpoint by ID
- **GET** `/api/v1/organizations/{orgId}/apps/{appId}/endpoints/{endpointId}` - Get Endpoint (Not Found) - Negative test

#### Endpoint Policies:
- **PUT** `/api/v1/organizations/{orgId}/apps/{appId}/endpoints/{endpointId}/policy` - Update Endpoint Policy
- **GET** `/api/v1/organizations/{orgId}/apps/{appId}/endpoints/{endpointId}/policy` - Get Endpoint Policy

#### Endpoint Lifecycle:
- **POST** `/api/v1/organizations/{orgId}/apps/{appId}/endpoints/{endpointId}/retire` - Retire Endpoint

### Scenarios:
- ✅ Define endpoints with all CRUD actions (list, get, create, update, delete)
- ✅ Endpoint ID auto-capture
- ✅ Endpoint listing with pagination
- ✅ Endpoint retrieval
- ✅ Policy configuration (permissions, rate limits, validation)
- ✅ Endpoint retirement
- ✅ Schema validation examples
- ✅ Custom table name configuration

---

## 8. RBAC (Role-Based Access Control)

### Endpoints:

#### Role Assignment:
- **POST** `/api/v1/users/{userId}/roles` - Assign Role to User
- **GET** `/api/v1/users/{userId}/roles` - List User Roles
- **DELETE** `/api/v1/users/{userId}/roles/{roleId}` - Remove Role from User

#### Role Management:
- **GET** `/api/v1/organizations/{orgId}/roles` - List Roles by Organization
- **GET** `/api/v1/organizations/{orgId}/apps/{appId}/roles` - List Roles by App
- **GET** `/api/v1/roles/{roleId}` - Get Role by ID

#### Permission Management:
- **GET** `/api/v1/users/{userId}/permissions` - Get User Permissions
- **GET** `/api/v1/roles/{roleId}/permissions` - Get Role Permissions
- **POST** `/api/v1/roles/{roleId}/permissions` - Add Permission to Role
- **DELETE** `/api/v1/roles/{roleId}/permissions/{permission}` - Remove Permission from Role

### Scenarios:
- ✅ Role assignment to users
- ✅ Role listing (organization/app scope)
- ✅ Permission management
- ✅ User permission aggregation
- ✅ Role-permission relationships
- ✅ Multi-scope role handling

---

## 9. Invoke Defined Endpoints

### Endpoints:

#### CRUD Operations:
- **GET** `/api/v1/products` - Invoke GET /products Endpoint (list action)
- **POST** `/api/v1/orders` - Invoke POST /orders Endpoint (create action)
- **GET** `/api/v1/orders/{orderId}` - Invoke GET /orders/{id} Endpoint (get action)
- **PUT** `/api/v1/orders/{orderId}` - Invoke PUT /orders/{id} Endpoint (update action)
- **DELETE** `/api/v1/orders/{orderId}` - Invoke DELETE /orders/{id} Endpoint (delete action)
- **GET** `/api/v1/orders?offset=0&limit=20` - Invoke GET /orders with Pagination

#### Authentication Methods:
- **GET** `/api/v1/products` - Invoke Endpoint with API Key Auth

#### Error Scenarios:
- **GET** `/api/v1/products` - Invoke Endpoint (Missing X-App-ID) - Negative test
- **GET** `/api/v1/products` - Invoke Endpoint (Insufficient Permissions) - Negative test
- **GET** `/api/v1/products` - Invoke Endpoint (Rate Limit Exceeded) - Negative test
- **POST** `/api/v1/orders` - Invoke Endpoint (Schema Validation Failed) - Negative test

### Scenarios:
- ✅ Dynamic CRUD operations (list, get, create, update, delete)
- ✅ Pagination support
- ✅ API key authentication
- ✅ JWT token authentication
- ✅ Gateway pipeline validation
- ✅ Multi-tenant isolation
- ✅ Error handling (missing headers, permissions, rate limits, validation)

---

## 10. Audit Logs

### Endpoints:
- **GET** `/api/v1/audit/organizations/{orgId}` - List Audit Logs by Organization
- **GET** `/api/v1/audit/users/{userId}` - List Audit Logs by User
- **GET** `/api/v1/audit` - List Audit Logs with Filters

### Scenarios:
- ✅ Audit log retrieval by organization
- ✅ Audit log retrieval by user
- ✅ Filtered audit log queries
- ✅ Pagination support
- ✅ Date range filtering

---

## 11. Error Scenarios

### Endpoints:
- **GET** `/api/v1/users/me` - Unauthorized (No Token) - Negative test
- **GET** `/api/v1/users/me` - Unauthorized (Invalid Token) - Negative test
- **GET** `/api/v1/organizations/00000000-0000-0000-0000-000000000000` - Not Found (404) - Negative test
- **POST** `/api/v1/organizations` - Validation Error (Empty Body) - Negative test
- **POST** `/api/v1/auth/register` - Duplicate Registration - Negative test

### Scenarios:
- ✅ Authentication error handling
- ✅ Authorization error handling
- ✅ Not found error handling
- ✅ Validation error handling
- ✅ Duplicate resource error handling

---

## Collection Variables

The collection uses the following variables (auto-populated):

- `base_url` - API base URL (default: `http://localhost:8080`)
- `jwt_token` - JWT authentication token (auto-captured from login)
- `user_id` - Current user ID (auto-captured)
- `org_id` - Organization ID (auto-captured)
- `app_id` - Application ID (auto-captured)
- `api_key_id` - API Key ID (auto-captured)
- `api_key_secret` - API Key secret (auto-captured)
- `endpoint_id` - Endpoint ID (auto-captured)
- `role_id` - Role ID (auto-captured)
- `order_id` - Order ID (for testing dynamic endpoints)

---

## Test Scripts

Each request includes automated test scripts that:

- ✅ Verify HTTP status codes
- ✅ Validate response structure
- ✅ Check required headers (X-Request-ID, etc.)
- ✅ Auto-capture IDs and tokens
- ✅ Verify response time (< 2s)
- ✅ Validate JSON structure
- ✅ Check multi-tenant isolation
- ✅ Verify pagination metadata

---

## Quick Start Workflow

1. **Health Check** → Verify service is running
2. **Register** → Create user account
3. **Login** → Get JWT token (auto-captured)
4. **Create Organization** → Get org_id (auto-captured)
5. **Create Application** → Get app_id (auto-captured)
6. **Create API Key** → Get api_key_secret (auto-captured)
7. **Define Endpoints** → Create endpoints with different actions
8. **Invoke Endpoints** → Test dynamic CRUD operations
9. **Configure Policies** → Set up permissions and rate limits
10. **Test RBAC** → Assign roles and test permissions

---

## Dynamic CRUD Actions Coverage

The collection includes examples for all 5 built-in actions:

1. **`list`** - GET with pagination
   - Example: `GET /products` with `?offset=0&limit=20`

2. **`get`** - GET by ID
   - Example: `GET /orders/{id}`

3. **`create`** - POST to create
   - Example: `POST /orders` with JSON body

4. **`update`** - PUT/PATCH to update
   - Example: `PUT /orders/{id}` with JSON body

5. **`delete`** - DELETE record
   - Example: `DELETE /orders/{id}`

---

## Notes

- All requests automatically include required headers (`X-Organization-ID`, `X-App-ID`) when available
- JWT tokens are automatically captured and reused
- IDs are auto-captured and stored in collection variables
- Negative test cases are included for comprehensive coverage
- The collection supports both JWT and API key authentication
- Multi-tenant isolation is tested throughout
