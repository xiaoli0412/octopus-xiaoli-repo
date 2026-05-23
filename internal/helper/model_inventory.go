package helper

import (
	"context"
	"strings"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
)

func EnsureReferencedLLMInfos(ctx context.Context) error {
	names, err := op.ReferencedModelNames(ctx)
	if err != nil {
		return err
	}
	missing := make([]string, 0, len(names))
	for _, name := range names {
		normalized := strings.ToLower(strings.TrimSpace(name))
		if normalized == "" {
			continue
		}
		if _, err := op.LLMGet(normalized); err == nil {
			continue
		}
		missing = append(missing, normalized)
	}
	return LLMPriceAddToDB(missing, ctx)
}
