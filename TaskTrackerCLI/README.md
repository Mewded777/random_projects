# Task Tracker CLI

A lightweight Command Line Interface (CLI) application built in Go to track and manage your daily tasks. This application uses a local JSON file as a persistent flat-file database to store, retrieve, update, and delete task objects without requiring an external database server.

The project demonstrates core Go concepts such as structural data modeling, command-line arguments parsing (`os.Args`), JSON serialization/deserialization (`encoding/json`), and local file-system operations (`os.ReadFile` and `os.WriteFile`).

## Features

- **Add Tasks**: Create new tasks with an automatically generated unique ID.
- **List Tasks**: View all tasks or filter them dynamically by status (todo, in-progress, done).
- **Update Tasks**: Modify descriptions of existing tasks by ID.
- **Delete Tasks**: Remove tasks permanently from the tracking history.
- **Status Updates**: Mark tasks as "in-progress" or "done" to track completion lifecycle.
- **Automatic Timestamps**: Tracks exact system times when tasks are created or updated.

---

## Prerequisites

To run this application, you only need:
- **Go** installed on your system.

---

## Installation & Setup

1. Save the code into a file named `main.go`.
2. Open your terminal in that folder and initialize your project tracking boundaries:
   ```bash
   go mod init task-tracker
   ```

---

## Usage Guide

Run the application by invoking the Go runtime toolchain and passing a specific sub-command alongside its arguments.

### 1. Add a New Task
```bash
go run main.go add "Buy groceries"
go run main.go add "Finish Go assignment"
```

### 2. List Tasks
You can list all tracked entries or filter them by passing a status criteria string:
```bash
# List all tasks
go run main.go list

# List only tasks that are in-progress
go run main.go list in-progress

# List only tasks that are done
go run main.go list done
```

### 3. Update a Task Description
Provide the specific task ID followed by your new text:
```bash
go run main.go update 1 "Buy groceries and meal prep"
```

### 4. Change Task Status
Transition tasks through your workflow using the marking commands:
```bash
# Mark a task as in-progress
go run main.go mark-in-progress 1

# Mark a task as done
go run main.go mark-done 1
```

### 5. Delete a Task
Remove an item by referencing its unique target identifier:
```bash
go run main.go delete 2
```

---

## Data Persistence
All operations automatically update a local file named `tasks.json` in your current working directory. If the file does not exist, the application creates it automatically upon adding your first task.

Example data structure inside `tasks.json`:
```json
{
  "tasks": [
    {
      "id": 1,
      "description": "Buy groceries and meal prep",
      "status": "done",
      "created_at": "2026-07-07T11:13:00Z",
      "updated_at": "2026-07-07T11:20:00Z"
    }
  ],
  "next_id": 2
}
```

---

## Repository Safeguards
To maintain a clean tracking environment, it is highly recommended to add a `.gitignore` file to your project folder before uploading it to GitHub. This keeps temporary local databases separate from your source code.

Create a `.gitignore` file and include:
```text
tasks.json
```
