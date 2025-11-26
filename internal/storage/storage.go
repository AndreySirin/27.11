package storage

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/AndreySirin/newProject-28-11/internal/entity"
	"go.etcd.io/bbolt"
	"time"
)

type Storage struct {
	db *bbolt.DB
}

func New(dbPath string) (*Storage, error) {
	db, err := bbolt.Open(dbPath, 0600, &bbolt.Options{
		Timeout: 1 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists([]byte("tasks"))
		if err != nil {
			return fmt.Errorf("create bucket %s: %s", "tasks", err)
		}
		if _, err = tx.CreateBucketIfNotExists([]byte("task_queue")); err != nil {
			return fmt.Errorf("create bucket task_queue: %w", err)
		}
		if _, err = tx.CreateBucketIfNotExists([]byte("meta")); err != nil {
			return fmt.Errorf("create bucket meta: %w", err)
		}
		return nil
	})
	return &Storage{
		db: db,
	}, nil
}
func (s *Storage) SaveTask(task *entity.Task) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		data, err := json.Marshal(task)
		if err != nil {
			return fmt.Errorf("failed to marshal task: %w", err)
		}

		tasksBucket := tx.Bucket([]byte("tasks"))
		if tasksBucket == nil {
			return fmt.Errorf("bucket tasks not found")
		}
		if err = tasksBucket.Put(integerToBytes(task.ID), data); err != nil {
			return err
		}
		queueBucket := tx.Bucket([]byte("task_queue"))
		if queueBucket == nil {
			return fmt.Errorf("bucket task_queue not found")
		}
		if err = queueBucket.Put(integerToBytes(task.ID), integerToBytes(task.ID)); err != nil {
			return fmt.Errorf("add task to queue: %w", err)
		}
		return nil
	})

}

func (s *Storage) UpdateTask(task *entity.Task) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		data, err := json.Marshal(task)
		if err != nil {
			return fmt.Errorf("failed to marshal task: %w", err)
		}
		tasksBucket := tx.Bucket([]byte("tasks"))
		if tasksBucket == nil {
			return fmt.Errorf("bucket tasks not found")
		}
		return tasksBucket.Put(integerToBytes(task.ID), data)
	})
}

func (s *Storage) GetTask(id uint64) (*entity.Task, error) {
	var task entity.Task
	err := s.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("tasks"))
		data := bucket.Get(integerToBytes(id))
		if data == nil {
			return fmt.Errorf("task not found: %s", id)
		}
		return json.Unmarshal(data, &task)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch task %s: %w", id, err)
	}
	return &task, nil
}

func (s *Storage) GetAllTaskId() ([]uint64, error) {
	var ids []uint64

	err := s.db.View(func(tx *bbolt.Tx) error {
		queueBucket := tx.Bucket([]byte("task_queue"))
		if queueBucket == nil {
			return fmt.Errorf("bucket task_queue not found")
		}
		c := queueBucket.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			ids = append(ids, binary.BigEndian.Uint64(k))
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("get all pending ids: %w", err)
	}
	return ids, nil
}

func (s *Storage) RemoveTaskIdFromQueue(taskID uint64) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		queueBucket := tx.Bucket([]byte("task_queue"))
		if queueBucket == nil {
			return fmt.Errorf("bucket task_queue not found")
		}
		if err := queueBucket.Delete(integerToBytes(taskID)); err != nil {
			return fmt.Errorf("failed to delete task %s from queue: %w", taskID, err)
		}
		return nil
	})
}
func (s *Storage) SetTaskID() (uint64, error) {
	var id uint64
	err := s.db.Update(func(tx *bbolt.Tx) error {
		meta := tx.Bucket([]byte("meta"))
		last := meta.Get([]byte("last_task_id"))
		if last != nil {
			id = binary.BigEndian.Uint64(last)
		}
		id++
		return meta.Put([]byte("last_task_id"), integerToBytes(id))
	})
	if err != nil {
		return 0, fmt.Errorf("failed to set task id: %w", err)
	}
	return id, nil
}

func integerToBytes(id uint64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, id)
	return buf
}
