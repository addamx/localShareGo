import { createRouter, createWebHistory } from "vue-router";

import DesktopPage from "../pages/DesktopPage.vue";
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
        component: DesktopPage,
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
