import { test, expect } from "@playwright/test";

test.describe("procpane UI (server build)", () => {
  test("smoke: shell, filters, and list pane", async ({ page }) => {
    await page.goto("/");

    await expect(page.getByRole("heading", { name: /procpane/i })).toBeVisible();
    await expect(
      page.getByPlaceholder("按端口或进程名筛选…"),
    ).toBeVisible();
    await expect(page.getByRole("button", { name: "全部" })).toBeVisible();
    await expect(page.getByRole("button", { name: "TCP" })).toBeVisible();
    await expect(page.getByRole("button", { name: "UDP" })).toBeVisible();
    await expect(page.getByRole("button", { name: "刷新" })).toBeVisible();

    await page.screenshot({
      path: "test-results/procpane-smoke.png",
      fullPage: true,
    });
  });

  test("service: LISTEN table or empty state after load", async ({ page }) => {
    await page.goto("/");

    const table = page.locator("table");
    const empty = page.getByText("没有匹配的 LISTEN 端口");
    await expect(table.or(empty)).toBeVisible({ timeout: 30_000 });

    await expect(page.getByRole("columnheader", { name: "Port" })).toBeVisible();
    await expect(page.getByRole("columnheader", { name: "Proto" })).toBeVisible();
    await expect(page.getByRole("columnheader", { name: "PID" })).toBeVisible();
    await expect(page.getByRole("columnheader", { name: "Name" })).toBeVisible();
  });

  test("interaction: proto chips toggle without error", async ({ page }) => {
    await page.goto("/");
    await expect(page.getByRole("button", { name: "全部" })).toBeVisible();

    for (const label of ["TCP", "UDP", "全部"] as const) {
      await page.getByRole("button", { name: label }).click();
      await expect(page.getByRole("heading", { name: /procpane/i })).toBeVisible();
    }
  });
});
