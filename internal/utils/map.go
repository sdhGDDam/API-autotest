package utils

// CopyMap создаёт поверхностную копию map[string]string.
// Полезно для тестов, когда нужно модифицировать копию исходной map,
// не затрагивая оригинал (например, для негативных тестов с некорректными значениями).
//
// Использование:
//
//	original := map[string]string{"a": "1", "b": "2"}
//	copy := utils.CopyMap(original)
//	copy["a"] = "99" // original не изменится
func CopyMap(original map[string]string) map[string]string {
	if original == nil {
		return nil
	}
	copied := make(map[string]string, len(original))
	for k, v := range original {
		copied[k] = v
	}
	return copied
}