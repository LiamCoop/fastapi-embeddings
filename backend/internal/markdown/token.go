package markdown

import "math"

// TokenBias tunes token estimation divisors.
type TokenBias int

const (
	BiasBalanced TokenBias = iota
	BiasProse
	BiasCode
)

// EstimateTokens returns a deterministic O(n) token estimate for a block.
func EstimateTokens(b Block, bias TokenBias) int {
	if b.Content == "" {
		return 0
	}

	proseDivisor := 4.0
	codeDivisor := 2.75
	switch bias {
	case BiasProse:
		proseDivisor = 4.4
		codeDivisor = 3.0
	case BiasCode:
		proseDivisor = 3.6
		codeDivisor = 2.4
	}

	chars := float64(len([]rune(b.Content)))
	divisor := proseDivisor
	if isCodeLike(b.Type) {
		divisor = codeDivisor
	}

	tokens := int(math.Ceil(chars / divisor))
	if tokens < 1 {
		return 1
	}
	return tokens
}

func isCodeLike(t BlockType) bool {
	switch t {
	case BlockCodeFence, BlockMDXImport, BlockMDXComponent:
		return true
	default:
		return false
	}
}
