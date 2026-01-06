package climerrors

import "strings"

var suggestionsByCategory = map[Category][]string{
	CategoryAuth: {
		"Your API token may be expired or invalid",
		"Try: deel auth login --account <account>",
	},
	CategoryForbidden: {
		"Your API token doesn't have permission for this action",
		"Check scopes with: deel auth status",
		"You may need to generate a new token with additional scopes",
	},
	CategoryNotFound: {
		"The resource was not found",
	},
	CategoryValidation: {
		"The request data was invalid",
		"Check required fields and formats",
	},
	CategoryRateLimit: {
		"You've hit Deel's rate limit",
		"Wait a few seconds and try again",
	},
	CategoryServer: {
		"Deel's API returned an error",
		"Check status.deel.com for outages",
		"Try again in a few minutes",
	},
	CategoryNetwork: {
		"Could not connect to Deel's API",
		"Check your internet connection",
		"If using a proxy/VPN, ensure api.letsdeel.com is accessible",
	},
	CategoryConfig: {
		"No account configured",
		"Run: deel auth login to set up an account",
	},
	CategoryUnknown: {
		"An unexpected error occurred",
	},
}

// SuggestionsFor returns helpful suggestions based on error category and operation
func SuggestionsFor(cat Category, operation string) []string {
	if cat == CategoryNotFound {
		return suggestionsForNotFound(operation)
	}

	if suggestions, ok := suggestionsByCategory[cat]; ok {
		return suggestions
	}
	return suggestionsByCategory[CategoryUnknown]
}

func suggestionsForNotFound(operation string) []string {
	op := strings.ToLower(operation)
	if strings.Contains(op, "getting") || strings.Contains(op, "get ") {
		return []string{
			"Check that the ID is correct",
			"The resource may have been deleted",
		}
	}
	if strings.Contains(op, "searching worker") {
		return []string{
			"The worker may not have signed their contract yet",
			"Workers only appear in the People directory after they complete signing",
		}
	}
	return []string{
		"This endpoint may not be available for your account type",
		"Check your API token has the required scopes",
	}
}
