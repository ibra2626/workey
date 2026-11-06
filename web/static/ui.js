(() => {
  "use strict";
  const root = document.documentElement;
  const status = document.getElementById("request-status");
  const error = document.getElementById("request-error");
  let pendingTimer;
  let navigationTimer;
  let focusID;
  let completing = false;
  const reduceMotion = window.matchMedia("(prefers-reduced-motion: reduce)");

  // The one authored payoff: a small momentum burst from the completed check.
  // Fixed overlay, pointer-events none, capped particle count, GPU transforms,
  // and skipped entirely when motion is reduced.
  function celebrate(anchor) {
    if (reduceMotion.matches || !anchor) return;
    const rect = anchor.getBoundingClientRect();
    if (!rect.width && !rect.height) return;
    const cx = rect.left + rect.width / 2;
    const cy = rect.top + rect.height / 2;
    const cs = getComputedStyle(document.documentElement);
    const colors = ["--color-success", "--color-accent", "--color-accent-2"].map((v) => cs.getPropertyValue(v).trim() || "#ff6b4a");
    const layer = document.createElement("div");
    layer.className = "celebrate-layer";
    layer.setAttribute("aria-hidden", "true");
    const count = 16;
    for (let i = 0; i < count; i++) {
      const dot = document.createElement("span");
      dot.className = "celebrate-dot";
      const angle = (Math.PI * 2 * i) / count + (Math.random() - 0.5) * 0.5;
      const dist = 56 + Math.random() * 72;
      const size = 6 + Math.random() * 6;
      dot.style.left = cx + "px";
      dot.style.top = cy + "px";
      dot.style.setProperty("--dx", Math.cos(angle) * dist + "px");
      dot.style.setProperty("--dy", Math.sin(angle) * dist + 38 + "px");
      dot.style.setProperty("--s", size + "px");
      dot.style.setProperty("--r", Math.random() < 0.5 ? "999px" : "2px");
      dot.style.setProperty("--rot", Math.random() * 180 - 90 + "deg");
      dot.style.setProperty("--c", colors[i % colors.length]);
      dot.style.animationDelay = Math.random() * 60 + "ms";
      layer.appendChild(dot);
    }
    document.body.appendChild(layer);
    setTimeout(() => layer.remove(), 1100);
  }

  function start(message) {
    clearTimeout(pendingTimer);
    if (error) error.hidden = true;
    // Fast requests should not flash a global loading indicator.
    pendingTimer = setTimeout(() => {
      root.classList.add("is-loading");
      if (status) status.textContent = message;
    }, 180);
  }

  function finish() {
    clearTimeout(pendingTimer);
    clearTimeout(navigationTimer);
    root.classList.remove("is-loading");
    if (status) status.textContent = "";
  }

  function showError(message) {
    finish();
    if (error) {
      error.textContent = message;
      error.hidden = false;
    }
  }

  document.addEventListener("htmx:beforeRequest", (event) => {
    const form = event.detail.elt;
    const button = form.querySelector("button");
    focusID = form.contains(document.activeElement) || button?.matches(":focus-visible") ? form.dataset.sessionId : null;
    completing = (form.getAttribute("action") || "").endsWith("/tamamla");
    form.setAttribute("aria-busy", "true");
    start("Değişiklik kaydediliyor…");
  });

  document.addEventListener("htmx:beforeSwap", (event) => {
    const xhr = event.detail.xhr;
    if (xhr.getResponseHeader("X-Workey-Study-Form") === "true" && [409, 422].includes(xhr.status)) {
      event.detail.shouldSwap = true;
      event.detail.isError = false;
    }
  });

  document.addEventListener("htmx:afterRequest", (event) => {
    event.detail.elt.removeAttribute("aria-busy");
    finish();
    const handled = event.detail.xhr.getResponseHeader("X-Workey-Study-Form") === "true";
    if (event.detail.failed || event.detail.xhr.status === 0) completing = false;
    if (!handled && (event.detail.failed || event.detail.xhr.status === 0)) {
      const code = event.detail.xhr.status;
      showError(code === 401 || code === 403
        ? "Oturum doğrulanamadı. Sayfayı yenileyip tekrar giriş yapın."
        : code === 409 ? "Çalışma başka bir sekmede değişti. Sayfayı yenileyip güncel durumu kontrol edin."
        : "İşlem doğrulanamadı. Bağlantınızı kontrol edip sayfayı yenileyin ve seansın durumunu kontrol edin.");
      if (focusID) event.detail.elt.querySelector("button")?.focus({ preventScroll: true });
      focusID = null;
    }
  });

  document.addEventListener("htmx:afterSettle", (event) => {
    const content = event.target;
    content.classList.add("content-updated");
    const feedback = content.querySelector("#study-feedback");
    if (feedback) feedback.focus({ preventScroll: false });
    if (!feedback && focusID && (document.activeElement === document.body || content.contains(document.activeElement))) {
      for (const form of content.querySelectorAll("form[data-session-id]")) {
        if (form.dataset.sessionId === focusID && form.getClientRects().length) {
          form.querySelector("button")?.focus({ preventScroll: true });
          break;
        }
      }
    }
    focusID = null;
    if (completing) {
      completing = false;
      const done = content.querySelector(".completion-status");
      if (done) {
        const panel = content.querySelector("#bugun-panel") || done;
        panel.classList.add("is-celebrating");
        setTimeout(() => panel.classList.remove("is-celebrating"), 900);
        const check = done.querySelector(".completion-check");
        requestAnimationFrame(() => celebrate(check || done));
      }
    }
    if (status) status.textContent = feedback ? feedback.textContent : content.querySelector(".completion-status")
      ? "Seans tamamlandı. İlerleme güncellendi." : "Seans ve ilerleme güncellendi.";
  });

  // Normal page loads retain real links/forms and their no-JavaScript fallback.
  document.addEventListener("click", (event) => {
    const link = event.target.closest("a[href]");
    if (event.defaultPrevented || event.button !== 0 || event.ctrlKey || event.metaKey || event.shiftKey || event.altKey || !link || link.hasAttribute("download") || (link.target && link.target !== "_self")) return;
    const next = new URL(link.href, location.href);
    if (next.origin !== location.origin || (next.pathname === location.pathname && next.search === location.search)) return;
    start("Sayfa yükleniyor…");
    // A cancelled navigation must not leave a permanent loading indicator.
    navigationTimer = setTimeout(finish, 15000);
  });

  document.addEventListener("submit", (event) => {
    const form = event.target;
    if (event.defaultPrevented || form.hasAttribute("hx-post")) return;
    if (form.classList.contains("is-submitting")) {
      event.preventDefault();
      return;
    }
    form.classList.add("is-submitting");
    form.setAttribute("aria-busy", "true");
    start("Sayfa yükleniyor…");
    navigationTimer = setTimeout(reset, 15000);
  });

  function reset() {
    finish();
    document.querySelectorAll("form.is-submitting").forEach((form) => {
      form.classList.remove("is-submitting");
      form.removeAttribute("aria-busy");
    });
  }
  window.addEventListener("pageshow", reset);
  document.addEventListener("visibilitychange", () => {
    root.classList.toggle("motion-paused", document.hidden);
  });

  // Theme toggle: reflect the effective day/night theme, flip and persist it.
  // The button is revealed only once JS is present; without it the OS
  // preference (prefers-color-scheme) still themes the app.
  (() => {
    const btn = document.getElementById("theme-toggle");
    if (!btn) return;
    const media = window.matchMedia("(prefers-color-scheme: dark)");
    const label = btn.querySelector(".theme-toggle-label");
    const effective = () => {
      const forced = root.getAttribute("data-theme");
      return forced === "dark" || forced === "light" ? forced : media.matches ? "dark" : "light";
    };
    const sync = () => {
      const eff = effective();
      btn.dataset.effective = eff;
      btn.setAttribute("aria-pressed", eff === "dark" ? "true" : "false");
      btn.setAttribute("aria-label", eff === "dark" ? "Gündüz temasına geç" : "Gece temasına geç");
      if (label) label.textContent = eff === "dark" ? "Gündüz teması" : "Gece teması";
    };
    btn.hidden = false;
    sync();
    btn.addEventListener("click", () => {
      const next = effective() === "dark" ? "light" : "dark";
      root.setAttribute("data-theme", next);
      try { localStorage.setItem("workey-theme", next); } catch (e) {}
      sync();
    });
    media.addEventListener("change", sync);
  })();
})();
