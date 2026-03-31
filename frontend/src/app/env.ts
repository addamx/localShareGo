export const desktopApp = window.go?.main?.App;
export const isDesktopRuntime = Boolean(desktopApp?.GetBootstrapContext);

export async function copyText(text: string) {
  const value = text.trim();
  if (!value) {
    return;
  }

  if (isDesktopRuntime && desktopApp) {
    await desktopApp.CopyText(value);
    return;
  }

  if (typeof navigator.clipboard?.writeText === "function") {
    await navigator.clipboard.writeText(value);
    return;
  }

  const textarea = document.createElement("textarea");
  textarea.value = value;
  textarea.setAttribute("readonly", "true");
  textarea.className = "fixed left-[-9999px] top-0 opacity-0";
  document.body.appendChild(textarea);
  textarea.select();
  document.execCommand("copy");
  document.body.removeChild(textarea);
}
