# /internal

Thư mục này chứa code riêng của ứng dụng, không được import bởi các package bên ngoài.

## Cấu Trúc

```
/domain
  /email          # Email domain logic
  /calendar       # Calendar domain logic
  /user          # User management
  /common        # Shared types

/repository
  /postgres      # PostgreSQL implementations
  /redis        # Redis cache & queue
  /mock         # Mock implementations

/service
  /auth         # Authentication service
  /processor    # Business logic
  /scheduler    # Task scheduling

/transport
  /http         # HTTP handlers
  /grpc         # gRPC handlers
  /middleware   # Shared middleware
```

## Các Package

### Domain
- Business logic interfaces
- Domain models & types
- Service contracts
- Error definitions

### Repository
- Data access layer
- Cache management
- Queue operations
- Test doubles

### Service
- Use case implementations
- External integrations
- Background tasks
- Error handling

### Transport
- Request handling
- Response formatting
- Authentication
- Validation

## Quy Tắc
1. Package Visibility
   - Internal only
   - No external imports
   - Clear dependencies

2. Error Handling
   - Domain errors
   - Wrapped errors
   - Error types

3. Testing
   - Unit tests
   - Mocks
   - Test helpers 