import { test, expect } from "@playwright/test";
import fs from "node:fs";

test("后台商品库存、上下架、删除和用户权限真实操作", async ({
  page,
  request,
}) => {
  const file = process.env.MALL_TEST_CREDENTIALS;
  test.skip(!file, "MALL_TEST_CREDENTIALS required");
  const credentials = JSON.parse(fs.readFileSync(file!, "utf8")) as {
    admin_email: string;
    admin_password: string;
  };
  const suffix = Date.now(),
    email = `managed-${suffix}@example.com`,
    name = `管理测试商品${suffix}`;
  const result = await request.post("/api/v1/auth/register", {
    data: { email, password: "ManagedPass123!", name: "待管理用户" },
  });
  expect(result.status()).toBe(201);
  await page.goto(
    `${process.env.MALL_ADMIN_URL || "http://127.0.0.1:5174"}/login`,
  );
  await page.getByPlaceholder("管理员邮箱").fill(credentials.admin_email);
  await page
    .getByPlaceholder("密码", { exact: true })
    .fill(credentials.admin_password);
  await page.getByRole("button", { name: "登录工作台" }).click();
  await page.getByRole("link", { name: "商品管理" }).click();
  await page.getByRole("button", { name: "新增商品", exact: true }).click();
  await page.getByPlaceholder("商品名称", { exact: true }).fill(name);
  await page.locator(".n-modal .n-base-selection").click();
  await page
    .getByText("居家生活", { exact: true })
    .filter({ visible: true })
    .last()
    .click();
  await page.getByPlaceholder("SKU", { exact: true }).fill(`TEST-${suffix}`);
  await page
    .locator(".n-modal")
    .getByRole("textbox", { name: "价格（分）", exact: true })
    .fill("12500");
  await page
    .locator(".n-modal")
    .getByRole("textbox", { name: "库存", exact: true })
    .fill("15");
  await page.getByRole("button", { name: "保存商品" }).click();
  const row = page.getByRole("row").filter({ hasText: name });
  await expect(row).toContainText("¥125.00");
  await expect(row).toContainText("15");
  await row.getByRole("button", { name: "下架", exact: true }).click();
  await expect(row).toContainText("已下架");
  await row.getByRole("button", { name: "上架", exact: true }).click();
  await expect(row).toContainText("销售中");
  await row.getByRole("button", { name: "编辑", exact: true }).click();
  await page
    .locator(".n-modal")
    .getByRole("textbox", { name: "库存", exact: true })
    .fill("12");
  await page.getByRole("button", { name: "保存商品" }).click();
  await expect(row).toContainText("12");
  await row.getByRole("button", { name: "删除", exact: true }).click();
  await page.getByRole("button", { name: "确认", exact: true }).click();
  await expect(row).toHaveCount(0);
  await page.getByRole("link", { name: "用户与权限" }).click();
  const userRow = page.getByRole("row").filter({ hasText: email });
  await userRow.getByRole("button", { name: "管理用户" }).click();
  await page.getByPlaceholder("昵称", { exact: true }).fill("运营测试人员");
  await page.locator(".n-modal .n-base-selection").click();
  await page.getByText("运营人员", { exact: true }).click();
  await page.getByRole("button", { name: "保存用户" }).click();
  await expect(userRow).toContainText("staff");
  await userRow.getByRole("button", { name: "管理用户" }).click();
  await page.locator(".n-modal").getByRole("switch").click();
  await page.getByRole("button", { name: "保存用户" }).click();
  await expect(userRow).toContainText("禁用");
  const login = await request.post("/api/v1/auth/login", {
    data: { email, password: "ManagedPass123!" },
  });
  expect(login.status()).toBe(401);
});

test("商城分类、搜索与商品详情加载", async ({ page }) => {
  await page.goto("/");
  await page.getByRole("button", { name: "数码配件", exact: true }).click();
  await expect(
    page.getByRole("link").filter({ hasText: "静音无线耳机" }),
  ).toBeVisible();
  await page.getByLabel("搜索商品").fill("静音");
  await page.getByRole("button", { name: "搜索", exact: true }).click();
  await expect(page.locator(".product-card")).toHaveCount(1);
  await page.getByRole("link").filter({ hasText: "静音无线耳机" }).click();
  await expect(
    page.getByRole("heading", { name: "静音无线耳机", exact: true }),
  ).toBeVisible();
  await expect(
    page.getByRole("button", { name: "深色 / 经典款", exact: true }),
  ).toBeVisible();
});
