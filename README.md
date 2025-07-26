# Hospital Tracker API

A comprehensive hospital management and tracking platform built with Go, featuring user management, clinic administration, and staff tracking capabilities.

## Features

### Core Functionality
- **Hospital Registration**: Register hospitals with unique constraints (Tax ID, Email, Phone)
- **User Authentication**: JWT-based authentication with email/phone login
- **Password Reset**: Phone-based verification system with temporary codes
- **User Management**: Sub-user creation with role-based access (Authorized/Employee)
- **Clinic Management**: Add clinics from predefined types with staff tracking
- **Staff Management**: Complete CRUD operations with pagination and filtering
- **Geographic Data**: Province and district management with relationships
- **Profession System**: Hierarchical profession groups and titles

### Technical Features
- **Caching**: Dragonfly (Redis-compatible) caching for geographic and profession data
- **Database**: PostgreSQL with GORM ORM
- **Documentation**: Auto-generated Swagger/OpenAPI documentation
- **Containerization**: Docker and Docker Compose support
- **Security**: Password hashing, JWT tokens, input validation

## Tech Stack

- **Language**: Go 1.24
- **Framework**: Gin
- **Database**: PostgreSQL
- **ORM**: GORM
- **Cache**: Dragonfly (Redis-compatible)
- **Documentation**: Swagger/OpenAPI with swaggo
- **Containerization**: Docker with distroless images

## Quick Start

```bash
git clone <repository-url>
cd hospital-tracker
docker-compose up -d
```

3. The API will be available at:
   - API: http://localhost:8080
   - Swagger Documentation: http://localhost:8080/swagger/index.html
   - Health Check: http://localhost:8080/health

### Manual Setup

**Prerequisites**:
   - Go 1.24+
   - PostgreSQL 15+
   - Dragonfly or Redis

```bash
cp .env.example .env
# edit .env with your configuration
go mod download
swag init
go run main.go
```

## API Endpoints

### Public Endpoints
- `POST /api/register` - Hospital registration
- `POST /api/login` - User login
- `POST /api/password-reset/request` - Request password reset
- `POST /api/password-reset/confirm` - Confirm password reset
- `GET /api/provinces` - Get provinces
- `GET /api/districts` - Get districts
- `GET /api/clinic-types` - Get clinic types
- `GET /api/profession-groups` - Get profession groups

### Protected Endpoints (Require Authentication)
- `GET /api/users` - List users
- `GET /api/users/:id` - Get user details
- `GET /api/clinics` - List hospital clinics
- `GET /api/staff` - List staff (with pagination/filtering)
- `GET /api/staff/:id` - Get staff details

### Authorized User Only (Admin Functions)
- `POST /api/users` - Create sub-user
- `PUT /api/users/:id` - Update user
- `DELETE /api/users/:id` - Delete user
- `POST /api/clinics` - Add clinic
- `DELETE /api/clinics/:id` - Remove clinic
- `POST /api/staff` - Add staff member
- `PUT /api/staff/:id` - Update staff member
- `DELETE /api/staff/:id` - Remove staff member

## Authentication & Admin Account Management

### **User Types & Roles**

The system has **2 user types**:
1. **`authorized`** - **Admin/Hospital Owner** (Full management access)
2. **`employee`** - **Regular Employee** (Read-only access)

### **How to Create Admin Accounts**

There are **2 ways** to create admin accounts:

#### **1. Hospital Registration (Primary Admin Creation)**
When a new hospital registers, the **first user automatically becomes an admin**:

**Endpoint:** `POST /api/register`

**Request Example:**
```json
{
  "hospital_name": "General Hospital",
  "tax_id": "1234567890",
  "email": "hospital@example.com",
  "phone": "+90-555-0101",
  "province_id": 1,
  "district_id": 1,
  "address": "123 Hospital St",
  
  "first_name": "John",
  "last_name": "Doe",
  "national_id": "12345678901",
  "user_email": "admin@example.com",
  "user_phone": "+90-555-0102",
  "password": "admin123"
}
```

**What happens:**
- Hospital is created
- User is created with `user_type: "authorized"` (admin)
- This user becomes the **hospital owner/primary admin**

