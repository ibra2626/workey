package views

import (
	"bytes"
	"context"
	"encoding/json"
	"html"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/ibra2626/workey/internal/apidoc"
	"github.com/ibra2626/workey/internal/clock"
	"github.com/ibra2626/workey/internal/plan"
)

func TestUsageDocumentRenderersShareAllContent(t *testing.T) {
	// UTC evening is already the next day in Istanbul.
	svc := plan.New(nil, clock.Fake{Current: time.Date(2030, 2, 4, 22, 30, 0, 0, time.UTC)})
	guide := svc.AgentGuide(clock.Date{Year: 2031, Month: 7, Day: 20})
	// Distinct values catch a second hard-coded catalog in either renderer.
	guide.Books = []string{"kitap-ç"}
	guide.SessionKinds = []string{"tür-ğ"}
	for _, key := range []string{"kullanim", "mcp", "api"} {
		t.Run(key, func(t *testing.T) {
			doc := NewUsageDocument(key, guide)
			var rendered bytes.Buffer
			if err := Usage(key, "csrf-shell-only", doc, false).Render(context.Background(), &rendered); err != nil {
				t.Fatal(err)
			}
			if key == "mcp" {
				for _, client := range []string{"codex-cli", "claude-code"} {
					if err := Usage(key, "csrf-shell-only", doc.SelectClient(client), false).Render(context.Background(), &rendered); err != nil {
						t.Fatal(err)
					}
				}
			}
			htmlText := html.UnescapeString(rendered.String())
			md := doc.Markdown()
			walkBlocks(doc.Blocks, func(b UsageBlock) {
				if b.Text != "" {
					if !strings.Contains(htmlText, b.Text) {
						t.Errorf("HTML missing %q", b.Text)
					}
					want := markdownText(b.Text)
					if b.Kind == UsageCode || b.Kind == UsagePrompt {
						want = b.Text
					}
					if !strings.Contains(md, want) {
						t.Errorf("Markdown missing %q", want)
					}
				}
				for _, item := range b.Items {
					if !strings.Contains(htmlText, item) || !strings.Contains(md, markdownText(item)) {
						t.Errorf("missing list item %q", item)
					}
				}
				if b.Kind == UsageLink && (!strings.Contains(htmlText, `href="`+b.URL+`"`) || !strings.Contains(md, "]("+b.URL+")")) {
					t.Errorf("missing link %q", b.URL)
				}
				if b.Kind == UsageCode && !strings.Contains(md, "```"+b.Language+"\n"+b.Text+"\n```") {
					t.Errorf("code or language changed: %s", b.Language)
				}
			})
			for _, text := range []string{"2030-02-05", "2031-07-20", "Europe/Istanbul", "kitap-ç", "tür-ğ"} {
				if !strings.Contains(htmlText, text) || !strings.Contains(md, markdownText(text)) {
					t.Errorf("missing dynamic guide value %q", text)
				}
			}
			for _, shell := range []string{"csrf-shell-only", "Kullanım rehberleri", "<nav", "<details", "Çıkış"} {
				if strings.Contains(md, shell) {
					t.Errorf("Markdown contains shell markup %q", shell)
				}
			}
			if strings.Contains(htmlText, "Kitaplar: M1, M2") {
				t.Fatal("hard-coded catalog")
			}
		})
	}
}

func walkBlocks(blocks []UsageBlock, visit func(UsageBlock)) {
	for _, b := range blocks {
		visit(b)
		walkBlocks(b.Children, visit)
	}
}

