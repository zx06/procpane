import { test, expect, type Page } from "@playwright/test";

async function selectSourceTab(
  page: Page,
  label: "本机" | "Docker" | "SSH",
) {
  const nav = page.getByRole("navigation", { name: "数据源" });
  await expect(nav).toBeVisible();
  await nav.getByRole("button", { name: label, exact: true }).click();
}

function activeTabPanel(page: Page) {
  return page.locator(".tab-panel-wrap:not(.hidden) .tab-panel");
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
    await expect(page.getByText("已发布端口（只读，无法结束容器进程）")).toBeVisible();

    const hintDisconnected = page.getByText(/Docker 未连接/);
    const hintOk = page.getByText(/^Engine · /);
    const emptyPublished = page.getByText("没有已发布的容器端口");
    const dockerPanel = activeTabPanel(page);
    const dockerCards = dockerPanel.locator(".cards-scroll .docker-card");
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
    await expect(
      page.getByText("远程主机 LISTEN（只读，无法结束进程）"),
    ).toBeVisible();

    const empty = page.getByText("尚未启用任何 SSH 主机");
    const settingsCta = page.getByRole("button", {
      name: "去设置启用 SSH 主机",
    });
    const sshPanel = page.locator('.tab-panel-wrap[data-tab="ssh"] .tab-panel');
    const hostChip = sshPanel.locator(".host-chips .host-chip").first();
    const sshEmptyList = sshPanel.getByText("没有 LISTEN 端口");
    const sshCards = sshPanel.locator(".cards-scroll .pcard");

    await expect
      .poll(
        async () => {
          if (await empty.isVisible().catch(() => false)) return "onboarding";
          if (await hostChip.isVisible().catch(() => false)) return "hosts";
          return null;
        },
        { timeout: 30_000 },
      )
      .not.toBeNull();

    if (await empty.isVisible().catch(() => false)) {
      await expect(settingsCta).toBeVisible();
    } else {
      await expect(hostChip).toBeVisible();
      await expect
        .poll(
          async () =>
            (await sshEmptyList.isVisible().catch(() => false)) ||
            (await sshCards.first().isVisible().catch(() => false)),
          { timeout: 30_000 },
        )
        .toBe(true);
    }
  });

  test("smoke: shell, filters, and card stack pane", async ({ page }) => {
    await page.goto("/");

    await expect(page.getByRole("heading", { name: /procpane/i })).toBeVisible();
    await expect(page.getByText("本地 LISTEN 端口")).toBeVisible();
    const panel = activeTabPanel(page);
    await expect(panel.getByPlaceholder("端口 / 进程…")).toBeVisible();
    await expect(panel.getByRole("button", { name: "全部" })).toBeVisible();
    await expect(panel.getByRole("button", { name: "TCP" })).toBeVisible();
    await expect(panel.getByRole("button", { name: "UDP" })).toBeVisible();
    await expect(panel.getByRole("button", { name: "刷新" })).toBeVisible();
    await expect(panel.locator(".cards-scroll")).toBeVisible();
    await expect(panel.locator(".count-pill")).toBeVisible();
    await expect(page.getByRole("button", { name: "退出" })).toBeVisible();

    await page.screenshot({
      path: "test-results/procpane-smoke.png",
      fullPage: true,
    });
  });

  test("service: listener cards or empty state after load", async ({ page }) => {
    await page.goto("/");

    const panel = activeTabPanel(page);
    const cards = panel.locator(".cards-scroll .pcard");
    const empty = panel.getByText(/当前没有 LISTEN 端口|没有匹配/);
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
    const panel = activeTabPanel(page);
    await expect(panel.locator(".cards-scroll")).toBeVisible();

    for (const label of ["TCP", "UDP", "全部"] as const) {
      await panel.getByRole("button", { name: label }).click();
      await expect(page.getByRole("heading", { name: /procpane/i })).toBeVisible();
      const cards = panel.locator(".cards-scroll .pcard");
      const empty = panel.getByText(/当前没有 LISTEN 端口|没有匹配/);
      await expect(cards.first().or(empty)).toBeVisible({ timeout: 30_000 });
    }
  });

  test("tabs: keep search when switching away and back", async ({ page }) => {
    await page.goto("/");
    await selectSourceTab(page, "SSH");
    const sshPanel = page.locator('.tab-panel-wrap[data-tab="ssh"] .tab-panel');
    const sshSearch = sshPanel.getByPlaceholder("端口 / 进程…");
    await sshSearch.fill("fixture-keepalive-query");
    await selectSourceTab(page, "Docker");
    await selectSourceTab(page, "SSH");
    await expect(sshSearch).toHaveValue("fixture-keepalive-query");
  });

  test("docker tab: cards use read-only styling", async ({ page }) => {
    await page.goto("/");
    await selectSourceTab(page, "Docker");
    const dockerPanel = activeTabPanel(page);
    await expect(
      dockerPanel.getByText("已发布端口（只读，无法结束容器进程）"),
    ).toBeVisible();
    const card = dockerPanel.locator(".docker-card.pcard-readonly").first();
    const settled = dockerPanel
      .getByText("没有已发布的容器端口")
      .or(dockerPanel.getByText(/无法连接 Docker|加载 Docker|DOCKER_HOST/));
    await expect(card.or(settled).first()).toBeVisible({ timeout: 30_000 });
    if ((await card.count()) > 0) {
      await expect(card).toHaveClass(/pcard-readonly/);
    }
  });

  test("visibility: manual refresh after hidden stops auto poll", async ({
    page,
  }) => {
    await page.goto("/");
    const panel = activeTabPanel(page);
    await expect(panel.getByRole("button", { name: "刷新" })).toBeVisible();

    await page.evaluate(() => {
      Object.defineProperty(document, "hidden", {
        configurable: true,
        get: () => true,
      });
      document.dispatchEvent(new Event("visibilitychange"));
    });

    await panel.getByRole("button", { name: "刷新" }).click();
    await expect(panel.getByText(/上次更新 \d{2}:\d{2}:\d{2}/)).toBeVisible({
      timeout: 30_000,
    });
  });

  test("stale-while-revalidate: cards stay visible while refresh spins", async ({
    page,
  }) => {
    await page.goto("/");

    const panel = activeTabPanel(page);
    const cards = panel.locator(".cards-scroll .pcard");
    const empty = panel.getByText(/当前没有 LISTEN 端口|没有匹配/);
    await expect(cards.first().or(empty)).toBeVisible({ timeout: 30_000 });

    const countBefore = await cards.count();
    if (countBefore === 0) {
      test.skip(true, "no listener cards in this environment");
    }

    const footerTime = panel.locator("footer.foot .live").getByText(/上次更新/);
    await expect(footerTime).toBeVisible({ timeout: 30_000 });

    const refreshBtn = panel.getByRole("button", { name: "刷新" });
    await refreshBtn.click();

    await expect(refreshBtn).toHaveClass(/spin/, { timeout: 5_000 });
    await expect(footerTime).toHaveText(/上次更新 \d{2}:\d{2}:\d{2}/);
    await expect(footerTime).not.toHaveText(/刷新中/);
    expect(await cards.count()).toBe(countBefore);

    await expect(refreshBtn).not.toHaveClass(/spin/, { timeout: 30_000 });
    expect(await cards.count()).toBeGreaterThan(0);
  });
});

