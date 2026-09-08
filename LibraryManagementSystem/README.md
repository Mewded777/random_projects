-- Library Management System

A full-stack Library Management System built with Go, PostgreSQL, HTML templates, and a layered Clean Architecture-inspired design.

The application provides separate workflows for administrators, librarians, and members, including book management, borrowing/returns, member management, dashboards, and a small JSON API.

-- Features

-- Administrator
- Manage users
- Register librarians and members
- Manage books
- View borrowing records
- View library statistics

-- Librarian
- Manage books
- Issue and receive books
- View active loans
- Search books and members

-- Member
- Browse available books
- Borrow and return books
- View borrowing history
- View currently borrowed books

-- Tech Stack

- **Go 1.26+**
- `net/http`
- `html/template`
- PostgreSQL
- `database/sql`
- HTML5 / CSS3
- Go `testing` + Testify

-- Architecture

```text
LibraryManagementSystem/
├── adapters/
│   ├── memory/          # In-memory repositories for local/demo use
│   └── postgres/        # PostgreSQL repositories
├── app/                 # Application use cases
├── database/
│   ├── library_management.sql
│   └── README.md
├── domain/              # Domain models and errors
├── httpapi/             # JSON API
├── ports/               # Repository interfaces
├── webui/               # HTML web interface
├── .env.example
├── .gitignore
├── LICENSE
├── go.mod
├── go.sum
├── main.go
└── README.md
```

-- Requirements

Install:

- Go 1.26 or later
- PostgreSQL
- Git

-- Getting Started

--# 1. Clone the repository

```bash
git clone https://github.com/YOUR_USERNAME/LibraryManagementSystem.git
cd LibraryManagementSystem
```

--# 2. Configure PostgreSQL

Create the database:

```sql
CREATE DATABASE library_management;
```

Load the schema and seed data:

```bash
psql -U postgres -d library_management -f database/library_management.sql
```

--# 3. Configure the application

Copy the example environment file:

```bash
cp .env.example .env
```

Edit `.env` and replace the placeholder PostgreSQL password:

```env
DATABASE_URL=postgres://postgres:your_password@localhost:5432/library_management?sslmode=disable
```

`.env` is ignored by Git and should **never be committed**.

--# 4. Install dependencies

```bash
go mod download
```

--# 5. Run the application

```bash
go run .
```

The application starts:

- Web interface: `http://localhost:8080`
- JSON API: `http://localhost:8081`

The web application opens the login page automatically on supported desktop systems. If it does not, open `http://localhost:8080/login` manually.

-- Database

The complete schema and demo seed data are in:

```text
database/library_management.sql
```

For PostgreSQL-specific details, see [`database/README.md`](database/README.md).

> **Demo data:** The SQL file contains development/demo accounts. Change or remove those credentials before using the project in a real environment.

-- API

The JSON API runs on port `8081`. Current endpoints are implemented in `httpapi/`.

Example:

```bash
curl http://localhost:8081/health
```

If you add or change endpoints, update the API tests and documentation accordingly.

-- Development

Format the code:

```bash
go fmt ./...
```

Run tests:

```bash
go test ./...
```

Run the complete local checks:

```bash
go test ./...
```

