import { createApp } from "vue";

import { isDesktopRuntime } from "./env";
import { createWorkbenchRouter } from "./router";
import AppShell from "./AppShell.vue";

export async function bootstrapApp() {
  const router = createWorkbenchRouter();

  if (isDesktopRuntime && !router.currentRoute.value.path.startsWith("/desktop")) {
    await router.replace("/desktop");
  }

  const app = createApp(AppShell);
  app.use(router);

  await router.isReady();
  app.mount("#app");
}
