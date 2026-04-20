package schema

import "fmt"

// FieldType — тип поля
type FieldType string

const (
	TypeString  FieldType = "string"
	TypeNumber  FieldType = "number"
	TypeBool    FieldType = "bool"
	TypeArray   FieldType = "array"
	TypeObject  FieldType = "object"
	TypeUnknown FieldType = "unknown"
)

// Field — описание одного поля схемы
type Field struct {
	Path     string    // полный путь: "address.city"
	Type     FieldType // тип поля
	Children []*Field  // вложенные поля (для объектов)
}

// Schema — схема данных, содержит все поля
type Schema struct {
	Fields []*Field          // поля верхнего уровня
	index  map[string]*Field // индекс всех полей по пути для быстрого доступа
}

// AllPaths возвращает все пути к полям включая вложенные
func (s *Schema) AllPaths() []string {
	paths := make([]string, 0, len(s.index))
	for path := range s.index {
		paths = append(paths, path)
	}
	return paths
}

// Get возвращает поле по пути
func (s *Schema) Get(path string) (*Field, bool) {
	f, ok := s.index[path]
	return f, ok
}

// ChildPaths возвращает пути дочерних полей для указанного префикса
// Например для "address" вернёт ["address.city", "address.zip"]
func (s *Schema) ChildPaths(prefix string) []string {
	var paths []string
	for path := range s.index {
		if len(path) > len(prefix)+1 &&
			path[:len(prefix)] == prefix &&
			path[len(prefix)] == '.' {
			// Берём только прямых потомков
			rest := path[len(prefix)+1:]
			hasNested := false
			for p := range s.index {
				if len(p) > len(path)+1 &&
					p[:len(path)] == path &&
					p[len(path)] == '.' {
					hasNested = true
					break
				}
			}
			_ = hasNested
			_ = rest
			paths = append(paths, path)
		}
	}
	return paths
}

// Infer анализирует записи и строит схему
// Анализирует первые maxRecords записей для определения типов
func Infer(records []map[string]any, maxRecords int) *Schema {
	schema := &Schema{
		index: map[string]*Field{},
	}

	limit := maxRecords
	if limit > len(records) || limit <= 0 {
		limit = len(records)
	}

	for i := 0; i < limit; i++ {
		collectFields(records[i], "", schema)
	}

	return schema
}

// collectFields рекурсивно обходит запись и собирает поля
func collectFields(record map[string]any, prefix string, schema *Schema) {
	for key, val := range record {
		path := key
		if prefix != "" {
			path = prefix + "." + key
		}

		fieldType := inferType(val)

		if _, exists := schema.index[path]; !exists {
			field := &Field{
				Path: path,
				Type: fieldType,
			}
			schema.index[path] = field

			if prefix == "" {
				schema.Fields = append(schema.Fields, field)
			}
		}

		// Рекурсивно обходим вложенные объекты
		if nested, ok := val.(map[string]any); ok {
			collectFields(nested, path, schema)
		}
	}
}

// inferType определяет тип значения
func inferType(val any) FieldType {
	if val == nil {
		return TypeUnknown
	}

	switch val.(type) {
	case string:
		return TypeString
	case float64, float32, int, int64, int32:
		return TypeNumber
	case bool:
		return TypeBool
	case []any:
		return TypeArray
	case map[string]any:
		return TypeObject
	default:
		return TypeUnknown
	}
}

// Print выводит схему в читаемом виде
func (s *Schema) Print() {
	printFields(s.Fields, s.index, 0)
}

func printFields(fields []*Field, index map[string]*Field, depth int) {
	indent := ""
	prefix := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}
	if depth > 0 {
		prefix = "└─ "
	}

	for _, f := range fields {
		parts := Split(f.Path)
		name := parts[len(parts)-1]

		if f.Type == TypeObject {
			fmt.Printf("%s%s%s\n", indent, prefix, name)
			// Собираем дочерние поля
			var children []*Field
			for path, field := range index {
				if len(path) > len(f.Path)+1 &&
					path[:len(f.Path)] == f.Path &&
					path[len(f.Path)] == '.' {
					rest := path[len(f.Path)+1:]
					if !containsDot(rest) {
						children = append(children, field)
					}
				}
			}
			printFields(children, index, depth+1)
		} else {
			fmt.Printf("%s%s%-20s %s\n", indent, prefix, name, f.Type)
		}
	}
}

func containsDot(s string) bool {
	for _, c := range s {
		if c == '.' {
			return true
		}
	}
	return false
}