#### **2. Admin Creates Another Admin**
An existing admin can create additional admin users:

**Endpoint:** `POST /api/users` (requires admin access)

**Request Example:**
```json
{
  "first_name": "Jane",
  "last_name": "Smith",
  "national_id": "98765432101",
  "email": "jane@example.com",
  "phone": "+90-555-0103",
  "password": faker.Password(),
  "user_type": "authorized"
}
```

### **Permission System**

#### **What Authorized Users (Admins) Can Do:**
- ✅ Create/update/delete users (both admins and employees)
- ✅ Create/delete clinics
- ✅ Create/update/delete staff
- ✅ All read operations

#### **What Employee Users Can Do:**
- ✅ Read operations only (users, clinics, staff)
- ❌ Cannot create/modify/delete anything

### **Authentication Flow**

1. **Login:** `POST /api/login`
   ```json
   {
     "identifier": "admin@example.com", // email OR phone
     "password": "admin123"
   }
   ```

2. **Get JWT Token:** Response includes token with user info
   ```json
   {
     "token": "eyJ...",
     "user_type": "authorized",
     "user": { ... }
   }
   ```

3. **Use Token:** Add to requests: `Authorization: Bearer <token>`

### **Quick Start: Create Your First Admin**

```bash
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{
    "hospital_name": "My Hospital",
    "tax_id": "1234567890",
    "email": "hospital@mycompany.com",
    "phone": "+90-555-0001",
    "province_id": 1,
    "district_id": 1,
    "address": "Hospital Address",
    "first_name": "Admin",
    "last_name": "User",
    "national_id": "12345678901",
    "user_email": "admin@mycompany.com",
    "user_phone": "+90-555-0002",
    "password": "admin123"
  }'
```

### **Key Points**

1. **No Default Admin:** The database doesn't seed any default admin accounts
2. **Hospital-Scoped:** Each admin only manages their own hospital
3. **First User Rule:** Hospital registration creates the first admin automatically
4. **Admin Hierarchy:** Admins can create more admins within their hospital
5. **Isolation:** Users from different hospitals cannot interact

## Data Models

### User Types
- **Authorized**: Full access to all hospital management functions
- **Employee**: Read-only access, cannot add staff

### Profession Groups
- **Doktor**: Asistan, Uzman
- **İdari Personel**: Başhekim, Müdür
- **Hizmet Personeli**: Danışman, Temizlik, Güvenlik

### Business Rules
- Only one Başhekim (Chief Physician) per hospital
- Staff can belong to only one clinic
- Some roles (like security) may not be assigned to clinics
- Phone numbers and national IDs must be unique across the system

## Staff Filtering

The staff endpoint supports the following filters:
- `first_name` - Filter by first name (partial match)
- `last_name` - Filter by last name (partial match)
- `national_id` - Filter by national ID (partial match)
- `profession_group_id` - Filter by profession group
- `title_id` - Filter by specific title
- `page` - Page number (default: 1)
- `limit` - Items per page (default: 10, fixed)

Example:
```
GET /api/staff?first_name=John&profession_group_id=1&page=2&limit=10
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| PORT | Server port | 8080 |
| ENV | Environment (development/production) | development |
| DB_HOST | PostgreSQL host | localhost |
| DB_PORT | PostgreSQL port | 5432 |
| DB_USER | PostgreSQL username | postgres |
| DB_PASSWORD | PostgreSQL password | password |
| DB_NAME | Database name | hospital_tracker |
| DB_SSLMODE | SSL mode | disable |
| REDIS_HOST | Dragonfly/Redis host | localhost |
| REDIS_PORT | Dragonfly/Redis port | 6379 |
| REDIS_PASSWORD | Dragonfly/Redis password | (empty) |
| REDIS_DB | Dragonfly/Redis database | 0 |
| JWT_SECRET | JWT signing secret | your-secret-key |
| JWT_EXPIRE_HOURS | JWT expiration hours | 24 |
| LOG_LEVEL | Logging level (debug/info/warn/error) | info |
| LOG_FORMAT | Log format (console/json) | console |
| LOG_CONSOLE | Enable colored console output | true |

## License

This project is licensed under the MIT License.