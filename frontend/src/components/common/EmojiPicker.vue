<script setup lang="ts">
import { ref } from 'vue'

const emit = defineEmits<{ pick: [emoji: string] }>()

const activeTab = ref(0)

const categories = [
  {
    icon: '😊',
    emojis: [
      '😀','😃','😄','😁','😆','😅','🤣','😂','🙂','🙃','😉','😊','😇',
      '🥰','😍','🤩','😘','😋','😛','😜','🤪','😝','🤑','🤗','🤭','🤫',
      '🤔','😐','😑','😶','😏','😒','🙄','😬','😌','😔','😢','😭','😤',
      '😠','🤬','😈','💩','🤡','👻','💀','😎','🥳','🤓','🧐','🥺','😳',
    ],
  },
  {
    icon: '👋',
    emojis: [
      '👋','🤚','✋','🖐','🖖','👌','✌️','🤞','🤟','🤘','🤙','👈','👉',
      '👆','👇','☝️','👍','👎','✊','👊','🤛','🤜','👏','🙌','🤝','🙏',
      '💪','🫶','🤲','✍️','💅','🤳',
    ],
  },
  {
    icon: '❤️',
    emojis: [
      '❤️','🧡','💛','💚','💙','💜','🖤','🤍','🤎','💔','❣️','💕','💞',
      '💓','💗','💖','💘','💝','🔥','⭐','✨','🌟','💫','💯','🎵','🎶',
      '👑','💎','🎯','🏅','🆗','🆒','🆕','💡','🔔','📣',
    ],
  },
  {
    icon: '🎉',
    emojis: [
      '🎉','🎊','🎈','🎁','🎄','🎃','🎆','🎇','🧨','🏆','🥇','🥈','🥉',
      '🎭','🎨','🎬','🎤','🎧','🎹','🎸','🎺','🎻','🥁','🎷','🍕','🍔',
      '🍟','🌮','🍦','🍰','🎂','🍫','🍬','☕','🍺','🥂','🍾',
    ],
  },
  {
    icon: '🐶',
    emojis: [
      '🐶','🐱','🐭','🐹','🐰','🦊','🐻','🐼','🐨','🐯','🦁','🐮','🐷',
      '🐸','🐵','🙈','🙉','🙊','🐔','🐧','🦆','🦉','🐺','🐴','🦄','🐝',
      '🦋','🐌','🐞','🐜','🐢','🦎','🐙','🐬','🐳','🦈','🌸','🌺','🌻',
      '🍀','🌈','⛅','❄️','🌊','🌙',
    ],
  },
]
</script>

<template>
  <div class="emoji-picker">
    <div class="emoji-tabs">
      <button
        v-for="(cat, i) in categories"
        :key="i"
        class="emoji-tab"
        :class="{ active: activeTab === i }"
        @mousedown.prevent
        @click="activeTab = i"
      >{{ cat.icon }}</button>
    </div>
    <div class="emoji-grid">
      <button
        v-for="emoji in categories[activeTab].emojis"
        :key="emoji"
        class="emoji-btn"
        @mousedown.prevent
        @click="emit('pick', emoji)"
      >{{ emoji }}</button>
    </div>
  </div>
</template>

<style scoped>
.emoji-picker {
  position: absolute;
  bottom: calc(100% + 6px);
  left: 0;
  width: 284px;
  background: var(--bg-surface);
  border: 1px solid var(--border-light);
  border-radius: 10px;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.14);
  z-index: 300;
  overflow: hidden;
  user-select: none;
}

.emoji-tabs {
  display: flex;
  border-bottom: 1px solid var(--border-light);
  padding: 4px 6px 0;
  gap: 2px;
}

.emoji-tab {
  flex: 1;
  height: 32px;
  font-size: 16px;
  background: none;
  border: none;
  border-radius: 6px 6px 0 0;
  cursor: pointer;
  transition: background 0.1s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.emoji-tab:hover { background: var(--bg-hover); }

.emoji-tab.active {
  background: var(--bg-hover);
  box-shadow: inset 0 -2px 0 var(--color-primary);
}

.emoji-grid {
  display: grid;
  grid-template-columns: repeat(8, 1fr);
  gap: 1px;
  padding: 6px;
  max-height: 200px;
  overflow-y: auto;
}

.emoji-btn {
  width: 30px;
  height: 30px;
  font-size: 19px;
  background: none;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.1s;
  line-height: 1;
}

.emoji-btn:hover { background: var(--bg-hover); }
</style>
