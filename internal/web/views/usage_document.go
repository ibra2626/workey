package views

import (
	"fmt"
	"strings"

	"github.com/ibra2626/workey/internal/apidoc"
	"github.com/ibra2626/workey/internal/plan"
)

// UsageDocument contains guide content only; navigation and view controls belong
// to the page shell. Both renderers traverse every block, including disclosures
// and client-specific content.
type UsageDocument struct {
	Title  string
	Blocks []UsageBlock
	Client string
}

// SelectClient normalizes the request value without storing a preference.
func (d UsageDocument) SelectClient(client string) UsageDocument {
	switch client {
	case "codex-cli", "claude-code":
		d.Client = client
	default:
		d.Client = "codex-desktop"
	}
	return d
}

func (d UsageDocument) ViewURL(key string, markdown bool) string {
	url := "/kullanim/" + key
	if key == "mcp" {
		url += "?client=" + d.SelectClient(d.Client).Client
		if markdown {
			url += "&view=markdown"
		}
	} else if markdown {
		url += "?view=markdown"
	}
	return url
}

func (d UsageDocument) GuideBlocks() []UsageBlock {
	return clientBlocks(d.Blocks, d.SelectClient(d.Client).Client)
}

func clientBlocks(blocks []UsageBlock, client string) []UsageBlock {
	var selected []UsageBlock
	for _, b := range blocks {
		if b.Client != "" && b.Client != client {
			continue
		}
		b.Children = clientBlocks(b.Children, client)
		selected = append(selected, b)
	}
	return selected
}

type UsageBlockKind string

const (
	UsageSection   UsageBlockKind = "section"
	UsageParagraph UsageBlockKind = "paragraph"
	UsageCode      UsageBlockKind = "code"
	UsageLink      UsageBlockKind = "link"
	UsageList      UsageBlockKind = "list"
	UsageDetails   UsageBlockKind = "details"
	UsagePrompt    UsageBlockKind = "prompt"
)

// Client and ID allow a presentation to select a client or expose an individual
// prompt for copying without maintaining a second copy of its text.
type UsageBlock struct {
	Kind                            UsageBlockKind
	Text, Language, URL, Client, ID string
	Ordered                         bool
	Items                           []string
	Children                        []UsageBlock
}

func paragraph(text string) UsageBlock { return UsageBlock{Kind: UsageParagraph, Text: text} }
func section(title string, children ...UsageBlock) UsageBlock {
	return UsageBlock{Kind: UsageSection, Text: title, Children: children}
}
func link(text, url string) UsageBlock { return UsageBlock{Kind: UsageLink, Text: text, URL: url} }
func code(text, language string) UsageBlock {
	return UsageBlock{Kind: UsageCode, Text: text, Language: language}
}

func exampleResponse(name string) string {
	return apidoc.OperationResponseJSON(apidoc.Operation{Output: name})
}

