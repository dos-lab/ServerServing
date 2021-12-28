import defaultSettings from '@/settings'

const title = defaultSettings.title || 'Server Serving'

export default function getPageTitle(pageTitle) {
  if (pageTitle) {
    return `${pageTitle} - ${title}`
  }
  return `${title}`
}
