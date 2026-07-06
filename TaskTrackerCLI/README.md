This Go program is a terminal-based Task Tracker CLI application that lets you manage a to-do list using simple command-line arguments. 
It stores all tasks locally inside a file named tasks.json, mimicking a persistent database.

Core ArchitectureData Structure: It maps tasks into a Task struct tracking a unique ID, a text Description, a Status (todo, in-progress, or done), and automated tracking timestamps (CreatedAt and UpdatedAt).

Storage Mechanism: The data is encoded to/from JSON (json.MarshalIndent/json.Unmarshal) and read or written directly to the file system using Go's os library.

Supported CLI Commands

add: Saves a new task into the JSON file with an auto-incrementing ID. 

The default status is set to todo.

list: Prints out saved tasks.

If you provide an extra status modifier (e.g., todo), it filters out everything else.

update: Locates a specific task by its numerical ID, edits its description, and refreshes its modification timestamp.

delete: Removes a task from the JSON list using its unique ID by slicing the underlying data array.

mark-in-progress / mark-done: Directly updates a target task's status flag to reflect workflow progression.
