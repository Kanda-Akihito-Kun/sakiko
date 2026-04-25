function pad2(value: number): string {
  return String(value).padStart(2, "0");
}

function parseDateValue(value?: string): Date | null {
  if (!value) {
    return null;
  }

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return null;
  }
  return date;
}

export function formatDateTimeForDisplay(value?: string): string {
  const date = parseDateValue(value);
  if (!date) {
    return value || "N/A";
  }

  return `${date.getFullYear()}-${pad2(date.getMonth() + 1)}-${pad2(date.getDate())} ${pad2(date.getHours())}:${pad2(date.getMinutes())}:${pad2(date.getSeconds())}`;
}

export function formatDateTimeForFileName(value?: string): string {
  const date = parseDateValue(value);
  if (!date) {
    return (value || "unknown").replace(/[<>:"/\\|?*\u0000-\u001F]/g, "_");
  }

  return `${date.getFullYear()}-${pad2(date.getMonth() + 1)}-${pad2(date.getDate())}_${pad2(date.getHours())}-${pad2(date.getMinutes())}-${pad2(date.getSeconds())}`;
}
