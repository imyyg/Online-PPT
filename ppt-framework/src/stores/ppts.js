import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { pptsApi } from '../api'

export const usePptsStore = defineStore('ppts', () => {
  const records = ref([])
  const currentRecord = ref(null)
  const loading = ref(false)
  const error = ref(null)
  const searchQuery = ref('')
  const selectedTag = ref('')
  const sortBy = ref('created_at_desc')

  const filteredRecords = computed(() => {
    let result = records.value

    if (searchQuery.value) {
      const query = searchQuery.value.toLowerCase()
      result = result.filter(record => 
        record.name?.toLowerCase().includes(query) ||
        record.title?.toLowerCase().includes(query) ||
        record.description?.toLowerCase().includes(query) ||
        record.tags?.some(tag => tag.toLowerCase().includes(query))
      )
    }

    if (selectedTag.value) {
      result = result.filter(record => 
        record.tags?.includes(selectedTag.value)
      )
    }

    return result
  })

  const allTags = computed(() => {
    const tagSet = new Set()
    for (const record of records.value) {
      if (record.tags) {
        for (const tag of record.tags) {
          tagSet.add(tag)
        }
      }
    }
    return Array.from(tagSet).sort((a, b) => a.localeCompare(b))
  })

  async function fetchRecords(params = {}) {
    try {
      loading.value = true
      error.value = null

      const queryParams = {
        q: searchQuery.value || undefined,
        tag: selectedTag.value || undefined,
        sort: sortBy.value,
        ...params
      }

      const response = await pptsApi.list(queryParams)
      records.value = response.items || []
      
      return response
    } catch (err) {
      error.value = err.response?.data?.message || 'Failed to fetch records'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function fetchRecord(id) {
    try {
      loading.value = true
      error.value = null

      const record = await pptsApi.get(id)
      currentRecord.value = record
      
      return record
    } catch (err) {
      error.value = err.response?.data?.message || 'Failed to fetch record'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function createRecord(data) {
    try {
      loading.value = true
      error.value = null

      const record = await pptsApi.create(data)
      records.value.push(record)
      
      return record
    } catch (err) {
      error.value = err.response?.data?.message || 'Failed to create record'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function updateRecord(id, data) {
    try {
      loading.value = true
      error.value = null

      const updated = await pptsApi.update(id, data)
      
      const index = records.value.findIndex(r => r.id === id)
      if (index !== -1) {
        records.value[index] = updated
      }
      
      if (currentRecord.value?.id === id) {
        currentRecord.value = updated
      }
      
      return updated
    } catch (err) {
      error.value = err.response?.data?.message || 'Failed to update record'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function deleteRecord(id) {
    try {
      loading.value = true
      error.value = null

      await pptsApi.delete(id)
      
      records.value = records.value.filter(r => r.id !== id)
      
      if (currentRecord.value?.id === id) {
        currentRecord.value = null
      }
    } catch (err) {
      error.value = err.response?.data?.message || 'Failed to delete record'
      throw err
    } finally {
      loading.value = false
    }
  }

  function setSearchQuery(query) {
    searchQuery.value = query
  }

  function setSelectedTag(tag) {
    selectedTag.value = tag
  }

  function setSortBy(sort) {
    sortBy.value = sort
  }

  function clearError() {
    error.value = null
  }

  return {
    records,
    currentRecord,
    loading,
    error,
    searchQuery,
    selectedTag,
    sortBy,
    filteredRecords,
    allTags,
    fetchRecords,
    fetchRecord,
    createRecord,
    updateRecord,
    deleteRecord,
    setSearchQuery,
    setSelectedTag,
    setSortBy,
    clearError
  }
})
