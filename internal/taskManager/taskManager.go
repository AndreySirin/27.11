package taskManager

import (
	"bytes"
	"context"
	"fmt"
	"github.com/AndreySirin/newProject-28-11/internal/client"
	"github.com/AndreySirin/newProject-28-11/internal/entity"
	"github.com/AndreySirin/newProject-28-11/internal/storage"
	"github.com/jung-kurt/gofpdf"
	"log/slog"
	"strings"
)

type TaskManager struct {
	lg     *slog.Logger
	Ch     chan uint64
	Db     *storage.Storage
	Client *client.Client
}

func New(lg *slog.Logger, db *storage.Storage, client *client.Client) *TaskManager {
	return &TaskManager{
		lg:     lg,
		Ch:     make(chan uint64, 100),
		Db:     db,
		Client: client,
	}
}

func (s *TaskManager) Init() error {
	taskId, err := s.Db.GetAllTaskId()
	if err != nil {
		return fmt.Errorf("get all task id error: %v", err)
	}
	for _, id := range taskId {
		s.Ch <- id
	}
	return nil
}

func (s *TaskManager) RunWorker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			s.lg.Info("worker stopped")
			return
		case id := <-s.Ch:
			task, err := s.Db.GetTask(id)
			if err != nil || task.IsCheck == true {
				s.lg.Error("get task %d error: %v", id, err)
				continue
			}
			for i, link := range task.Links {
				status, errCheck := s.Client.CheckLink(link.Url)
				if errCheck != nil {
					s.lg.Error("check link %s error: %v", link.Url, errCheck)
				}
				task.Links[i].Status = status
			}
			task.IsCheck = true
			err = s.Db.UpdateTask(task)
			if err != nil {
				s.lg.Error("update task %s error: %v", id, err)
			}
			err = s.Db.RemoveTaskIdFromQueue(task.ID)
			if err != nil {
				s.lg.Error("remove task %s error: %v", task.ID, err)
			}
		}
	}
}

func (s *TaskManager) RequestToService(task entity.RequestTask) *entity.Task {
	links := make([]entity.Link, len(task.Links))

	for i, url := range task.Links {

		links[i] = entity.Link{
			Url:    url,
			Status: "",
		}
	}
	id, err := s.Db.SetTaskID()
	if err != nil {
		s.lg.Error("set task id error: %v", err)
		return nil
	}

	return &entity.Task{
		ID:      id,
		IsCheck: false,
		Links:   links,
	}
}

func (s *TaskManager) GenerateReport(taskId entity.RequestTaskList) ([]byte, error) {

	taskList := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(taskId)), ","), "[]")

	pdf := CreatePDF(taskList)

	for _, id := range taskId.Ids {
		task, err := s.Db.GetTask(id)
		if err != nil || task.IsCheck == false {
			s.lg.Error("get task %s error: %v", id, err)
			continue
		}
		for _, link := range task.Links {

			line := fmt.Sprintf("task number:%d.%s - %s)", task.ID, link.Url, link.Status)

			pdf.MultiCell(0, 8, line, "", "", false)
		}

	}
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return buf.Bytes(), nil
}
func CreatePDF(taskList string) *gofpdf.Fpdf {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 16)

	pdf.Cell(40, 10, fmt.Sprintf("Task Report #%s", taskList))
	pdf.Ln(12)
	pdf.SetFont("Arial", "", 12)

	return pdf
}
