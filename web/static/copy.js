(() => {
  "use strict";
  // Each control points at a readonly text target and its own live status.
  document.querySelectorAll("[data-copy-target]").forEach((button) => {
    const target = document.getElementById(button.dataset.copyTarget);
    const status = document.getElementById(button.dataset.copyStatus);
    if (!target || !status) return;
    button.hidden = false;
    button.addEventListener("click", async () => {
      status.textContent = "";
      button.disabled = true;
      try {
        if (!navigator.clipboard?.writeText) throw new Error("Clipboard unavailable");
        await navigator.clipboard.writeText(target.value);
        status.textContent = "Kopyalandı";
      } catch {
        const panel = target.closest("[data-copy-panel]");
        if (panel) panel.hidden = false;
        target.hidden = false;
        target.focus();
        target.select();
        status.textContent = "Otomatik kopyalama kullanılamadı. Metin seçildi; Ctrl+C (macOS: ⌘C) ile elle kopyalayın.";
      } finally {
        button.disabled = false;
      }
    });
  });
})();
