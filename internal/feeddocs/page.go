package feeddocs

import (
	"fmt"
	"html"
	"strings"
	"unicode"
)

func Render(markdown string) string {
	body := renderMarkdown(markdown)
	return `<!doctype html>
<html lang="cs">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Zpracování dodavatelských feedů</title>
  <style>
    :root {
      --bg: #f5f7f8;
      --surface: #ffffff;
      --text: #172024;
      --muted: #5d6a70;
      --line: #d9e0e4;
      --accent: #0d5f66;
      --accent-soft: #e3f2f3;
      --code-bg: #102126;
      --code-text: #e8f5f6;
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      background: var(--bg);
      color: var(--text);
      font: 16px/1.62 -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
    }
    .page-header, main {
      width: min(1040px, calc(100% - 32px));
      margin: 0 auto;
    }
    .page-header {
      margin-top: 28px;
    }
    .page-header-inner { padding: 8px 0; }
    h1, h2, h3 { line-height: 1.2; letter-spacing: 0; }
    h1 {
      margin: 0;
      max-width: 760px;
      font-size: clamp(32px, 6vw, 52px);
    }
    main { padding: 28px 0 56px; }
    .status-link, article {
      background: var(--surface);
      border: 1px solid var(--line);
      border-radius: 8px;
      box-shadow: 0 12px 30px rgba(16, 33, 38, .06);
    }
    .status-link {
      margin-bottom: 24px;
      padding: 16px 20px;
    }
    .status-link p { margin: 0; }
    article { padding: clamp(20px, 4vw, 40px); }
    article > h1:first-child { display: none; }
    h2 {
      margin: 40px 0 14px;
      padding-top: 10px;
      font-size: 26px;
      border-top: 1px solid var(--line);
    }
    h2:first-of-type { margin-top: 0; border-top: 0; padding-top: 0; }
    h3 {
      margin: 34px 0 12px;
      padding: 16px 0 0;
      font-size: 21px;
      border-top: 1px solid var(--line);
    }
    p { margin: 0 0 14px; color: var(--muted); }
    ul { margin: 0 0 20px; padding-left: 24px; }
    li { margin: 6px 0; }
    code {
      border-radius: 5px;
      background: var(--accent-soft);
      color: #0a4e55;
      padding: 2px 5px;
      font: 0.92em "SFMono-Regular", Consolas, monospace;
      white-space: nowrap;
    }
    .code-block {
      position: relative;
      margin: 8px 0 18px;
    }
    pre {
      margin: 0;
      padding: 14px 58px 14px 16px;
      overflow-x: auto;
      border-radius: 8px;
      background: var(--code-bg);
      color: var(--code-text);
    }
    pre code {
      display: block;
      padding: 0;
      background: transparent;
      color: inherit;
      white-space: pre;
    }
    .copy-code {
      position: absolute;
      top: 8px;
      right: 8px;
      z-index: 1;
      display: inline-flex;
      align-items: center;
      justify-content: center;
      width: 32px;
      height: 32px;
      border: 1px solid #d9e0e4;
      border-radius: 6px;
      background: #f8fbfc;
      color: #102126;
      cursor: pointer;
      box-shadow: 0 2px 8px rgba(16, 33, 38, .24);
    }
    .copy-code:hover {
      border-color: #b8cbd0;
      background: var(--accent-soft);
    }
    .copy-code svg {
      width: 16px;
      height: 16px;
      stroke-width: 2;
    }
    .copy-code .icon-check { display: none; }
    .copy-code.is-copied {
      border-color: #86d79c;
      background: #dcf8e3;
      color: #14532d;
    }
    .copy-code.is-copied .icon-copy { display: none; }
    .copy-code.is-copied .icon-check { display: block; }
    a { color: var(--accent); }
    @media (max-width: 640px) {
      .page-header, main { width: min(100% - 20px, 1040px); }
      article { padding: 18px; }
      code { white-space: normal; }
    }
  </style>
</head>
<body>
  <header class="page-header">
    <div class="page-header-inner">
      <h1>Zpracování dodavatelských feedů</h1>
    </div>
  </header>
  <main>
    <section class="status-link">
      <p>Stav posledních rebuild běhů je dostupný jako JSON na <a href="/status">/status</a>.</p>
    </section>
    <article>` + body + `</article>
  </main>
  <script>
    (() => {
      const timers = new WeakMap();
      const copyText = async (text) => {
        if (navigator.clipboard && window.isSecureContext) {
          await navigator.clipboard.writeText(text);
          return;
        }
        const textarea = document.createElement("textarea");
        textarea.value = text;
        textarea.setAttribute("readonly", "");
        textarea.style.position = "fixed";
        textarea.style.top = "-1000px";
        document.body.appendChild(textarea);
        textarea.select();
        const copied = document.execCommand("copy");
        textarea.remove();
        if (!copied) throw new Error("copy command failed");
      };

      document.addEventListener("click", async (event) => {
        const button = event.target.closest("[data-copy-code]");
        if (!button) return;

        const code = button.closest(".code-block")?.querySelector("code");
        if (!code) return;

        try {
          await copyText(code.textContent.replace(/\n$/, ""));
          button.classList.add("is-copied");
          button.setAttribute("aria-label", "Zkopírováno");
          button.title = "Zkopírováno";
          clearTimeout(timers.get(button));
          timers.set(button, setTimeout(() => {
            button.classList.remove("is-copied");
            button.setAttribute("aria-label", "Zkopírovat kód");
            button.title = "Zkopírovat kód";
          }, 2000));
        } catch (error) {
          console.error("Kopírování se nepodařilo.", error);
        }
      });
    })();
  </script>
</body>
</html>
`
}

