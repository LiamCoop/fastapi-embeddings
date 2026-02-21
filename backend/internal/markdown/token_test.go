package markdown

import "testing"

func TestEstimateTokens_Biases(t *testing.T) {
	prose := Block{Type: BlockParagraph, Content: "This is a moderately sized paragraph for testing token estimation."}
	code := Block{Type: BlockCodeFence, Content: "```go\nfor i := 0; i < 100; i++ {\nfmt.Println(i)\n}\n```"}

	proseBalanced := EstimateTokens(prose, BiasBalanced)
	proseProse := EstimateTokens(prose, BiasProse)
	if proseProse > proseBalanced {
		t.Fatalf("expected prose bias to estimate fewer/equal tokens for prose")
	}

	codeBalanced := EstimateTokens(code, BiasBalanced)
	codeCode := EstimateTokens(code, BiasCode)
	if codeCode < codeBalanced {
		t.Fatalf("expected code bias to estimate more/equal tokens for code")
	}
}
