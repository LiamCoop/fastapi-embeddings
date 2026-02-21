package kb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"ragtime-backend/internal/domain"
)

var (
	ErrNotFound     = errors.New("knowledge base not found")
	ErrInvalidInput = errors.New("invalid input")
)

// KnowledgeBase re-exports the domain model for handler convenience.
type KnowledgeBase = domain.KnowledgeBase

// Service manages knowledge base lifecycle.
type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
		now:  func() time.Time { return time.Now().UTC() },
	}
}

func NewServiceWithPostgres(db *sql.DB) *Service {
	repo := NewPostgresRepository(db)
	return NewService(repo)
}

type CreateRequest struct {
	Name     string
	Metadata map[string]any
}

type UpdateRequest struct {
	Name     *string
	Metadata map[string]any
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (*KnowledgeBase, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("repository is required")
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, validationError("name is required")
	}
	metadata := req.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}

	now := s.now()
	record := KnowledgeBaseRecord{
		ID:        uuid.NewString(),
		Name:      name,
		Metadata:  metadata,
		CreatedAt: now,
		UpdatedAt: now,
	}

	inserted, err := s.repo.InsertKnowledgeBase(ctx, record)
	if err != nil {
		return nil, err
	}

	kb := toDomain(inserted)
	return &kb, nil
}

func (s *Service) List(ctx context.Context) ([]KnowledgeBase, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("repository is required")
	}

	records, err := s.repo.ListKnowledgeBases(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]KnowledgeBase, 0, len(records))
	for _, record := range records {
		record := record
		items = append(items, toDomain(&record))
	}

	return items, nil
}

func (s *Service) Get(ctx context.Context, id string) (*KnowledgeBase, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("repository is required")
	}
	if err := validateID(id); err != nil {
		return nil, err
	}

	record, err := s.repo.GetKnowledgeBase(ctx, id)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, nil
	}

	kb := toDomain(record)
	return &kb, nil
}

func (s *Service) Update(ctx context.Context, id string, req UpdateRequest) (*KnowledgeBase, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("repository is required")
	}
	if err := validateID(id); err != nil {
		return nil, err
	}
	if req.Name == nil && req.Metadata == nil {
		return nil, validationError("no fields to update")
	}

	existing, err := s.repo.GetKnowledgeBase(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrNotFound
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, validationError("name cannot be empty")
		}
		existing.Name = name
	}
	if req.Metadata != nil {
		existing.Metadata = req.Metadata
	}

	existing.UpdatedAt = s.now()
	updated, err := s.repo.UpdateKnowledgeBase(ctx, *existing)
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, ErrNotFound
	}

	kb := toDomain(updated)
	return &kb, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if s.repo == nil {
		return fmt.Errorf("repository is required")
	}
	if err := validateID(id); err != nil {
		return err
	}

	ok, err := s.repo.DeleteKnowledgeBase(ctx, id)
	if err != nil {
		return err
	}
	if !ok {
		return ErrNotFound
	}

	return nil
}

func validateID(id string) error {
	if strings.TrimSpace(id) == "" {
		return validationError("kb_id is required")
	}
	if _, err := uuid.Parse(id); err != nil {
		return validationError("kb_id must be a valid UUID")
	}
	return nil
}

func validationError(message string) error {
	return fmt.Errorf("%w: %s", ErrInvalidInput, message)
}

func toDomain(record *KnowledgeBaseRecord) KnowledgeBase {
	metadata := record.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}

	return domain.KnowledgeBase{
		ID:        record.ID,
		Name:      record.Name,
		Metadata:  metadata,
		CreatedAt: record.CreatedAt,
		UpdatedAt: record.UpdatedAt,
	}
}
