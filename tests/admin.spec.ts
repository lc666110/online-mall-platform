import { test, expect } from "@playwright/test";
import fs from "node:fs";
test("后台登录、真实统计和分类增改删", async ({ page }) => {
  const file = process.env.MALL_TEST_CREDENTIALS;
  test.skip(
    !file,
    "MALL_TEST_CREDENTIALS required for admin browser integration",
  );
  const credentials = JSON.parse(fs.readFileSync(file!, "utf8")) as {
    admin_email: string;
    admin_password: string;
  };
  await page.goto(
    `${process.env.MALL_ADMIN_URL || "http://127.0.0.1:5174"}/login`,
  );
  await page.getByPlaceholder("管理员邮箱").fill(credentials.admin_email);
  await page
    .getByPlaceholder("密码", { exact: true })
    .fill(credentials.admin_password);
  await page.getByRole("button", { name: "登录工作台" }).click();
  await expect(page.getByRole("heading", { name: "经营概览" })).toBeVisible();
  await expect(page.getByText("累计销售额", { exact: true })).toBeVisible();
  await page.getByRole("link", { name: "分类管理" }).click();
  await page.getByRole("button", { name: "新增分类" }).click();
  const name = `浏览器分类${Date.now()}`;
  await page.getByPlaceholder("分类名称").fill(name);
  await page.getByRole("button", { name: "保存分类" }).click();
  await expect(page.getByRole("cell", { name, exact: true })).toBeVisible();
  const row = page.getByRole("row").filter({ hasText: name });
  await row.getByRole("button", { name: "编辑" }).click();
  await page.getByPlaceholder("分类名称").fill(name + "已编辑");
  await page.getByRole("button", { name: "保存分类" }).click();
  await expect(
    page.getByRole("cell", { name: name + "已编辑", exact: true }),
  ).toBeVisible();
  await page
    .getByRole("row")
    .filter({ hasText: name + "已编辑" })
    .getByRole("button", { name: "删除" })
    .click();
  await page.getByRole("button", { name: "确认", exact: true }).click();
  await expect(
    page.getByRole("cell", { name: name + "已编辑", exact: true }),
  ).toHaveCount(0);
});
