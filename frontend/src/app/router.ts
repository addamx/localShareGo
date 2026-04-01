import { createRouter, createWebHistory } from "vue-router";

import HomePage from "../pages/HomePage.vue";
import WebPage from "../pages/WebPage.vue";

export function createWorkbenchRouter() {
  return createRouter({
    history: createWebHistory(),
    routes: [
      {
        path: "/",
        name: "home",
        component: HomePage,
      },
      {
        path: "/desktop",
        name: "desktop",
        component: () => import("../pages/DesktopPage.vue"),
      },
      {
        path: "/desktop/settings",
        name: "desktop-settings",
        component: () => import("../pages/DesktopSettingsPage.vue"),
      },
      {
        path: "/web",
        name: "web",
        component: WebPage,
        alias: ["/web/:pathMatch(.*)*"],
      },
      {
        path: "/:pathMatch(.*)*",
        redirect: "/",
      },
    ],
  });
}
