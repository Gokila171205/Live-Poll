export function Alert({ type = 'error', message, title, onClose, className = '' }) {
  if (!message) return null

  return (
    <div className={`alert-box alert-${type} ${className}`.trim()} role="alert">
      <div className="alert-content">
        {title && <h4 className="alert-title">{title}</h4>}
        <p className="alert-message">{message}</p>
      </div>
      {onClose && (
        <button type="button" onClick={onClose} className="alert-close-btn" aria-label="Close alert">
          &times;
        </button>
      )}
    </div>
  )
}