func TestUsageAPIExamplesComeFromContract(t *testing.T) {
	doc := NewUsageDocument("api", plan.AgentGuide{})
	var details []UsageBlock
	var codes []string
	walkBlocks(doc.Blocks, func(b UsageBlock) {
		if b.Kind == UsageDetails && (strings.HasPrefix(b.Text, "GET /api/v1") || strings.HasPrefix(b.Text, "PUT /api/v1") || strings.HasPrefix(b.Text, "POST /api/v1") || strings.HasPrefix(b.Text, "DELETE /api/v1")) {
			details = append(details, b)
		}
		if b.Kind == UsageCode && b.Language == "json" {
			if !json.Valid([]byte(b.Text)) {
				t.Errorf("invalid example JSON: %s", b.Text)
			}
			codes = append(codes, b.Text)
		}
	})
	operations := apidoc.Operations()
	if len(details) != len(operations) {
		t.Fatalf("got %d endpoint disclosures, want %d", len(details), len(operations))
	}
	for i, op := range operations {
		b := details[i]
		if b.Text != op.Method+" /api/v1"+op.Path+" — "+op.Summary {
			t.Errorf("endpoint drift: %q", b.Text)
		}
		if op.Output != "" && b.Children[1].Text != apidoc.OperationResponseJSON(op) {
			t.Errorf("response drift: %s", op.ID)
		}
	}
	for _, want := range []string{apidoc.ExampleOutline, apidoc.ExamplePlan, exampleResponse("Plan"), exampleResponse("PlanResult")} {
		found := false
		for _, got := range codes {
			if got == want {
				found = true
			}
		}
		if !found {
			t.Errorf("missing exact contract example %q", want)
		}
	}
}

func TestUsageMarkdownSyntaxAndCodeSafety(t *testing.T) {
	snippet := "Türkçe şğİıöüç\r\n```markdown\n<script>alert('örnek')</script>\n````\n"
	doc := UsageDocument{Title: "Rehber [örnek]", Blocks: []UsageBlock{
		section("Başlık", paragraph("Metin <etiket> & Türkçe"),
			UsageBlock{Kind: UsageList, Items: []string{"Bir", "İki\nDevam"}},
			UsageBlock{Kind: UsageList, Ordered: true, Items: []string{"İlk", "Son"}},
			link("Bağlantı [tam]", "https://example.com/a(b)"),
			UsageBlock{Kind: UsageDetails, Text: "Kapalı bölüm", Children: []UsageBlock{code(snippet, "markdown")}},
			UsageBlock{Kind: UsagePrompt, ID: "sample", Text: "Ayrı istem"},
		),
	}}
	md := doc.Markdown()
	for _, want := range []string{
		"# Rehber \\[örnek\\]\n", "## Başlık\n", "### Kapalı bölüm\n",
		"- Bir\n- İki\n  Devam\n", "1. İlk\n2. Son\n",
		"[Bağlantı \\[tam\\]](https://example.com/a%28b%29)",
		"`````markdown\n" + snippet + "`````\n", "```\nAyrı istem\n```",
	} {
		if !strings.Contains(md, want) {
			t.Errorf("missing syntax %q in %s", want, md)
		}
	}
	var rendered bytes.Buffer
	if err := usageBlocks(doc.Blocks).Render(context.Background(), &rendered); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(rendered.String(), "<script>") || !strings.Contains(rendered.String(), "&lt;script&gt;") {
		t.Fatal("code rendered as active HTML")
	}
}

func TestUsageMCPClientsAndPrompts(t *testing.T) {
	doc := NewUsageDocument("mcp", plan.AgentGuide{})
	clients := map[string]bool{}
	ids := map[string]bool{}
	walkBlocks(doc.Blocks, func(b UsageBlock) {
		if b.Client != "" {
			clients[b.Client] = true
		}
		if b.Kind == UsagePrompt {
			if b.ID == "" || ids[b.ID] {
				t.Errorf("invalid prompt ID %q", b.ID)
			}
			ids[b.ID] = true
			if !strings.Contains(doc.Markdown(), b.Text) {
				t.Errorf("missing prompt %s", b.ID)
			}
		}
	})
	for _, client := range []string{"claude-code", "codex-cli", "codex-desktop"} {
		if !clients[client] {
			t.Errorf("missing client %s", client)
		}
	}
	if len(ids) != 8 {
		t.Errorf("got %d prompts", len(ids))
	}
}

