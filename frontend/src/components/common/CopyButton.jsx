import { useState } from 'react'
import { CopyIcon, CheckIcon } from '../icons/Icons'

export function CopyButton({ text, label = 'Copy Link', size = 'sm', className = '' }) {
  const [copied, setCopied] = useState(false)

  const handleCopy = async (e) => {
    e.stopPropagation()
    e.preventDefault()
    try {
      if (navigator?.clipboard?.writeText) {
        await navigator.clipboard.writeText(text)
      } else {
        // Fallback for older environments
        const textArea = document.createElement('textarea')
        textArea.value = text
        document.body.appendChild(textArea)
        textArea.select()
        document.execCommand('copy')
        document.body.removeChild(textArea)
      }
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch (err) {
      console.error('Failed to copy text:', err)
    }
  }

  return (
    <button
      type="button"
      onClick={handleCopy}
      className={`copy-btn copy-btn-${size} ${copied ? 'copied' : ''} ${className}`.trim()}
      title={copied ? 'Copied to clipboard!' : 'Copy to clipboard'}
    >
      {copied ? <CheckIcon size={16} /> : <CopyIcon size={16} />}
      <span>{copied ? 'Copied!' : label}</span>
    </button>
  )
}