// NewUsageDocument uses only public documentation and the service's agent guide;
// it never receives a request token, environment values or a user's plan.
func NewUsageDocument(key string, guide plan.AgentGuide) UsageDocument {
	doc := UsageDocument{Title: usageTitle(key), Blocks: []UsageBlock{
		paragraph("Yapay zeka asistanınız planı hazırlar; Workey kitap, bölüm ve sayfa referanslarını saklar ve ilerlemeyi izler."),
	}}
	switch key {
	case "mcp":
		doc.Blocks = []UsageBlock{paragraph("Asistanınız Workey’e bağlanır, çalışma planınızı kaydeder ve ilerlemenizi okur.")}
		doc.Blocks = append(doc.Blocks, mcpSections()...)
		tools := section("Araçlar")
		tools.Kind = UsageDetails
		for _, tool := range mcpToolDescriptions {
			tools.Children = append(tools.Children, paragraph(tool.Name+" — "+tool.Description))
		}
		examples := section("Örnek istemler",
			UsageBlock{Kind: UsagePrompt, ID: "week", Text: "Önce ajan rehberini ve mevcut planı oku. Verdiğim kitap/bölüm/sayfa referanslarıyla bu haftayı düzenle. Korunacak seansları koru, yazdıktan sonra tekrar okuyup sonucu bildir."},
			UsageBlock{Kind: UsagePrompt, ID: "makeup", Text: "Yapılamayan seansı hafta sonuna telafi olarak ekle. Var olan hafta sonu seanslarını koru ve bağlantıyı doğrula."},
			UsageBlock{Kind: UsagePrompt, ID: "complete", Text: "Bugünkü seansımı tamamla; ardından güncel ilerlemeyi oku."},
			UsageBlock{Kind: UsagePrompt, ID: "undo", Text: "Son tamamlamayı geri al."},
			link("OpenAI resmî MCP dokümantasyonu", "https://learn.chatgpt.com/docs/extend/mcp"),
		)
		examples.Kind = UsageDetails
		doc.Blocks = append(doc.Blocks, tools, examples)
	case "api":
		contract := section("Sözleşme ve kimlik doğrulama",
			paragraph("Taban yol /api/v1. Tüm endpoint’ler Authorization: Bearer token ister. Gövdeler JSON, üst sınır 1 MiB. Yanıtlar önbelleğe alınmaz."),
			link("OpenAPI 3.1 dosyasını indir", "/kullanim/api/openapi.json"),
			paragraph("Ajan aynı sözleşmeyi Bearer doğrulamasıyla GET /api/v1/openapi.json üzerinden alabilir. Şema tüm istek/yanıt alanlarını ve zorunluluklarını içerir."),
		)
		for _, op := range apidoc.Operations() {
			b := UsageBlock{Kind: UsageDetails, Text: op.Method + " /api/v1" + op.Path + " — " + op.Summary}
			if op.Output != "" {
				b.Children = []UsageBlock{
					paragraph("26 Ekim 2026 tarihli temsili senaryonun tam yanıtı. POST tamamlamada completed=true ve state=done olur; DELETE geri almada tarihine göre state tekrar hesaplanır."),
					code(apidoc.OperationResponseJSON(op), "json"),
				}
			} else {
				b.Children = []UsageBlock{link("Sözleşmeyi indir", "/kullanim/api/openapi.json")}
			}
			contract.Children = append(contract.Children, b)
		}
		doc.Blocks = append(doc.Blocks, contract)
		doc.Blocks = append(doc.Blocks, apiSections()...)
	default:
		doc.Blocks = append(doc.Blocks, section("Nasıl bağlanacaksınız?",
			paragraph("Claude veya Codex kullanıyorsanız MCP rehberiyle başlayın. Kendi istemciniz veya otomasyonunuz için REST API rehberini kullanın. İkisi de aynı plan üzerinde çalışır."),
			paragraph("Web sayfaları oturum çereziyle, MCP ve API ise Bearer token ile korunur. Web oturumu API tokenı yerine geçmez."),
		))
	}
	doc.Blocks = append(doc.Blocks, section("Ajanın izleyeceği işlem sırası",
		paragraph("Sunucu tarihi: "+guide.Today+" · "+guide.Timezone+" · Sınav: "+guide.ExamDate),
		UsageBlock{Kind: UsageList, Ordered: true, Items: guide.Workflow},
	))
	rules := section("Ortak işlem kuralları", paragraph("Kitaplar: "+strings.Join(guide.Books, ", ")+". Seans türleri: "+strings.Join(guide.SessionKinds, ", ")+"."))
	for _, rule := range guide.Rules {
		rules.Children = append(rules.Children, paragraph(rule))
	}
	doc.Blocks = append(doc.Blocks, rules)
	if key == "mcp" {
		for i := range doc.Blocks {
			if doc.Blocks[i].Text == "Ajanın izleyeceği işlem sırası" || doc.Blocks[i].Text == "Ortak işlem kuralları" {
				doc.Blocks[i].Kind = UsageDetails
			}
		}
	}
	return doc
}

func usageSection(title string, paragraphs []string, example, language, client string) UsageBlock {
	b := section(title)
	b.Client = client
	for _, p := range paragraphs {
		b.Children = append(b.Children, paragraph(p))
	}
	if example != "" {
		b.Children = append(b.Children, code(example, language))
	}
	return b
}

// Markdown includes all clients and closed details. Code is emitted verbatim;
// fences grow to remain valid even when a snippet contains Markdown fences.
func (d UsageDocument) Markdown() string {
	var out strings.Builder
	out.WriteString("# " + markdownText(d.Title) + "\n\n")
	writeMarkdown(&out, d.Blocks, 2)
	return out.String()
}

func writeMarkdown(out *strings.Builder, blocks []UsageBlock, level int) {
	for _, b := range blocks {
		switch b.Kind {
		case UsageSection, UsageDetails:
			out.WriteString(strings.Repeat("#", min(level, 6)) + " " + markdownText(b.Text) + "\n\n")
			writeMarkdown(out, b.Children, level+1)
		case UsageParagraph:
			out.WriteString(markdownText(b.Text) + "\n\n")
		case UsageCode, UsagePrompt:
			fence := "```"
			for strings.Contains(b.Text, fence) {
				fence += "`"
			}
			out.WriteString(fence + b.Language + "\n" + b.Text)
			if !strings.HasSuffix(b.Text, "\n") {
				out.WriteByte('\n')
			}
			out.WriteString(fence + "\n\n")
		case UsageLink:
			url := strings.NewReplacer(" ", "%20", "(", "%28", ")", "%29", "<", "%3C", ">", "%3E", "\n", "%0A").Replace(b.URL)
			out.WriteString("[" + markdownText(b.Text) + "](" + url + ")\n\n")
		case UsageList:
			for i, item := range b.Items {
				prefix := "- "
				if b.Ordered {
					prefix = fmt.Sprintf("%d. ", i+1)
				}
				text := strings.ReplaceAll(markdownText(item), "\n", "\n"+strings.Repeat(" ", len(prefix)))
				out.WriteString(prefix + text + "\n")
			}
			out.WriteByte('\n')
		}
	}
}

func markdownText(s string) string {
	return strings.NewReplacer("\\", "\\\\", "`", "\\`", "*", "\\*", "_", "\\_", "[", "\\[", "]", "\\]", "<", "&lt;", ">", "&gt;", "#", "\\#", "!", "\\!", "|", "\\|", ".", "\\.", "-", "\\-", "+", "\\+").Replace(s)
}
