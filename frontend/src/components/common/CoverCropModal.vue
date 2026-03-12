<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

const props = defineProps<{ file: File }>()
const emit = defineEmits<{
  confirm: [blob: Blob]
  cancel: []
}>()

// 目标宽高比（封面区域 680:290，参考微信朋友圈）
const ASPECT = 680 / 290

const canvasRef = ref<HTMLCanvasElement | null>(null)
const containerRef = ref<HTMLDivElement | null>(null)

// 原始图像
const img = new Image()
const imgUrl = URL.createObjectURL(props.file)

// 图像在容器内的绘制区域（像素）
let drawX = 0
let drawY = 0
let drawW = 0
let drawH = 0

// 裁剪框位置（相对于容器）
const cropX = ref(0)
const cropY = ref(0)
const cropW = ref(0)
const cropH = ref(0)

// 拖拽状态
let dragging = false
let dragStartX = 0
let dragStartY = 0
let dragStartCropX = 0
let dragStartCropY = 0

function initCrop() {
  const container = containerRef.value
  if (!container) return
  const cW = container.clientWidth
  const cH = container.clientHeight

  // 计算图像在容器中的绘制区域（contain 方式）
  const imgRatio = img.naturalWidth / img.naturalHeight
  if (imgRatio > cW / cH) {
    drawW = cW
    drawH = cW / imgRatio
    drawX = 0
    drawY = (cH - drawH) / 2
  } else {
    drawH = cH
    drawW = cH * imgRatio
    drawX = (cW - drawW) / 2
    drawY = 0
  }

  // 初始裁剪框：在图像区域内尽可能大，保持目标比例
  let cropWVal = drawW
  let cropHVal = drawW / ASPECT
  if (cropHVal > drawH) {
    cropHVal = drawH
    cropWVal = drawH * ASPECT
  }
  cropW.value = cropWVal
  cropH.value = cropHVal
  cropX.value = drawX + (drawW - cropWVal) / 2
  cropY.value = drawY + (drawH - cropHVal) / 2

  renderCanvas()
}

function renderCanvas() {
  const canvas = canvasRef.value
  const container = containerRef.value
  if (!canvas || !container) return
  const cW = container.clientWidth
  const cH = container.clientHeight
  canvas.width = cW
  canvas.height = cH
  const ctx = canvas.getContext('2d')!

  // 绘制图像
  ctx.drawImage(img, drawX, drawY, drawW, drawH)

  // 遮罩（裁剪框外半透明黑色）
  ctx.save()
  ctx.fillStyle = 'rgba(0,0,0,0.5)'
  ctx.fillRect(0, 0, cW, cH)
  ctx.clearRect(cropX.value, cropY.value, cropW.value, cropH.value)
  ctx.restore()

  // 重新绘制裁剪框内的图像（使得框内清晰）
  ctx.save()
  ctx.beginPath()
  ctx.rect(cropX.value, cropY.value, cropW.value, cropH.value)
  ctx.clip()
  ctx.drawImage(img, drawX, drawY, drawW, drawH)
  ctx.restore()

  // 裁剪框边框
  ctx.strokeStyle = 'white'
  ctx.lineWidth = 2
  ctx.strokeRect(cropX.value, cropY.value, cropW.value, cropH.value)

  // 网格线（三等分）
  ctx.strokeStyle = 'rgba(255,255,255,0.4)'
  ctx.lineWidth = 1
  for (let i = 1; i < 3; i++) {
    const x = cropX.value + (cropW.value / 3) * i
    const y = cropY.value + (cropH.value / 3) * i
    ctx.beginPath(); ctx.moveTo(x, cropY.value); ctx.lineTo(x, cropY.value + cropH.value); ctx.stroke()
    ctx.beginPath(); ctx.moveTo(cropX.value, y); ctx.lineTo(cropX.value + cropW.value, y); ctx.stroke()
  }
}

function clampCrop(x: number, y: number) {
  const minX = drawX
  const maxX = drawX + drawW - cropW.value
  const minY = drawY
  const maxY = drawY + drawH - cropH.value
  return {
    x: Math.max(minX, Math.min(maxX, x)),
    y: Math.max(minY, Math.min(maxY, y)),
  }
}

function onMouseDown(e: MouseEvent) {
  const rect = canvasRef.value!.getBoundingClientRect()
  const mx = e.clientX - rect.left
  const my = e.clientY - rect.top
  if (
    mx >= cropX.value && mx <= cropX.value + cropW.value &&
    my >= cropY.value && my <= cropY.value + cropH.value
  ) {
    dragging = true
    dragStartX = mx
    dragStartY = my
    dragStartCropX = cropX.value
    dragStartCropY = cropY.value
  }
}

