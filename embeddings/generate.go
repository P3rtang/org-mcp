package embeddings

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type Embedding [768]float64

type Response struct {
	Model             string      `json:"model"`
	Embeddings        []Embedding `json:"embeddings"`
	Total_duration    float64     `json:"total_duration"`
	Load_duration     float64     `json:"load_duration"`
	Prompt_eval_count int         `json:"prompt_eval_count"`
}

func Generate(text []string) ([]Embedding, error) {
	body := map[string]any{
		"model": "embeddinggemma",
		"input": text,
	}

	b, err := json.Marshal(body)
	if err != nil {
		return []Embedding{}, err
	}

	resp, err := http.Post("http://localhost:11434/api/embed", "application/json", bytes.NewReader(b))
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return []Embedding{}, err
	}

	content := Response{}
	json.Unmarshal(bytes, &content)

	return content.Embeddings, nil
}

func dotProduct(vec1 []float64, vec2 []float64) (product float64) {
	for i, val1 := range vec1 {
		val2 := vec2[i]
		product += val1 * val2
	}

	return
}

func magnitude(vec []float64) (mag float64) {
	for _, val := range vec {
		mag += val * val
	}

	return
}

func Similarity(embed1 []float64, embed2 []float64) float64 {
	dot := dotProduct(embed1, embed2)
	mag1 := magnitude(embed1)
	mag2 := magnitude(embed2)

	return dot / (mag1 * mag2)
}
