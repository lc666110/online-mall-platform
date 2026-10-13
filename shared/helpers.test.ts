import { describe, it, expect } from "vitest";
import { money, images, statusLabel } from "./helpers";
describe("商城展示契约", () => {
  it("金额按分展示，保留两位", () => {
    expect(money(1234)).toBe("¥12.34");
    expect(money(0)).toBe("¥0.00");
  });
  it("恶意或损坏图片字段不会产生不安全图片", () => {
    expect(images("broken")).toEqual([]);
    expect(
      images('["javascript:alert(1)","https://example.com/a.jpg"]'),
    ).toEqual(["https://example.com/a.jpg"]);
  });
  it("订单状态为用户可读中文", () => {
    expect(statusLabel("pending")).toBe("待确认");
    expect(statusLabel("shipped")).toBe("已发货");
  });
});
