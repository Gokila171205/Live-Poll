import { QRCodeSVG } from 'qrcode.react'
import { CopyButton } from './CopyButton'
import { Button } from './Button'

export function QRCodeModal({ url, question, isOpen, onClose }) {
  if (!isOpen) return null

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content glass-panel" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h3 className="modal-title">Scan to Vote</h3>
          <button type="button" className="modal-close-btn" onClick={onClose} aria-label="Close modal">
            &times;
          </button>
        </div>

        <p className="modal-subtitle">
          Audience members can scan this QR code with their mobile camera to vote immediately.
        </p>

        {question && <p className="modal-question-preview">"{question}"</p>}

        <div className="qr-code-wrapper">
          <div className="qr-code-box">
            <QRCodeSVG
              value={url}
              size={220}
              bgColor="#ffffff"
              fgColor="#090d16"
              level="M"
              includeMargin={true}
            />
          </div>
        </div>

        <div className="modal-url-row">
          <code className="modal-url-text">{url}</code>
          <CopyButton text={url} label="Copy Link" size="sm" />
        </div>

        <div className="modal-footer">
          <Button variant="outline" size="md" onClick={onClose} className="w-full">
            Close
          </Button>
        </div>
      </div>
    </div>
  )
}
