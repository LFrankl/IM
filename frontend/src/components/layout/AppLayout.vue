<script setup lang="ts">
import { computed, watch } from 'vue'
import { useRoute } from 'vue-router'
import NavBar from './NavBar.vue'
import UserCard from '@/components/common/UserCard.vue'
import { useUserCard } from '@/composables/useUserCard'

const { cardUserId, closeCard } = useUserCard()
const route = useRoute()

// Close any open UserCard whenever the route changes (covers direct nav via NavBar etc.)
watch(() => route.path, () => closeCard())

const showCard = computed(() => cardUserId.value !== null && cardUserId.value > 0)
const activeUserId = computed(() => cardUserId.value ?? 0)
</script>

<template>
  <div class="app-layout">
    <NavBar />
    <div class="app-content">
      <RouterView />
    </div>
  </div>
  <UserCard v-if="showCard" :user-id="activeUserId" @close="closeCard" />
</template>

<style scoped>
.app-layout {
  display: flex;
  height: 100vh;
  overflow: hidden;
}

.app-content {
  flex: 1;
  display: flex;
  overflow: hidden;
}
</style>