// Every separate copy target must preserve the shared source exactly, including
// prompts nested in initially closed disclosures and Turkish punctuation.
func TestAllUsagePromptTargets(t *testing.T) {
	doc := NewUsageDocument("mcp", plan.AgentGuide{})
	want := map[string]string{}
	walkBlocks(doc.Blocks, func(b UsageBlock) {
		if b.Kind == UsagePrompt {
			want["usage-prompt-"+b.ID] = b.Text
		}
	})
	pattern := regexp.MustCompile(`(?s)<textarea[^>]*id="(usage-prompt-[^"]+)"[^>]*>(.*?)</textarea>`)
	for _, client := range []string{"codex-desktop", "codex-cli", "claude-code"} {
		var rendered bytes.Buffer
		if err := Usage("mcp", "csrf", doc.SelectClient(client), false).Render(context.Background(), &rendered); err != nil {
			t.Fatal(err)
		}
		matches := pattern.FindAllStringSubmatch(rendered.String(), -1)
		if len(matches) != len(want) {
			t.Fatalf("%s: got %d copy targets, want %d", client, len(matches), len(want))
		}
		for _, match := range matches {
			if text, ok := want[match[1]]; !ok || html.UnescapeString(match[2]) != text {
				t.Errorf("%s: changed prompt %s", client, match[1])
			}
		}
	}
}

func TestExamPlanPrompt(t *testing.T) {
	svc := plan.New(nil, clock.Fake{Current: time.Date(2030, 2, 4, 22, 30, 0, 0, time.UTC)})
	doc := NewUsageDocument("mcp", svc.AgentGuide(clock.Date{Year: 2031, Month: 7, Day: 20}))
	var prompt string
	walkBlocks(doc.Blocks, func(b UsageBlock) {
		if b.Kind == UsagePrompt && b.ID == "exam-plan" {
			prompt = b.Text
		}
	})
	if prompt == "" {
		t.Fatal("exam plan prompt missing")
	}
	for _, want := range []string{
		"Hedefim ve mevcut seviyem: [", "Uygun gün ve saatlerim: [", "Seans sürelerim: [", "Dinlenme günlerim: [", "Kitap/bölüm/sayfa referanslarım: [",
		"dört akşam × 90 dakika", "bir hafta sonu × 180 dakika", "today, timezone ve exam_date", "get_outline → ilgili get_plan → get_progress",
		"Doldurulmamış alanlarla yazmaya başlama", "referansı uydurma", "yanıtımı bekle", "save_outline", "korunacak mevcut faz ve grupları da içeren TAM liste",
		"Pazartesi–Pazar", "çakışmayan", "İlk aralığın başlangıcı today, son aralığın bitişi exam_date", "HER haftayı yazmadan hemen önce", "listeden çıkarılan başlanmamış açık seanslar silinir",
		"Geçmiş, tamamlanmış ve başlanmış kayıtları koru", "study ve yer imini incele", "skipped reason=started", "eksikse korunur", "notation ve study okuma alanlarını gönderme", "topic_group_key", "makeup_of", "created/updated/deleted/unchanged/skipped", "skipped kayıtlarını başarı diye sunma",
		"Her yazmadan sonra ilgili get_plan ve get_progress", "doğrulama hatasında DUR", "Sonraki haftaları yazma", "Tamamlanan ve kalan aralıkları", "Türkçe özetle", "complete_session çağırma",
	} {
		if !strings.Contains(prompt, want) {
			t.Errorf("missing planning instruction %q", want)
		}
	}
	if regexp.MustCompile(`\d{4}-\d{2}-\d{2}`).MatchString(prompt) {
		t.Fatal("prompt must read dates at execution time")
	}
	for _, client := range []string{"codex-desktop", "codex-cli", "claude-code"} {
		var rendered bytes.Buffer
		if err := Usage("mcp", "csrf", doc.SelectClient(client), false).Render(context.Background(), &rendered); err != nil {
			t.Fatal(err)
		}
		body := rendered.String()
		target := regexp.MustCompile(`(?s)<textarea[^>]*id="usage-prompt-exam-plan"[^>]*>(.*?)</textarea>`).FindStringSubmatch(body)
		if len(target) != 2 || html.UnescapeString(target[1]) != prompt || !strings.Contains(doc.Markdown(), "```\n"+prompt+"\n```") {
			t.Fatal("displayed/copy prompt and Markdown differ")
		}
		// Check visible guidance independently of the hidden full Markdown.
		visible := regexp.MustCompile(`(?s)<div id="usage-markdown-panel".*?</div>`).ReplaceAllString(html.UnescapeString(body), "")
		for _, want := range []string{"Sınav tarihi: 2031-07-20", "Sunucu tarihi: 2030-02-05", "açık seansları tam listeyle değiştirir", `data-copy-target="usage-prompt-exam-plan"`, `data-copy-status="usage-prompt-status-exam-plan"`, "İstemi kopyala"} {
			if !strings.Contains(visible, want) {
				t.Errorf("missing visible guidance/control %q", want)
			}
		}
		if strings.Contains(visible, "hx-post") || strings.Contains(visible, "hx-put") {
			t.Fatal("guide must not write a plan")
		}
	}
}