func renderMarkdown(markdown string) string {
	lines := strings.Split(strings.ReplaceAll(markdown, "\r\n", "\n"), "\n")
	var builder strings.Builder
	inList := false
	inCode := false

	closeList := func() {
		if inList {
			builder.WriteString("</ul>\n")
			inList = false
		}
	}

	for _, rawLine := range lines {
		line := strings.TrimRight(rawLine, " \t")
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "```") {
			if inCode {
				builder.WriteString("</code></pre></div>\n")
				inCode = false
				continue
			}
			closeList()
			builder.WriteString(`<div class="code-block">` + copyButtonHTML + `<pre><code>`)
			inCode = true
			continue
		}
		if inCode {
			builder.WriteString(html.EscapeString(rawLine))
			builder.WriteByte('\n')
			continue
		}
		if trimmed == "" {
			closeList()
			continue
		}

		switch {
		case strings.HasPrefix(trimmed, "### "):
			closeList()
			writeHeading(&builder, 3, strings.TrimSpace(strings.TrimPrefix(trimmed, "### ")))
		case strings.HasPrefix(trimmed, "## "):
			closeList()
			writeHeading(&builder, 2, strings.TrimSpace(strings.TrimPrefix(trimmed, "## ")))
		case strings.HasPrefix(trimmed, "# "):
			closeList()
			writeHeading(&builder, 1, strings.TrimSpace(strings.TrimPrefix(trimmed, "# ")))
		case strings.HasPrefix(trimmed, "- "):
			if !inList {
				builder.WriteString("<ul>\n")
				inList = true
			}
			fmt.Fprintf(&builder, "<li>%s</li>\n", renderInline(strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))))
		default:
			closeList()
			fmt.Fprintf(&builder, "<p>%s</p>\n", renderInline(trimmed))
		}
	}

	closeList()
	if inCode {
		builder.WriteString("</code></pre></div>\n")
	}
	return builder.String()
}

const copyButtonHTML = `<button class="copy-code" type="button" data-copy-code aria-label="Zkopírovat kód" title="Zkopírovat kód">
  <svg class="icon-copy" aria-hidden="true" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round">
    <rect width="14" height="14" x="8" y="8" rx="2" ry="2"></rect>
    <path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2"></path>
  </svg>
  <svg class="icon-check" aria-hidden="true" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round">
    <path d="M20 6 9 17l-5-5"></path>
  </svg>
</button>`

func writeHeading(builder *strings.Builder, level int, text string) {
	fmt.Fprintf(builder, `<h%d id="%s">%s</h%d>`+"\n", level, slug(text), renderInline(text), level)
}

func renderInline(value string) string {
	parts := strings.Split(value, "`")
	if len(parts) == 1 {
		return html.EscapeString(value)
	}

	var builder strings.Builder
	for index, part := range parts {
		if index%2 == 1 {
			fmt.Fprintf(&builder, "<code>%s</code>", html.EscapeString(part))
		} else {
			builder.WriteString(html.EscapeString(part))
		}
	}
	return builder.String()
}

func slug(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var builder strings.Builder
	lastDash := false
	for _, r := range value {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r):
			builder.WriteRune(r)
			lastDash = false
		case !lastDash && builder.Len() > 0:
			builder.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(builder.String(), "-")
}
