import { ref } from 'vue'

const cardUserId = ref<number | null>(null)

export function useUserCard() {
  function openCard(userId: number) {
    cardUserId.value = userId
  }
  function closeCard() {
    cardUserId.value = null
  }
  return { cardUserId, openCard, closeCard }
}
