import { test, expect } from "@playwright/test";
import fs from "node:fs";
test("商城下单、后台确认发货、收货评价、售后", async ({ page, browser }) => {
  const file = process.env.MALL_TEST_CREDENTIALS;
  test.skip(!file, "MALL_TEST_CREDENTIALS required");
  const credentials = JSON.parse(fs.readFileSync(file!, "utf8")) as {
    admin_email: string;
    admin_password: string;
  };
  const email = `checkout-${Date.now()}@example.com`;
  await page.goto("/login");
  await page.getByRole("button", { name: "创建账号", exact: true }).click();
  await page.getByLabel("昵称").fill("商城体验官");
  await page.getByLabel("邮箱").fill(email);
  await page.getByLabel("密码", { exact: true }).fill("CheckoutPass123!");
  await page.getByRole("button", { name: "注册并登录" }).click();
  await page
    .getByRole("link")
    .filter({ hasText: "静音无线耳机" })
    .first()
    .click();
  await page
    .getByRole("button", { name: "深色 / 经典款", exact: true })
    .click();
  await page.getByRole("button", { name: /加入购物袋/ }).click();
  await expect(page.getByRole("status")).toContainText("已加入购物袋");
  await page.getByRole("link", { name: "购物袋", exact: true }).click();
  await expect(
    page.getByRole("heading", { name: "喜欢的，都在这里。" }),
  ).toBeVisible();
  await page.getByLabel("数量").fill("2");
  await page.getByLabel("数量").press("Tab");
  await expect(
    page.getByText("¥638.00", { exact: true }).first(),
  ).toBeVisible();
  await expect(
    page.getByRole("button", { name: "提交订单", exact: true }),
  ).toBeEnabled();
  await page.getByLabel("收货人").fill("测试收货人");
  await page.getByLabel("联系电话").fill("13800138000");
  await page.getByLabel("收货地址").fill("测试省测试市测试地址");
  await page.getByRole("button", { name: "提交订单", exact: true }).click();
  await expect(
    page.getByRole("heading", { name: "订单详情", exact: true }),
  ).toBeVisible();
  await expect(
    page.getByText("¥638.00", { exact: true }).first(),
  ).toBeVisible();
  const number = await page.locator(".order-heading h2").innerText();
  const admin = await browser.newPage();
  await admin.goto(
    `${process.env.MALL_ADMIN_URL || "http://127.0.0.1:5174"}/login`,
  );
  await admin.getByPlaceholder("管理员邮箱").fill(credentials.admin_email);
  await admin
    .getByPlaceholder("密码", { exact: true })
    .fill(credentials.admin_password);
  await admin.getByRole("button", { name: "登录工作台" }).click();
  await admin.getByRole("link", { name: "订单与发货" }).click();
  await admin
    .getByRole("row")
    .filter({ hasText: number })
    .getByRole("button", { name: "查看 / 处理" })
    .click();
  await admin.locator(".n-modal .n-base-selection").click();
  await admin.getByText("已确认", { exact: true }).click();
  await admin.getByRole("button", { name: "保存订单状态" }).click();
  await admin
    .getByRole("row")
    .filter({ hasText: number })
    .getByRole("button", { name: "查看 / 处理" })
    .click();
  await admin.locator(".n-modal .n-base-selection").click();
  await admin.getByText("已发货", { exact: true }).click();
  await admin.getByPlaceholder("物流公司").fill("测试快递");
  await admin.getByPlaceholder("运单号").fill("TEST123456");
  await admin.getByRole("button", { name: "保存订单状态" }).click();
  await page.reload();
  await page.getByRole("button", { name: "确认收货" }).click();
  await expect(page.getByText("已完成", { exact: true })).toBeVisible();
  await page.getByRole("button", { name: "评价商品" }).click();
  await page
    .getByLabel("评价内容")
    .fill("真实联调评价：规格、库存与订单流程正常。");
  await page.getByRole("button", { name: "提交评价" }).click();
  await expect(page.getByRole("status")).toContainText("感谢你的评价");
  await page.getByRole("button", { name: "申请售后", exact: true }).click();
  await page.getByLabel("申请原因").fill("测试售后流程");
  await page.getByRole("button", { name: "提交售后申请" }).click();
  await expect(page.getByRole("status")).toContainText("售后申请已提交");
  await admin.getByRole("link", { name: "售后管理" }).click();
  await admin
    .getByRole("row")
    .filter({ hasText: number })
    .getByRole("button", { name: "处理售后" })
    .click();
  await admin.getByPlaceholder("处理备注").fill("已核实演示退货，线下处理");
  await admin.getByRole("button", { name: "提交处理结果" }).click();
  await expect(
    admin.getByRole("row").filter({ hasText: number }),
  ).toContainText("已批准");
  await page.reload();
  await expect(page.getByText("售后已完成", { exact: true })).toBeVisible();
  await admin.close();
});
