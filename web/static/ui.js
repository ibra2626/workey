(() => {
  "use strict";
  const root = document.documentElement;
  const status = document.getElementById("request-status");
  const error = document.getElementById("request-error");
  let pendingTimer;
  let navigationTimer;
  let focusID;

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
    focusID = button?.matches(":focus-visible") ? form.dataset.sessionId : null;
    form.setAttribute("aria-busy", "true");
    start("Değişiklik kaydediliyor…");
  });

  document.addEventListener("htmx:afterRequest", (event) => {
    event.detail.elt.removeAttribute("aria-busy");
    finish();
    if (event.detail.failed || event.detail.xhr.status === 0) {
      const code = event.detail.xhr.status;
      showError(code === 401 || code === 403
        ? "Oturum doğrulanamadı. Sayfayı yenileyip tekrar giriş yapın."
        : "İşlem doğrulanamadı. Bağlantınızı kontrol edip sayfayı yenileyin ve seansın durumunu kontrol edin.");
      if (focusID) event.detail.elt.querySelector("button")?.focus({ preventScroll: true });
      focusID = null;
    }
  });

  document.addEventListener("htmx:afterSettle", (event) => {
    const content = event.target;
    content.classList.add("content-updated");
    if (focusID && (document.activeElement === document.body || content.contains(document.activeElement))) {
      for (const form of content.querySelectorAll("form[data-session-id]")) {
        if (form.dataset.sessionId === focusID && form.getClientRects().length) {
          form.querySelector("button")?.focus({ preventScroll: true });
          break;
        }
      }
    }
    focusID = null;
    if (status) status.textContent = content.querySelector(".completion-status")
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
})();
