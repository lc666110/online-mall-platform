export function money(cents: number): string {
  return `¥${(cents / 100).toFixed(2)}`;
}
export function images(value: string): string[] {
  try {
    const a: unknown = JSON.parse(value);
    return Array.isArray(a)
      ? a.filter(
          (x): x is string =>
            typeof x === "string" &&
            (x.startsWith("https://") || x.startsWith("/images/")),
        )
      : [];
  } catch {
    return [];
  }
}
export function statusLabel(status: string): string {
  return (
    (
      {
        pending: "待确认",
        confirmed: "已确认",
        shipped: "已发货",
        completed: "已完成",
        cancelled: "已取消",
        refunded: "售后已完成",
        approved: "已批准",
        rejected: "已拒绝",
      } as Record<string, string>
    )[status] ?? status
  );
}
