<template>
  <div class="space-y-6">
    <PageHeader
      eyebrow="Administration"
      title="Namespaces"
      subtitle="Manage tenant isolation and security policies"
    >
      <template #actions>
        <UButton icon="i-heroicons-plus" @click="openCreateModal">Create Namespace</UButton>
      </template>
    </PageHeader>

    <UCard>
      <UInput v-model="search" icon="i-heroicons-magnifying-glass" placeholder="Search namespaces..." class="max-w-xs mb-4" @input="debouncedFetch" />

      <UTable :rows="namespaces" :columns="columns" :loading="loading">
        <template #config-data="{ row }">
          <div class="flex flex-wrap gap-1">
            <UBadge size="xs" :color="row.config?.mfa_policy === 'MFA_POLICY_REQUIRED' ? 'orange' : row.config?.mfa_policy === 'MFA_POLICY_OPTIONAL' ? 'blue' : 'gray'" variant="soft">MFA {{ {MFA_POLICY_REQUIRED: 'required', MFA_POLICY_OPTIONAL: 'optional', MFA_POLICY_DISABLED: 'disabled'}[row.config?.mfa_policy] || 'disabled' }}</UBadge>
            <template v-if="row.config?.password_policy">
              <UBadge size="xs" color="primary" variant="soft">min {{ row.config.password_policy.min_length || 0 }} chars</UBadge>
              <UBadge v-if="row.config.password_policy.require_uppercase" size="xs" color="primary" variant="soft">A-Z</UBadge>
              <UBadge v-if="row.config.password_policy.require_lowercase" size="xs" color="primary" variant="soft">a-z</UBadge>
              <UBadge v-if="row.config.password_policy.require_number" size="xs" color="primary" variant="soft">0-9</UBadge>
              <UBadge v-if="row.config.password_policy.require_special" size="xs" color="primary" variant="soft">!@#</UBadge>
              <UBadge v-if="row.config.password_policy.password_history" size="xs" color="primary" variant="soft">history {{ row.config.password_policy.password_history }}</UBadge>
            </template>
            <UBadge v-for="p in row.config?.allowed_social_providers" :key="p" size="xs" color="blue" variant="soft">{{ p }}</UBadge>
            <UBadge v-if="row.config?.ip_allowlist?.length" size="xs" color="green" variant="soft">IP allowlist</UBadge>
            <UBadge v-if="row.config?.ip_denylist?.length" size="xs" color="red" variant="soft">IP denylist</UBadge>
            <UBadge v-if="row.config?.webhook_url" size="xs" color="purple" variant="soft">Webhook</UBadge>
          </div>
        </template>
        <template #actions-data="{ row }">
          <div class="flex gap-1">
            <UTooltip text="Edit config">
              <UButton size="xs" variant="ghost" icon="i-heroicons-pencil-square" :aria-label="`Edit ${row.name}`" @click="openEditModal(row)" />
            </UTooltip>
            <UTooltip text="Delete">
              <UButton size="xs" variant="ghost" color="red" icon="i-heroicons-trash" :aria-label="`Delete ${row.name}`" @click="handleDelete(row.id)" />
            </UTooltip>
          </div>
        </template>
      </UTable>
    </UCard>

    <!-- Create/Edit Modal -->
    <UModal v-model="isModalOpen">
      <UCard>
        <template #header>
          <div class="flex items-center justify-between">
            <h3 class="text-base font-semibold">{{ editingId ? 'Edit Namespace Config' : 'New Namespace' }}</h3>
            <UButton color="gray" variant="ghost" icon="i-heroicons-x-mark-20-solid" @click="isModalOpen = false" />
          </div>
        </template>

        <form @submit.prevent="editingId ? handleUpdateConfig() : handleCreate()" class="space-y-4 py-2">
          <UFormGroup label="Name" v-if="!editingId">
            <UInput v-model="form.name" placeholder="e.g. acme-corp" />
          </UFormGroup>
          <UFormGroup label="Security">
            <USelectMenu v-model="form.config.mfa_policy" :options="[
              { label: 'Disabled', value: 'MFA_POLICY_DISABLED' },
              { label: 'Optional (challenge enrolled users)', value: 'MFA_POLICY_OPTIONAL' },
              { label: 'Required (challenge all users)', value: 'MFA_POLICY_REQUIRED' },
            ]" value-attribute="value" option-attribute="label" />
          </UFormGroup>
          <UFormGroup label="Social Providers">
            <USelectMenu v-model="form.config.allowed_social_providers" multiple :options="['google', 'github']" />
          </UFormGroup>
          <UFormGroup label="Password Policy">
            <div class="space-y-2">
              <UInput v-model.number="form.config.password_policy.min_length" type="number" placeholder="Min length" />
              <UCheckbox v-model="form.config.password_policy.require_uppercase" label="Require uppercase" />
              <UCheckbox v-model="form.config.password_policy.require_lowercase" label="Require lowercase" />
              <UCheckbox v-model="form.config.password_policy.require_number" label="Require number" />
              <UCheckbox v-model="form.config.password_policy.require_special" label="Require special character" />
              <UInput v-model.number="form.config.password_policy.password_history" type="number" placeholder="Password history count" />
            </div>
          </UFormGroup>
          <UFormGroup label="Notification">
            <div class="space-y-4">
              <div class="grid grid-cols-2 gap-4">
                <div>
                  <label class="block text-xs font-medium text-slate-500 mb-1">Email Provider</label>
                  <USelectMenu
                    v-model="form.config.notification.email_provider"
                    :options="emailProviderOptions"
                    value-attribute="value"
                    option-attribute="label"
                  />
                </div>
                <div>
                  <label class="block text-xs font-medium text-slate-500 mb-1">SMS Provider</label>
                  <USelectMenu
                    v-model="form.config.notification.sms_provider"
                    :options="smsProviderOptions"
                    value-attribute="value"
                    option-attribute="label"
                  />
                </div>
              </div>

              <div>
                <div class="flex items-center justify-between mb-2">
                  <label class="text-xs font-medium text-slate-500">Template Overrides</label>
                  <UDropdown :items="addableTemplateItems" :popper="{ placement: 'bottom-end' }">
                    <UButton size="xs" variant="ghost" icon="i-heroicons-plus" label="Add override" :disabled="addableTemplateItems[0].length === 0" />
                  </UDropdown>
                </div>

                <div v-for="(override, name) in form.config.notification.email_templates" :key="'email-' + name" class="border border-slate-200 dark:border-slate-700 rounded-lg mb-3">
                  <div class="flex items-center justify-between px-3 py-2 bg-slate-50 dark:bg-slate-800 border-b border-slate-200 dark:border-slate-700 rounded-t-lg">
                    <div class="flex items-center gap-2">
                      <UBadge size="xs" color="blue" variant="soft">email</UBadge>
                      <span class="text-sm font-mono font-medium">{{ name }}</span>
                    </div>
                    <UButton size="xs" variant="ghost" color="red" @click="delete form.config.notification.email_templates[name]">Remove</UButton>
                  </div>
                  <div class="p-3 space-y-3">
                    <div>
                      <label class="block text-xs text-slate-500 mb-1">Subject</label>
                      <UInput v-model="override.subject" placeholder="Leave empty to use default" class="font-mono text-sm" />
                    </div>
                    <div>
                      <label class="block text-xs text-slate-500 mb-1">HTML Body</label>
                      <UTextarea v-model="override.html_body" placeholder="Leave empty to use default" class="font-mono text-sm" :rows="4" />
                      <p class="text-xs text-slate-400 mt-1" v-pre>Variables: <code class="bg-slate-100 dark:bg-slate-800 px-1 rounded">{{.Code}}</code> <code class="bg-slate-100 dark:bg-slate-800 px-1 rounded">{{.TTLMinutes}}</code> <code class="bg-slate-100 dark:bg-slate-800 px-1 rounded">{{.AppName}}</code></p>
                    </div>
                    <div>
                      <label class="block text-xs text-slate-500 mb-1">Text Body</label>
                      <UTextarea v-model="override.text_body" placeholder="Leave empty to use default" class="font-mono text-sm" :rows="2" />
                    </div>
                  </div>
                </div>

                <div v-for="(override, name) in form.config.notification.sms_templates" :key="'sms-' + name" class="border border-slate-200 dark:border-slate-700 rounded-lg mb-3">
                  <div class="flex items-center justify-between px-3 py-2 bg-slate-50 dark:bg-slate-800 border-b border-slate-200 dark:border-slate-700 rounded-t-lg">
                    <div class="flex items-center gap-2">
                      <UBadge size="xs" color="amber" variant="soft">sms</UBadge>
                      <span class="text-sm font-mono font-medium">{{ name }}</span>
                    </div>
                    <UButton size="xs" variant="ghost" color="red" @click="delete form.config.notification.sms_templates[name]">Remove</UButton>
                  </div>
                  <div class="p-3">
                    <label class="block text-xs text-slate-500 mb-1">Body</label>
                    <UTextarea v-model="override.body" placeholder="Leave empty to use default" class="font-mono text-sm" :rows="2" />
                    <p class="text-xs text-slate-400 mt-1" v-pre>Variables: <code class="bg-slate-100 dark:bg-slate-800 px-1 rounded">{{.Code}}</code> <code class="bg-slate-100 dark:bg-slate-800 px-1 rounded">{{.TTLMinutes}}</code></p>
                  </div>
                </div>
              </div>
            </div>
          </UFormGroup>
          <div class="flex justify-end gap-2 mt-4">
            <UButton variant="ghost" @click="isModalOpen = false">Cancel</UButton>
            <UButton type="submit" :loading="saving">{{ editingId ? 'Save' : 'Create' }}</UButton>
          </div>
        </form>
      </UCard>
    </UModal>
  </div>
