package executor_test

import (
	"strings"
	"testing"

	"github.com/alex-ermolaxe/qdata/internal/executor"
	"github.com/alex-ermolaxe/qdata/internal/format"
	formatjson "github.com/alex-ermolaxe/qdata/internal/format/json"
	"github.com/alex-ermolaxe/qdata/internal/schema"
	"github.com/iancoleman/orderedmap"
)

func TestSelectFieldsFromDecodedJSONObjectArray(t *testing.T) {
	input := `[{"id":1,"items":[{"name":"first","value":10},{"name":"second","value":20}]}]`
	records, err := (&formatjson.JSONFormat{}).Decode(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	inferred := schema.Infer(records, 100)
	hasItemsName := false
	for _, path := range inferred.AllPaths() {
		if path == "items.name" {
			hasItemsName = true
			break
		}
	}
	if !hasItemsName {
		t.Fatal("schema did not infer field items.name from JSON")
	}

	result := executor.Select(records, []string{"items.name"})
	value, ok := result[0].Get("items")
	if !ok {
		t.Fatal("Select omitted items")
	}
	items, ok := value.([]any)
	if !ok || len(items) != 2 {
		t.Fatalf("items = %#v, want two-element []any", value)
	}

	for i, want := range []string{"first", "second"} {
		item, ok := items[i].(*orderedmap.OrderedMap)
		if !ok {
			t.Fatalf("items[%d] has type %T, want *orderedmap.OrderedMap", i, items[i])
		}
		if got, ok := item.Get("name"); !ok || got != want {
			t.Errorf("items[%d].name = %v, %v; want %q, true", i, got, ok, want)
		}
		if _, ok := item.Get("value"); ok {
			t.Errorf("Select retained unrequested items[%d].value", i)
		}
	}
}

func TestSelectFieldsFromObjectArray(t *testing.T) {
	firstItem := orderedmap.New()
	firstItem.Set("name", "keyboard")
	firstItem.Set("price", 99.5)
	firstItem.Set("sku", "kb-1")

	secondItem := orderedmap.New()
	secondItem.Set("name", "mouse")
	secondItem.Set("price", 25.0)
	secondItem.Set("sku", "ms-1")

	products := []any{firstItem, secondItem}
	record := format.NewRecord()
	record.Set("id", 1)
	record.Set("products", products)
	record.Set("ignored", "value")

	result := executor.Select([]format.Record{record}, []string{"id", "products.name", "products.price"})
	if len(result) != 1 {
		t.Fatalf("Select returned %d records, want 1", len(result))
	}

	selected := result[0]
	if value, ok := selected.Get("id"); !ok || value != 1 {
		t.Errorf("selected id = %v, %v; want 1, true", value, ok)
	}
	if _, ok := selected.Get("ignored"); ok {
		t.Error("Select retained unrequested top-level field \"ignored\"")
	}

	value, ok := selected.Get("products")
	if !ok {
		t.Fatal("Select omitted products")
	}
	selectedProducts, ok := value.([]any)
	if !ok {
		t.Fatalf("products has type %T, want []any", value)
	}
	if len(selectedProducts) != 2 {
		t.Fatalf("selected products has length %d, want 2", len(selectedProducts))
	}

	assertSelectedProduct := func(index int, name string, price float64) {
		t.Helper()
		product, ok := selectedProducts[index].(*orderedmap.OrderedMap)
		if !ok {
			t.Fatalf("product %d has type %T, want *orderedmap.OrderedMap", index, selectedProducts[index])
		}
		if value, ok := product.Get("name"); !ok || value != name {
			t.Errorf("product %d name = %v, %v; want %q, true", index, value, ok, name)
		}
		if value, ok := product.Get("price"); !ok || value != price {
			t.Errorf("product %d price = %v, %v; want %v, true", index, value, ok, price)
		}
		if _, ok := product.Get("sku"); ok {
			t.Errorf("Select retained unrequested field sku from product %d", index)
		}
	}
	assertSelectedProduct(0, "keyboard", 99.5)
	assertSelectedProduct(1, "mouse", 25.0)

	// Projection must not modify the source objects.
	if _, ok := firstItem.Get("sku"); !ok {
		t.Error("Select modified the source product")
	}
}

func TestSelectNestedFieldInsideObjectArray(t *testing.T) {
	details := orderedmap.New()
	details.Set("color", "black")
	details.Set("weight", 1.2)
	item := orderedmap.New()
	item.Set("name", "keyboard")
	item.Set("details", details)

	record := format.NewRecord()
	record.Set("items", []any{item})

	result := executor.Select([]format.Record{record}, []string{"items.details.color"})
	value, ok := result[0].Get("items")
	if !ok {
		t.Fatal("Select omitted items")
	}
	items, ok := value.([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("items = %#v, want one-element []any", value)
	}
	selectedItem, ok := items[0].(*orderedmap.OrderedMap)
	if !ok {
		t.Fatalf("item has type %T, want *orderedmap.OrderedMap", items[0])
	}
	selectedDetails, ok := selectedItem.Get("details")
	if !ok {
		t.Fatal("Select omitted details")
	}
	selectedDetailsMap, ok := selectedDetails.(*orderedmap.OrderedMap)
	if !ok {
		t.Fatalf("details has type %T, want *orderedmap.OrderedMap", selectedDetails)
	}
	if value, ok := selectedDetailsMap.Get("color"); !ok || value != "black" {
		t.Errorf("details.color = %v, %v; want black, true", value, ok)
	}
	if _, ok := selectedDetailsMap.Get("weight"); ok {
		t.Error("Select retained unrequested nested field weight")
	}
}
