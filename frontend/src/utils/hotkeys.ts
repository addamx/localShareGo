const modifierKeys = new Set(["Control", "Shift", "Alt", "Meta"]);

export function formatHotkeyFromKeyboardEvent(event: KeyboardEvent) {
  const parts: string[] = [];

  if (event.ctrlKey) {
    parts.push("Ctrl");
  }
  if (event.altKey) {
    parts.push("Alt");
  }
  if (event.shiftKey) {
    parts.push("Shift");
  }
  if (event.metaKey) {
    parts.push("Win");
  }

  const mainKey = normalizeMainKey(event.key);
  if (!mainKey || modifierKeys.has(event.key)) {
    return "";
  }

  parts.push(mainKey);
  return parts.join("+");
}

function normalizeMainKey(key: string) {
  if (!key) {
    return "";
  }

  if (key.length === 1) {
    const upper = key.toUpperCase();
    if ((upper >= "A" && upper <= "Z") || (upper >= "0" && upper <= "9")) {
      return upper;
    }
  }

  switch (key) {
    case " ":
      return "Space";
    case "Tab":
    case "Enter":
    case "Escape":
    case "Backspace":
    case "Insert":
    case "Delete":
    case "Home":
    case "End":
    case "PageUp":
    case "PageDown":
      return key;
    case "ArrowLeft":
      return "Left";
    case "ArrowUp":
      return "Up";
    case "ArrowRight":
      return "Right";
    case "ArrowDown":
      return "Down";
    default:
      if (/^F([1-9]|1\d|2[0-4])$/.test(key)) {
        return key;
      }
      return "";
  }
}