</template>

<script setup>
import { useApi } from '~/composables/useApi'

definePageMeta({ middleware: 'auth' })

const api = useApi()
const toast = useToast()

const loading = ref(true)
const saving = ref(false)
const isModalOpen = ref(false)
const namespaces = ref([])
const editingId = ref(null)
const search = ref('')

const defaultConfig = () => ({
  mfa_policy: 'MFA_POLICY_DISABLED',
  allowed_social_providers: [],
  password_policy: { min_length: 8, require_uppercase: false, require_lowercase: false, require_number: false, require_special: false, password_history: 0 },
  notification: {
    email_provider: '',
    sms_provider: '',
    email_templates: {},
    sms_templates: {},
  },
})

const form = reactive({ name: '', config: defaultConfig() })

const availableProviders = ref({ email_providers: [], sms_providers: [] })

const columns = [
  { key: 'name', label: 'Name', sortable: true },
  { key: 'id', label: 'ID' },
  { key: 'config', label: 'Security Policy' },
  { key: 'actions', label: '' }
]

let fetchTimeout = null
function debouncedFetch() {
  clearTimeout(fetchTimeout)
  fetchTimeout = setTimeout(fetchNamespaces, 300)
}

function openCreateModal() {
  editingId.value = null
  form.name = ''
  Object.assign(form.config, defaultConfig())
  isModalOpen.value = true
}

