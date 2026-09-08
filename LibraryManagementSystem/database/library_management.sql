
---------------------------------------------------
-- ROLES
---------------------------------------------------
CREATE TABLE roles (
    role_id SERIAL PRIMARY KEY,
    role_name VARCHAR(30) UNIQUE NOT NULL
);

INSERT INTO roles(role_name)
VALUES
('Admin'),
('Librarian'),
('Student');

---------------------------------------------------
-- USERS
---------------------------------------------------
CREATE TABLE users (
    user_id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role_id INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY(role_id)
        REFERENCES roles(role_id)
);

---------------------------------------------------
-- AUTHORS
---------------------------------------------------
CREATE TABLE authors (
    author_id SERIAL PRIMARY KEY,
    first_name VARCHAR(50),
    last_name VARCHAR(50)
);

---------------------------------------------------
-- CATEGORIES
---------------------------------------------------
-- A category is a (genre, language) pair — e.g. ('Romance', 'English')
-- and ('የፍቅር (Romance)', 'Amharic') are two distinct rows.
CREATE TABLE categories (
    category_id SERIAL PRIMARY KEY,
    category_name VARCHAR(100) NOT NULL,
    language VARCHAR(20) NOT NULL DEFAULT 'English',
    UNIQUE(category_name, language)
);

---------------------------------------------------
-- BOOKS
---------------------------------------------------
CREATE TABLE books (
    book_id SERIAL PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    isbn VARCHAR(30) UNIQUE,
    publication_year INT,
    category_id INT,
    total_copies INT DEFAULT 1,
    available_copies INT DEFAULT 1,

    FOREIGN KEY(category_id)
        REFERENCES categories(category_id)
);

---------------------------------------------------
-- BOOK AUTHORS (Many-to-Many)
---------------------------------------------------
CREATE TABLE book_authors (
    book_id INT,
    author_id INT,

    PRIMARY KEY(book_id, author_id),

    FOREIGN KEY(book_id)
        REFERENCES books(book_id)
        ON DELETE CASCADE,

    FOREIGN KEY(author_id)
        REFERENCES authors(author_id)
        ON DELETE CASCADE
);

---------------------------------------------------
-- LOANS
---------------------------------------------------
CREATE TABLE loans (
    loan_id VARCHAR(64) PRIMARY KEY,

    user_id INT NOT NULL,
    book_id INT NOT NULL,

    borrow_date DATE NOT NULL,
    due_date DATE NOT NULL,
    return_date DATE,

    status VARCHAR(20) DEFAULT 'Borrowed',

    FOREIGN KEY(user_id)
        REFERENCES users(user_id),

    FOREIGN KEY(book_id)
        REFERENCES books(book_id)
);


---------------------------------------------------
-- LOGIN AUDIT
---------------------------------------------------
CREATE TABLE login_logs (
    log_id SERIAL PRIMARY KEY,

    user_id INT,

    login_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    ip_address VARCHAR(100),

    FOREIGN KEY(user_id)
        REFERENCES users(user_id)
);

---------------------------------------------------
-- SEED DATA
---------------------------------------------------


INSERT INTO categories(category_name, language) VALUES ('Software Engineering', 'English');

INSERT INTO authors(first_name, last_name) VALUES
('Eric', 'Evans'),
('Robert C.', 'Martin');

INSERT INTO books(title, isbn, category_id, total_copies, available_copies) VALUES
('Domain-Driven Design', '9780321125217', 1, 2, 2),
('Clean Architecture', '9780134494166', 1, 1, 1);

INSERT INTO book_authors(book_id, author_id) VALUES
(1, 1),
(2, 2);

-- Hardcoded admin, seeded on every fresh database, mirroring the
-- old in-memory bootstrap.
INSERT INTO users(username, first_name, last_name, email, password_hash, role_id) VALUES
('SUPRADMN', 'Super', 'Admin', 'supradmn@library.local', 'password', (SELECT role_id FROM roles WHERE role_name = 'Admin'));

-- Two demo members with login credentials, mirroring the old in-memory
-- seed of member-1 (Alice) and member-2 (Bob).
INSERT INTO users(username, first_name, last_name, email, password_hash, role_id) VALUES
('alice', 'Alice', '', 'alice@library.local', 'password', (SELECT role_id FROM roles WHERE role_name = 'Student')),
('bob', 'Bob', '', 'bob@library.local', 'password', (SELECT role_id FROM roles WHERE role_name = 'Student'));