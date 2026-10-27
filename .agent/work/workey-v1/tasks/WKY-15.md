# WKY-15 — Kullanım rehberleri için ortak HTML/Markdown içerik modeli

## Goal
MCP ve API rehberlerini aynı yapılandırılmış içerikten HTML ve ham Markdown olarak üret; iki gösterimin birbirinden kopmasını önle.

## Scope
- Ortak kurallar: `../README.md`. Onaylanan kapsam: 2026-10-27 kullanım rehberleri ve MCP çalışma planı promptu planı.
- Başlangıç noktaları: `internal/web/views/usage_content.go`, `usage.templ`, `internal/web/usage.go`.
- Başlık, paragraf, dil etiketli kod bloğu, bağlantı, liste ve açılır bölüm içeriklerini temsil eden ortak bir rehber modeli oluştur. İstemciye özel bölümler ve ayrı kopyalanabilir istemler de aynı modelde temsil edilebilsin.
- HTML sunumunu ve Markdown üretimini bu modele bağla. Menüleri ve görünüm kontrollerini Markdown'a dahil etme; kapalı bölümleri ve tüm istemcilerin içeriklerini dahil et.
- API endpoint bilgileri ve örnekleri mevcut `internal/apidoc` kaynağından gelsin. Mevcut API açıklamaları, örnekler ve OpenAPI indirme bağlantısı korunsun.
- Tarih, sınav tarihi, desteklenen kitap/tür değerleri, işlem sırası ve kurallar mevcut `plan.AgentGuide` kaynağından gelsin; sabit ikinci bir kopya oluşturma.
- Türkçe karakterleri, kodların satır sonlarını ve Markdown kod çitlerini doğru üret. Kullanıcıya ait gerçek token, ortam değeri veya plan verisi rehbere eklenmesin.
- REST/MCP sözleşmesi, iş kuralları ve veritabanı değişmeyecek. Yeni paket/dependency gerekmiyor.

## Acceptance criteria
- [x] MCP/API HTML ve Markdown aynı içerik modelini kullanıyor; ayrı elle yazılmış rehber kopyaları yok.
- [x] Markdown kapalı bölümleri kapsıyor; başlıklar, listeler, bağlantılar ve kod blokları geçerli.
- [x] API örnekleri mevcut sözleşme kaynağıyla tutarlı; OpenAPI indirmesi korunuyor.
- [x] Sahte saatle ajan rehberindeki tarih ve sınav tarihi iki çıktıda da doğrulanıyor.

## Checks
- İçerik üretimi ve HTML/Markdown tutarlılığı için odaklı testler ekle; gerçek örnek kodların bozulmadığını doğrula.
- `go tool templ generate`
- `go test ./internal/web/... ./internal/apidoc/...`
- `go build ./...`

## Progress
Completed 2026-10-27. `UsageDocument` / `UsageBlock` is the shared source for
the HTML renderer and full Markdown output. Client sections and individually
identified prompts are represented in the model; Markdown traverses every
client and closed endpoint disclosure, excluding page navigation and controls.

Endpoint inventory, request bodies and plan/endpoint response examples come from
`internal/apidoc`. Runtime dates, exam date, catalog, workflow and rules use
`plan.AgentGuide`. No transport, domain, database or dependency changes.

Focused tests cover HTML/Markdown content parity, fake-clock Istanbul date rollover,
exam date, injected catalog values, all endpoint examples, valid JSON, exact code
preservation, Turkish characters, Markdown fences/lists/links, HTML escaping,
client metadata and unique prompt IDs. Existing auth, no-store and OpenAPI
download coverage is retained. `templ generate`, Tailwind rebuild,
`go test ./internal/web/... ./internal/apidoc/... ./internal/api/...`, focused
`go vet`, normal and CGO-disabled builds, gofmt and diff whitespace checks passed.

Sonraki görev: WKY-16 (Markdown görünümü ve kopyalama kontrolleri).
