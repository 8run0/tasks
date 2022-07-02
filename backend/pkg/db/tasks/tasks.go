package tasks

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

const (
	tableName = "tasks"
	table     = `CREATE TABLE IF NOT EXISTS 'tasks' (
		'id' INTEGER PRIMARY KEY AUTOINCREMENT, 
		'title' VARCHAR(255), 
		'description' VARCHAR(1024),
		'created_on' TIMESTAMP, 
		'completed' INTEGER, 
		'completed_on' TIMESTAMP);`
)

type Task struct {
	ID          uint64    `db:"id"`
	Title       string    `db:"title"`
	Description string    `db:"description"`
	CreatedOn   time.Time `db:"created_on"`
	Completed   bool      `db:"completed"`
	CompletedOn time.Time `db:"completed_on"`
}

type taskDB interface {
	Create(ctx context.Context, task Task) (uint64, error)
	Update(ctx context.Context, task Task) error
	Delete(ctx context.Context, id uint64) error
	Complete(ctx context.Context, id uint64) error
	List(ctx context.Context, limit int64) ([]Task, error)
	GetByID(ctx context.Context, id uint64) (Task, error)
}

type DB struct {
	Conn *sqlx.DB
}

// Complete implements taskDB
func (db *DB) Complete(ctx context.Context, id uint64) error {
	tx, err := db.prepareTX(ctx)
	if err != nil {
		return err
	}
	stmt, err := db.prepareStatement(ctx, tx, "UPDATE tasks SET completed = ?, completed_on = ? WHERE id = ?;")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(1, time.Now(), id)
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}

var _ taskDB = &DB{}

func NewDB(dbURI string) *DB {
	db, _ := sqlx.Connect("sqlite3", dbURI)
	initTable(db)
	return &DB{
		Conn: db,
	}
}

func initTable(db *sqlx.DB) {
	_, err := db.Exec(table)
	if err != nil {
		log.Printf("failed to initalise table %s", tableName)
	}
	fmt.Printf("%s table initalised\n", tableName)
}

// Create implements taskDB
func (db *DB) Create(ctx context.Context, task Task) (uint64, error) {
	dbTask := Task{
		Title:       task.Title,
		Description: task.Description,
		CreatedOn:   time.Now(),
		Completed:   false,
		CompletedOn: time.Time{},
	}
	dbTask.CreatedOn = time.Now()

	tx, err := db.prepareTX(ctx)
	if err != nil {
		return 0, err
	}
	stmt, err := db.prepareStatement(ctx, tx, "INSERT INTO tasks (title, description, created_on, completed, completed_on) VALUES (?,?,?,?,?);")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	res, err := stmt.Exec(task.Title, task.Description, task.CreatedOn, task.Completed, task.CompletedOn)
	if err != nil {
		return 0, err
	}
	tx.Commit()
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return uint64(id), nil
}

func (db *DB) prepareStatement(ctx context.Context, tx *sql.Tx, sql string) (*sql.Stmt, error) {
	stmt, err := tx.PrepareContext(ctx, sql)
	if err != nil {
		return nil, err
	}
	return stmt, nil
}
func (db *DB) prepareTX(ctx context.Context) (*sql.Tx, error) {
	tx, err := db.Conn.BeginTx(ctx, &sql.TxOptions{
		Isolation: 0,
		ReadOnly:  false,
	})
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// Delete implements taskDB
func (db *DB) Delete(ctx context.Context, id uint64) error {
	tx, err := db.prepareTX(ctx)
	if err != nil {
		return err
	}
	stmt, err := db.prepareStatement(ctx, tx, "DELETE FROM tasks WHERE id = ?;")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(id)
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}

// Get1 implements taskDB
func (db *DB) GetByID(ctx context.Context, id uint64) (Task, error) {
	task := Task{}
	err := db.Conn.GetContext(ctx, &task, "SELECT * FROM tasks WHERE ID = ?;", id)
	return task, err
}

// List implements taskDB
func (db *DB) List(ctx context.Context, limit int64) ([]Task, error) {
	tasks := []Task{}
	err := db.Conn.SelectContext(ctx, &tasks, "SELECT * FROM tasks LIMIT ?;", limit)
	return tasks, err
}

// Update implements taskDB
func (db *DB) Update(ctx context.Context, task Task) error {
	tx, err := db.prepareTX(ctx)
	if err != nil {
		return err
	}
	stmt, err := db.prepareStatement(ctx, tx, "UPDATE tasks SET title = ?, description = ? WHERE id = ?;")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(task.Title, task.Description, task.ID)
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}
