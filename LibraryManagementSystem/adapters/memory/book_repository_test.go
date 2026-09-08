package memory

import (
	"context"
	"errors"
	"library/domain"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBookRepository_IncrementAvailable_CannotExceedCopies(t *testing.T) {
	ctx := context.Background()
	repo := NewBookRepository(map[string]*domain.Book{
		"book-1": {ID: "book-1", Title: "Test Book", Copies: 1, Available: 1}, // already "full"
	})
	_ = ctx
	_ = repo

	incremented, err := repo.IncrementAvailable(ctx, "book-1") // This should be rejected.
	assert.True(t, errors.Is(err, domain.ErrAvailableExceedsCopies))
	found, err := repo.FindByID(ctx, "book-1") // Verify that the book's Available count has not changed.
	assert.NoError(t, err)
	assert.Equal(t, 1, found.Available) // The Available count should still be 1, not 2.
	_ = incremented
}

func TestBookRepository_IncrementAvailable_NormalCaseStillWorks(t *testing.T) {
	ctx := context.Background()
	repo := NewBookRepository(map[string]*domain.Book{
		"book-1": {ID: "book-1", Title: "Test Book", Copies: 2, Available: 1},
	})
	_ = ctx
	_ = repo

	incremented, err := repo.IncrementAvailable(ctx, "book-1") // This should be fine
	assert.NoError(t, err)
	assert.Equal(t, domain.Book{ID: "book-1", Title: "Test Book", Copies: 2, Available: 2}, incremented) // Verify that the book's Available count has increased to 2.
	_, err = repo.IncrementAvailable(ctx, "book-1")                                                      // This should be rejected
	assert.True(t, errors.Is(err, domain.ErrAvailableExceedsCopies))
	var _ = assert.Equal
}

func TestBookRepository_Create_StartsFullyAvailable(t *testing.T) {
	ctx := context.Background()
	repo := NewBookRepository(nil)

	book, err := repo.Create(ctx, domain.Book{
		Title:    "Things Fall Apart",
		Author:   "Chinua Achebe",
		Language: domain.LanguageEnglish,
		Genre:    "Drama",
		Copies:   3,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, book.ID)
	assert.Equal(t, 3, book.Copies)
	assert.Equal(t, 3, book.Available)

	found, err := repo.FindByID(ctx, book.ID)
	assert.NoError(t, err)
	assert.Equal(t, book, found)
}

func TestBookRepository_Create_AssignsDistinctIDs(t *testing.T) {
	ctx := context.Background()
	repo := NewBookRepository(nil)

	first, err := repo.Create(ctx, domain.Book{Title: "A", Author: "A", Copies: 1})
	assert.NoError(t, err)
	second, err := repo.Create(ctx, domain.Book{Title: "B", Author: "B", Copies: 1})
	assert.NoError(t, err)

	assert.NotEqual(t, first.ID, second.ID)
}

func TestBookRepository_IncreaseCopies_RaisesBothCopiesAndAvailable(t *testing.T) {
	ctx := context.Background()
	repo := NewBookRepository(map[string]*domain.Book{
		"book-1": {ID: "book-1", Title: "Test Book", Copies: 2, Available: 1},
	})

	book, err := repo.IncreaseCopies(ctx, "book-1", 3)
	assert.NoError(t, err)
	assert.Equal(t, 5, book.Copies)
	assert.Equal(t, 4, book.Available)
}

func TestBookRepository_IncreaseCopies_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := NewBookRepository(nil)

	_, err := repo.IncreaseCopies(ctx, "does-not-exist", 1)
	assert.True(t, errors.Is(err, domain.ErrNotFound))
}

func TestBookRepository_IncreaseCopies_RejectsNonPositiveAmount(t *testing.T) {
	ctx := context.Background()
	repo := NewBookRepository(map[string]*domain.Book{
		"book-1": {ID: "book-1", Title: "Test Book", Copies: 2, Available: 1},
	})

	_, err := repo.IncreaseCopies(ctx, "book-1", 0)
	assert.True(t, errors.Is(err, domain.ErrInvalidBookInput))

	found, err := repo.FindByID(ctx, "book-1")
	assert.NoError(t, err)
	assert.Equal(t, 1, found.Available)
	assert.Equal(t, 2, found.Copies)
}
