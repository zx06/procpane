import { test, expect, type Page } from "@playwright/test";

async function selectSourceTab(
  page: Page,
  label: "本机" | "Docker" | "SSH",
) {
  const nav = page.getByRole("navigation", { name: "数据源" });
  await expect(nav).toBeVisible();
  await nav.getByRole("button", { name: label, exact: true }).click();
}

test.describe("procpane UI (server build)", () => {
  test("smoke: source tabs 本机 / Docker / SSH", async ({ page }) => {
    await page.goto("/");

    const tabs = page.locator("nav.source-tabs");
    await expect(tabs).toBeVisible();
    await expect(tabs.getByRole("button", { name: "本机", exact: true })).toBeVisible();
    await expect(tabs.getByRole("button", { name: "Docker", exact: true })).toBeVisible();
    await expect(tabs.getByRole("button", { name: "SSH", exact: true })).toBeVisible();

    // Default tab remains local.
    await expect(page.getByText("本地 LISTEN 端口")).toBeVisible();
    await expect(tabs.getByRole("button", { name: "本机", exact: true })).toHaveClass(/on/);
  });

  test("docker tab: panel and unreachable or empty state (no daemon required)", async ({
    page,
  }) => {
    await page.goto("/");
    await selectSourceTab(page, "Docker");

    await expect(page.getByRole("heading", { name: "Docker", exact: true })).toBeVisible();
    await expect(page.getByText("已发布端口（Engine API）")).toBeVisible();

    const hintDisconnected = page.getByText(/Docker 未连接/);
    const hintOk = page.getByText(/^Engine · /);
    const emptyPublished = page.getByText("没有已发布的容器端口");
    const dockerCards = page.locator(".cards-scroll .docker-card");
    const dockerHint = page.getByText(/DOCKER_HOST|本地 socket/);
    const emptyError = page.getByText(/无法连接 Docker Engine|加载失败/);

    // Wait until Docker status probe finishes (connected or not).
    await expect
      .poll(
        async () =>
          (await hintDisconnected.isVisible()) || (await hintOk.isVisible()),
        { timeout: 30_000 },
      )
      .toBe(true);

    if (await hintDisconnected.isVisible()) {
      await expect(emptyError.or(dockerHint).first()).toBeVisible();
    } else {
      // Engine reachable: empty list and/or published-port cards (both may show).
      await expect(hintOk).toBeVisible();
      await expect(emptyPublished.or(dockerCards.first())).toBeVisible();
    }
  });

  test("ssh tab: empty state or enabled host chips", async ({ page }) => {
    await page.goto("/");
    await selectSourceTab(page, "SSH");

    await expect(page.getByRole("heading", { name: "SSH", exact: true })).toBeVisible();
    await expect(page.getByText("已启用主机的 LISTEN")).toBeVisible();

    const empty = page.getByText("尚未启用任何 SSH 主机");
    const settingsCta = page.getByRole("button", {
      name: "去设置启用 SSH 主机",
    });
    const hostChip = page.locator(".host-chips .host-chip").first();
    const sshEmptyList = page.getByText("没有 LISTEN 端口");
    const sshCards = page.locator(".cards-scroll .pcard");

    await expect(empty.or(hostChip).or(sshEmptyList).or(sshCards.first())).toBeVisible({
      timeout: 30_000,
    });

    if (await empty.isVisible().catch(() => false)) {
      await expect(settingsCta).toBeVisible();
    }
  });

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

  test("interaction: chain node opens process detail sheet", async ({ page }) => {
    await page.goto("/");

    const appCard = page.locator(".pcard").filter({ hasText: "procpane-app" });
    await expect(appCard).toBeVisible({ timeout: 30_000 });

    await appCard.scrollIntoViewIfNeeded();
    await appCard.locator(".pcard-top").click({ force: true });
    await expect(appCard.locator(".pcard-detail")).toBeVisible({ timeout: 10_000 });

    const chainNode = appCard.locator(".chain .n.clickable").first();
    await expect(chainNode).toBeVisible();
    await chainNode.click();

    const detailSheet = page.locator(".sheet-detail");
    await expect(detailSheet).toBeVisible({ timeout: 15_000 });
    await expect(detailSheet.locator("#proc-detail-title")).toBeVisible();
    await expect(page.getByRole("button", { name: "关闭" })).toBeVisible();

    await page.getByRole("button", { name: "关闭" }).click();
    await expect(detailSheet).toBeHidden();
  });

  test("interaction: PID link opens process detail sheet", async ({ page }) => {
    await page.goto("/");

    const appCard = page.locator(".pcard").filter({ hasText: "procpane-app" });
    await expect(appCard).toBeVisible({ timeout: 30_000 });

    await appCard.scrollIntoViewIfNeeded();
    await appCard.locator(".pcard-top").click({ force: true });
    await expect(appCard.locator(".pcard-detail")).toBeVisible({ timeout: 10_000 });

    const pidLink = appCard.locator(".pid-link");
    await expect(pidLink).toBeVisible();
    const pidText = await pidLink.innerText();
    const pidMatch = pidText.match(/^PID (\d+)$/);
    expect(pidMatch).not.toBeNull();
    const pid = pidMatch![1];

    await pidLink.click();

    const detailSheet = page.locator(".sheet-detail");
    await expect(detailSheet).toBeVisible({ timeout: 15_000 });
    await expect(detailSheet.locator("#proc-detail-title")).toContainText(
      "procpane-app",
    );
    await expect(detailSheet.locator(".detail-sub")).toHaveText(`PID ${pid}`);
    await expect(
      detailSheet.locator(".kv .r").filter({ hasText: "cmdline" }),
    ).toBeVisible();

    await page.keyboard.press("Escape");
    await expect(detailSheet).toBeHidden();
  });

  test("interaction: proto chips toggle without error", async ({ page }) => {
    await page.goto("/");
    await selectSourceTab(page, "本机");
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

test.describe("procpane settings (server build, /settings.html)", () => {
  test("smoke: settings page loads Docker read-only and SSH host list", async ({
    page,
  }) => {
    await page.goto("/settings.html");

    await expect(page.getByRole("heading", { name: /procpane 设置/i })).toBeVisible({
      timeout: 30_000,
    });
    await expect(page.getByRole("heading", { name: "Docker", exact: true })).toBeVisible();
    await expect(page.getByText("Endpoint")).toBeVisible();
    await expect(page.getByRole("heading", { name: "SSH 主机", exact: true })).toBeVisible();
    await expect(page.getByText(/~\/\.ssh\/config/)).toBeVisible();
  });
});
