import axios from 'axios'

const api = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080',
  timeout: 10000,
})

export interface ClusterStatus {
  leader: {
    id: string
    is_leader: boolean
    status: string
    url: string
  } | null
  followers: Array<{
    id: string
    is_leader: boolean
    status: string
    url: string
  }>
  healthy: boolean
}

export interface Video {
  id: string
  title: string
  bucket: string
  object: string
  thumbnail_url: string
  size: number
  content_type: string
  uploaded_at: string
  resolutions: string[]
}

export interface UploadResponse {
  id: string
  title: string
  bucket: string
  object: string
  thumbnail_url: string
  size: number
  content_type: string
  uploaded_at: string
  resolutions: string[]
}

export const getClusterStatus = async (): Promise<ClusterStatus> => {
  const response = await api.get('/cluster/status')
  return response.data
}

export const getVideos = async (): Promise<Video[]> => {
  const response = await api.get('/videos')
  return response.data
}

export const getVideo = async (id: string): Promise<Video> => {
  const response = await api.get(`/videos/${id}`)
  return response.data
}

export const uploadVideo = async (file: File, onProgress?: (progress: number) => void): Promise<UploadResponse> => {
  const formData = new FormData()
  formData.append('file', file)

  const response = await api.post('/upload', formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
    onUploadProgress: (progressEvent) => {
      if (onProgress && progressEvent.total) {
        const progress = Math.round((progressEvent.loaded * 100) / progressEvent.total)
        onProgress(progress)
      }
    },
  })

  return response.data
}

export const getVideoStreamUrl = (id: string): string => {
  return `${api.defaults.baseURL}/videos/${id}/stream`
}

export const getVideoThumbnailUrl = (id: string): string => {
  return `${api.defaults.baseURL}/videos/${id}/thumbnail`
}

export default api
