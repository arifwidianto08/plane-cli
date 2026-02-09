## What's the Main Feature

The Suppliers module is a comprehensive supplier management system that allows merchants to manage their supplier relationships within the GloryX ERP platform. This feature enables merchants to maintain detailed supplier information, associate suppliers with inventory items and recipes, and track supplier performance.

### Core Functionality

1. **Supplier Profile Management**
   - Create, read, update, and delete supplier records
   - Store comprehensive supplier information including:
     - Unique supplier code (e.g., SUP001)
     - Company name and address details
     - City and province information
     - Shipment/delivery codes
     - Primary and secondary phone numbers
     - Email address and website
     - Additional notes

2. **Supplier-Item Association**
   - Link suppliers to specific inventory items
   - Support for primary/default supplier assignment
   - Item references stored as JSONB array with optional notes per item

3. **Supplier Status Management**
   - Active/inactive status toggle
   - Soft delete functionality with audit trail
   - Supplier filtering by status

4. **Multi-Tenant Architecture**
   - Scoped by merchant (tenant isolation)
   - Per-schema database design
   - UUID-based identification system

### API Endpoints Implemented

- **POST** `/api/v1/merchant/business/suppliers` - Create new supplier
- **GET** `/api/v1/merchant/business/suppliers/{id}` - Get supplier details
- **PUT** `/api/v1/merchant/business/suppliers/{id}` - Update supplier
- **DELETE** `/api/v1/merchant/business/suppliers/{id}` - Delete supplier (soft delete)
- **GET** `/api/v1/merchant/business/suppliers` - List suppliers with filtering

### Database Schema

**Table:** `tenant_base_schema.suppliers`

| Column                 | Type        | Description                          |
| ---------------------- | ----------- | ------------------------------------ |
| id                     | UUID        | Primary key, auto-generated          |
| merchant_id            | UUID        | Foreign key to merchants table       |
| code                   | TEXT        | Unique supplier code per merchant    |
| name                   | TEXT        | Supplier company name                |
| address                | TEXT        | Full address                         |
| city                   | TEXT        | City name                            |
| province               | TEXT        | Province/state                       |
| shipment_code          | TEXT        | Delivery/shipment identifier         |
| primary_phone_number   | TEXT        | Main contact number                  |
| secondary_phone_number | TEXT        | Alternative contact (optional)       |
| webpage                | TEXT        | Company website (optional)           |
| email                  | TEXT        | Contact email                        |
| is_active              | BOOLEAN     | Status flag (default: true)          |
| note                   | TEXT        | Additional notes (optional)          |
| items                  | JSONB       | Array of associated items with notes |
| created_by             | UUID        | User who created the record          |
| updated_by             | UUID        | User who last updated                |
| deleted_by             | UUID        | User who deleted (soft delete)       |
| created_at             | TIMESTAMPTZ | Creation timestamp                   |
| updated_at             | TIMESTAMPTZ | Last update timestamp                |
| deleted_at             | TIMESTAMPTZ | Soft delete timestamp                |

### Related Database Changes

Migration `00016_add_main_supplier_to_items_and_recipes.sql` adds:

- `main_supplier_id` column to items table
- `main_supplier_id` column to recipes table
- Foreign key constraints linking to suppliers table
- Indexes for performance optimization

---

## Definition of Done

### Backend Implementation

- [x] Database schema created with proper indexes and constraints
- [x] CRUD API endpoints implemented with full validation
- [x] Supplier model with all required fields
- [x] Request/response schemas defined
- [x] Service layer with business logic
- [x] Repository layer for database operations
- [x] Handler layer with Echo framework
- [x] Swagger/OpenAPI documentation
- [x] Audit logging for all operations
- [x] Soft delete implementation
- [x] Multi-tenant support (schema-based)
- [x] JWT authentication and authorization
- [x] Input validation with go-playground/validator
- [x] Error handling and proper HTTP status codes

### API Features

- [x] Create supplier with validation
- [x] Get supplier by ID with expanded items
- [x] Update supplier with change tracking
- [x] Soft delete supplier with audit
- [x] List suppliers with pagination
- [x] Search/filter by name
- [x] Filter by active status
- [x] Filter by city/province
- [x] Code uniqueness validation per merchant
- [x] Item association support

### Database

- [x] Suppliers table with all columns
- [x] Proper indexes (merchant_id, is_active, deleted_at, name, code)
- [x] Unique constraint on merchant_id + code
- [x] Foreign key relationships
- [x] Trigger for updated_at timestamp
- [x] Migration files created and tested
- [x] Integration with items and recipes tables

### Testing & Quality

- [x] Swagger documentation complete
- [x] API tested via Swagger UI
- [x] All CRUD operations working
- [x] Validation working correctly
- [x] Error responses properly formatted
- [x] Audit logs recording correctly

---

## Figma/Design

