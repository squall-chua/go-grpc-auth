<template>
  <div class="space-y-6">
    <PageHeader
      eyebrow="Administration"
      title="OIDC Clients"
      subtitle="Manage OAuth2/OIDC client applications"
    >
      <template #actions>
        <UButton icon="i-heroicons-plus" @click="openCreateModal">Register Client</UButton>
      </template>
    </PageHeader>

    <UCard>
      <div class="flex gap-2 mb-4">
        <UInput v-model="search" icon="i-heroicons-magnifying-glass" placeholder="Search clients..." class="flex-1 max-w-xs" @input="debouncedFetch" />
        <UInput v-model="namespace" placeholder="Namespace" class="w-48" @input="debouncedFetch" />
      </div>

      <UTable :rows="clients" :columns="columns" :loading="loading">
        <template #allowed_scopes-data="{ row }">
          <div class="flex flex-wrap gap-1">
            <UBadge v-for="s in row.allowed_scopes" :key="s" color="blue" variant="soft" size="xs">{{ s }}</UBadge>
          </div>
        </template>
        <template #skip_consent-data="{ row }">
          <UBadge :color="row.skip_consent ? 'green' : 'gray'" variant="soft" size="xs">{{ row.skip_consent ? 'Yes' : 'No' }}</UBadge>
        </template>
        <template #actions-data="{ row }">
          <div class="flex gap-1">
            <UTooltip text="Edit">
              <UButton size="xs" variant="ghost" icon="i-heroicons-pencil-square" @click="openEditModal(row)" />
            </UTooltip>
            <UTooltip text="Rotate Secret">
              <UButton size="xs" variant="ghost" color="orange" icon="i-heroicons-arrow-path" @click="rotateSecret(row.client_id)" />
            </UTooltip>
            <UTooltip text="Delete">
              <UButton size="xs" variant="ghost" color="red" icon="i-heroicons-trash" @click="handleDelete(row.client_id)" />
            </UTooltip>
          </div>
        </template>
      </UTable>
    </UCard>

    <!-- Create Modal -->
    <UModal v-model="isCreateOpen">
      <UCard>
        <template #header>
          <h3 class="text-base font-semibold">Register OIDC Client</h3>
        </template>
        <form @submit.prevent="handleCreate" class="space-y-4">
          <UFormGroup label="Name">
            <UInput v-model="form.name" placeholder="My App" />
          </UFormGroup>
          <UFormGroup label="Namespace">
            <UInput v-model="form.namespace" placeholder="default" />
          </UFormGroup>
          <UFormGroup label="Redirect URIs (one per line)">
            <UTextarea v-model="redirectUrisInput" placeholder="http://localhost:3000/callback" />
          </UFormGroup>
          <UFormGroup label="Allowed Scopes (comma-separated)">
            <UInput v-model="scopesInput" placeholder="openid, profile, email" />
          </UFormGroup>
          <UCheckbox v-model="form.skip_consent" label="Skip consent screen" />
          <div class="flex justify-end gap-2">
            <UButton variant="ghost" @click="isCreateOpen = false">Cancel</UButton>
            <UButton type="submit" :loading="saving">Register</UButton>
          </div>
        </form>
      </UCard>
    </UModal>

    <!-- Secret Display Modal -->
    <UModal v-model="isSecretOpen" :prevent-close="true">
      <UCard>
        <template #header>
          <h3 class="text-base font-semibold text-orange-600">Client Secret</h3>
        </template>
        <div class="space-y-4">
          <UAlert color="orange" variant="soft" title="Copy this secret now" description="This secret will not be shown again." />
          <div class="p-3 bg-slate-950 rounded font-mono text-sm text-green-400 break-all select-all">{{ clientSecret }}</div>
          <UButton block @click="isSecretOpen = false">I've copied the secret</UButton>
        </div>
      </UCard>
    </UModal>

    <!-- Edit Modal -->
    <UModal v-model="isEditOpen">
      <UCard>
        <template #header>
          <h3 class="text-base font-semibold">Edit Client</h3>
        </template>
        <form @submit.prevent="handleUpdate" class="space-y-4">
          <UFormGroup label="Name">
            <UInput v-model="editForm.name" />
          </UFormGroup>
          <UFormGroup label="Redirect URIs (one per line)">
            <UTextarea v-model="editRedirectUrisInput" />
          </UFormGroup>
          <UFormGroup label="Allowed Scopes (comma-separated)">
            <UInput v-model="editScopesInput" />
          </UFormGroup>
          <UCheckbox v-model="editForm.skip_consent" label="Skip consent screen" />
          <div class="flex justify-end gap-2">
            <UButton variant="ghost" @click="isEditOpen = false">Cancel</UButton>
            <UButton type="submit" :loading="saving">Save</UButton>
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
const clients = ref([])
const search = ref('')
const namespace = ref('')

