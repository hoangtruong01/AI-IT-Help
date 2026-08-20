package middleware

import (
	"regexp"
	"strings"
)

var (
	sensitiveKeyPattern = regexp.MustCompile(`(?i)(password|token|secret|credit_card|api_key|private_key|auth_key|ssn|cvv|access_key)`)
	creditCardPattern   = regexp.MustCompile(`\b(?:\d[ -]*?){13,16}\b`)
	jwtPattern          = regexp.MustCompile(`eyJ[A-Za-z0-9-_=]+\.[A-Za-z0-9-_=]+\.?[A-Za-z0-9-_.+/=]*`)
)

// MaskSensitiveData scans a map and masks values corresponding to sensitive keys.
func MaskSensitiveData(data map[string]interface{}) map[string]interface{} {
	if data == nil {
		return nil
	}

	result := make(map[string]interface{}, len(data))
	for k, v := range data {
		if sensitiveKeyPattern.MatchString(k) {
			result[k] = "********"
			continue
		}

		switch val := v.(type) {
		case string:
			result[k] = MaskString(val)
		case map[string]interface{}:
			result[k] = MaskSensitiveData(val)
		case []interface{}:
			var maskedList []interface{}
			for _, item := range val {
				if itemMap, ok := item.(map[string]interface{}); ok {
					maskedList = append(maskedList, MaskSensitiveData(itemMap))
				} else if itemStr, ok := item.(string); ok {
					maskedList = append(maskedList, MaskString(itemStr))
				} else {
					maskedList = append(maskedList, item)
				}
			}
			result[k] = maskedList
		default:
			result[k] = v
		}
	}
	return result
}

// MaskString sanitizes credit card numbers, JWT tokens and sensitive strings.
func MaskString(s string) string {
	if s == "" {
		return s
	}

	// Mask JWT tokens
	s = jwtPattern.ReplaceAllStringFunc(s, func(match string) string {
		if len(match) > 12 {
			return match[:6] + "..." + match[len(match)-4:]
		}
		return "********"
	})

	// Mask Credit Cards
	s = creditCardPattern.ReplaceAllString(s, "****-****-****-****")

	// Lowercase keyword check
	lower := strings.ToLower(s)
	if strings.Contains(lower, "bearer ") {
		parts := strings.Split(s, " ")
		if len(parts) >= 2 {
			return parts[0] + " ********"
		}
	}

	return s
}