### API Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    API Endpoints                             │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  POST /api/v1/merchant/business/suppliers                   │
│  ├── Authentication: Bearer Token                           │
│  ├── Headers: X-Tenant-Slug or X-Tenant-ID                  │
│  └── Body: CreateSupplierRequest                            │
│                                                              │
│  GET /api/v1/merchant/business/suppliers/{id}               │
│  ├── Authentication: Bearer Token                           │
│  └── Returns: SupplierResponse with items                   │
│                                                              │
│  PUT /api/v1/merchant/business/suppliers/{id}               │
│  ├── Authentication: Bearer Token                           │
│  ├── Headers: X-Tenant-Slug or X-Tenant-ID                  │
│  └── Body: UpdateSupplierRequest (partial updates)          │
│                                                              │
│  DELETE /api/v1/merchant/business/suppliers/{id}            │
│  ├── Authentication: Bearer Token                           │
│  └── Soft delete with audit logging                         │
│                                                              │
│  GET /api/v1/merchant/business/suppliers                    │
│  ├── Query Params:                                          │
│  │   ├── search (by name)                                   │
│  │   ├── is_active (boolean)                                │
│  │   ├── city (string)                                      │
│  │   ├── province (string)                                  │
│  │   ├── page (int)                                         │
│  │   └── limit (int)                                        │
│  └── Returns: ListSuppliersResponse with pagination         │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Data Flow

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   Handler    │────▶│   Service    │────▶│  Repository  │
└──────────────┘     └──────────────┘     └──────────────┘
       │                    │                    │
       │                    │                    │
       ▼                    ▼                    ▼
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  Validation  │     │   Business   │     │  Database    │
│  Swagger Doc │     │    Logic     │     │   (PostgreSQL)│
└──────────────┘     └──────────────┘     └──────────────┘
       │                    │                    │
       │                    │                    │
       ▼                    ▼                    ▼
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  Response    │     │  Audit Log   │     │  Tenant      │
│  DTO         │     │  (Common)    │     │  Schema      │
└──────────────┘     └──────────────┘     └──────────────┘
```

### Entity Relationship

```
┌─────────────────┐         ┌──────────────────┐
│    Merchants    │         │    Suppliers     │
├─────────────────┤         ├──────────────────┤
│ id (PK)         │◄────────┤ merchant_id (FK) │
│ ...             │         │ id (PK)          │
└─────────────────┘         │ code             │
                            │ name             │
                            │ address          │
                            │ city             │
                            │ province         │
                            │ email            │
                            │ is_active        │
                            │ items[]          │
                            │ ...              │
                            └──────────────────┘
                                     │
                                     │
                    ┌────────────────┼────────────────┐
                    │                │                │
                    ▼                ▼                ▼
            ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
            │    Items     │  │   Recipes    │  │  Audit Log   │
            ├──────────────┤  ├──────────────┤  ├──────────────┤
            │main_supplier │  │main_supplier │  │  Reference   │
            │    _id       │  │    _id       │  │    Type      │
            └──────────────┘  └──────────────┘  └──────────────┘
```

---

## Notes

### Implementation Details

1. **Multi-Tenancy**: The module uses a schema-based multi-tenant architecture where each tenant has their own PostgreSQL schema (`tenant_base_schema`). This ensures complete data isolation between merchants.

2. **Soft Delete**: All delete operations are soft deletes (setting `deleted_at` timestamp), preserving data integrity and enabling audit trails. Unique constraints exclude soft-deleted records.

3. **Audit Logging**: Every create, update, and delete operation is logged with user ID, IP address, user agent, old values, and new values for compliance and debugging.

4. **Code Uniqueness**: Supplier codes must be unique per merchant but can be reused across different merchants.

5. **Item Association**: Suppliers can be linked to items via a JSONB array that stores item IDs with optional notes. This allows flexible many-to-many relationships.

6. **Search Capability**: Full-text search is enabled on supplier names using PostgreSQL's GIN index with `to_tsvector`.

7. **Validation**: Comprehensive input validation using `go-playground/validator` with custom validation rules for:
   - Required fields (code, name, address, city, province, shipment_code, primary_phone, email)
   - Field length limits
   - Email format validation
   - URL validation for webpage

### Technical Stack

- **Framework**: Echo (Go)
- **Database**: PostgreSQL with pgx driver
- **ORM/Query Builder**: Jet (type-safe SQL builder)
- **Validation**: go-playground/validator
- **Authentication**: JWT Bearer tokens
- **Documentation**: Swagger/OpenAPI 2.0
- **UUID**: gofrs/uuid v7 (time-ordered)

### Integration Points

- **Items Module**: Suppliers can be set as main supplier for items
- **Recipes Module**: Suppliers can be associated with recipes
- **Audit Module**: All operations are audited via common audit service
- **Auth Module**: JWT middleware validates tokens and extracts user/merchant context

### Performance Considerations

- Database indexes on frequently queried columns (merchant_id, is_active, deleted_at)
- GIN index for full-text search on supplier names
- Pagination implemented for list endpoints
- JSONB storage for items array (flexible schema, good performance)

### Security

- All endpoints require Bearer authentication
- Tenant isolation enforced at database schema level
- Input validation prevents SQL injection
- Audit logging tracks all data modifications
- Soft delete prevents accidental data loss

---

**Status**: ✅ Fully Implemented and Tested

**Last Updated**: 2026-02-09

**Module Path**: `/modules/merchant/business/suppliers`

**Database Migrations**:

- `00014_create_suppliers_table.sql`
- `00016_add_main_supplier_to_items_and_recipes.sql`
