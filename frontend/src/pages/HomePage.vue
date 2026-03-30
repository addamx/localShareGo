<template>
  <div class="page-shell page-shell--center">
    <section class="panel-surface poster-card">
      <p class="eyebrow">Route Gateway</p>
      <h1>LocalShareGo</h1>
      <p class="lead">
        NaiveDesktop 现在使用 <strong>/desktop</strong> 路由，Web 入口使用
        <strong>/web</strong> 路由。
      </p>
      <p class="muted-copy">
        浏览器请从 NaiveDesktop 生成的访问链接进入，通常形式为
        <code>/web?token=...</code>。
      </p>
    </section>
  </div>
</template>

<style scoped>
.poster-card {
  max-width: 720px;
  padding: 3rem;
  text-align: center;
}

.poster-card h1 {
  margin: 0;
  font-size: clamp(2.5rem, 5vw, 4.5rem);
  line-height: 0.95;
}

.lead {
  margin: 1.25rem 0 0;
  font-size: 1.05rem;
  line-height: 1.7;
}

.muted-copy {
  margin: 1rem auto 0;
  max-width: 36rem;
  line-height: 1.7;
}

code {
  padding: 0.12rem 0.45rem;
  border-radius: 999px;
  background: rgba(24, 37, 31, 0.08);
  font-family: "Cascadia Code", "SFMono-Regular", Consolas, monospace;
}
</style>
