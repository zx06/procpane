import { test, expect } from "@playwright/test";

test.describe("procpane UI (server build)", () => {
  test("smoke: shell, filters, and card stack pane", async ({ page }) => {
    await page.goto("/");

    await expect(page.getByRole("heading", { name: /procpane/i })).toBeVisible();
    await expect(page.getByText("本地 LISTEN 端口")).toBeVisible();
    await expect(page.getByPlaceholder("端口 / 进程…")).toBeVisible();
    await expect(page.getByRole("button", { name: "全部" })).toBeVisible();
    await expect(page.getByRole("button", { name: "TCP" })).toBeVisible();
    await expect(page.getByRole("button", { name: "UDP" })).toBeVisible();
    await expect(page.getByRole("button", { name: "刷新" })).toBeVisible();
    await expect(page.locator(".cards-scroll")).toBeVisible();
    await expect(page.locator(".count-pill")).toBeVisible();
    await expect(page.getByRole("button", { name: "退出" })).toBeVisible();

    await page.screenshot({
      path: "test-results/procpane-smoke.png",
      fullPage: true,
    });
  });

  test("service: listener cards or empty state after load", async ({ page }) => {
    await page.goto("/");

    const cards = page.locator(".cards-scroll .pcard");
    const empty = page.getByText("没有匹配的 LISTEN 端口");
    await expect(cards.first().or(empty)).toBeVisible({ timeout: 30_000 });

    const count = await cards.count();
    if (count > 0) {
      const firstCard = cards.first();
      const ports = firstCard.locator(".pcard-ports .port");
      const badges = firstCard.locator(".pcard-ports .badge");
      expect(await ports.count()).toBeGreaterThanOrEqual(1);
      expect(await badges.count()).toBeGreaterThanOrEqual(1);
      await expect(ports.first()).toBeVisible();
      await expect(badges.first()).toBeVisible();
      await expect(firstCard.getByText(/^PID \d+/)).toBeVisible();
    } else {
      await expect(empty).toBeVisible();
    }
  });

  test("screenshots: expanded listener card for PR comment", async ({ page }) => {
    await page.goto("/");

    const appCard = page.locator(".pcard").filter({ hasText: "procpane-app" });
    await expect(appCard).toBeVisible({ timeout: 30_000 });

    await appCard.scrollIntoViewIfNeeded();
    await appCard.locator(".pcard-top").click({ force: true });
    await expect(appCard.locator(".pcard-detail")).toBeVisible({ timeout: 10_000 });

    await page.screenshot({
      path: "test-results/procpane-expanded.png",
      fullPage: true,
    });
  });

  test("interaction: proto chips toggle without error", async ({ page }) => {
    await page.goto("/");
    await expect(page.locator(".cards-scroll")).toBeVisible();

    for (const label of ["TCP", "UDP", "全部"] as const) {
      await page.getByRole("button", { name: label }).click();
      await expect(page.getByRole("heading", { name: /procpane/i })).toBeVisible();
      const cards = page.locator(".cards-scroll .pcard");
      const empty = page.getByText("没有匹配的 LISTEN 端口");
      await expect(cards.first().or(empty)).toBeVisible({ timeout: 30_000 });
    }
  });

  test("visibility: manual refresh after hidden stops auto poll", async ({
    page,
  }) => {
    await page.goto("/");
    await expect(page.getByRole("button", { name: "刷新" })).toBeVisible();

    await page.evaluate(() => {
      Object.defineProperty(document, "hidden", {
        configurable: true,
        get: () => true,
      });
      document.dispatchEvent(new Event("visibilitychange"));
    });

    await page.getByRole("button", { name: "刷新" }).click();
    await expect(page.getByText(/已刷新 · .*s 前/)).toBeVisible({
      timeout: 30_000,
    });
  });
});