function onMouseMove(e: MouseEvent) {
  if (!dragging) return
  const rect = canvasRef.value!.getBoundingClientRect()
  const mx = e.clientX - rect.left
  const my = e.clientY - rect.top
  const dx = mx - dragStartX
  const dy = my - dragStartY
  const { x, y } = clampCrop(dragStartCropX + dx, dragStartCropY + dy)
  cropX.value = x
  cropY.value = y
  renderCanvas()
}

function onMouseUp() {
  dragging = false
}

// 触摸支持
let touchStartX = 0
let touchStartY = 0

function onTouchStart(e: TouchEvent) {
  const touch = e.touches[0]
  const rect = canvasRef.value!.getBoundingClientRect()
  const mx = touch.clientX - rect.left
  const my = touch.clientY - rect.top
  if (
    mx >= cropX.value && mx <= cropX.value + cropW.value &&
    my >= cropY.value && my <= cropY.value + cropH.value
  ) {
    dragging = true
    touchStartX = mx
    touchStartY = my
    dragStartCropX = cropX.value
    dragStartCropY = cropY.value
  }
}

function onTouchMove(e: TouchEvent) {
  if (!dragging) return
  e.preventDefault()
  const touch = e.touches[0]
  const rect = canvasRef.value!.getBoundingClientRect()
  const mx = touch.clientX - rect.left
  const my = touch.clientY - rect.top
  const dx = mx - touchStartX
  const dy = my - touchStartY
  const { x, y } = clampCrop(dragStartCropX + dx, dragStartCropY + dy)
  cropX.value = x
  cropY.value = y
  renderCanvas()
}

function onConfirm() {
  // 将裁剪框坐标映射回原始图像像素
  const scaleX = img.naturalWidth / drawW
  const scaleY = img.naturalHeight / drawH
  const sx = (cropX.value - drawX) * scaleX
  const sy = (cropY.value - drawY) * scaleY
  const sw = cropW.value * scaleX
  const sh = cropH.value * scaleY

  const offscreen = document.createElement('canvas')
  offscreen.width = 680
  offscreen.height = 290
  const ctx = offscreen.getContext('2d')!
  ctx.drawImage(img, sx, sy, sw, sh, 0, 0, 680, 290)
  offscreen.toBlob((blob) => {
    if (blob) emit('confirm', blob)
  }, 'image/jpeg', 0.92)
}

onMounted(() => {
  img.onload = () => initCrop()
  img.src = imgUrl
  window.addEventListener('mousemove', onMouseMove)
  window.addEventListener('mouseup', onMouseUp)
})

onUnmounted(() => {
  URL.revokeObjectURL(imgUrl)
  window.removeEventListener('mousemove', onMouseMove)
  window.removeEventListener('mouseup', onMouseUp)
})
</script>

<template>
  <Teleport to="body">
    <div class="crop-mask">
      <div class="crop-dialog">
        <div class="crop-title">裁剪封面</div>
        <div class="crop-tip">拖动选框选择显示区域</div>
        <div ref="containerRef" class="crop-canvas-wrap">
          <canvas
            ref="canvasRef"
            class="crop-canvas"
            @mousedown="onMouseDown"
            @touchstart.passive="onTouchStart"
            @touchmove="onTouchMove"
          />
        </div>
        <div class="crop-actions">
          <button class="crop-btn-cancel" @click="emit('cancel')">取消</button>
          <button class="crop-btn-confirm" @click="onConfirm">确认裁剪</button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.crop-mask {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
  z-index: 2000;
  display: flex;
  align-items: center;
  justify-content: center;
}

.crop-dialog {
  background: #fff;
  border-radius: 12px;
  padding: 24px;
  width: min(680px, 96vw);
  display: flex;
  flex-direction: column;
  gap: 14px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.18);
}

.crop-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.crop-tip {
  font-size: 12px;
  color: var(--text-tertiary);
  margin-top: -6px;
}

.crop-canvas-wrap {
  width: 100%;
  aspect-ratio: 680 / 400;
  border-radius: 8px;
  overflow: hidden;
  background: #f0f2f5;
  position: relative;
}

.crop-canvas {
  width: 100%;
  height: 100%;
  display: block;
  cursor: move;
}

.crop-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding-top: 4px;
  border-top: 1px solid var(--border-light);
}

.crop-btn-cancel {
  padding: 7px 20px;
  border-radius: 6px;
  border: 1px solid var(--border-normal);
  background: #fff;
  color: var(--text-secondary);
  font-size: 13px;
  cursor: pointer;
}
.crop-btn-cancel:hover { background: var(--bg-hover); }

.crop-btn-confirm {
  padding: 7px 20px;
  border-radius: 6px;
  border: none;
  background: var(--qq-blue-primary);
  color: #fff;
  font-size: 13px;
  cursor: pointer;
}
.crop-btn-confirm:hover { opacity: 0.88; }
</style>
