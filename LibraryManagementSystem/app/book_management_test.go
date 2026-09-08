package app

import (
	"context"
	"errors"
	"testing"

	"library/adapters/memory"
	"library/domain"

	"github.com/stretchr/testify/assert"
)

func TestAddBookHandler_Success(t *testing.T) {
	bookRepo := memory.NewBookRepository(nil)
	add := NewAddBookHandler(bookRepo)
	ctx := context.Background()

	book, err := add.Handle(ctx, AddBookCommand{
		Title:    "Oliver Twist",
		Author:   "Charles Dickens",
		Language: domain.LanguageEnglish,
		Genre:    "Drama",
		Copies:   4,
	})

	assert.NoError(t, err)
	assert.NotEmpty(t, book.ID)
	assert.Equal(t, 4, book.Copies)
	assert.Equal(t, 4, book.Available)

	all, err := bookRepo.FindAll(ctx)
	assert.NoError(t, err)
	assert.Len(t, all, 1)
}

func TestAddBookHandler_RejectsMissingTitleOrAuthor(t *testing.T) {
	bookRepo := memory.NewBookRepository(nil)
	add := NewAddBookHandler(bookRepo)
	ctx := context.Background()

	_, err := add.Handle(ctx, AddBookCommand{
		Title:    "",
		Author:   "Someone",
		Language: domain.LanguageEnglish,
		Genre:    "Drama",
		Copies:   1,
	})
	assert.True(t, errors.Is(err, domain.ErrInvalidBookInput))

	_, err = add.Handle(ctx, AddBookCommand{
		Title:    "Something",
		Author:   "",
		Language: domain.LanguageEnglish,
		Genre:    "Drama",
		Copies:   1,
	})
	assert.True(t, errors.Is(err, domain.ErrInvalidBookInput))
}

func TestAddBookHandler_RejectsNonPositiveCopies(t *testing.T) {
	bookRepo := memory.NewBookRepository(nil)
	add := NewAddBookHandler(bookRepo)
	ctx := context.Background()

	_, err := add.Handle(ctx, AddBookCommand{
		Title:    "Something",
		Author:   "Someone",
		Language: domain.LanguageEnglish,
		Genre:    "Drama",
		Copies:   0,
	})
	assert.True(t, errors.Is(err, domain.ErrInvalidBookInput))
}

func TestAddBookHandler_RejectsUnknownLanguage(t *testing.T) {
	bookRepo := memory.NewBookRepository(nil)
	add := NewAddBookHandler(bookRepo)
	ctx := context.Background()

	_, err := add.Handle(ctx, AddBookCommand{
		Title:    "Something",
		Author:   "Someone",
		Language: "French",
		Genre:    "Drama",
		Copies:   1,
	})
	assert.True(t, errors.Is(err, domain.ErrInvalidBookInput))
}

func TestAddBookHandler_RejectsGenreFromWrongLanguage(t *testing.T) {
	bookRepo := memory.NewBookRepository(nil)
	add := NewAddBookHandler(bookRepo)
	ctx := context.Background()

	// "Comedy" is an English genre, not one of the Amharic genres.
	_, err := add.Handle(ctx, AddBookCommand{
		Title:    "Something",
		Author:   "Someone",
		Language: domain.LanguageAmharic,
		Genre:    "Comedy",
		Copies:   1,
	})
	assert.True(t, errors.Is(err, domain.ErrInvalidBookInput))
}

func TestIncreaseBookCopiesHandler_Success(t *testing.T) {
	bookRepo := memory.NewBookRepository(map[string]*domain.Book{
		"book-1": {ID: "book-1", Title: "Test Book", Copies: 2, Available: 1},
	})
	increase := NewIncreaseBookCopiesHandler(bookRepo)
	ctx := context.Background()

	book, err := increase.Handle(ctx, IncreaseBookCopiesCommand{BookID: "book-1", Amount: 2})
	assert.NoError(t, err)
	assert.Equal(t, 4, book.Copies)
	assert.Equal(t, 3, book.Available)
}

func TestIncreaseBookCopiesHandler_BookNotFound(t *testing.T) {
	bookRepo := memory.NewBookRepository(nil)
	increase := NewIncreaseBookCopiesHandler(bookRepo)
	ctx := context.Background()

	_, err := increase.Handle(ctx, IncreaseBookCopiesCommand{BookID: "does-not-exist", Amount: 1})
	assert.True(t, errors.Is(err, domain.ErrBookNotFound))
}

func TestIncreaseBookCopiesHandler_RejectsNonPositiveAmount(t *testing.T) {
	bookRepo := memory.NewBookRepository(map[string]*domain.Book{
		"book-1": {ID: "book-1", Title: "Test Book", Copies: 2, Available: 1},
	})
	increase := NewIncreaseBookCopiesHandler(bookRepo)
	ctx := context.Background()

	_, err := increase.Handle(ctx, IncreaseBookCopiesCommand{BookID: "book-1", Amount: 0})
	assert.True(t, errors.Is(err, domain.ErrInvalidBookInput))
}
