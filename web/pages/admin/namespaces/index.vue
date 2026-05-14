<template>
  <div class="space-y-6">
    <header class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold tracking-tight">Namespaces</h1>
        <p class="text-sm text-slate-500">Manage tenant isolation and security policies</p>
      </div>
      <UButton icon="i-heroicons-plus" @click="isCreateModalOpen = true">Create Namespace</UButton>
    </header>

    <UCard>
      <UTable :rows="namespaces" :columns="columns" :loading="loading">
        <template #config-data="{ row }">
          <div class="flex gap-1">
            <UBadge v-if="row.config?.mfa_required" size="xs" color="orange" variant="soft">MFA</UBadge>
            <UBadge v-for="p in row.config?.allowed_social_providers" :key="p" size="xs" color="blue" variant="soft">
              {{ p }}
            </UBadge>
          </div>
        </template>
        
        <template #actions-data="{ row }">
          <UDropdown :items="actions(row)">
            <UButton color="gray" variant="ghost" icon="i-heroicons-ellipsis-horizontal-20-solid" />
          </UDropdown>
        </template>
      </UTable>
    </UCard>

    <UModal v-model="isCreateModalOpen">
      <UCard :ui="{ ring: '', divide: 'divide-y divide-gray-100 dark:divide-gray-800' }">
        <template #header>
          <div class="flex items-center justify-between">
            <h3 class="text-base font-semibold leading-6 text-gray-900 dark:text-white">
              New Namespace
            </h3>
            <UButton color="gray" variant="ghost" icon="i-heroicons-x-mark-20-solid" class="-my-1" @click="isCreateModalOpen = false" />
          </div>
        </template>

        <form @submit.prevent="handleCreateNamespace" class="space-y-4 py-4">
          <UFormGroup label="Name" name="name" required>
            <UInput v-model="newNamespace.name" placeholder="e.g. acme-corp" />
          </UFormGroup>
          <UFormGroup label="Security Settings">
            <UCheckbox v-model="newNamespace.config.mfa_required" label="Require MFA for all users" />
          </UFormGroup>
          <UFormGroup label="Social Providers">
            <USelectMenu v-model="newNamespace.config.allowed_social_providers" multiple :options="['google', 'github']" />
          </UFormGroup>
          
          <div class="flex justify-end gap-2 mt-6">
            <UButton variant="ghost" @click="isCreateModalOpen = false">Cancel</UButton>
            <UButton type="submit" :loading="creating">Create</UButton>
          </div>
        </form>
      </UCard>
    </UModal>
  </div>
</template>

<script setup>
import { useApi } from '~/composables/useApi'

definePageMeta({
  middleware: 'auth'
})

const api = useApi()
const toast = useToast()

const loading = ref(true)
const creating = ref(false)
const isCreateModalOpen = ref(false)
const namespaces = ref([])

const newNamespace = reactive({
  name: '',
  config: {
    mfa_required: false,
    allowed_social_providers: [],
    password_policy: {
      min_length: 8,
      require_uppercase: true,
      require_lowercase: true,
      require_number: true,
      require_special: false
    }
  }
})

const columns = [
  { key: 'name', label: 'Name', sortable: true },
  { key: 'id', label: 'ID' },
  { key: 'config', label: 'Security Policy' },
  { key: 'actions', label: '' }
]

const actions = (row) => [
  [{
    label: 'Edit Config',
    icon: 'i-heroicons-pencil-square',
    click: () => console.log('Edit', row.id)
  }],
  [{
    label: 'Delete',
    icon: 'i-heroicons-trash',
    color: 'red',
    click: () => handleDeleteNamespace(row.id)
  }]
]

async function fetchNamespaces() {
  loading.value = true
  try {
    const res = await api.fetch('/v1/admin/namespaces?page_size=100')
    namespaces.value = res.namespaces || []
  } catch (err) {
    toast.add({ title: 'Error fetching namespaces', color: 'red' })
  } finally {
    loading.value = false
  }
}

async function handleCreateNamespace() {
  creating.value = true
  try {
    await api.fetch('/v1/admin/namespaces', {
      method: 'POST',
      body: newNamespace
    })
    toast.add({ title: 'Namespace created!', color: 'green' })
    isCreateModalOpen.value = false
    fetchNamespaces()
  } catch (err) {
    toast.add({ title: 'Creation failed', color: 'red' })
  } finally {
    creating.value = false
  }
}

async function handleDeleteNamespace(id) {
  if (!confirm('Are you sure? This action is permanent.')) return
  
  try {
    await api.fetch(`/v1/admin/namespaces/${id}`, { method: 'DELETE' })
    toast.add({ title: 'Namespace deleted', color: 'green' })
    fetchNamespaces()
  } catch (err) {
    toast.add({ title: 'Deletion failed', color: 'red' })
  }
}

onMounted(fetchNamespaces)
</script>
