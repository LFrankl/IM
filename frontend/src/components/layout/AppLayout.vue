<script setup lang="ts">
import NavBar from './NavBar.vue'
</script>

<template>
  <div class="app-layout">
    <NavBar />
    <div class="app-content">
      <RouterView />
    </div>
  </div>
</template>

<style scoped>
.app-layout {
  display: flex;
  height: 100vh;
  overflow: hidden;
  min-width: 960px;
}

.app-content {
  flex: 1;
  display: flex;
  overflow: hidden;
}
</style>
