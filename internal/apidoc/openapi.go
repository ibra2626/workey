// Package apidoc builds the REST contract from the actual transport types.
package apidoc

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/ibra2626/workey/internal/httpx"
	"github.com/ibra2626/workey/internal/plan"
)

type object = map[string]any

// Operation is also the human-readable endpoint inventory.
type Operation struct{ Method, Path, ID, Summary, Input, Output string }

func Operations() []Operation {
	return []Operation{
		{"GET", "/agent-guide", "getAgentGuide", "Ajan rehberini ve sunucu tarihini oku", "", "AgentGuide"},
		{"GET", "/outline", "getOutline", "Fazları ve konu gruplarını oku", "", "Outline"},
		{"PUT", "/outline", "saveOutline", "Faz ve grupların tam listesini değiştir", "SaveOutline", "OutlineResult"},
		{"GET", "/plan", "getPlan", "from–to dahil aralıktaki seansları oku", "", "Plan"},
		{"PUT", "/plan", "savePlan", "Aralıktaki açık seansların tam listesini değiştir", "SavePlan", "PlanResult"},
		{"GET", "/today", "getToday", "Bugünün planını oku", "", "Day"},
		{"GET", "/week", "getWeek", "Verilen tarihin haftasını oku", "", "Week"},
		{"GET", "/progress", "getProgress", "İlerlemeyi oku", "", "Progress"},
		{"POST", "/sessions/{id}/complete", "completeSession", "Seansı tamamla", "", "Session"},
		{"DELETE", "/sessions/{id}/complete", "undoSession", "Tamamlamayı geri al", "", "Session"},
		{"GET", "/openapi.json", "getOpenAPI", "OpenAPI 3.1 sözleşmesini oku", "", ""},
	}
}

// ExamplePlan is synthetic and intentionally uses a fixed date for contract tests.
const ExamplePlan = `{
  "from": "2026-10-26", "to": "2026-10-26",
  "sessions": [{
    "date": "2026-10-26", "start_time": "20:30", "duration_min": 90,
    "kind": "aksam", "topic_group": "temel",
    "items": [{"book": "M1", "chapter_no": 1, "chapter_title": "Temel kavramlar", "page_from": 1, "page_to": 10}]
  }]
}`

const ExampleOutline = `{
  "phases": [{"key": "faz-1", "ordinal": 1, "name": "İşlem temeli", "start_date": "2026-10-26", "end_date": "2026-12-13"}],
  "topic_groups": [{"key": "temel", "name": "Temel kavramlar", "ordinal": 1}]
}`

var document = sync.OnceValues(build)

