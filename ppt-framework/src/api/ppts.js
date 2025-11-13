import apiClient from './client'

export const pptsApi = {
  async list(params = {}) {
    const response = await apiClient.get('/ppts', { params })
    return response.data
  },

  async get(id) {
    const response = await apiClient.get(`/ppts/${id}`)
    return response.data
  },

  async create(data) {
    const response = await apiClient.post('/ppts', {
      name: data.name,
      title: data.title || null,
      description: data.description || null,
      tags: data.tags || []
    })
    return response.data
  },

  async update(id, data) {
    const response = await apiClient.patch(`/ppts/${id}`, {
      name: data.name,
      title: data.title || null,
      description: data.description || null,
      tags: data.tags || []
    })
    return response.data
  },

  async delete(id) {
    const response = await apiClient.delete(`/ppts/${id}`)
    return response.data
  }
}
