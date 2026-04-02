const modifierKeys = new Set(["Control", "Shift", "Alt", "Meta"]);
const modifierCodes = new Set([
  "ControlLeft",
  "ControlRight",
  "ShiftLeft",
  "ShiftRight",
  "AltLeft",
  "AltRight",
  "MetaLeft",
  "MetaRight",
]);

export function formatHotkeyFromKeyboardEvent(event: KeyboardEvent, allowModifiersOnly = false) {
  const parts = collectModifierParts(event);

  const mainKey = normalizeMainKey(event);
  if (!mainKey) {
    if (!allowModifiersOnly) {
      return "";
    }
    return parts.join("+");
  }

  parts.push(mainKey);
  return parts.join("+");
}

function collectModifierParts(event: KeyboardEvent) {
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

  return parts;
}

function normalizeMainKey(event: KeyboardEvent) {
  const codeKey = normalizeMainKeyFromCode(event.code);
  if (codeKey) {
    return codeKey;
  }

  if (modifierKeys.has(event.key) || modifierCodes.has(event.code)) {
    return "";
  }

  return normalizeMainKeyFromKey(event.key);
}

function normalizeMainKeyFromCode(code: string) {
  if (!code || modifierCodes.has(code)) {
    return "";
  }

  if (code.startsWith("Key") && code.length === 4) {
    return code.slice(3);
  }
  if (code.startsWith("Digit") && code.length === 6) {
    return code.slice(5);
  }

  switch (code) {
    case "Backquote":
      return "`";
    case "Space":
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
      return code;
    case "ArrowLeft":
      return "Left";
    case "ArrowUp":
      return "Up";
    case "ArrowRight":
      return "Right";
    case "ArrowDown":
      return "Down";
    default:
      if (/^F([1-9]|1\d|2[0-4])$/.test(code)) {
        return code;
      }
      return "";
  }
}

function normalizeMainKeyFromKey(key: string) {
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
    case "`":
      return "`";
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