func TestRawUsageEscapesHTML(t *testing.T) {
	doc := UsageDocument{Title: "Test", Blocks: []UsageBlock{{Kind: UsageCode, Text: "</textarea><script>alert('ç')</script>"}}}
	var b bytes.Buffer
	if err := Usage("mcp", "csrf", doc, true).Render(context.Background(), &b); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(b.String(), "<script>alert") || !strings.Contains(b.String(), "&lt;/textarea&gt;") {
		t.Fatal("raw Markdown interpreted as HTML")
	}
}

// Check real navigation/copy contracts rather than visual classes: unique IDs,
// valid anchors, verbatim snippets, and selected-client targets.
func TestUsageSectionNavigationAndCodeTargets(t *testing.T) {
	for _, key := range []string{"mcp", "api"} {
		for _, client := range []string{"codex-desktop", "codex-cli", "claude-code"} {
			doc := NewUsageDocument(key, plan.AgentGuide{}).SelectClient(client)
			var rendered bytes.Buffer
			if err := Usage(key, "csrf", doc, false).Render(context.Background(), &rendered); err != nil {
				t.Fatal(err)
			}
			body := rendered.String()
			ids := map[string]bool{}
			for _, m := range regexp.MustCompile(`\bid="([^"]+)"`).FindAllStringSubmatch(body, -1) {
				if ids[m[1]] {
					t.Errorf("duplicate ID: %s", m[1])
				}
				ids[m[1]] = true
			}
			for _, m := range regexp.MustCompile(`href="#([^"]+)"`).FindAllStringSubmatch(body, -1) {
				if !ids[m[1]] {
					t.Errorf("broken fragment: %s", m[1])
				}
			}
			codes := map[string]string{}
			walkBlocks(doc.GuideBlocks(), func(b UsageBlock) {
				if b.Kind == UsageCode {
					codes["usage-code-"+b.ID] = b.Text
				}
			})
			matches := regexp.MustCompile(`(?s)<pre[^>]*id="(usage-code-[^"]+)"[^>]*><code[^>]*>(.*?)</code></pre>`).FindAllStringSubmatch(body, -1)
			if len(matches) != len(codes) {
				t.Fatalf("copy target count: %d, want %d", len(matches), len(codes))
			}
			for _, m := range matches {
				if html.UnescapeString(m[2]) != codes[m[1]] {
					t.Errorf("changed code %s", m[1])
				}
				if !strings.Contains(body, `data-copy-target="`+m[1]+`"`) {
					t.Errorf("missing code control %s", m[1])
				}
			}
			if key == "mcp" {
				start := strings.Index(body, `id="connect"`)
				picker := strings.Index(body, `aria-label="MCP istemcisi"`)
				if start < 0 || picker < start {
					t.Fatal("client selector must be in the connection section")
				}
				if !strings.Contains(body, `#connect"`) || !strings.Contains(body, "üç istemcinin kurulumunu") {
					t.Fatal("missing client navigation/scope")
				}
				preview := regexp.MustCompile(`(?s)data-prompt="exam-plan".*?<details[^>]*>.*?<textarea[^>]*id="usage-prompt-exam-plan"`)
				if !preview.MatchString(body) {
					t.Fatal("exam prompt preview missing")
				}
				if strings.Contains(body, "Promptu kopyala") {
					t.Fatal("inconsistent prompt label")
				}
			} else if strings.Index(body, `id="first-request"`) > strings.Index(body, `id="endpoints"`) {
				t.Fatal("first request must precede endpoint reference")
			}
		}
	}
}
