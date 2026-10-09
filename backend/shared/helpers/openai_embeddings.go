package helpers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const defaultOpenAIEmbeddingModel = "text-embedding-3-small"

func useOpenAIEmbeddings() bool {
	return strings.TrimSpace(os.Getenv("OPENAI_API_KEY")) != ""
}

func openAIEmbeddingModel() string {
	model := strings.TrimSpace(os.Getenv("OPENAI_EMBEDDING_MODEL"))
	if model == "" {
		model = defaultOpenAIEmbeddingModel
	}
	return model
}

func GenerateOpenAIEmbedding(ctx context.Context, input string) ([]float64, error) {
	apiKey := strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
	if apiKey == "" {
		return nil, errors.New("openai api key is empty")
	}

	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return nil, errors.New("embedding input kosong")
	}

	body := map[string]any{
		"model": openAIEmbeddingModel(),
		"input": trimmed,
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/embeddings", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("openai embeddings error: %s", strings.TrimSpace(string(rawBody)))
	}

	var decoded struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rawBody, &decoded); err != nil {
		return nil, err
	}
	if len(decoded.Data) == 0 || len(decoded.Data[0].Embedding) == 0 {
		return nil, errors.New("openai embedding kosong")
	}

	return decoded.Data[0].Embedding, nil
}

func EmbeddingToVectorLiteral(embedding []float64) string {
	if len(embedding) == 0 {
		return ""
	}

	var builder strings.Builder
	builder.WriteByte('[')
	for i, value := range embedding {
		if i > 0 {
			builder.WriteByte(',')
		}
		builder.WriteString(strconv.FormatFloat(value, 'f', -1, 64))
	}
	builder.WriteByte(']')
	return builder.String()
}

func BuildEmbeddingText(parts ...string) string {
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			values = append(values, trimmed)
		}
	}
	return strings.Join(values, " | ")
}
