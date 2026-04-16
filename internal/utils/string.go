package utils

// ContainsSubstring проверяет, содержится ли подстрока substr в строке s.
// Регистр учитывается. Если substr пустая, возвращает true.
func ContainsSubstring(s, substr string) bool {
	if substr == "" {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}