export function LoadingSpinner({ text = 'Loading...', size = 'md' }) {
  return (
    <div className={`loading-container loading-${size}`}>
      <div className="spinner-ring"></div>
      {text && <p className="loading-text">{text}</p>}
    </div>
  )
}
