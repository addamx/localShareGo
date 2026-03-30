<template>
  <n-config-provider :theme-overrides="themeOverrides">
    <n-dialog-provider>
      <n-message-provider placement="top-right" :max="4">
        <router-view />
      </n-message-provider>
    </n-dialog-provider>
  </n-config-provider>
</template>

<script setup lang="ts">
import {
  NConfigProvider,
  NDialogProvider,
  NMessageProvider,
  type GlobalThemeOverrides,
} from "naive-ui";

const themeOverrides: GlobalThemeOverrides = {
  common: {
    primaryColor: "#1f7a5a",
    primaryColorHover: "#2b8b69",
    primaryColorPressed: "#176348",
    primaryColorSuppl: "#2b8b69",
    successColor: "#1f7a5a",
    warningColor: "#d08a24",
    errorColor: "#d65f5f",
    borderRadius: "18px",
    fontFamily: '"Nunito", "Segoe UI", "PingFang SC", "Microsoft YaHei", sans-serif',
    bodyColor: "#fffdfa",
    cardColor: "#fffcf7",
    modalColor: "#fffcf7",
    popoverColor: "#fffaf1",
  },
  Button: {
    borderRadiusMedium: "14px",
  },
  Input: {
    borderRadius: "14px",
  },
  Card: {
    borderRadius: "24px",
  },
};
</script>
