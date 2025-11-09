# Project Documentation

## Overview

This Go codebase is designed for an iOS application backend.

## Architecture Guidelines

### Clean Architecture Principles

All code in this project must follow **Clean Architecture** principles:

#### Layer Structure

1. **Entities (Domain Layer)**
   - Core business logic and domain models
   - Independent of frameworks and external agencies
   - Contains enterprise-wide business rules

2. **Use Cases (Application Layer)**
   - Application-specific business rules
   - Orchestrates data flow between entities and interfaces
   - Independent of UI and database implementations

3. **Interface Adapters (Presentation Layer)**
   - Converts data between use cases and external agencies
   - Contains controllers, presenters, and gateways
   - Adapts data formats for different purposes

4. **Frameworks & Drivers (Infrastructure Layer)**
   - External frameworks and tools
   - Database implementations
   - Web frameworks
   - iOS-specific integrations

#### Key Principles

- **Dependency Rule**: Dependencies should point inward. Inner layers should not know about outer layers.
- **Independence**: Business rules can be tested without UI, database, web server, or external elements.
- **Testability**: Each layer can be tested independently.
- **Framework Independence**: Architecture doesn't depend on specific frameworks.

#### Directory Structure Example

```
/domain          - Entities and business logic
/usecase         - Application business rules
/repository      - Data access interfaces
/delivery        - HTTP handlers, controllers
/infrastructure  - Database, external services
```

## Development Guidelines

- All new features must follow clean architecture patterns
- Keep dependencies pointing inward
- Write unit tests for each layer independently
- Use dependency injection for loose coupling
- Maintain clear separation of concerns
