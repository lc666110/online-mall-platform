import { test, expect } from "@playwright/test";
import fs from "node:fs";
test("注册、资料与退出可在真实商城完成", async ({ page }) => {
  const email = `browser-${Date.now()}@example.com`;
  await page.goto("/login");
  await page.getByRole("button", { name: "创建账号", exact: true }).click();
  await page.getByLabel("昵称").fill("浏览器用户");
  await page.getByLabel("邮箱").fill(email);
  await page.getByLabel("密码", { exact: true }).fill("BrowserPass123!");
  await page.getByRole("button", { name: "注册并登录", exact: true }).click();
  await expect(page).toHaveURL(/\/$/);
  await page.getByRole("link", { name: "浏览器用户", exact: true }).click();
  await page.getByLabel("昵称").fill("好物体验官");
  await page.getByRole("button", { name: "保存资料" }).click();
  await expect(page.getByRole("status")).toContainText("资料已保存");
  await page.getByRole("button", { name: "退出", exact: true }).click();
  await expect(page.getByRole("link", { name: "登录 / 注册" })).toBeVisible();
});

test("移动端长分类名称仍保持页面宽度", async ({ page, request }) => {
  const file = process.env.MALL_TEST_CREDENTIALS;
  test.skip(!file, "MALL_TEST_CREDENTIALS required");
  const credentials = JSON.parse(fs.readFileSync(file!, "utf8"));
  const login = await request.post("/api/v1/auth/login", {
    data: {
      email: credentials.admin_email,
      password: credentials.admin_password,
    },
  });
  expect(login.status()).toBe(200);
  const token = (await login.json()).data.token;
  const headers = { Authorization: `Bearer ${token}` };
  const added = await request.post("/api/v1/admin/categories", {
    headers,
    data: { name: "LongCategory".repeat(5) + Date.now() },
  });
  expect(added.status()).toBe(200);
  const id = (await added.json()).data.id;
  try {
    await page.setViewportSize({ width: 390, height: 844 });
    await page.goto("/");
    await expect(
      page.getByRole("button", { name: /LongCategory/ }),
    ).toBeVisible();
    expect(
      await page.evaluate(
        () => document.documentElement.scrollWidth <= window.innerWidth,
      ),
    ).toBeTruthy();
  } finally {
    expect(
      (
        await request.delete(`/api/v1/admin/categories/${id}`, { headers })
      ).status(),
    ).toBe(200);
  }
});
test("移动端首页没有横向溢出", async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 844 });
  await page.goto("/");
  await expect(
    page.getByRole("heading", { name: "发现你的下一件好物" }),
  ).toBeVisible();
  expect(
    await page.evaluate(
      () => document.documentElement.scrollWidth <= window.innerWidth,
    ),
  ).toBeTruthy();
});