const isCreateOpen = ref(false)
const isSecretOpen = ref(false)
const isEditOpen = ref(false)
const clientSecret = ref('')

const form = reactive({ name: '', namespace: 'default', skip_consent: false })
const redirectUrisInput = ref('')
const scopesInput = ref('')

const editForm = reactive({ client_id: '', name: '', skip_consent: false })
const editRedirectUrisInput = ref('')
const editScopesInput = ref('')

const columns = [
  { key: 'client_id', label: 'Client ID' },
  { key: 'name', label: 'Name' },
  { key: 'namespace', label: 'Namespace' },
  { key: 'allowed_scopes', label: 'Scopes' },
  { key: 'skip_consent', label: 'Skip Consent' },
  { key: 'actions', label: '' }
]


let fetchTimeout = null
function debouncedFetch() { clearTimeout(fetchTimeout); fetchTimeout = setTimeout(fetchClients, 300) }

function openCreateModal() {
  form.name = ''
  form.namespace = 'default'
  form.skip_consent = false
  redirectUrisInput.value = ''
  scopesInput.value = 'openid, profile, email'
  isCreateOpen.value = true
}

function openEditModal(row) {
  editForm.client_id = row.client_id
  editForm.name = row.name
  editForm.skip_consent = row.skip_consent
  editRedirectUrisInput.value = (row.redirect_uris || []).join('\n')
  editScopesInput.value = (row.allowed_scopes || []).join(', ')
  isEditOpen.value = true
}

async function fetchClients() {
  loading.value = true
  try {
    const params = new URLSearchParams({ page_size: '100' })
    if (namespace.value) params.set('namespace', namespace.value)
    if (search.value) params.set('query', search.value)
    const res = await api.fetch(`/v1/admin/oidc/clients?${params}`)
    clients.value = res.clients || []
  } catch { toast.add({ title: 'Error fetching clients', color: 'red' }) }
  finally { loading.value = false }
}

async function handleCreate() {
  saving.value = true
  try {
    const body = {
      ...form,
      redirect_uris: redirectUrisInput.value.split('\n').map(s => s.trim()).filter(Boolean),
      allowed_scopes: scopesInput.value.split(',').map(s => s.trim()).filter(Boolean),
    }
    const res = await api.fetch('/v1/admin/oidc/clients', { method: 'POST', body })
    clientSecret.value = res.client_secret || ''
    isCreateOpen.value = false
    if (clientSecret.value) isSecretOpen.value = true
    toast.add({ title: 'Client registered', color: 'green' })
    fetchClients()
  } catch { toast.add({ title: 'Registration failed', color: 'red' }) }
  finally { saving.value = false }
}

async function handleUpdate() {
  saving.value = true
  try {
    await api.fetch(`/v1/admin/oidc/clients/${editForm.client_id}`, {
      method: 'PATCH',
      body: {
        name: editForm.name,
        redirect_uris: editRedirectUrisInput.value.split('\n').map(s => s.trim()).filter(Boolean),
        allowed_scopes: editScopesInput.value.split(',').map(s => s.trim()).filter(Boolean),
        skip_consent: editForm.skip_consent,
      }
    })
    toast.add({ title: 'Client updated', color: 'green' })
    isEditOpen.value = false
    fetchClients()
  } catch { toast.add({ title: 'Update failed', color: 'red' }) }
  finally { saving.value = false }
}

async function rotateSecret(clientId) {
  if (!confirm('Rotate secret? The old secret will stop working immediately.')) return
  try {
    const res = await api.fetch(`/v1/admin/oidc/clients/${clientId}/rotate-secret`, { method: 'POST' })
    clientSecret.value = res.client_secret || ''
    if (clientSecret.value) isSecretOpen.value = true
    toast.add({ title: 'Secret rotated', color: 'green' })
  } catch { toast.add({ title: 'Rotation failed', color: 'red' }) }
}

async function handleDelete(clientId) {
  if (!confirm('Delete this client?')) return
  try {
    await api.fetch(`/v1/admin/oidc/clients/${clientId}`, { method: 'DELETE' })
    toast.add({ title: 'Client deleted', color: 'green' })
    fetchClients()
  } catch { toast.add({ title: 'Deletion failed', color: 'red' }) }
}

onMounted(fetchClients)
</script>
