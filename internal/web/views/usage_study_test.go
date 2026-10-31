package views

import (
	"bytes"
	"context"
	"html"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/ibra2626/workey/internal/plan"
)

func TestStudyContractDocsStayInSync(t *testing.T) {
	want := StudyContractMarkdown()
	for _, path := range []string{"../../../docs/api.md", "../../../docs/mcp.md"} {
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		s := strings.ReplaceAll(string(b), "\r\n", "\n")
		if !strings.Contains(s, "<!-- study-contract:start -->\n"+want+"<!-- study-contract:end -->") {
			t.Errorf("study contract drift: %s", path)
		}
	}
}

func TestPlanningInputsSharedAndCopyable(t *testing.T) {
	var md strings.Builder
	writeMarkdown(&md, planningInputSections(), 2)
	for _, key := range []string{"api", "mcp"} {
		doc := NewUsageDocument(key, plan.AgentGuide{})
		var out bytes.Buffer
		if err := Usage(key, "private-csrf", doc, false).Render(context.Background(), &out); err != nil {
			t.Fatal(err)
		}
		body := out.String()
		for id, prompt := range map[string]string{"first-week": firstWeekPrompt, "replan": replanPrompt} {
			target := regexp.MustCompile(`(?s)<textarea[^>]*id="usage-prompt-` + id + `"[^>]*>(.*?)</textarea>`).FindStringSubmatch(body)
			if len(target) != 2 || html.UnescapeString(target[1]) != prompt || !strings.Contains(body, `data-copy-target="usage-prompt-`+id+`"`) {
				t.Fatalf("%s missing copy target %s", key, id)
			}
			for _, want := range []string{"get_agent_guide", "get_outline", "get_plan", "get_progress", "TAM liste", "skipped reason=started", "yeniden oku", "uydurma", "yanıtımı bekle"} {
				if !strings.Contains(prompt, want) {
					t.Errorf("%s missing safeguard %q", id, want)
				}
			}
			if strings.Contains(prompt, "private-csrf") || regexp.MustCompile(`\d{4}-\d{2}-\d{2}`).MatchString(prompt) {
				t.Fatal("private context or fixed date in prompt")
			}
		}
		if !strings.Contains(doc.Markdown(), md.String()) || strings.Contains(body, "hx-post") {
			t.Fatal("guide parity or copy-only scope broken")
		}
		b, err := os.ReadFile("../../../docs/" + key + ".md")
		if err != nil || !strings.Contains(strings.ReplaceAll(string(b), "\r\n", "\n"), "<!-- planning-inputs:start -->\n"+md.String()+"<!-- planning-inputs:end -->") {
			t.Fatalf("planning inputs drift: %s %v", key, err)
		}
	}
}