// ServeHTTP serves the same bytes behind REST Bearer auth and web session auth.
func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b, err := document()
	if err != nil {
		http.Error(w, "sözleşme oluşturulamadı", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	_, _ = w.Write(b)
}

func schema[T any]() (any, error) {
	s, err := jsonschema.For[T](nil)
	if err != nil {
		return nil, err
	}
	b, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	var v object
	if err := json.Unmarshal(b, &v); err != nil {
		return nil, err
	}
	// encoding/json emits nil slices as null. Reflect this existing wire format.
	nullableArrays(v)
	return v, nil
}

func nullableArrays(v any) {
	switch x := v.(type) {
	case map[string]any:
		if x["type"] == "array" {
			x["type"] = []string{"array", "null"}
		}
		for _, child := range x {
			nullableArrays(child)
		}
		if props, ok := x["properties"].(map[string]any); ok {
			for name, raw := range props {
				p, ok := raw.(map[string]any)
				if !ok {
					continue
				}
				switch name {
				case "date", "today", "from", "to", "start_date", "end_date":
					p["format"] = "date"
				case "start_time":
					p["pattern"] = `^([01]\d|2[0-3]):[0-5]\d$`
				case "book":
					p["enum"] = []string{"M1", "M2", "M3", "G1", "G2"}
				case "kind":
					p["enum"] = []string{plan.KindAksam, plan.KindUzunBlok, plan.KindDeneme}
				case "chapter_no", "page_from", "page_to", "duration_min", "id":
					p["minimum"] = 1
				case "chapter_title":
					p["minLength"] = 1
				case "state":
					p["enum"] = []string{"planned", "today", "done", "missed", "moved"}
				}
			}
		}
	case []any:
		for _, child := range x {
			nullableArrays(child)
		}
	}
}

// REST's JSON decoder accepts unknown fields and zero-valued ordinal/collections.
// Keep that existing behavior visible while strongly recommending explicit full lists.
func requestSchema(v any) {
	switch x := v.(type) {
	case map[string]any:
		if x["type"] == "object" {
			x["additionalProperties"] = true
		}
		if required, ok := x["required"].([]any); ok {
			kept := []any{}
			for _, name := range required {
				if name != "ordinal" && name != "sessions" && name != "phases" && name != "topic_groups" {
					kept = append(kept, name)
				}
			}
			x["required"] = kept
		}
		for _, child := range x {
			requestSchema(child)
		}
	case []any:
		for _, child := range x {
			requestSchema(child)
		}
	}
}

func ref(name string) object { return object{"$ref": "#/components/schemas/" + name} }
func content(s any) object   { return object{"application/json": object{"schema": s}} }

func build() ([]byte, error) {
	schemas := object{}
	builders := map[string]func() (any, error){
		"AgentGuide": schema[plan.AgentGuide], "SavePlan": schema[plan.SavePlanRequest],
		"SaveOutline": schema[plan.SaveOutlineRequest], "PlanResult": schema[plan.SavePlanResult],
		"OutlineResult": schema[plan.SaveOutlineResult], "Outline": schema[plan.OutlineView],
		"Session": schema[plan.SessionView], "Day": schema[plan.DayView],
		"Week": schema[plan.WeekView], "Progress": schema[plan.ProgressView],
		"Plan": schema[struct {
			Sessions []plan.SessionView `json:"sessions"`
		}],
		"Error": schema[struct {
			Error httpx.Error `json:"error"`
		}],
	}
	for name, build := range builders {
		s, err := build()
		if err != nil {
			return nil, err
		}
		schemas[name] = s
	}
	// Domain constraints and full-replacement semantics supplement inferred shapes.
	requestSchema(schemas["SavePlan"])
	requestSchema(schemas["SaveOutline"])
	schemas["SavePlan"].(object)["description"] = strings.Join(plan.GuideRules(), "\n")
	schemas["SaveOutline"].(object)["description"] = "Tam faz/grup listesi; eksik key silinir. Kullanılan grup silinemez. Boş olmayan Idempotency-Key başlığı gövdedeki anahtardan önceliklidir."
	paths := object{}
	for _, op := range Operations() {
		responseSchema := any(object{"type": "object"})
		if op.Output != "" {
			responseSchema = ref(op.Output)
		}
		responses := object{"200": object{"description": "Başarılı", "content": content(responseSchema)}}
		if example := OperationResponseExample(op); example != nil {
			responses["200"].(object)["content"].(object)["application/json"].(object)["example"] = example
		}
		for code, desc := range map[string]string{"400": "Bozuk JSON veya 1 MiB gövde sınırı", "401": "Bearer token eksik veya geçersiz", "404": "Kayıt bulunamadı", "409": "idempotency_conflict", "422": "validation_failed; alan hatalarını düzelt", "500": "internal"} {
			responses[code] = object{"description": desc, "content": content(ref("Error"))}
		}
		entry := object{"operationId": op.ID, "summary": op.Summary, "responses": responses}
		params := []any{}
		param := func(name, in string, required bool, s object) {
			params = append(params, object{"name": name, "in": in, "required": required, "schema": s})
		}
		if op.Method == "GET" && op.Path == "/plan" {
			for _, name := range []string{"from", "to"} {
				param(name, "query", true, object{"type": "string", "format": "date"})
			}
		}
		if op.Path == "/week" {
			param("date", "query", false, object{"type": "string", "format": "date", "description": "Varsayılan sunucunun İstanbul bugünü"})
		}
		if strings.Contains(op.Path, "{id}") {
			param("id", "path", true, object{"type": "integer", "minimum": 1})
		}
		if op.Input != "" {
			param("Idempotency-Key", "header", false, object{"type": "string", "description": "Boş değilse gövdenin idempotency_key alanından önceliklidir."})
			c := content(ref(op.Input))
			example := ExamplePlan
			if op.Input == "SaveOutline" {
				example = ExampleOutline
			}
			c["application/json"].(object)["example"] = json.RawMessage(example)
			entry["requestBody"] = object{"required": true, "content": c}
		}
		if len(params) > 0 {
			entry["parameters"] = params
		}
		if paths[op.Path] == nil {
			paths[op.Path] = object{}
		}
		paths[op.Path].(object)[strings.ToLower(op.Method)] = entry
	}
	return json.MarshalIndent(object{
		"openapi": "3.1.0", "info": object{"title": "Workey REST API", "version": "v1", "description": "Kitap referanslarına dayalı çalışma planı. Önce agent-guide okuyun. Tarihler Europe/Istanbul. Gövdeler en fazla 1 MiB."},
		"servers": []any{object{"url": "/api/v1"}}, "security": []any{object{"bearerAuth": []string{}}},
		"paths": paths, "components": object{"schemas": schemas, "securitySchemes": object{"bearerAuth": object{"type": "http", "scheme": "bearer"}}},
	}, "", "  ")
}