test.describe("procpane settings (server build, /settings.html)", () => {
  test("smoke: settings page loads Docker context select and SSH host list", async ({
    page,
  }) => {
    await page.goto("/settings.html");

    await expect(page.getByRole("heading", { name: /procpane 设置/i })).toBeVisible({
      timeout: 30_000,
    });
    await expect(page.getByRole("heading", { name: "常规", exact: true })).toBeVisible();
    await expect(page.getByText("开机自启", { exact: true })).toBeVisible();
    await expect(page.getByRole("heading", { name: "自动刷新", exact: true })).toBeVisible();
    await expect(page.getByRole("heading", { name: "Docker", exact: true })).toBeVisible();
    await expect(page.locator("select.ctx-select").first()).toBeVisible();
    await expect(page.locator("select.ctx-select")).toHaveCount(2);
    await expect(page.getByText("Endpoint")).toBeVisible();
    await expect(page.getByRole("heading", { name: "SSH 主机", exact: true })).toBeVisible();
    await expect(page.getByText(/~\/\.ssh\/config/)).toBeVisible();
  });

  test("settings scroll root allows overflow (matches fixed settings window)", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 420, height: 480 });
    await page.goto("/settings.html");

    await expect(page.getByRole("heading", { name: /procpane 设置/i })).toBeVisible({
      timeout: 30_000,
    });

    const scrollRoot = page.locator(".settings-app");
    await expect(scrollRoot).toBeVisible();

    const overflowY = await scrollRoot.evaluate(
      (el) => getComputedStyle(el).overflowY,
    );
    expect(overflowY).toBe("auto");

    await page.evaluate(() => {
      const root = document.querySelector(".settings-app");
      if (!root) throw new Error("missing .settings-app");
      let ul = root.querySelector(".host-list");
      if (!ul) {
        ul = document.createElement("ul");
        ul.className = "host-list";
        root.appendChild(ul);
      }
      for (let i = 0; i < 40; i++) {
        const li = document.createElement("li");
        li.textContent = `scroll-fixture-${i}`;
        li.style.padding = "12px 0";
        ul.appendChild(li);
      }
    });

    const scrolled = await scrollRoot.evaluate((el) => {
      const before = el.scrollTop;
      el.scrollTop = el.scrollHeight;
      return el.scrollTop > before;
    });
    expect(scrolled).toBe(true);
  });
});

test.describe("procpane conn graph (server build, /conn-graph.html)", () => {
  test("smoke: connection graph page loads toolbar and table", async ({ page }) => {
    await page.setViewportSize({ width: 960, height: 640 });
    await page.goto("/conn-graph.html");

    await expect(page.getByPlaceholder("进程名 / PID / 端口")).toBeVisible({
      timeout: 30_000,
    });
    await expect(page.getByRole("button", { name: "只看本地互联" })).toBeVisible();
    await expect(page.getByRole("button", { name: "隐藏浏览器类" })).toHaveClass(/on/);
    const groupSelect = page.getByLabel("本地互联组");
    if (await groupSelect.count()) {
      await expect(groupSelect).toBeVisible();
    }
    await expect(page.getByText("TCP ESTABLISHED · 本机")).toBeVisible();
    await expect(page.locator("table.cg-table")).toBeVisible();
    await expect(page.locator(".svelte-flow")).toBeVisible();
  });
});
