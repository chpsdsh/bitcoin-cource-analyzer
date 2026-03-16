package domain

type Category int

const (
	PoliticsCategory Category = iota
	EnvironmentCategory
	EconomyCategory
	TechnologyCategory
	CryptoCategory
)

var AllCategories = []Category{
	PoliticsCategory,
	EnvironmentCategory,
	EconomyCategory,
	TechnologyCategory,
	CryptoCategory,
}

func CategoryToString(category Category) string {
	switch category {
	case PoliticsCategory:
		return "politics"
	case EnvironmentCategory:
		return "environment"
	case EconomyCategory:
		return "economy"
	case TechnologyCategory:
		return "technology"
	case CryptoCategory:
		return "crypto"
	default:
		return "unknown"
	}
}
