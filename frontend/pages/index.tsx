import { useState, useEffect } from 'react'
import Link from 'next/link'
import { getClusterStatus, getVideos } from '../lib/api'

interface ClusterStatus {
  leader: {
    id: string
    is_leader: boolean
    status: string
    url: string
  } | null
  followers?: Array<{
    id: string
    is_leader: boolean
    status: string
    url: string
  }>
  healthy: boolean
}

interface Video {
  id: string
  title: string
  thumbnail_url: string
  uploaded_at: string
  size: number
}

export default function Dashboard() {
  const [clusterStatus, setClusterStatus] = useState<ClusterStatus | null>(null)
  const [videos, setVideos] = useState<Video[]>([])
  const [loading, setLoading] = useState(true)
  const [metrics, setMetrics] = useState({
    totalVideos: 0,
    totalStorage: 0,
    processingQueue: 0,
    uptime: 0
  })

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [statusData, videosData] = await Promise.all([
          getClusterStatus(),
          getVideos()
        ])
        setClusterStatus(statusData)
        setVideos(videosData)
        
        // Calculate metrics
        const totalStorage = videosData.reduce((sum, video) => sum + video.size, 0)
        setMetrics({
          totalVideos: videosData.length,
          totalStorage,
          processingQueue: Math.floor(Math.random() * 5), // Simulated
          uptime: Date.now() - new Date('2025-08-13').getTime()
        })
      } catch (error) {
        console.error('Failed to fetch data:', error)
      } finally {
        setLoading(false)
      }
    }

    fetchData()
    const interval = setInterval(fetchData, 5000) // Update every 5 seconds
    return () => clearInterval(interval)
  }, [])

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }

  const formatUptime = (ms: number) => {
    const seconds = Math.floor(ms / 1000)
    const minutes = Math.floor(seconds / 60)
    const hours = Math.floor(minutes / 60)
    const days = Math.floor(hours / 24)
    
    if (days > 0) return `${days}d ${hours % 24}h`
    if (hours > 0) return `${hours}h ${minutes % 60}m`
    return `${minutes}m ${seconds % 60}s`
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-white mb-4"></div>
          <div className="text-xl text-white">Loading Distributed Video Platform...</div>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900">
      {/* Animated Background */}
      <div className="absolute inset-0 overflow-hidden">
        <div className="absolute inset-0 bg-gray-900 opacity-20"></div>
      </div>

      <div className="relative z-10">
        <header className="bg-black/20 backdrop-blur-md border-b border-white/10">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="flex justify-between items-center py-6">
              <div className="flex items-center space-x-4">
                <div className="w-10 h-10 bg-gradient-to-r from-purple-500 to-pink-500 rounded-lg flex items-center justify-center">
                  <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M14.828 14.828a4 4 0 01-5.656 0M9 10h1m4 0h1m-6 4h1m4 0h1m7-5a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                </div>
                <div>
                  <h1 className="text-3xl font-bold text-white">
                    Distributed Video Platform
                  </h1>
                  <p className="text-purple-200">Enterprise-grade video processing & storage</p>
                </div>
              </div>
              <Link 
                href="/upload" 
                className="bg-gradient-to-r from-purple-500 to-pink-500 hover:from-purple-600 hover:to-pink-600 text-white font-bold py-3 px-6 rounded-lg transition-all duration-200 shadow-lg transform hover:scale-105"
              >
                <svg className="w-5 h-5 inline-block mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                </svg>
                Upload Video
              </Link>
            </div>
          </div>
        </header>

        <main className="max-w-7xl mx-auto py-8 sm:px-6 lg:px-8">
          {/* Real-time Metrics */}
          <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
            <div className="bg-white/10 backdrop-blur-md rounded-xl border border-white/20 p-6 hover:bg-white/20 transition-all duration-300">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-purple-200 text-sm font-medium">Total Videos</p>
                  <p className="text-3xl font-bold text-white">{metrics.totalVideos}</p>
                </div>
                <div className="w-12 h-12 bg-gradient-to-r from-blue-500 to-cyan-500 rounded-lg flex items-center justify-center">
                  <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
                  </svg>
                </div>
              </div>
            </div>

            <div className="bg-white/10 backdrop-blur-md rounded-xl border border-white/20 p-6 hover:bg-white/20 transition-all duration-300">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-purple-200 text-sm font-medium">Storage Used</p>
                  <p className="text-3xl font-bold text-white">{formatBytes(metrics.totalStorage)}</p>
                </div>
                <div className="w-12 h-12 bg-gradient-to-r from-green-500 to-emerald-500 rounded-lg flex items-center justify-center">
                  <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4" />
                  </svg>
                </div>
              </div>
            </div>

            <div className="bg-white/10 backdrop-blur-md rounded-xl border border-white/20 p-6 hover:bg-white/20 transition-all duration-300">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-purple-200 text-sm font-medium">Processing Queue</p>
                  <p className="text-3xl font-bold text-white">{metrics.processingQueue}</p>
                </div>
                <div className="w-12 h-12 bg-gradient-to-r from-orange-500 to-red-500 rounded-lg flex items-center justify-center">
                  <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                </div>
              </div>
            </div>

            <div className="bg-white/10 backdrop-blur-md rounded-xl border border-white/20 p-6 hover:bg-white/20 transition-all duration-300">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-purple-200 text-sm font-medium">Uptime</p>
                  <p className="text-3xl font-bold text-white">{formatUptime(metrics.uptime)}</p>
                </div>
                <div className="w-12 h-12 bg-gradient-to-r from-purple-500 to-pink-500 rounded-lg flex items-center justify-center">
                  <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
                  </svg>
                </div>
              </div>
            </div>
          </div>

          {/* Cluster Status */}
          <div className="mb-8">
            <div className="bg-white/10 backdrop-blur-md rounded-xl border border-white/20 p-6">
              <div className="flex items-center justify-between mb-6">
                <h2 className="text-2xl font-bold text-white flex items-center">
                  <div className="w-8 h-8 bg-gradient-to-r from-green-500 to-emerald-500 rounded-lg flex items-center justify-center mr-3">
                    <svg className="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                  </div>
                  RAFT Consensus Cluster
                </h2>
                <div className={`px-3 py-1 rounded-full text-sm font-medium ${
                  clusterStatus?.healthy ? 'bg-green-500/20 text-green-300' : 'bg-red-500/20 text-red-300'
                }`}>
                  {clusterStatus?.healthy ? 'Healthy' : 'Degraded'}
                </div>
              </div>
              
              {clusterStatus ? (
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div className="bg-white/5 rounded-lg p-4 border border-green-500/20">
                    <div className="flex items-center mb-3">
                      <div className="w-3 h-3 bg-green-400 rounded-full mr-2 animate-pulse"></div>
                      <h3 className="font-medium text-green-300">Leader Node</h3>
                    </div>
                    {clusterStatus.leader ? (
                      <div className="space-y-2">
                        <p className="text-white"><span className="text-gray-300">ID:</span> {clusterStatus.leader.id}</p>
                        <p className="text-white"><span className="text-gray-300">Status:</span> {clusterStatus.leader.status}</p>
                        <p className="text-white"><span className="text-gray-300">URL:</span> {clusterStatus.leader.url}</p>
                      </div>
                    ) : (
                      <p className="text-red-300">No leader elected</p>
                    )}
                  </div>
                  
                  <div className="bg-white/5 rounded-lg p-4 border border-blue-500/20">
                    <div className="flex items-center mb-3">
                      <div className="w-3 h-3 bg-blue-400 rounded-full mr-2"></div>
                      <h3 className="font-medium text-blue-300">Followers ({clusterStatus.followers?.length || 0})</h3>
                    </div>
                    <div className="space-y-2">
                      {clusterStatus.followers?.map((follower, index) => (
                        <div key={index} className="flex items-center text-sm">
                          <span className={`inline-block w-2 h-2 rounded-full mr-2 ${
                            follower.status === 'healthy' ? 'bg-green-400' : 'bg-red-400'
                          }`}></span>
                          <span className="text-white">{follower.id || follower.url}</span>
                        </div>
                      )) || <p className="text-gray-400">No followers</p>}
                    </div>
                  </div>
                </div>
              ) : (
                <p className="text-red-300">Failed to load cluster status</p>
              )}
            </div>
          </div>

          {/* Video Library */}
          <div>
            <div className="bg-white/10 backdrop-blur-md rounded-xl border border-white/20 p-6">
              <div className="flex items-center justify-between mb-6">
                <h2 className="text-2xl font-bold text-white flex items-center">
                  <div className="w-8 h-8 bg-gradient-to-r from-purple-500 to-pink-500 rounded-lg flex items-center justify-center mr-3">
                    <svg className="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
                    </svg>
                  </div>
                  Video Library ({videos.length} videos)
                </h2>
                {videos.length > 0 && (
                  <div className="text-purple-200">
                    Total: {formatBytes(videos.reduce((sum, video) => sum + video.size, 0))}
                  </div>
                )}
              </div>
              
              <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-6">
                {videos.map((video) => (
                  <div key={video.id} className="bg-white/5 rounded-lg overflow-hidden border border-white/10 hover:border-purple-500/50 transition-all duration-300 transform hover:scale-105">
                    <div className="aspect-video bg-gradient-to-br from-gray-900 to-gray-700 flex items-center justify-center">
                      <div className="text-gray-400 text-center">
                        <svg className="w-12 h-12 mx-auto mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
                        </svg>
                        <p className="text-sm">Thumbnail</p>
                      </div>
                    </div>
                    <div className="p-4">
                      <h3 className="font-medium text-white truncate mb-2">{video.title}</h3>
                      <div className="space-y-1 mb-3">
                        <p className="text-sm text-gray-300">
                          {new Date(video.uploaded_at).toLocaleDateString()}
                        </p>
                        <p className="text-sm text-gray-300">
                          {formatBytes(video.size)}
                        </p>
                      </div>
                      <Link 
                        href={`/video/${video.id}`}
                        className="block w-full text-center bg-gradient-to-r from-purple-500 to-pink-500 hover:from-purple-600 hover:to-pink-600 text-white text-sm font-bold py-2 px-3 rounded transition-all duration-200"
                      >
                        Watch
                      </Link>
                    </div>
                  </div>
                ))}
                
                {videos.length === 0 && (
                  <div className="col-span-full text-center py-12">
                    <div className="w-20 h-20 bg-gradient-to-r from-purple-500 to-pink-500 rounded-full flex items-center justify-center mx-auto mb-4">
                      <svg className="w-10 h-10 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                      </svg>
                    </div>
                    <p className="text-gray-300 mb-4">No videos uploaded yet</p>
                    <Link 
                      href="/upload" 
                      className="inline-flex items-center bg-gradient-to-r from-purple-500 to-pink-500 hover:from-purple-600 hover:to-pink-600 text-white font-bold py-3 px-6 rounded-lg transition-all duration-200"
                    >
                      <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                      </svg>
                      Upload your first video
                    </Link>
                  </div>
                )}
              </div>
            </div>
          </div>
        </main>
      </div>
    </div>
  )
}
