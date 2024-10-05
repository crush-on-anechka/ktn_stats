package essentialshandler

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/crush-on-anechka/ktn_stats/config"
	"github.com/crush-on-anechka/ktn_stats/db"
)

type EssentialsHandler struct {
	storage *db.SqliteDB
}

func New(storage *db.SqliteDB) *EssentialsHandler {
	return &EssentialsHandler{storage: storage}
}

func (handler *EssentialsHandler) UpdateEssentialsByDate(date string) error {
	inscriptions, err := handler.storage.GetInscriptionsByDate(date)
	if err != nil {
		return fmt.Errorf("failed to fetch essentials from db: %w", err)
	}

	essentialWordsCount := make(map[string]int)

	for _, inscription := range inscriptions {
		words := handler.extractEssentialWords(inscription)
		for _, word := range words {
			essentialWordsCount[word]++
		}
	}

	essentialsJson, err := json.Marshal(essentialWordsCount)
	if err != nil {
		return fmt.Errorf("failed to marshal essentials map: %w", err)
	}

	if err = handler.storage.UpdateWords(date, string(essentialsJson)); err != nil {
		return fmt.Errorf("failed to update essentials for %s: %w", date, err)
	}

	return nil
}

func (handler *EssentialsHandler) extractEssentialWords(inscription string) []string {
	var essentialWords []string

	inscription = strings.ReplaceAll(strings.ReplaceAll(inscription, "\n", " "), "\r", " ")

	for _, word := range strings.Split(inscription, " ") {
		word = handler.cleanEssentialWord(word)
		if utf8.RuneCountInString(word) > 1 && !containsLowercaseCyrillic(word) {
			essentialWords = append(essentialWords, word)
		}
	}

	return essentialWords
}

func (handler *EssentialsHandler) cleanEssentialWord(s string) string {
	s = strings.ReplaceAll(s, "\"", "")
	s = strings.TrimRight(s, ".,:;!?")
	return s
}

func containsLowercaseCyrillic(s string) bool {
	return config.LowercaseCyrillicRegex.MatchString(s)
}
