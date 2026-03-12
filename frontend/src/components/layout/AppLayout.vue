<script setup lang="ts">
import NavBar from './NavBar.vue'
import UserCard from '@/components/common/UserCard.vue'
import { useUserCard } from '@/composables/useUserCard'

const { cardUserId, closeCard } = useUserCard()
</script>

<template>
  <div class="app-layout">
    <NavBar />
    <div class="app-content">
      <RouterView />
    </div>
  </div>
  <UserCard v-if="cardUserId" :user-id="cardUserId" @close="closeCard" />
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
