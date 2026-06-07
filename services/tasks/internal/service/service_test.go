package service

import (
	"context"
	"testing"
	"time"

	"pz8/services/tasks/internal/repository"
)

type taskRepoStub struct {
	createFn        func(ctx context.Context, params repository.CreateTaskParams) (repository.Task, error)
	listFn          func(ctx context.Context) ([]repository.Task, error)
	getFn           func(ctx context.Context, id string) (repository.Task, error)
	searchByTitleFn func(ctx context.Context, title string) ([]repository.Task, error)
	updateFn        func(ctx context.Context, id string, params repository.UpdateTaskParams) (repository.Task, error)
	deleteFn        func(ctx context.Context, id string) error
}

func (s *taskRepoStub) Create(ctx context.Context, params repository.CreateTaskParams) (repository.Task, error) {
	return s.createFn(ctx, params)
}

func (s *taskRepoStub) List(ctx context.Context) ([]repository.Task, error) {
	return s.listFn(ctx)
}

func (s *taskRepoStub) Get(ctx context.Context, id string) (repository.Task, error) {
	return s.getFn(ctx, id)
}

func (s *taskRepoStub) SearchByTitle(ctx context.Context, title string) ([]repository.Task, error) {
	return s.searchByTitleFn(ctx, title)
}

func (s *taskRepoStub) Update(ctx context.Context, id string, params repository.UpdateTaskParams) (repository.Task, error) {
	return s.updateFn(ctx, id, params)
}

func (s *taskRepoStub) Delete(ctx context.Context, id string) error {
	return s.deleteFn(ctx, id)
}

func TestCreateSanitizesInput(t *testing.T) {
	now := time.Date(2026, 4, 2, 12, 0, 0, 0, time.UTC)

	repo := &taskRepoStub{
		createFn: func(ctx context.Context, params repository.CreateTaskParams) (repository.Task, error) {
			if params.Title != "Task title" {
				t.Fatalf("unexpected title: %q", params.Title)
			}
			if params.Description != "&lt;b&gt;unsafe&lt;/b&gt;" {
				t.Fatalf("unexpected description: %q", params.Description)
			}
			if params.DueDate == nil || params.DueDate.Format("2006-01-02") != "2026-05-01" {
				t.Fatalf("unexpected due date: %#v", params.DueDate)
			}

			return repository.Task{
				ID:          "task-1",
				Title:       params.Title,
				Description: params.Description,
				DueDate:     params.DueDate,
				Done:        false,
				CreatedAt:   now,
			}, nil
		},
		listFn: func(ctx context.Context) ([]repository.Task, error) { return nil, nil },
		getFn:  func(ctx context.Context, id string) (repository.Task, error) { return repository.Task{}, nil },
		searchByTitleFn: func(ctx context.Context, title string) ([]repository.Task, error) {
			return nil, nil
		},
		updateFn: func(ctx context.Context, id string, params repository.UpdateTaskParams) (repository.Task, error) {
			return repository.Task{}, nil
		},
		deleteFn: func(ctx context.Context, id string) error { return nil },
	}

	svc := New(repo)

	task, err := svc.Create(context.Background(), CreateTaskInput{
		Title:       "   Task title   ",
		Description: "<b>unsafe</b>",
		DueDate:     "2026-05-01",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if task.ID != "task-1" {
		t.Fatalf("unexpected id: %q", task.ID)
	}
	if task.Description != "&lt;b&gt;unsafe&lt;/b&gt;" {
		t.Fatalf("unexpected description in dto: %q", task.Description)
	}
	if task.DueDate != "2026-05-01" {
		t.Fatalf("unexpected due date in dto: %q", task.DueDate)
	}
}

func TestCreateRequiresTitle(t *testing.T) {
	repo := &taskRepoStub{
		createFn: func(ctx context.Context, params repository.CreateTaskParams) (repository.Task, error) {
			return repository.Task{}, nil
		},
		listFn: func(ctx context.Context) ([]repository.Task, error) { return nil, nil },
		getFn:  func(ctx context.Context, id string) (repository.Task, error) { return repository.Task{}, nil },
		searchByTitleFn: func(ctx context.Context, title string) ([]repository.Task, error) {
			return nil, nil
		},
		updateFn: func(ctx context.Context, id string, params repository.UpdateTaskParams) (repository.Task, error) {
			return repository.Task{}, nil
		},
		deleteFn: func(ctx context.Context, id string) error { return nil },
	}

	svc := New(repo)

	_, err := svc.Create(context.Background(), CreateTaskInput{
		Title: "   ",
	})
	if err == nil {
		t.Fatalf("expected validation error")
	}
	if err.Error() != "title is required" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateChangesFields(t *testing.T) {
	now := time.Date(2026, 4, 2, 12, 0, 0, 0, time.UTC)

	repo := &taskRepoStub{
		createFn: func(ctx context.Context, params repository.CreateTaskParams) (repository.Task, error) {
			return repository.Task{}, nil
		},
		listFn: func(ctx context.Context) ([]repository.Task, error) { return nil, nil },
		getFn: func(ctx context.Context, id string) (repository.Task, error) {
			return repository.Task{
				ID:          id,
				Title:       "Old title",
				Description: "Old description",
				Done:        false,
				CreatedAt:   now,
			}, nil
		},
		searchByTitleFn: func(ctx context.Context, title string) ([]repository.Task, error) {
			return nil, nil
		},
		updateFn: func(ctx context.Context, id string, params repository.UpdateTaskParams) (repository.Task, error) {
			if id != "task-1" {
				t.Fatalf("unexpected id: %q", id)
			}
			if params.Title != "New title" {
				t.Fatalf("unexpected title: %q", params.Title)
			}
			if params.Description != "&lt;script&gt;1&lt;/script&gt;" {
				t.Fatalf("unexpected description: %q", params.Description)
			}
			if !params.Done {
				t.Fatalf("expected done=true")
			}

			return repository.Task{
				ID:          id,
				Title:       params.Title,
				Description: params.Description,
				Done:        params.Done,
				CreatedAt:   now,
			}, nil
		},
		deleteFn: func(ctx context.Context, id string) error { return nil },
	}

	svc := New(repo)

	title := "  New title  "
	description := "<script>1</script>"
	done := true

	task, err := svc.Update(context.Background(), "task-1", UpdateTaskInput{
		Title:       &title,
		Description: &description,
		Done:        &done,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if task.Title != "New title" {
		t.Fatalf("unexpected dto title: %q", task.Title)
	}
	if task.Description != "&lt;script&gt;1&lt;/script&gt;" {
		t.Fatalf("unexpected dto description: %q", task.Description)
	}
	if !task.Done {
		t.Fatalf("expected dto done=true")
	}
}
