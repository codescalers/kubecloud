import { ref } from 'vue'
import { api } from '../utils/api'
import { useNotificationStore } from '../stores/notifications'

export function useKubeconfig() {
  const downloading = ref<string | null>(null)
  const notifications = useNotificationStore()

  async function download(projectName: string) {
    if (!projectName) return
    
    downloading.value = projectName
    try {
      const response = await api.get(`/v1/deployments/${projectName}/kubeconfig`, { 
        requiresAuth: true, 
        showNotifications: false,
        timeout: 120000
      })
      
      const data = response.data as any
      if (data.kubeconfig) {
        downloadFile(data.kubeconfig, `${projectName}-kubeconfig.yaml`)
        notifications.success('Download Successful', 'Kubeconfig file downloaded.')
      } else {
        notifications.error('Download Failed', 'No kubeconfig content available.')
      }
    } catch (err: any) {
      notifications.error('Download Failed', err?.message || 'Failed to download kubeconfig')
    } finally {
      downloading.value = null
    }
  }

  function downloadFile(content: string, filename: string) {
    if (!content || !filename) return
    
    const blob = new Blob([content], { type: 'application/x-yaml' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    
    link.href = url
    link.download = filename
    document.body.appendChild(link)
    link.click()
    
    setTimeout(() => {
      document.body.removeChild(link)
      URL.revokeObjectURL(url)
    }, 100)
  }

  return {
    downloading,
    download,
    downloadFile
  }
} 