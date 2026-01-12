package wish

import (
	"context"
	"errors"
	"time"
)

// Core
type Wish struct {
	ID          int
	OwnerEmail  string
	Title       string
	Description string
	IsBought    bool
	BoughtAt    *time.Time
	CreatedAt   time.Time
}

// Errors
var (
	ErrOwnerEmailRequired = errors.New("email not provided")
	ErrTitleRequired      = errors.New("title not provided")
	ErrWishAlreadyBought  = errors.New("wish already bought")
	ErrWishNotFound       = errors.New("wish not found")
)

// Interfaces

// Abstraction over repository
type WishRepository interface {
	Create(ctx context.Context, wish *Wish) error
	GetByID(ctx context.Context, id int) (*Wish, error)
	List(ctx context.Context, ownerEmail string, bought *bool) ([]*Wish, error)
	Update(ctx context.Context, wish *Wish) error
	Delete(ctx context.Context, id int) error
	Stats(ctx context.Context, ownerEmail string) (Stats, error)
}

// Business logic for handlers
type WishService interface {
	CreateWish(ctx context.Context, ownerEmail, title, description string) (*Wish, error)
	GetWish(ctx context.Context, id int) (*Wish, error)
	ListWishes(ctx context.Context, ownerEmail string, bought *bool) ([]*Wish, error)
	UpdateWish(ctx context.Context, id int, title, description string) (*Wish, error)
	DeleteWish(ctx context.Context, id int) error
	BuyWish(ctx context.Context, id int) error
	GetStats(ctx context.Context, ownerEmail string) (Stats, error)
}

// DTO for stats
type Stats struct {
	Total   int `json:"total"`
	Bought  int `json:"bought"`
	Pending int `json:"pending"`
}

type service struct {
	repo WishRepository
}

func NewService(repo WishRepository) WishService {
	return &service{repo: repo}
}

func (s *service) CreateWish(ctx context.Context, ownerEmail, title, description string) (*Wish, error) {
	if ownerEmail == "" {
		return nil, ErrOwnerEmailRequired
	}
	if title == "" {
		return nil, ErrTitleRequired
	}

	wish := &Wish{
		OwnerEmail:  ownerEmail,
		Title:       title,
		Description: description,
		CreatedAt:   time.Now(),
	}

	if err := s.repo.Create(ctx, wish); err != nil {
		return nil, err
	}

	return wish, nil
}

func (s *service) GetWish(ctx context.Context, id int) (*Wish, error) {
	wish, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return wish, nil
}

func (s *service) ListWishes(ctx context.Context, ownerEmail string, bought *bool) ([]*Wish, error) {
	return s.repo.List(ctx, ownerEmail, bought)
}

func (s *service) UpdateWish(ctx context.Context, id int, title, description string) (*Wish, error) {
	wish, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	wish.Title = title
	wish.Description = description
	err = s.repo.Update(ctx, wish)

	return wish, err
}

func (s *service) DeleteWish(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *service) BuyWish(ctx context.Context, id int) error {
	wish, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if wish.IsBought {
		return ErrWishAlreadyBought
	}

	wish.IsBought = true
	now := time.Now()
	wish.BoughtAt = &now

	return s.repo.Update(ctx, wish)
}

func (s *service) GetStats(ctx context.Context, ownerEmail string) (Stats, error) {
	return s.repo.Stats(ctx, ownerEmail)
}
