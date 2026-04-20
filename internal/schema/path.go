package schema

import "strings"

// Split разбивает путь к полю на части
// "address.city" -> ["address", "city"]
// "name" -> ["name"]
func Split(path string) []string {
	return strings.Split(path, ".")
}

// Join собирает путь из частей
// ["address", "city"] -> "address.city"
func Join(parts []string) string {
	return strings.Join(parts, ".")
}

// GetNested извлекает значение из вложенной map по пути
// path: "address.city", record: {"address": {"city": "Moscow"}} -> "Moscow", true
func GetNested(record map[string]any, path string) (any, bool) {
	parts := Split(path)
	current := record

	for i, part := range parts {
		val, ok := current[part]
		if !ok {
			return nil, false
		}

		// Если это последняя часть пути — возвращаем значение
		if i == len(parts)-1 {
			return val, true
		}

		// Иначе пробуем спуститься глубже
		nested, ok := val.(map[string]any)
		if !ok {
			return nil, false
		}
		current = nested
	}

	return nil, false
}

// SetNested устанавливает значение в вложенной map по пути
// Создаёт промежуточные map если они не существуют
func SetNested(record map[string]any, path string, value any) {
	parts := Split(path)
	current := record

	for i, part := range parts {
		if i == len(parts)-1 {
			current[part] = value
			return
		}

		nested, ok := current[part].(map[string]any)
		if !ok {
			nested = map[string]any{}
			current[part] = nested
		}
		current = nested
	}
}

// DeleteNested удаляет поле из вложенной map по пути
func DeleteNested(record map[string]any, path string) {
	parts := Split(path)
	current := record

	for i, part := range parts {
		if i == len(parts)-1 {
			delete(current, part)
			return
		}

		nested, ok := current[part].(map[string]any)
		if !ok {
			return
		}
		current = nested
	}
}
