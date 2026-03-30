export const desktopApp = window.go?.main?.App;
export const isDesktopRuntime = Boolean(desktopApp?.GetBootstrapContext);

export async function copyText(text: string) {
  const value = text.trim();
  if (!value) return;
  if (isDesktopRuntime && desktopApp) {
    await desktopApp.CopyText(value);
    return;
  }
  await navigator.clipboard.writeText(value);
}
