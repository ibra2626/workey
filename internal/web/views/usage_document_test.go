package views

import (
	"bytes"
	"context"
	"encoding/json"
	"html"
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
		if b.Kind == UsageDetails {
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
	if len(ids) != 4 {
		t.Errorf("got %d prompts", len(ids))
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
