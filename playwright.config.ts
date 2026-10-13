import { defineConfig } from "@playwright/test";
export default defineConfig({
  testDir: "./tests",
  timeout: 30000,
  workers: 1,
  use: {
    baseURL: process.env.MALL_STOREFRONT_URL || "http://127.0.0.1:5173",
    trace: "retain-on-failure",
  },
  reporter: [["list"], ["html", { open: "never" }]],
});