function openEditModal(row) {
  editingId.value = row.id
  form.name = row.name
  Object.assign(form.config, defaultConfig(), row.config || {})
  if (row.config?.password_policy) Object.assign(form.config.password_policy, row.config.password_policy)
  if (row.config?.notification) {
    form.config.notification = {
      email_provider: row.config.notification.email_provider || '',
      sms_provider: row.config.notification.sms_provider || '',
      email_templates: JSON.parse(JSON.stringify(row.config.notification.email_templates || {})),
      sms_templates: JSON.parse(JSON.stringify(row.config.notification.sms_templates || {})),
    }
  } else {
    form.config.notification = defaultConfig().notification
  }
  isModalOpen.value = true
}

async function fetchNamespaces() {
  loading.value = true
  try {
    const params = new URLSearchParams({ page_size: '100' })
    if (search.value) params.set('query', search.value)
    const res = await api.fetch(`/v1/admin/namespaces?${params}`)
    namespaces.value = res.namespaces || []
  } catch { toast.add({ title: 'Error fetching namespaces', color: 'red' }) }
  finally { loading.value = false }
}

async function handleCreate() {
  saving.value = true
  try {
    await api.fetch('/v1/admin/namespaces', { method: 'POST', body: form })
    toast.add({ title: 'Namespace created', color: 'green' })
    isModalOpen.value = false
    fetchNamespaces()
  } catch { toast.add({ title: 'Creation failed', color: 'red' }) }
  finally { saving.value = false }
}

async function handleUpdateConfig() {
  saving.value = true
  try {
    await api.fetch(`/v1/admin/namespaces/${editingId.value}/config`, { method: 'PATCH', body: { config: form.config } })
    toast.add({ title: 'Config updated', color: 'green' })
    isModalOpen.value = false
    fetchNamespaces()
  } catch { toast.add({ title: 'Update failed', color: 'red' }) }
  finally { saving.value = false }
}

const emailProviderOptions = computed(() => [
  { label: 'Use server default', value: '' },
  ...availableProviders.value.email_providers.map(p => ({ label: p, value: p })),
])

const smsProviderOptions = computed(() => [
  { label: 'Use server default', value: '' },
  ...availableProviders.value.sms_providers.map(p => ({ label: p, value: p })),
])

const templateRegistry = [
  { name: 'mfa_email_otp', channel: 'email' },
  { name: 'mfa_sms_otp', channel: 'sms' },
]

const addableTemplateItems = computed(() => {
  const existing = new Set([
    ...Object.keys(form.config.notification.email_templates || {}),
    ...Object.keys(form.config.notification.sms_templates || {}),
  ])
  return [templateRegistry.filter(t => !existing.has(t.name)).map(t => ({
    label: `${t.channel}: ${t.name}`,
    click: () => {
      if (t.channel === 'email') {
        form.config.notification.email_templates[t.name] = { subject: '', html_body: '', text_body: '' }
      } else {
        form.config.notification.sms_templates[t.name] = { body: '' }
      }
    },
  }))]
})

async function handleDelete(id) {
  if (!confirm('Are you sure? This action is permanent.')) return
  try {
    await api.fetch(`/v1/admin/namespaces/${id}`, { method: 'DELETE' })
    toast.add({ title: 'Namespace deleted', color: 'green' })
    fetchNamespaces()
  } catch { toast.add({ title: 'Deletion failed', color: 'red' }) }
}

onMounted(async () => {
  fetchNamespaces()
  try {
    availableProviders.value = await api.fetch('/v1/admin/notification/providers')
  } catch { /* ignore */ }
})
</script>
