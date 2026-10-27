# WKY-16 — Markdown görünümü ve kopyalama kontrolleri

## Goal
MCP/API rehberlerini sayfa içinde ham Markdown olarak görüntüle ve tek düğmeyle kopyala.

## Scope
- Önkoşul: WKY-15. Ortak kurallar: `../README.md`.
- `/kullanim/mcp` ve `/kullanim/api` üst bölümüne Rehber / Markdown bağlantıları ve Markdown olarak kopyala düğmesi ekle.
- Varsayılan biçimlendirilmiş rehber; `?view=markdown` sayfa kabuğu içinde etiketli, salt okunur, seçilebilir ham metin alanını açsın. Bilinmeyen view değeri normal rehbere dönsün.
- Görünüm URL üzerinden çalışsın; JavaScript kapalıyken de ham metin açılabilsin. Tercihi hesaba veya yerel depolamaya kaydetme.
- Kopyalama her iki görünümde WKY-15'in tam Markdown çıktısını kullansın; yalnızca görünür DOM metnini kopyalama.
- Küçük, bağımlılıksız bir yerel JavaScript ile Clipboard API kullan. Aynı mekanizma WKY-17/18'deki ayrı istemleri de kopyalayabilsin; bir kopyalama hedefi ve erişilebilir durum bildirimi kullansın.
- Başarıda aria-live üzerinden Kopyalandı bildir. Pano desteklenmiyorsa veya erişim reddedilirse hedef metni görünür yapıp seç; elle kopyalama açıklaması göster. Başarısızlığı başarı diye bildirme.
- Mevcut oturum koruması ve `Cache-Control: no-store` her iki görünümde korunsun. OpenAPI indirme davranışı değişmesin.
- DESIGN.md ile uyumlu mobil/masaüstü kontrolleri ve görünür klavye odağı sağla; yeni framework veya dış script ekleme.

## Acceptance criteria
- [x] İki rehberin normal ve Markdown URL'leri oturumsuz kullanıcıyı girişe yönlendiriyor.
- [x] Markdown görünümü `text/html` sayfa içinde ham, HTML olarak çalıştırılmayan metin gösteriyor; no-store korunuyor.
- [x] Kopyalanan içerik tam Markdown çıktısına eşit; aktif olmayan/kapalı bölümler kaybolmuyor.
- [x] Başarı, desteklenmeyen pano ve reddedilen erişim senaryoları doğru geri bildirim veriyor.
- [x] JavaScript olmadan görünüm geçişi ve elle metin seçme çalışıyor.

## Checks
- `internal/web/usage_test.go`: görünüm, auth, başlıklar, içerik ve mevcut OpenAPI indirmesi için kapsamı genişlet.
- `go tool templ generate`; sınıf değişikliklerinde Tailwind CSS'i yeniden üret.
- `go test ./internal/web/...`
- Tarayıcıda klavyeyle geçiş/kopyalama ve pano hatasında elle kopyalama kontrolü; son kabul WKY-19'da.

## Progress
Completed 2026-10-27. Both guides expose URL-based Rehber / Markdown views
and a progressive-enhancement copy button. Unknown view values render the guide.
A labelled readonly textarea holds the exact full UsageDocument.Markdown output
in either view, including every client and closed API section. templ escapes the
textarea contents; the raw view remains an authenticated, no-store HTML page.

Local dependency-free web/static/copy.js accepts data-copy-target and
data-copy-status IDs for reuse with individual prompts in WKY-17/18. It announces
Kopyalandı only after writeText resolves. Missing Clipboard API or denied access
reveals, focuses and selects the target with manual-copy instructions. The
button stays hidden without JS; view links and manual selection still work.
Controls wrap on mobile, use design tokens and have explicit keyboard focus.

Validation passed: templ generate, Tailwind rebuild, go test ./internal/web/...,
go vet ./internal/web/..., gofmt and git diff --check. Normal and CGO_ENABLED=0 go build ./... passed; a temporary browser-test binary was also built. HTTP tests cover normal/raw/unknown views,
auth redirects, text/html, no-store, exact full Markdown target equality,
HTML escaping and existing OpenAPI download behavior.

Headless Edge browser acceptance passed at 390x844 and 1280x844: both guides,
keyboard Enter navigation/copy, visible focus, actual clipboard writes/readback,
raw view parity and no horizontal overflow. Clipboard writeText argument exactly
matches textarea.value; Windows clipboard readback normalizes LF to CRLF.
Separate unsupported/NotAllowedError scenarios reveal/select the entire text,
focus the textarea and announce manual instructions without false success.
JavaScript-disabled navigation and keyboard text selection passed. Screenshots
were visually inspected at both widths. Browser tests used a temporary dev-open
server and disposable database; production auth is tested by the Go suite.
No external dependencies, persistence preference, commit, push or deploy.

Sonraki görev: WKY-17. Final combined guide acceptance remains WKY-19.

Temporary browser-test servers stopped after verification.
