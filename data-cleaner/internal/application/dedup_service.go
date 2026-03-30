package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"strings"

	"golang.org/x/text/unicode/norm"

	"data-cleaner/internal/domain"
)

type Storage interface {
	AddNewArticle(ctx context.Context, key string) (bool, error)
}

type Sender interface {
	SendArticle(dto domain.ArticleDto)
}

type DedupService struct {
	Storage Storage
	Sender  Sender
}

func (d DedupService) HandleNewArticle(ctx context.Context, article domain.ArticleDto) {
	key := buildDedupKey(article)
	res, err := d.Storage.AddNewArticle(ctx, key)
	if err != nil {
		slog.Error("error adding new article to storage", slog.String("error", err.Error()), slog.String("key", key))
		return
	}
	if !res {
		slog.Error("article exists", slog.String("key", key))
		return
	}
	d.Sender.SendArticle(article)
}

func buildDedupKey(article domain.ArticleDto) string {
	raw := normalize(article.Category) + "|" +
		normalize(article.Title) + "|" +
		normalize(article.Text)

	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func normalize(s string) string {
	if s == "" {
		return ""
	}

	s = norm.NFKC.String(s)
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.Join(strings.Fields(s), " ")
	s = strings.ReplaceAll(s, " ,", ",")
	s = strings.ReplaceAll(s, " .", ".")
	s = strings.ReplaceAll(s, " :", ":")
	s = strings.ReplaceAll(s, " ;", ";")

	return strings.TrimSpace(s)
}
