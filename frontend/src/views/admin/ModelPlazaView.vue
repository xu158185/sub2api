<template>
  <AppLayout>
    <div class="mx-auto max-w-7xl space-y-6 px-4 py-6 sm:px-6 lg:px-8">
      <div class="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">模型广场价格</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
            查看实际计费价格，并维护国模官网人民币原价。价格单位为元 / 百万 token。
          </p>
        </div>
        <button class="btn btn-secondary" :disabled="loading" title="刷新" @click="load">
          <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
        </button>
      </div>

      <ModelPlazaContent :response="response" :loading="loading" :error="loadError" embedded />

      <section class="rounded-2xl border border-gray-200 bg-white shadow-card dark:border-dark-700 dark:bg-dark-800/50">
        <div class="flex flex-wrap items-center justify-between gap-3 border-b border-gray-100 px-5 py-4 dark:border-dark-700/60">
          <div>
            <h2 class="font-semibold text-gray-900 dark:text-white">国模官网原价</h2>
            <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">保存后会覆盖公开模型广场中的官方参考价。</p>
          </div>
          <button class="btn btn-primary" :disabled="saving || loading" @click="save">
            <Icon name="check" size="sm" class="mr-2" />
            {{ saving ? '保存中...' : '保存价格' }}
          </button>
        </div>

        <div v-if="models.length === 0" class="px-5 py-10 text-center text-sm text-gray-500 dark:text-dark-400">
          暂无可维护的模型价格
        </div>
        <div v-else class="divide-y divide-gray-100 dark:divide-dark-700/60">
          <div v-for="platformGroup in platformGroups" :key="platformGroup.platform" class="overflow-x-auto">
            <div class="flex items-center gap-2 bg-gray-50 px-5 py-3 dark:bg-dark-800/70">
              <PlatformIcon :platform="platformGroup.platform as GroupPlatform" size="sm" :class="platformTextClass(platformGroup.platform)" />
              <span class="font-semibold" :class="platformTextClass(platformGroup.platform)">{{ platformGroup.label }}</span>
              <span class="text-xs text-gray-400 dark:text-dark-500">{{ platformGroup.models.length }}</span>
            </div>
          <table class="w-full min-w-[980px] text-sm">
            <thead class="bg-gray-50 text-left text-xs text-gray-500 dark:bg-dark-800 dark:text-dark-400">
              <tr>
                <th class="px-5 py-3">模型</th>
                <th v-for="field in priceFields" :key="field" class="px-3 py-3">{{ priceLabels[field] }}</th>
                <th class="px-3 py-3">来源</th>
                <th class="px-3 py-3">生效日期</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="model in platformGroup.models" :key="`${platformGroup.platform}:${model.name}`" class="border-t border-gray-100 dark:border-dark-700/60">
                <td class="px-5 py-3 align-top">
                  <div class="font-medium text-gray-900 dark:text-white">{{ model.name }}</div>
                  <div class="mt-1 text-xs text-gray-400">{{ model.platforms.join(', ') }}</div>
                </td>
                <td v-for="field in priceFields" :key="field" class="px-3 py-3 align-top">
                  <input v-model.number="draft[model.name][field]" type="number" min="0" step="0.000001" class="input w-32" :placeholder="'-'" />
                </td>
                <td class="px-3 py-3 align-top">
                  <input v-model="draft[model.name].source" type="text" class="input w-44" placeholder="官网链接或来源" />
                </td>
                <td class="px-3 py-3 align-top">
                  <input v-model="draft[model.name].effective_date" type="date" class="input w-36" />
                </td>
              </tr>
            </tbody>
          </table>
          </div>
        </div>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import ModelPlazaContent from '@/components/modelPlaza/ModelPlazaContent.vue'
import {
  getAdminModelPlaza,
  getModelPlazaDomesticPricing,
  updateModelPlazaDomesticPricing,
  type ModelPlazaDomesticPricing,
  type ModelPlazaResponse
} from '@/api/modelPlaza'
import { useAppStore } from '@/stores/app'
import type { GroupPlatform } from '@/types'
import { platformLabel, platformTextClass } from '@/utils/platformColors'

const appStore = useAppStore()
const response = ref<ModelPlazaResponse | null>(null)
const loading = ref(true)
const saving = ref(false)
const loadError = ref(false)
const draft = reactive<Record<string, ModelPlazaDomesticPricing>>({})
const priceFields = ['input_price', 'output_price', 'cache_write_price', 'cache_write_1h_price', 'cache_write_12h_price', 'cache_read_price'] as const
const priceLabels: Record<(typeof priceFields)[number], string> = {
  input_price: '输入',
  output_price: '输出',
  cache_write_price: '缓存写入',
  cache_write_1h_price: '缓存写入 1h',
  cache_write_12h_price: '缓存写入 12h',
  cache_read_price: '缓存读取'
}

const models = computed(() => {
  const seen = new Map<string, { name: string; platforms: string[] }>()
  for (const group of response.value?.groups ?? []) {
    for (const model of group.models) {
      const existing = seen.get(model.name)
      if (existing) {
        if (!existing.platforms.includes(model.platform)) existing.platforms.push(model.platform)
      } else {
        seen.set(model.name, { name: model.name, platforms: [model.platform] })
      }
    }
  }
  return [...seen.values()].sort((a, b) => a.name.localeCompare(b.name))
})

const platformGroups = computed(() => {
  const grouped = new Map<string, typeof models.value>()
  for (const model of models.value) {
    for (const platform of model.platforms) {
      const list = grouped.get(platform) ?? []
      list.push(model)
      grouped.set(platform, list)
    }
  }
  return [...grouped.entries()]
    .sort(([a], [b]) => platformLabel(a).localeCompare(platformLabel(b)))
    .map(([platform, groupedModels]) => ({
      platform,
      label: platformLabel(platform),
      models: groupedModels
    }))
})

function ensureDraft(model: string, value: ModelPlazaDomesticPricing = {}) {
  if (!draft[model]) draft[model] = { ...value }
}

async function load() {
  loading.value = true
  loadError.value = false
  try {
    const [plaza, prices] = await Promise.all([getAdminModelPlaza(), getModelPlazaDomesticPricing()])
    response.value = plaza
    for (const [name, value] of Object.entries(prices)) ensureDraft(name, value)
    for (const group of plaza.groups) for (const model of group.models) ensureDraft(model.name)
  } catch (error) {
    loadError.value = true
    appStore.showError('加载模型广场价格失败')
    console.error(error)
  } finally {
    loading.value = false
  }
}

async function save() {
  saving.value = true
  try {
    const payload: Record<string, ModelPlazaDomesticPricing> = {}
    for (const model of models.value) {
      const value = draft[model.name]
      if (!value || Object.values(value).every((item) => item == null || item === '')) continue
      payload[model.name] = { ...value }
    }
    await updateModelPlazaDomesticPricing(payload)
    appStore.showSuccess('国模官网原价已保存')
    await load()
  } catch (error) {
    appStore.showError('保存国模官网原价失败')
    console.error(error)
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>
