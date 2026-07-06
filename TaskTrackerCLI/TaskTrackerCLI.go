package main

// JSON which stands for JavaScript Object Notation is often used to transmit data between a server and a web application as text.
// the json here is used as a simple database to store tasks for a task tracker CLI application.
import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"
)

// This struct represents a task in the task tracker
// The json: tags are used to specify how the struct fields should be serialized to JSON
type Task struct {
	ID          int       `json:"id"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Database struct {
	Tasks  []Task `json:"tasks"`
	NextID int    `json:"next_id"`
}

// The database is stored in a JSON file named tasks.json
const dbFile = "tasks.json"

func loadDB() *Database {
	db := &Database{Tasks: []Task{}, NextID: 1}
	if file, err := os.ReadFile(dbFile); err == nil {
		json.Unmarshal(file, db)
	}
	return db
}

func saveDB(db *Database) {
	data, _ := json.MarshalIndent(db, "", "  ")
	os.WriteFile(dbFile, data, 0644)
}

func main() {
	fmt.Println("Task Tracker CLI")
	if len(os.Args) < 2 {
		fmt.Println("Usage: task-cli [command] [args]")
		return
	}

	db := loadDB()
	command := os.Args[1]

	switch command {
	case "add":
		if len(os.Args) < 3 {
			fmt.Println("Please provide a task description")
			return
		}
		task := Task{
			ID:          db.NextID,
			Description: os.Args[2],
			Status:      "todo",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		db.Tasks = append(db.Tasks, task)
		db.NextID++
		saveDB(db)
		fmt.Printf("Task added successfully (ID: %d)\n", task.ID)

	case "list":
		for _, t := range db.Tasks {
			if len(os.Args) > 2 && t.Status != os.Args[2] {
				continue
			}
			fmt.Printf("[%d] %s - %s\n", t.ID, t.Description, t.Status)
		}

	case "update":
		if len(os.Args) < 4 {
			fmt.Println("Usage: task-cli update [id] [new_description]")
			return
		}
		id, _ := strconv.Atoi(os.Args[2])
		for i := range db.Tasks {
			if db.Tasks[i].ID == id {
				db.Tasks[i].Description = os.Args[3]
				db.Tasks[i].UpdatedAt = time.Now()
				saveDB(db)
				fmt.Printf("Task %d updated\n", id)
				return
			}
		}

	case "delete":
		if len(os.Args) < 3 {
			fmt.Println("Please provide a task id")
			return
		}
		id, _ := strconv.Atoi(os.Args[2])
		for i, t := range db.Tasks {
			if t.ID == id {
				db.Tasks = append(db.Tasks[:i], db.Tasks[i+1:]...)
				saveDB(db)
				fmt.Printf("Task %d deleted\n", id)
				return
			}
		}

	case "mark-in-progress", "mark-done":
		if len(os.Args) < 3 {
			fmt.Println("Please provide a task id")
			return
		}
		id, _ := strconv.Atoi(os.Args[2])
		status := "in-progress"
		if command == "mark-done" {
			status = "done"
		}
		for i := range db.Tasks {
			if db.Tasks[i].ID == id {
				db.Tasks[i].Status = status
				db.Tasks[i].UpdatedAt = time.Now()
				saveDB(db)
				fmt.Printf("Task %d marked as %s\n", id, status)
				return
			}
		}

	default:
		fmt.Println("Unknown command")
	}
}
